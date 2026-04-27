package proxy

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"icoo_proxy/internal/api"
	"icoo_proxy/internal/catalog"
	"icoo_proxy/internal/config"
	"icoo_proxy/internal/consts"
)

type Service struct {
	cfg     config.Config
	catalog *catalog.Catalog
	client  *http.Client
	logger  *slog.Logger
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

func (s *Service) SetChainLogger(logger *slog.Logger) {
	s.logger = logger
}

func (s *Service) Handle(w http.ResponseWriter, r *http.Request, downstream consts.Protocol) {
	requestID := newRequestID()
	start := time.Now()
	w.Header().Set("X-ICOO-Request-ID", requestID)

	if r.Method != http.MethodPost {
		s.logChain("downstream.request.rejected",
			"request_id", requestID,
			"downstream", downstream.ToString(),
			"method", r.Method,
			"path", r.URL.Path,
			"headers", sanitizedHeaders(r.Header),
			"error", "method not allowed",
		)
		s.fail(w, downstream, api.RequestView{
			RequestID:  requestID,
			Downstream: downstream.ToString(),
			StatusCode: http.StatusMethodNotAllowed,
			DurationMS: time.Since(start).Milliseconds(),
			Error:      "method not allowed",
			CreatedAt:  time.Now().Format(time.RFC3339),
		})
		return
	}
	if err := s.authorize(r); err != nil {
		s.logChain("downstream.request.rejected",
			"request_id", requestID,
			"downstream", string(downstream),
			"method", r.Method,
			"path", r.URL.Path,
			"headers", sanitizedHeaders(r.Header),
			"error", err.Error(),
		)
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
	s.logChain("downstream.request",
		"request_id", requestID,
		"downstream", string(downstream),
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
		"headers", sanitizedHeaders(r.Header),
		"body", s.logBody(body),
	)

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
	s.logChain("route.resolved",
		"request_id", requestID,
		"downstream", string(downstream),
		"requested_model", requestModel,
		"upstream", string(route.Upstream),
		"target_model", route.Model,
		"route_name", route.Name,
	)

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
	downstreamWantsStream := requestUsesStreaming(body)
	forcedUpstreamStream := false
	if s.shouldForceUpstreamStream(route.Upstream) && !requestUsesStreaming(preparedBody) {
		preparedBody, err = forceStreamRequest(preparedBody)
		if err != nil {
			s.fail(w, downstream, api.RequestView{
				RequestID:  requestID,
				Downstream: string(downstream),
				Upstream:   string(route.Upstream),
				Model:      route.Model,
				StatusCode: http.StatusBadRequest,
				DurationMS: time.Since(start).Milliseconds(),
				Error:      err.Error(),
				CreatedAt:  time.Now().Format(time.RFC3339),
			})
			return
		}
		forcedUpstreamStream = true
	}
	s.logChain("conversion.request",
		"request_id", requestID,
		"downstream", string(downstream),
		"upstream", string(route.Upstream),
		"target_model", route.Model,
		"translated", route.Upstream != downstream,
		"forced_stream", forcedUpstreamStream,
		"input_body", s.logBody(body),
		"output_body", s.logBody(preparedBody),
	)

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
	s.logChain("upstream.request",
		"request_id", requestID,
		"downstream", string(downstream),
		"upstream", string(route.Upstream),
		"method", req.Method,
		"url", upstreamURL,
		"headers", sanitizedHeaders(req.Header),
		"body", s.logBody(preparedBody),
	)

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
		if isEventStream(resp.Header) {
			s.logChain("upstream.response",
				"request_id", requestID,
				"downstream", string(downstream),
				"upstream", string(route.Upstream),
				"status_code", resp.StatusCode,
				"headers", sanitizedHeaders(resp.Header),
				"body", "<event-stream body not captured>",
			)
			if downstreamWantsStream {
				w.WriteHeader(resp.StatusCode)
				s.logChain("downstream.response",
					"request_id", requestID,
					"downstream", string(downstream),
					"upstream", string(route.Upstream),
					"status_code", resp.StatusCode,
					"headers", sanitizedHeaders(w.Header()),
					"body", "<event-stream body not captured>",
				)
				copyStream(w, resp.Body)
			} else if route.Upstream == consts.ProtocolOpenAIResponses {
				aggregatedBody, aggregateErr := aggregateResponsesStreamToJSON(resp.Body)
				if aggregateErr != nil {
					s.fail(w, downstream, api.RequestView{
						RequestID:  requestID,
						Downstream: string(downstream),
						Upstream:   string(route.Upstream),
						Model:      route.Model,
						StatusCode: http.StatusBadGateway,
						DurationMS: time.Since(start).Milliseconds(),
						Error:      aggregateErr.Error(),
						CreatedAt:  time.Now().Format(time.RFC3339),
					})
					return
				}
				s.logChain("conversion.stream.aggregate",
					"request_id", requestID,
					"downstream", string(downstream),
					"upstream", string(route.Upstream),
					"target_model", route.Model,
					"output_body", s.logBody(aggregatedBody),
				)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(resp.StatusCode)
				_, _ = w.Write(aggregatedBody)
				s.logChain("downstream.response",
					"request_id", requestID,
					"downstream", string(downstream),
					"upstream", string(route.Upstream),
					"status_code", resp.StatusCode,
					"headers", sanitizedHeaders(w.Header()),
					"body", s.logBody(aggregatedBody),
				)
			} else {
				s.fail(w, downstream, api.RequestView{
					RequestID:  requestID,
					Downstream: string(downstream),
					Upstream:   string(route.Upstream),
					Model:      route.Model,
					StatusCode: http.StatusNotImplemented,
					DurationMS: time.Since(start).Milliseconds(),
					Error:      "stream-only upstream aggregation is not implemented for this protocol",
					CreatedAt:  time.Now().Format(time.RFC3339),
				})
				return
			}
		} else {
			upstreamBody, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
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
			s.logChain("upstream.response",
				"request_id", requestID,
				"downstream", string(downstream),
				"upstream", string(route.Upstream),
				"status_code", resp.StatusCode,
				"headers", sanitizedHeaders(resp.Header),
				"body", s.logBody(upstreamBody),
			)
			s.logChain("downstream.response",
				"request_id", requestID,
				"downstream", string(downstream),
				"upstream", string(route.Upstream),
				"status_code", resp.StatusCode,
				"headers", sanitizedHeaders(w.Header()),
				"body", s.logBody(upstreamBody),
			)
			w.WriteHeader(resp.StatusCode)
			_, _ = w.Write(upstreamBody)
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
		s.logChain("upstream.response",
			"request_id", requestID,
			"downstream", string(downstream),
			"upstream", string(route.Upstream),
			"status_code", resp.StatusCode,
			"headers", sanitizedHeaders(resp.Header),
			"body", "<event-stream body not captured>",
		)
		switch {
		case downstream == consts.ProtocolAnthropic && route.Upstream == consts.ProtocolOpenAIResponses && downstreamWantsStream:
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Del("Content-Length")
			w.WriteHeader(resp.StatusCode)
			err = s.translateResponsesStreamToAnthropic(w, resp.Body, route.Model, requestID)
			s.logChain("downstream.response",
				"request_id", requestID,
				"downstream", string(downstream),
				"upstream", string(route.Upstream),
				"status_code", resp.StatusCode,
				"headers", sanitizedHeaders(w.Header()),
				"body", "<translated event-stream body not captured>",
			)
			item := api.RequestView{
				RequestID:  requestID,
				Downstream: string(downstream),
				Upstream:   string(route.Upstream),
				Model:      route.Model,
				StatusCode: resp.StatusCode,
				DurationMS: time.Since(start).Milliseconds(),
				CreatedAt:  time.Now().Format(time.RFC3339),
			}
			if err != nil {
				item.Error = err.Error()
				s.logChain("conversion.stream.error",
					"request_id", requestID,
					"downstream", string(downstream),
					"upstream", string(route.Upstream),
					"error", err.Error(),
				)
			}
			s.logRequest(item)
			return
		case !downstreamWantsStream && route.Upstream == consts.ProtocolOpenAIResponses:
			aggregatedBody, aggregateErr := aggregateResponsesStreamToJSON(resp.Body)
			if aggregateErr != nil {
				s.fail(w, downstream, api.RequestView{
					RequestID:  requestID,
					Downstream: string(downstream),
					Upstream:   string(route.Upstream),
					Model:      route.Model,
					StatusCode: http.StatusBadGateway,
					DurationMS: time.Since(start).Milliseconds(),
					Error:      aggregateErr.Error(),
					CreatedAt:  time.Now().Format(time.RFC3339),
				})
				return
			}
			s.logChain("conversion.stream.aggregate",
				"request_id", requestID,
				"downstream", string(downstream),
				"upstream", string(route.Upstream),
				"target_model", route.Model,
				"output_body", s.logBody(aggregatedBody),
			)
			translated, translateErr := translateResponseBody(downstream, route.Upstream, route.Model, aggregatedBody)
			if translateErr != nil {
				s.fail(w, downstream, api.RequestView{
					RequestID:  requestID,
					Downstream: string(downstream),
					Upstream:   string(route.Upstream),
					Model:      route.Model,
					StatusCode: http.StatusBadGateway,
					DurationMS: time.Since(start).Milliseconds(),
					Error:      translateErr.Error(),
					CreatedAt:  time.Now().Format(time.RFC3339),
				})
				return
			}
			s.logChain("conversion.response",
				"request_id", requestID,
				"downstream", string(downstream),
				"upstream", string(route.Upstream),
				"target_model", route.Model,
				"translated", route.Upstream != downstream,
				"input_body", s.logBody(aggregatedBody),
				"output_body", s.logBody(translated),
			)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(resp.StatusCode)
			_, _ = w.Write(translated)
			s.logChain("downstream.response",
				"request_id", requestID,
				"downstream", string(downstream),
				"upstream", string(route.Upstream),
				"status_code", resp.StatusCode,
				"headers", sanitizedHeaders(w.Header()),
				"body", s.logBody(translated),
			)
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
		default:
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
	s.logChain("upstream.response",
		"request_id", requestID,
		"downstream", string(downstream),
		"upstream", string(route.Upstream),
		"status_code", resp.StatusCode,
		"headers", sanitizedHeaders(resp.Header),
		"body", s.logBody(upstreamBody),
	)

	if resp.StatusCode >= 400 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(upstreamBody)
		s.logChain("downstream.response",
			"request_id", requestID,
			"downstream", string(downstream),
			"upstream", string(route.Upstream),
			"status_code", resp.StatusCode,
			"headers", sanitizedHeaders(w.Header()),
			"body", s.logBody(upstreamBody),
		)
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
	s.logChain("conversion.response",
		"request_id", requestID,
		"downstream", string(downstream),
		"upstream", string(route.Upstream),
		"target_model", route.Model,
		"translated", route.Upstream != downstream,
		"input_body", s.logBody(upstreamBody),
		"output_body", s.logBody(translated),
	)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(translated)
	s.logChain("downstream.response",
		"request_id", requestID,
		"downstream", string(downstream),
		"upstream", string(route.Upstream),
		"status_code", resp.StatusCode,
		"headers", sanitizedHeaders(w.Header()),
		"body", s.logBody(translated),
	)
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

func (s *Service) shouldForceUpstreamStream(protocol consts.Protocol) bool {
	switch protocol {
	case consts.ProtocolOpenAIResponses:
		return s.cfg.OpenAIResponsesOnlyStreamValue()
	default:
		return false
	}
}

func (s *Service) authorize(r *http.Request) error {
	expected := s.cfg.AuthKeys()
	if len(expected) == 0 && s.cfg.AllowUnauthenticatedLocal {
		if isLocalRequest(r) {
			return nil
		}
		return fmt.Errorf("proxy api key is required")
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

func isLocalRequest(r *http.Request) bool {
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		host = strings.TrimSpace(r.RemoteAddr)
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// upstreamURL 获取指定协议的上游URL
func (s *Service) upstreamURL(protocol consts.Protocol) (string, error) {
	switch protocol {
	case consts.ProtocolAnthropic:
		if strings.TrimSpace(s.cfg.AnthropicAPIKey) == "" {
			return "", fmt.Errorf("anthropic upstream is not configured")
		}
		return strings.TrimRight(s.cfg.AnthropicBaseURL, "/") + "/v1/messages", nil
	case consts.ProtocolOpenAIChat:
		if strings.TrimSpace(s.cfg.OpenAIChatAPIKeyValue()) == "" {
			return "", fmt.Errorf("openai chat upstream is not configured")
		}
		return strings.TrimRight(s.cfg.OpenAIChatBaseURLValue(), "/") + "/v1/chat/completions", nil
	case consts.ProtocolOpenAIResponses:
		if strings.TrimSpace(s.cfg.OpenAIResponsesAPIKeyValue()) == "" {
			return "", fmt.Errorf("openai responses upstream is not configured")
		}
		return strings.TrimRight(s.cfg.OpenAIResponsesBaseURLValue(), "/") + "/v1/responses", nil
	default:
		return "", fmt.Errorf("unsupported upstream protocol %q", protocol)
	}
}

// applyRequestHeaders 应用请求头到目标请求
func (s *Service) applyRequestHeaders(target *http.Request, source *http.Request, protocol consts.Protocol) {
	target.Header.Set("Content-Type", "application/json")
	if accept := strings.TrimSpace(source.Header.Get("Accept")); accept != "" {
		target.Header.Set("Accept", accept)
	}
	switch protocol {
	case consts.ProtocolAnthropic:
		target.Header.Set("x-api-key", s.cfg.AnthropicAPIKey)
		target.Header.Set("anthropic-version", s.cfg.AnthropicVersion)
		if userAgent := strings.TrimSpace(s.cfg.AnthropicUserAgent); userAgent != "" {
			target.Header.Set("User-Agent", userAgent)
		}
		if beta := strings.TrimSpace(source.Header.Get("anthropic-beta")); beta != "" {
			target.Header.Set("anthropic-beta", beta)
		}
	case consts.ProtocolOpenAIChat:
		target.Header.Set("Authorization", "Bearer "+s.cfg.OpenAIChatAPIKeyValue())
		if userAgent := strings.TrimSpace(s.cfg.OpenAIChatUserAgentValue()); userAgent != "" {
			target.Header.Set("User-Agent", userAgent)
		}
		if value := strings.TrimSpace(source.Header.Get("OpenAI-Beta")); value != "" {
			target.Header.Set("OpenAI-Beta", value)
		}
	case consts.ProtocolOpenAIResponses:
		target.Header.Set("Authorization", "Bearer "+s.cfg.OpenAIResponsesAPIKeyValue())
		if userAgent := strings.TrimSpace(s.cfg.OpenAIResponsesUserAgentValue()); userAgent != "" {
			target.Header.Set("User-Agent", userAgent)
		}
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

func rewriteResponsesRequest(body []byte, model string) ([]byte, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid json body")
	}
	payload["model"] = model
	applyDefaultResponsesReasoning(payload)
	rewritten, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to rewrite request body")
	}
	return rewritten, nil
}

func requestUsesStreaming(body []byte) bool {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return false
	}
	stream, _ := payload["stream"].(bool)
	return stream
}

func forceStreamRequest(body []byte) ([]byte, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid json body")
	}
	payload["stream"] = true
	rewritten, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to rewrite request body")
	}
	return rewritten, nil
}

func (s *Service) prepareRequestBody(downstream consts.Protocol, route catalog.Route, body []byte) ([]byte, error) {
	if route.Upstream == downstream {
		if downstream == consts.ProtocolOpenAIResponses {
			return rewriteResponsesRequest(body, route.Model)
		}
		return rewriteModel(body, route.Model)
	}
	switch {
	case downstream == consts.ProtocolOpenAIChat && route.Upstream == consts.ProtocolOpenAIResponses:
		return translateChatToResponsesRequest(body, route.Model)
	case downstream == consts.ProtocolOpenAIResponses && route.Upstream == consts.ProtocolOpenAIChat:
		return translateResponsesToChatRequest(body, route.Model)
	case downstream == consts.ProtocolAnthropic && route.Upstream == consts.ProtocolOpenAIResponses:
		return translateAnthropicToResponsesRequest(body, route.Model)
	case downstream == consts.ProtocolOpenAIResponses && route.Upstream == consts.ProtocolAnthropic:
		return translateResponsesToAnthropicRequest(body, route.Model)
	case downstream == consts.ProtocolAnthropic && route.Upstream == consts.ProtocolOpenAIChat:
		return translateAnthropicToChatRequest(body, route.Model)
	case downstream == consts.ProtocolOpenAIChat && route.Upstream == consts.ProtocolAnthropic:
		return translateChatToAnthropicRequest(body, route.Model)
	default:
		return nil, fmt.Errorf("cross protocol translation from %s to %s is not implemented yet", downstream, route.Upstream)
	}
}

func translateResponseBody(downstream, upstream consts.Protocol, model string, body []byte) ([]byte, error) {
	switch {
	case downstream == consts.ProtocolOpenAIChat && upstream == consts.ProtocolOpenAIResponses:
		return translateResponsesToChatResponse(body, model)
	case downstream == consts.ProtocolOpenAIResponses && upstream == consts.ProtocolOpenAIChat:
		return translateChatToResponsesResponse(body, model)
	case downstream == consts.ProtocolAnthropic && upstream == consts.ProtocolOpenAIResponses:
		return translateResponsesToAnthropicResponse(body, model)
	case downstream == consts.ProtocolOpenAIResponses && upstream == consts.ProtocolAnthropic:
		return translateAnthropicToResponsesResponse(body, model)
	case downstream == consts.ProtocolAnthropic && upstream == consts.ProtocolOpenAIChat:
		return translateChatToAnthropicResponse(body, model)
	case downstream == consts.ProtocolOpenAIChat && upstream == consts.ProtocolAnthropic:
		return translateAnthropicToChatResponse(body, model)
	default:
		return nil, fmt.Errorf("cross protocol response translation from %s to %s is not implemented yet", upstream, downstream)
	}
}

func writeProtocolError(w http.ResponseWriter, protocol consts.Protocol, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	switch protocol {
	case consts.ProtocolAnthropic:
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

func (s *Service) fail(w http.ResponseWriter, protocol consts.Protocol, item api.RequestView) {
	s.logChain("downstream.response.error",
		"request_id", item.RequestID,
		"downstream", item.Downstream,
		"upstream", item.Upstream,
		"model", item.Model,
		"status_code", item.StatusCode,
		"duration_ms", item.DurationMS,
		"error", item.Error,
	)
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

func (s *Service) logChain(event string, attrs ...interface{}) {
	if s == nil || s.logger == nil {
		return
	}
	s.logger.Info(event, attrs...)
}

func (s *Service) logBody(body []byte) string {
	if s == nil || !s.cfg.ChainLogBodies {
		return "<body logging disabled>"
	}
	if body == nil {
		return ""
	}
	result := redactJSONBody(body)
	if max := s.cfg.ChainLogMaxBodyBytes; max > 0 && len([]byte(result)) > max {
		return string([]byte(result)[:max]) + "...<truncated>"
	}
	return result
}

func redactJSONBody(body []byte) string {
	var payload interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return string(body)
	}
	redacted := redactJSONValue(payload)
	data, err := json.Marshal(redacted)
	if err != nil {
		return string(body)
	}
	return string(data)
}

func redactJSONValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			if isSensitiveName(key) {
				result[key] = "<redacted>"
				continue
			}
			result[key] = redactJSONValue(item)
		}
		return result
	case []interface{}:
		result := make([]interface{}, 0, len(typed))
		for _, item := range typed {
			result = append(result, redactJSONValue(item))
		}
		return result
	default:
		return value
	}
}

func sanitizedHeaders(headers http.Header) map[string][]string {
	result := make(map[string][]string, len(headers))
	for key, values := range headers {
		if isSensitiveName(key) {
			result[key] = []string{"<redacted>"}
			continue
		}
		result[key] = slices.Clone(values)
	}
	return result
}

func isSensitiveName(name string) bool {
	normalized := strings.ToLower(strings.NewReplacer("-", "", "_", "", ".", "").Replace(name))
	for _, marker := range []string{"authorization", "apikey", "token", "secret", "password", "credential"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return normalized == "key"
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
	s.logChain("request.completed",
		"request_id", item.RequestID,
		"downstream", item.Downstream,
		"upstream", item.Upstream,
		"model", item.Model,
		"status_code", item.StatusCode,
		"duration_ms", item.DurationMS,
		"error", item.Error,
	)
}

func (s *Service) recordRequest(item api.RequestView) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recent = append([]api.RequestView{item}, s.recent...)
	if len(s.recent) > 12 {
		s.recent = s.recent[:12]
	}
}
