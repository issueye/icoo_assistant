package llm

import (
	"encoding/json"
	"fmt"
	"strings"

	"icoo_assistant/internal/config"

	anthropic "github.com/anthropics/anthropic-sdk-go"
)

type AnthropicConfig struct {
	Model             string
	MaxTokens         int64
	EnablePromptCache bool
	EnableThinking    bool
}

type AnthropicClient struct {
	client anthropic.Client
	config AnthropicConfig
}

func NewAnthropicClient(config AnthropicConfig) *AnthropicClient {
	model := strings.TrimSpace(config.Model)
	if model == "" {
		model = "claude-opus-4-7"
	}
	maxTokens := config.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 16000
	}
	config.Model = model
	config.MaxTokens = maxTokens
	return &AnthropicClient{
		client: anthropic.NewClient(),
		config: config,
	}
}

func NewClientFromConfig(cfg config.Config) (Client, string, error) {
	if strings.TrimSpace(cfg.AnthropicAPIKey) == "" {
		return &FakeClient{}, "fake", nil
	}
	return NewAnthropicClient(AnthropicConfig{
		Model:             cfg.AnthropicModel,
		MaxTokens:         cfg.AnthropicMaxTokens,
		EnablePromptCache: cfg.EnablePromptCache,
		EnableThinking:    cfg.EnableThinking,
	}), "anthropic", nil
}

func (c *AnthropicClient) CreateMessage(system string, messages []Message, tools []Tool) (Response, error) {
	requestMessages, err := c.buildMessages(messages)
	if err != nil {
		return Response{}, err
	}
	requestTools := c.buildTools(tools)
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(c.config.Model),
		MaxTokens: c.config.MaxTokens,
		Messages:  requestMessages,
		Tools:     requestTools,
	}
	if strings.TrimSpace(system) != "" {
		block := anthropic.TextBlockParam{Text: system}
		if c.config.EnablePromptCache {
			block.CacheControl = anthropic.NewCacheControlEphemeralParam()
		}
		params.System = []anthropic.TextBlockParam{block}
	}
	if c.config.EnableThinking {
		params.Thinking = anthropic.ThinkingConfigParamUnion{
			OfAdaptive: &anthropic.ThinkingConfigAdaptiveParam{
				Display: anthropic.ThinkingConfigAdaptiveDisplaySummarized,
			},
		}
	}
	resp, err := c.client.Messages.New(nil, params)
	if err != nil {
		return Response{}, err
	}
	result := Response{
		StopReason: string(resp.StopReason),
		Raw:        resp.ToParam(),
	}
	texts := make([]string, 0)
	for _, block := range resp.Content {
		switch variant := block.AsAny().(type) {
		case anthropic.TextBlock:
			texts = append(texts, variant.Text)
		case anthropic.ToolUseBlock:
			var input map[string]interface{}
			if err := json.Unmarshal([]byte(variant.JSON.Input.Raw()), &input); err != nil {
				return Response{}, fmt.Errorf("decode tool input for %s: %w", variant.Name, err)
			}
			result.ToolUses = append(result.ToolUses, ToolUse{
				ID:    variant.ID,
				Name:  variant.Name,
				Input: input,
			})
		}
	}
	result.Text = strings.TrimSpace(strings.Join(texts, "\n"))
	return result, nil
}

func (c *AnthropicClient) buildMessages(messages []Message) ([]anthropic.MessageParam, error) {
	result := make([]anthropic.MessageParam, 0, len(messages))
	for _, message := range messages {
		switch content := message.Content.(type) {
		case string:
			if message.Role != "user" {
				return nil, fmt.Errorf("string content only supported for user messages")
			}
			result = append(result, anthropic.NewUserMessage(anthropic.NewTextBlock(content)))
		case anthropic.MessageParam:
			result = append(result, content)
		case []ToolResultBlock:
			blocks := make([]anthropic.ContentBlockParamUnion, 0, len(content))
			for _, block := range content {
				blocks = append(blocks, anthropic.NewToolResultBlock(block.ToolUseID, block.Content, block.IsError))
			}
			result = append(result, anthropic.NewUserMessage(blocks...))
		default:
			return nil, fmt.Errorf("unsupported message content type %T", message.Content)
		}
	}
	return result, nil
}

func (c *AnthropicClient) buildTools(tools []Tool) []anthropic.ToolUnionParam {
	result := make([]anthropic.ToolUnionParam, 0, len(tools))
	for _, tool := range tools {
		param := anthropic.ToolParam{
			Name:        tool.Name,
			Description: anthropic.String(tool.Description),
			InputSchema: anthropic.ToolInputSchemaParam{},
		}
		if properties, ok := tool.InputSchema["properties"].(map[string]interface{}); ok {
			param.InputSchema.Properties = properties
		}
		if required, ok := tool.InputSchema["required"].([]string); ok {
			param.InputSchema.Required = required
		} else if requiredAny, ok := tool.InputSchema["required"].([]interface{}); ok {
			required := make([]string, 0, len(requiredAny))
			for _, value := range requiredAny {
				if item, ok := value.(string); ok {
					required = append(required, item)
				}
			}
			param.InputSchema.Required = required
		}
		result = append(result, anthropic.ToolUnionParam{OfTool: &param})
	}
	return result
}
