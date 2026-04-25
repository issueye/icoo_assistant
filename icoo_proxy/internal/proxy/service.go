package proxy

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"icoo_proxy/internal/api"
	"icoo_proxy/internal/catalog"
	"icoo_proxy/internal/config"
)

type Service struct {
	cfg     config.Config
	catalog *catalog.Catalog
	client  *http.Client
	mu      sync.RWMutex
	recent  []api.RequestView
}

func New(cfg config.Config, catalog *catalog.Catalog) *Service {
	return &Service{
		cfg:     cfg,
		catalog: catalog,
		client:  &http.Client{},
	}
}

func (s *Service) Handle(w http.ResponseWriter, r *http.Request, downstream catalog.Protocol) {
	requestID := newRequestID()
	start := time.Now()
	w.Header().Set("X-ICOO-Request-ID", requestID)

	if r.Method != http.MethodPost {
		s.fail(w, downstream, api.RequestView{
			RequestID:  requestID,
			Downstream: string(downstream),
			StatusCode: http.StatusMethodNotAllowed,
			DurationMS: time.Since(start).Milliseconds(),
			Error:      "method not allowed",
			CreatedAt:  time.Now().Format(time.RFC3339),
		})
		return
	}
	if err := s.authorize(r); err != nil {
		s.fail(w, downstream, api.RequestView{
			RequestID:  requestID,
			Downstream: string(downstream),
			StatusCode: http.StatusUnauthorized,
			DurationMS: time.Since(start).Milliseconds(),
			Error:      err.Error(),
			CreatedAt:  time.Now().Format(time.RFC3339),
		})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.fail(w, downstream, api.RequestView{
			RequestID:  requestID,
			Downstream: string(downstream),
			StatusCode: http.StatusBadRequest,
			DurationMS: time.Since(start).Milliseconds(),
			Error:      "failed to read request body",
			CreatedAt:  time.Now().Format(time.RFC3339),
		})
		return
	}
	defer r.Body.Close()

	requestModel, err := extractModel(body)
	if err != nil {
		s.fail(w, downstream, api.RequestView{
			RequestID:  requestID,
			Downstream: string(downstream),
			StatusCode: http.StatusBadRequest,
			DurationMS: time.Since(start).Milliseconds(),
			Error:      err.Error(),
			CreatedAt:  time.Now().Format(time.RFC3339),
		})
		return
	}

	route, err := s.catalog.Resolve(downstream, requestModel)
	if err != nil {
		s.fail(w, downstream, api.RequestView{
			RequestID:  requestID,
			Downstream: string(downstream),
			Model:      requestModel,
			StatusCode: http.StatusBadRequest,
			DurationMS: time.Since(start).Milliseconds(),
			Error:      err.Error(),
			CreatedAt:  time.Now().Format(time.RFC3339),
		})
		return
	}

	preparedBody, err := s.prepareRequestBody(downstream, route, body)
	if err != nil {
		status := mapPrepareErrorStatus(err)
		s.fail(w, downstream, api.RequestView{
			RequestID:  requestID,
			Downstream: string(downstream),
			Upstream:   string(route.Upstream),
			Model:      route.Model,
			StatusCode: status,
			DurationMS: time.Since(start).Milliseconds(),
			Error:      err.Error(),
			CreatedAt:  time.Now().Format(time.RFC3339),
		})
		return
	}

	upstreamURL, err := s.upstreamURL(route.Upstream)
	if err != nil {
		s.fail(w, downstream, api.RequestView{
			RequestID:  requestID,
			Downstream: string(downstream),
			Upstream:   string(route.Upstream),
			Model:      route.Model,
			StatusCode: http.StatusBadGateway,
			DurationMS: time.Since(start).Milliseconds(),
			Error:      err.Error(),
			CreatedAt:  time.Now().Format(time.RFC3339),
		})
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstreamURL, strings.NewReader(string(preparedBody)))
	if err != nil {
		s.fail(w, downstream, api.RequestView{
			RequestID:  requestID,
			Downstream: string(downstream),
			Upstream:   string(route.Upstream),
			Model:      route.Model,
			StatusCode: http.StatusBadGateway,
			DurationMS: time.Since(start).Milliseconds(),
			Error:      "failed to build upstream request",
			CreatedAt:  time.Now().Format(time.RFC3339),
		})
		return
	}
	s.applyRequestHeaders(req, r, route.Upstream)

	resp, err := s.client.Do(req)
	if err != nil {
		s.fail(w, downstream, api.RequestView{
			RequestID:  requestID,
			Downstream: string(downstream),
			Upstream:   string(route.Upstream),
			Model:      route.Model,
			StatusCode: http.StatusBadGateway,
			DurationMS: time.Since(start).Milliseconds(),
			Error:      fmt.Sprintf("upstream request failed: %v", err),
			CreatedAt:  time.Now().Format(time.RFC3339),
		})
		return
	}
	defer resp.Body.Close()

	copyResponseHeaders(w.Header(), resp.Header)
	w.Header().Set("X-ICOO-Request-ID", requestID)
	w.Header().Set("X-ICOO-Upstream-Protocol", string(route.Upstream))

	if route.Upstream == downstream {
		w.WriteHeader(resp.StatusCode)
		if isEventStream(resp.Header) {
			copyStream(w, resp.Body)
		} else {
			_, _ = io.Copy(w, resp.Body)
		}
		s.logRequest(api.RequestView{
			RequestID:  requestID,
			Downstream: string(downstream),
			Upstream:   string(route.Upstream),
			Model:      route.Model,
			StatusCode: resp.StatusCode,
			DurationMS: time.Since(start).Milliseconds(),
			CreatedAt:  time.Now().Format(time.RFC3339),
		})
		return
	}

	if isEventStream(resp.Header) {
		s.fail(w, downstream, api.RequestView{
			RequestID:  requestID,
			Downstream: string(downstream),
			Upstream:   string(route.Upstream),
			Model:      route.Model,
			StatusCode: http.StatusNotImplemented,
			DurationMS: time.Since(start).Milliseconds(),
			Error:      "streaming cross protocol translation is not implemented yet",
			CreatedAt:  time.Now().Format(time.RFC3339),
		})
		return
	}

	upstreamBody, err := io.ReadAll(resp.Body)
	if err != nil {
		s.fail(w, downstream, api.RequestView{
			RequestID:  requestID,
			Downstream: string(downstream),
			Upstream:   string(route.Upstream),
			Model:      route.Model,
			StatusCode: http.StatusBadGateway,
			DurationMS: time.Since(start).Milliseconds(),
			Error:      "failed to read upstream response body",
			CreatedAt:  time.Now().Format(time.RFC3339),
		})
		return
	}

	if resp.StatusCode >= 400 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(upstreamBody)
		s.logRequest(api.RequestView{
			RequestID:  requestID,
			Downstream: string(downstream),
			Upstream:   string(route.Upstream),
			Model:      route.Model,
			StatusCode: resp.StatusCode,
			DurationMS: time.Since(start).Milliseconds(),
			Error:      "upstream returned error",
			CreatedAt:  time.Now().Format(time.RFC3339),
		})
		return
	}

	translated, err := translateResponseBody(downstream, route.Upstream, route.Model, upstreamBody)
	if err != nil {
		s.fail(w, downstream, api.RequestView{
			RequestID:  requestID,
			Downstream: string(downstream),
			Upstream:   string(route.Upstream),
			Model:      route.Model,
			StatusCode: http.StatusBadGateway,
			DurationMS: time.Since(start).Milliseconds(),
			Error:      err.Error(),
			CreatedAt:  time.Now().Format(time.RFC3339),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(translated)
	s.logRequest(api.RequestView{
		RequestID:  requestID,
		Downstream: string(downstream),
		Upstream:   string(route.Upstream),
		Model:      route.Model,
		StatusCode: resp.StatusCode,
		DurationMS: time.Since(start).Milliseconds(),
		CreatedAt:  time.Now().Format(time.RFC3339),
	})
}

func (s *Service) authorize(r *http.Request) error {
	expected := s.cfg.AuthKeys()
	if len(expected) == 0 && s.cfg.AllowUnauthenticatedLocal {
		return nil
	}
	if len(expected) == 0 {
		return fmt.Errorf("proxy api key is required")
	}
	if slices.Contains(expected, strings.TrimSpace(r.Header.Get("x-api-key"))) {
		return nil
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") && slices.Contains(expected, strings.TrimSpace(auth[7:])) {
		return nil
	}
	return fmt.Errorf("invalid proxy api key")
}

func (s *Service) upstreamURL(protocol catalog.Protocol) (string, error) {
	switch protocol {
	case catalog.ProtocolAnthropic:
		if strings.TrimSpace(s.cfg.AnthropicAPIKey) == "" {
			return "", fmt.Errorf("anthropic upstream is not configured")
		}
		return strings.TrimRight(s.cfg.AnthropicBaseURL, "/") + "/v1/messages", nil
	case catalog.ProtocolOpenAIChat:
		if strings.TrimSpace(s.cfg.OpenAIApiKey) == "" {
			return "", fmt.Errorf("openai upstream is not configured")
		}
		return strings.TrimRight(s.cfg.OpenAIBaseURL, "/") + "/v1/chat/completions", nil
	case catalog.ProtocolOpenAIResponse:
		if strings.TrimSpace(s.cfg.OpenAIApiKey) == "" {
			return "", fmt.Errorf("openai upstream is not configured")
		}
		return strings.TrimRight(s.cfg.OpenAIBaseURL, "/") + "/v1/responses", nil
	default:
		return "", fmt.Errorf("unsupported upstream protocol %q", protocol)
	}
}

func (s *Service) applyRequestHeaders(target *http.Request, source *http.Request, protocol catalog.Protocol) {
	target.Header.Set("Content-Type", "application/json")
	if accept := strings.TrimSpace(source.Header.Get("Accept")); accept != "" {
		target.Header.Set("Accept", accept)
	}
	switch protocol {
	case catalog.ProtocolAnthropic:
		target.Header.Set("x-api-key", s.cfg.AnthropicAPIKey)
		target.Header.Set("anthropic-version", s.cfg.AnthropicVersion)
		if beta := strings.TrimSpace(source.Header.Get("anthropic-beta")); beta != "" {
			target.Header.Set("anthropic-beta", beta)
		}
	case catalog.ProtocolOpenAIChat, catalog.ProtocolOpenAIResponse:
		target.Header.Set("Authorization", "Bearer "+s.cfg.OpenAIApiKey)
		if value := strings.TrimSpace(source.Header.Get("OpenAI-Beta")); value != "" {
			target.Header.Set("OpenAI-Beta", value)
		}
	}
}

func extractModel(body []byte) (string, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("invalid json body")
	}
	model, _ := payload["model"].(string)
	return strings.TrimSpace(model), nil
}

func rewriteModel(body []byte, model string) ([]byte, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid json body")
	}
	payload["model"] = model
	rewritten, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to rewrite request body")
	}
	return rewritten, nil
}

