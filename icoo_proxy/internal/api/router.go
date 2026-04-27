package api

import (
	"encoding/json"
	"icoo_proxy/internal/consts"
	"net/http"
)

type State struct {
	Service                   string                 `json:"service"`
	Version                   string                 `json:"version"`
	Running                   bool                   `json:"running"`
	ListenAddr                string                 `json:"listen_addr,omitempty"`
	ProxyURL                  string                 `json:"proxy_url,omitempty"`
	LastError                 string                 `json:"last_error,omitempty"`
	AuthRequired              bool                   `json:"auth_required"`
	AuthKeyCount              int                    `json:"auth_key_count"`
	AllowUnauthenticatedLocal bool                   `json:"allow_unauthenticated_local"`
	SupportedPaths            []string               `json:"supported_paths"`
	Defaults                  []RouteView            `json:"defaults"`
	Aliases                   []RouteView            `json:"aliases"`
	Upstreams                 []UpstreamView         `json:"upstreams"`
	Endpoints                 []EndpointView         `json:"endpoints"`
	RoutePolicies             []RoutePolicyView      `json:"route_policies"`
	RecentRequests            []RequestView          `json:"recent_requests"`
	Notes                     []string               `json:"notes"`
	Checks                    map[string]interface{} `json:"checks"`
}

type RouteView struct {
	Name     string `json:"name"`
	Upstream string `json:"upstream"`
	Model    string `json:"model"`
}

type UpstreamView struct {
	Protocol   consts.Protocol `json:"protocol"`
	BaseURL    string          `json:"base_url,omitempty"`
	Configured bool            `json:"configured"`
}

type RequestView struct {
	RequestID  string `json:"request_id"`
	Downstream string `json:"downstream"`
	Upstream   string `json:"upstream"`
	Model      string `json:"model"`
	StatusCode int    `json:"status_code"`
	DurationMS int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
	CreatedAt  string `json:"created_at"`
}

type RoutePolicyView struct {
	ID                 string          `json:"id"`
	DownstreamProtocol consts.Protocol `json:"downstream_protocol"`
	SupplierID         string          `json:"supplier_id"`
	SupplierName       string          `json:"supplier_name"`
	UpstreamProtocol   consts.Protocol `json:"upstream_protocol"`
	Enabled            bool            `json:"enabled"`
	UpdatedAt          string          `json:"updated_at"`
	CreatedAt          string          `json:"created_at"`
}

type EndpointView struct {
	ID          string          `json:"id"`
	Path        string          `json:"path"`
	Protocol    consts.Protocol `json:"protocol"`
	Description string          `json:"description"`
	Enabled     bool            `json:"enabled"`
	BuiltIn     bool            `json:"built_in"`
	UpdatedAt   string          `json:"updated_at"`
	CreatedAt   string          `json:"created_at"`
}

type StateProvider interface {
	State() State
}

type ProxyHandler interface {
	Handle(w http.ResponseWriter, r *http.Request, downstream consts.Protocol)
}

type EndpointRoute struct {
	Path     string
	Protocol consts.Protocol
}

func NewMux(provider StateProvider, proxy ProxyHandler, endpoints []EndpointRoute) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, provider.State())
	})
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"service": provider.State().Service,
			"status":  "ok",
		})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		state := provider.State()
		statusCode := http.StatusOK
		if !state.Running {
			statusCode = http.StatusServiceUnavailable
		}
		writeJSON(w, statusCode, map[string]interface{}{
			"service": state.Service,
			"ready":   state.Running,
			"checks":  state.Checks,
		})
	})
	mux.HandleFunc("/admin/models", func(w http.ResponseWriter, r *http.Request) {
		state := provider.State()
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"defaults": state.Defaults,
			"aliases":  state.Aliases,
		})
	})
	mux.HandleFunc("/admin/routes", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"supported_paths": provider.State().SupportedPaths,
			"notes":           provider.State().Notes,
		})
	})
	mux.HandleFunc("/admin/requests", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"items": provider.State().RecentRequests,
		})
	})
	for _, endpoint := range endpoints {
		route := endpoint
		mux.HandleFunc(route.Path, func(w http.ResponseWriter, r *http.Request) {
			proxy.Handle(w, r, route.Protocol)
		})
	}
	return mux
}

func writeJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
