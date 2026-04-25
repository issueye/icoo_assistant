package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"icoo_proxy/internal/api"
	"icoo_proxy/internal/catalog"
	"icoo_proxy/internal/config"
	"icoo_proxy/internal/proxy"
	"icoo_proxy/internal/server"
	"icoo_proxy/internal/supplier"
)

type App struct {
	ctx        context.Context
	mu         sync.RWMutex
	root       string
	cfg        config.Config
	catalog    *catalog.Catalog
	service    *proxy.Service
	suppliers  *supplier.Service
	httpServer *http.Server
	listenAddr string
	running    bool
	lastError  string
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	root, err := os.Getwd()
	if err != nil {
		a.setLastError(err.Error())
		return
	}
	a.root = root
	suppliers, err := supplier.NewService(root)
	if err != nil {
		a.setLastError(err.Error())
		return
	}
	a.suppliers = suppliers
	if err := a.startProxy(); err != nil {
		a.setLastError(err.Error())
	}
}

func (a *App) shutdown(ctx context.Context) {
	_ = a.stopProxy(ctx)
}

func (a *App) GetOverview() map[string]interface{} {
	return stateToMap(a.State())
}

func (a *App) ReloadProxy() (map[string]interface{}, error) {
	if err := a.stopProxy(context.Background()); err != nil {
		a.setLastError(err.Error())
		return stateToMap(a.State()), err
	}
	if err := a.startProxy(); err != nil {
		a.setLastError(err.Error())
		return stateToMap(a.State()), err
	}
	return stateToMap(a.State()), nil
}

func (a *App) ListSuppliers() []supplier.Record {
	if a.suppliers == nil {
		return nil
	}
	return a.suppliers.List()
}

func (a *App) SaveSupplier(input supplier.UpsertInput) ([]supplier.Record, error) {
	if a.suppliers == nil {
		return nil, context.Canceled
	}
	if _, err := a.suppliers.Upsert(input); err != nil {
		return nil, err
	}
	return a.suppliers.List(), nil
}

func (a *App) DeleteSupplier(id string) ([]supplier.Record, error) {
	if a.suppliers == nil {
		return nil, context.Canceled
	}
	if err := a.suppliers.Delete(id); err != nil {
		return nil, err
	}
	return a.suppliers.List(), nil
}

func (a *App) State() api.State {
	a.mu.RLock()
	defer a.mu.RUnlock()

	state := api.State{
		Service:                   "icoo_proxy",
		Version:                   Version,
		Running:                   a.running,
		ListenAddr:                a.listenAddr,
		ProxyURL:                  proxyURL(a.listenAddr),
		LastError:                 a.lastError,
		AuthRequired:              strings.TrimSpace(a.cfg.ProxyAPIKey) != "",
		AllowUnauthenticatedLocal: a.cfg.AllowUnauthenticatedLocal,
		SupportedPaths: []string{
			"/healthz",
			"/readyz",
			"/admin/models",
			"/admin/routes",
			"/admin/requests",
			"/v1/messages",
			"/anthropic/v1/messages",
			"/v1/chat/completions",
			"/openai/v1/chat/completions",
			"/v1/responses",
			"/openai/v1/responses",
		},
		Upstreams: []api.UpstreamView{
			{
				Protocol:   string(catalog.ProtocolAnthropic),
				BaseURL:    a.cfg.AnthropicBaseURL,
				Configured: strings.TrimSpace(a.cfg.AnthropicAPIKey) != "",
			},
			{
				Protocol:   string(catalog.ProtocolOpenAIChat),
				BaseURL:    a.cfg.OpenAIBaseURL,
				Configured: strings.TrimSpace(a.cfg.OpenAIApiKey) != "",
			},
			{
				Protocol:   string(catalog.ProtocolOpenAIResponse),
				BaseURL:    a.cfg.OpenAIBaseURL,
				Configured: strings.TrimSpace(a.cfg.OpenAIApiKey) != "",
			},
		},
		Notes: []string{
			"Current build supports same-protocol forwarding and model alias rewriting.",
			"Current build also supports non-streaming chat/completions <-> responses translation.",
			"Current build also supports non-streaming anthropic messages <-> responses translation.",
			"Basic function tool definitions and non-streaming tool call/result mapping are now supported.",
			"Streaming cross-protocol translation is still planned.",
			"The desktop app starts the local proxy automatically during startup.",
		},
		Checks: map[string]interface{}{
			"proxy_running":       a.running,
			"anthropic_ready":     strings.TrimSpace(a.cfg.AnthropicAPIKey) != "",
			"openai_ready":        strings.TrimSpace(a.cfg.OpenAIApiKey) != "",
			"route_catalog_ready": a.catalog != nil,
			"supplier_store_ready": a.suppliers != nil,
		},
	}
	if a.catalog != nil {
		for _, route := range a.catalog.Defaults() {
			state.Defaults = append(state.Defaults, api.RouteView{
				Name:     route.Name,
				Upstream: string(route.Upstream),
				Model:    route.Model,
			})
		}
		for _, route := range a.catalog.Aliases() {
			state.Aliases = append(state.Aliases, api.RouteView{
				Name:     route.Name,
				Upstream: string(route.Upstream),
				Model:    route.Model,
			})
		}
	}
	if a.service != nil {
		state.RecentRequests = a.service.RecentRequests()
	}
	return state
}

func (a *App) startProxy() error {
	cfg, err := config.Load(a.root)
	if err != nil {
		return err
	}
	cat, err := catalog.New(cfg)
	if err != nil {
		return err
	}
	service := proxy.New(cfg, cat)
	handler := api.NewMux(a, service)
	srv := server.New(cfg, handler)
	listener, err := net.Listen("tcp", cfg.Addr())
	if err != nil {
		return err
	}
	listenAddr := listener.Addr().String()

	a.mu.Lock()
	a.cfg = cfg
	a.catalog = cat
	a.service = service
	a.httpServer = srv
	a.listenAddr = listenAddr
	a.running = true
	a.lastError = ""
	a.mu.Unlock()

	go func() {
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			a.setLastError(err.Error())
		}
	}()
	return nil
}

func (a *App) stopProxy(ctx context.Context) error {
	a.mu.Lock()
	srv := a.httpServer
	a.httpServer = nil
	a.running = false
	a.listenAddr = ""
	a.mu.Unlock()

	if srv == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return srv.Shutdown(ctx)
}

func (a *App) setLastError(message string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lastError = message
}

func proxyURL(addr string) string {
	if strings.TrimSpace(addr) == "" {
		return ""
	}
	return "http://" + addr
}

func stateToMap(state api.State) map[string]interface{} {
	return map[string]interface{}{
		"service":                     state.Service,
		"version":                     state.Version,
		"running":                     state.Running,
		"listen_addr":                 state.ListenAddr,
		"proxy_url":                   state.ProxyURL,
		"last_error":                  state.LastError,
		"auth_required":               state.AuthRequired,
		"allow_unauthenticated_local": state.AllowUnauthenticatedLocal,
		"supported_paths":             state.SupportedPaths,
		"defaults":                    state.Defaults,
		"aliases":                     state.Aliases,
		"upstreams":                   state.Upstreams,
		"recent_requests":             state.RecentRequests,
		"notes":                       state.Notes,
		"checks":                      state.Checks,
	}
}