func (s *Service) prepareRequestBody(downstream catalog.Protocol, route catalog.Route, body []byte) ([]byte, error) {
	if route.Upstream == downstream {
		return rewriteModel(body, route.Model)
	}
	switch {
	case downstream == catalog.ProtocolOpenAIChat && route.Upstream == catalog.ProtocolOpenAIResponse:
		return translateChatToResponsesRequest(body, route.Model)
	case downstream == catalog.ProtocolOpenAIResponse && route.Upstream == catalog.ProtocolOpenAIChat:
		return translateResponsesToChatRequest(body, route.Model)
	case downstream == catalog.ProtocolAnthropic && route.Upstream == catalog.ProtocolOpenAIResponse:
		return translateAnthropicToResponsesRequest(body, route.Model)
	case downstream == catalog.ProtocolOpenAIResponse && route.Upstream == catalog.ProtocolAnthropic:
		return translateResponsesToAnthropicRequest(body, route.Model)
	case downstream == catalog.ProtocolAnthropic && route.Upstream == catalog.ProtocolOpenAIChat:
		return translateAnthropicToChatRequest(body, route.Model)
	case downstream == catalog.ProtocolOpenAIChat && route.Upstream == catalog.ProtocolAnthropic:
		return translateChatToAnthropicRequest(body, route.Model)
	default:
		return nil, fmt.Errorf("cross protocol translation from %s to %s is not implemented yet", downstream, route.Upstream)
	}
}

