package catalog

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"icoo_proxy/internal/config"
	"icoo_proxy/internal/consts"
)

type Route struct {
	Name     string          `json:"name"`
	Upstream consts.Protocol `json:"upstream"`
	Model    string          `json:"model"`
}

type Catalog struct {
	defaults map[consts.Protocol]Route
	aliases  map[string]Route
}

func New(cfg config.Config) (*Catalog, error) {
	defaults := make(map[consts.Protocol]Route)
	if strings.TrimSpace(cfg.DefaultAnthropicRoute) != "" {
		route, err := parseTarget("default:anthropic", cfg.DefaultAnthropicRoute)
		if err != nil {
			return nil, err
		}
		defaults[consts.ProtocolAnthropic] = route
	}
	if strings.TrimSpace(cfg.DefaultChatRoute) != "" {
		route, err := parseTarget("default:chat", cfg.DefaultChatRoute)
		if err != nil {
			return nil, err
		}
		defaults[consts.ProtocolOpenAIChat] = route
	}
	if strings.TrimSpace(cfg.DefaultResponsesRoute) != "" {
		route, err := parseTarget("default:responses", cfg.DefaultResponsesRoute)
		if err != nil {
			return nil, err
		}
		defaults[consts.ProtocolOpenAIResponses] = route
	}

	aliases := make(map[string]Route)
	for _, entry := range splitEntries(cfg.ModelRoutes) {
		alias, target, found := strings.Cut(entry, "=")
		if !found {
			return nil, fmt.Errorf("invalid model route entry %q", entry)
		}
		alias = strings.TrimSpace(alias)
		if alias == "" {
			return nil, fmt.Errorf("invalid empty alias in model routes")
		}
		route, err := parseTarget(alias, target)
		if err != nil {
			return nil, err
		}
		route.Name = alias
		aliases[alias] = route
	}
	return &Catalog{
		defaults: defaults,
		aliases:  aliases,
	}, nil
}

func (c *Catalog) Resolve(downstream consts.Protocol, requestedModel string) (Route, error) {
	model := strings.TrimSpace(requestedModel)
	slog.Info("resolve model", "model", model, "downstream", downstream)

	route, ok := c.defaults[downstream]
	if model == "" {
		if !ok {
			return Route{}, fmt.Errorf("missing model and no default route for %s", downstream)
		}
		return route, nil
	}
	if route, ok := c.aliases[model]; ok {
		return route, nil
	}
	if !ok {
		return Route{}, fmt.Errorf("requested model %q has no default route for %s", model, downstream)
	}
	copyRoute := route
	copyRoute.Name = model
	copyRoute.Model = model
	return copyRoute, nil
}

func (c *Catalog) Defaults() []Route {
	items := make([]Route, 0, len(c.defaults))
	for protocol, route := range c.defaults {
		copyRoute := route
		copyRoute.Name = string(protocol)
		items = append(items, copyRoute)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return items
}

func (c *Catalog) Aliases() []Route {
	items := make([]Route, 0, len(c.aliases))
	for _, route := range c.aliases {
		items = append(items, route)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return items
}

func splitEntries(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n' || r == ';'
	})
	items := make([]string, 0, len(fields))
	for _, field := range fields {
		value := strings.TrimSpace(field)
		if value != "" {
			items = append(items, value)
		}
	}
	return items
}

func parseTarget(name, raw string) (Route, error) {
	value := strings.TrimSpace(raw)
	protocolRaw, model, found := strings.Cut(value, ":")
	if !found {
		return Route{}, fmt.Errorf("invalid route %q for %s", raw, name)
	}
	protocol := consts.Protocol(strings.TrimSpace(protocolRaw))
	switch protocol {
	case consts.ProtocolAnthropic, consts.ProtocolOpenAIChat, consts.ProtocolOpenAIResponses:
	default:
		return Route{}, fmt.Errorf("unsupported upstream protocol %q for %s", protocol, name)
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return Route{}, fmt.Errorf("missing model for %s", name)
	}
	return Route{
		Name:     name,
		Upstream: protocol,
		Model:    model,
	}, nil
}
