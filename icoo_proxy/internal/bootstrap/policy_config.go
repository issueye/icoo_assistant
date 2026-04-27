package bootstrap

import (
	"fmt"
	"strings"

	"icoo_proxy/internal/config"
	"icoo_proxy/internal/consts"
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
		defaultModel := strings.TrimSpace(snapshot.DefaultModel)
		if defaultModel == "" {
			return cfg, fmt.Errorf("route policy supplier %q default model is required", snapshot.Name)
		}
		target := snapshot.Protocol.ToString() + ":" + defaultModel
		switch policy.DownstreamProtocol {
		case consts.ProtocolAnthropic:
			cfg.DefaultAnthropicRoute = target
		case consts.ProtocolOpenAIChat:
			cfg.DefaultChatRoute = target
		case consts.ProtocolOpenAIResponses:
			cfg.DefaultResponsesRoute = target
		}
		switch snapshot.Protocol {
		case consts.ProtocolAnthropic:
			cfg.AnthropicBaseURL = snapshot.BaseURL
			cfg.AnthropicAPIKey = snapshot.APIKey
			cfg.AnthropicOnlyStream = snapshot.OnlyStream
			cfg.AnthropicUserAgent = snapshot.UserAgent
		case consts.ProtocolOpenAIChat:
			cfg.OpenAIChatBaseURL = snapshot.BaseURL
			cfg.OpenAIChatAPIKey = snapshot.APIKey
			cfg.OpenAIChatOnlyStream = snapshot.OnlyStream
			cfg.OpenAIChatUserAgent = snapshot.UserAgent
		case consts.ProtocolOpenAIResponses:
			cfg.OpenAIResponsesBaseURL = snapshot.BaseURL
			cfg.OpenAIResponsesAPIKey = snapshot.APIKey
			cfg.OpenAIResponsesOnlyStream = snapshot.OnlyStream
			cfg.OpenAIResponsesUserAgent = snapshot.UserAgent
		}
	}
	return cfg, nil
}