func translateResponseBody(downstream, upstream catalog.Protocol, model string, body []byte) ([]byte, error) {
	switch {
	case downstream == catalog.ProtocolOpenAIChat && upstream == catalog.ProtocolOpenAIResponse:
		return translateResponsesToChatResponse(body, model)
	case downstream == catalog.ProtocolOpenAIResponse && upstream == catalog.ProtocolOpenAIChat:
		return translateChatToResponsesResponse(body, model)
	case downstream == catalog.ProtocolAnthropic && upstream == catalog.ProtocolOpenAIResponse:
		return translateResponsesToAnthropicResponse(body, model)
	case downstream == catalog.ProtocolOpenAIResponse && upstream == catalog.ProtocolAnthropic:
		return translateAnthropicToResponsesResponse(body, model)
	case downstream == catalog.ProtocolAnthropic && upstream == catalog.ProtocolOpenAIChat:
		return translateChatToAnthropicResponse(body, model)
	case downstream == catalog.ProtocolOpenAIChat && upstream == catalog.ProtocolAnthropic:
		return translateAnthropicToChatResponse(body, model)
	default:
		return nil, fmt.Errorf("cross protocol response translation from %s to %s is not implemented yet", upstream, downstream)
	}
}

func writeProtocolError(w http.ResponseWriter, protocol catalog.Protocol, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	switch protocol {
	case catalog.ProtocolAnthropic:
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"type": "error",
			"error": map[string]string{
				"type":    "invalid_request_error",
				"message": message,
			},
		})
	default:
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"type":    "invalid_request_error",
				"message": message,
			},
		})
	}
}

