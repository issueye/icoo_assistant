package bootstrap

import (
	"fmt"
	"strings"

	"icoo_proxy/internal/config"
	"icoo_proxy/internal/routepolicy"
	"icoo_proxy/internal/supplier"
)

func ApplyRoutePolicies(cfg config.Config, suppliers *supplier.Service, policies *routepolicy.Service) (config.Config, error) {
	if suppliers == nil || policies == nil {
		return cfg, nil
	}
	for _, policy := range policies.Enabled() {
		snapshot, ok := suppliers.Resolve(policy.SupplierID)
		if !ok {
			return cfg, fmt.Errorf("route policy supplier %q not found", policy.SupplierID)
		}
		target := strings.TrimSpace(policy.UpstreamProtocol) + ":" + strings.TrimSpace(policy.TargetModel)
		switch policy.DownstreamProtocol {
		case "anthropic":
			cfg.DefaultAnthropicRoute = target
		case "openai-chat":
			cfg.DefaultChatRoute = target
		case "openai-responses":
			cfg.DefaultResponsesRoute = target
		}
		switch strings.TrimSpace(snapshot.Protocol) {
		case "anthropic":
			cfg.AnthropicBaseURL = snapshot.BaseURL
			cfg.AnthropicAPIKey = snapshot.APIKey
			cfg.AnthropicOnlyStream = snapshot.OnlyStream
			cfg.AnthropicUserAgent = snapshot.UserAgent
		case "openai-chat":
			cfg.OpenAIChatBaseURL = snapshot.BaseURL
			cfg.OpenAIChatAPIKey = snapshot.APIKey
			cfg.OpenAIChatOnlyStream = snapshot.OnlyStream
			cfg.OpenAIChatUserAgent = snapshot.UserAgent
		case "openai-responses":
			cfg.OpenAIResponsesBaseURL = snapshot.BaseURL
			cfg.OpenAIResponsesAPIKey = snapshot.APIKey
			cfg.OpenAIResponsesOnlyStream = snapshot.OnlyStream
			cfg.OpenAIResponsesUserAgent = snapshot.UserAgent
		}
	}
	return cfg, nil
}