func (s *Service) fail(w http.ResponseWriter, protocol catalog.Protocol, item api.RequestView) {
	s.recordRequest(item)
	writeProtocolError(w, protocol, item.StatusCode, item.Error)
}

func mapPrepareErrorStatus(err error) int {
	if strings.Contains(strings.ToLower(err.Error()), "not implemented") {
		return http.StatusNotImplemented
	}
	return http.StatusBadRequest
}

func copyResponseHeaders(dst, src http.Header) {
	for key, values := range src {
		switch strings.ToLower(key) {
		case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization", "te", "trailers", "transfer-encoding", "upgrade", "content-length":
			continue
		}
		dst.Del(key)
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func isEventStream(header http.Header) bool {
	return strings.Contains(strings.ToLower(header.Get("Content-Type")), "text/event-stream")
}

func copyStream(w http.ResponseWriter, body io.Reader) {
	flusher, _ := w.(http.Flusher)
	buffer := make([]byte, 4096)
	for {
		n, err := body.Read(buffer)
		if n > 0 {
			_, _ = w.Write(buffer[:n])
			if flusher != nil {
				flusher.Flush()
			}
		}
		if err != nil {
			if err != io.EOF {
				log.Printf("icoo_proxy stream relay error: %v", err)
			}
			return
		}
	}
}

func newRequestID() string {
	var data [8]byte
	if _, err := rand.Read(data[:]); err != nil {
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return "req-" + hex.EncodeToString(data[:])
}

func (s *Service) RecentRequests() []api.RequestView {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Clone(s.recent)
}

func (s *Service) logRequest(item api.RequestView) {
	s.recordRequest(item)
	log.Printf("icoo_proxy request_id=%s downstream=%s upstream=%s model=%s status=%d duration_ms=%d", item.RequestID, item.Downstream, item.Upstream, item.Model, item.StatusCode, item.DurationMS)
}

func (s *Service) recordRequest(item api.RequestView) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recent = append([]api.RequestView{item}, s.recent...)
	if len(s.recent) > 12 {
		s.recent = s.recent[:12]
	}
}
