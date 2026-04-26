package proxy

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const defaultResponsesReasoningEffort = "medium"

func translateAnthropicToResponsesRequest(body []byte, model string) ([]byte, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid json body")
	}

	messages, _ := payload["messages"].([]interface{})
	request := map[string]interface{}{
		"model": model,
		"input": normalizeAnthropicMessages(messages),
	}
	if stream, _ := payload["stream"].(bool); stream {
		request["stream"] = true
	}
	if system := anthropicSystemToInstructions(payload["system"]); system != "" {
		request["instructions"] = system
	}
	copyIfExists(payload, request, "temperature", "top_p")
	if value, ok := payload["max_tokens"]; ok {
		request["max_output_tokens"] = value
	}
	if value, ok := payload["tools"]; ok {
		request["tools"] = anthropicToolsToResponsesTools(value)
	}
	applyDefaultResponsesReasoning(request)
	return json.Marshal(request)
}

func translateResponsesToAnthropicRequest(body []byte, model string) ([]byte, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid json body")
	}
	if stream, _ := payload["stream"].(bool); stream {
		return nil, fmt.Errorf("streaming cross protocol translation is not implemented yet")
	}

	request := map[string]interface{}{
		"model":      model,
		"messages":   normalizeResponsesInput(payload["input"]),
		"max_tokens": intValue(payload["max_output_tokens"]),
	}
	if instructions, ok := payload["instructions"].(string); ok && strings.TrimSpace(instructions) != "" {
		request["system"] = instructions
	}
	copyIfExists(payload, request, "temperature", "top_p")
	if value, ok := payload["tools"]; ok {
		request["tools"] = responsesToolsToAnthropicTools(value)
	}
	return json.Marshal(request)
}

func translateAnthropicToChatRequest(body []byte, model string) ([]byte, error) {
	responsesBody, err := translateAnthropicToResponsesRequest(body, model)
	if err != nil {
		return nil, err
	}
	return translateResponsesToChatRequest(responsesBody, model)
}

func translateChatToAnthropicRequest(body []byte, model string) ([]byte, error) {
	responsesBody, err := translateChatToResponsesRequest(body, model)
	if err != nil {
		return nil, err
	}
	return translateResponsesToAnthropicRequest(responsesBody, model)
}

func translateChatToResponsesRequest(body []byte, model string) ([]byte, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid json body")
	}
	if stream, _ := payload["stream"].(bool); stream {
		return nil, fmt.Errorf("streaming cross protocol translation is not implemented yet")
	}

	messages, _ := payload["messages"].([]interface{})
	instructions := collectSystemMessages(messages)
	input := make([]map[string]interface{}, 0, len(messages))
	for _, raw := range messages {
		msg, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := msg["role"].(string)
		if strings.EqualFold(role, "system") {
			continue
		}
		input = append(input, chatMessageToResponsesInput(msg)...)
	}

	request := map[string]interface{}{
		"model": model,
		"input": input,
	}
	if instructions != "" {
		request["instructions"] = instructions
	}
	copyIfExists(payload, request, "temperature", "top_p", "tool_choice")
	if value, ok := payload["max_tokens"]; ok {
		request["max_output_tokens"] = value
	}
	if value, ok := payload["tools"]; ok {
		request["tools"] = chatToolsToResponsesTools(value)
	}
	applyDefaultResponsesReasoning(request)
	return json.Marshal(request)
}

func translateResponsesToChatRequest(body []byte, model string) ([]byte, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid json body")
	}
	if stream, _ := payload["stream"].(bool); stream {
		return nil, fmt.Errorf("streaming cross protocol translation is not implemented yet")
	}

	messages := make([]map[string]interface{}, 0)
	if instructions, ok := payload["instructions"].(string); ok && strings.TrimSpace(instructions) != "" {
		messages = append(messages, map[string]interface{}{
			"role":    "system",
			"content": instructions,
		})
	}
	for _, item := range normalizeResponsesInputToChatMessages(payload["input"]) {
		messages = append(messages, item)
	}

	request := map[string]interface{}{
		"model":    model,
		"messages": messages,
	}
	copyIfExists(payload, request, "temperature", "top_p", "tool_choice")
	if value, ok := payload["max_output_tokens"]; ok {
		request["max_tokens"] = value
	}
	if value, ok := payload["tools"]; ok {
		request["tools"] = responsesToolsToChatTools(value)
	}
	return json.Marshal(request)
}

func translateResponsesToChatResponse(body []byte, model string) ([]byte, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid upstream json body")
	}

	content := extractResponsesText(payload)
	toolCalls := extractResponsesFunctionCalls(payload["output"])
	message := map[string]interface{}{
		"role":    "assistant",
		"content": content,
	}
	finishReason := mapResponsesStatusToFinishReason(payload)
	if len(toolCalls) > 0 {
		message["tool_calls"] = toolCalls
		finishReason = "tool_calls"
		if content == "" {
			message["content"] = nil
		}
	}
	response := map[string]interface{}{
		"id":      stringValue(payload["id"], "chatcmpl-proxy"),
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []map[string]interface{}{
			{
				"index":         0,
				"message":       message,
				"finish_reason": finishReason,
			},
		},
	}
	if usage, ok := payload["usage"]; ok {
		response["usage"] = mapUsageToChat(usage)
	}
	return json.Marshal(response)
}

func translateChatToResponsesResponse(body []byte, model string) ([]byte, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid upstream json body")
	}

	text := extractChatAssistantText(payload["choices"])
	output := make([]map[string]interface{}, 0)
	if text != "" {
		output = append(output, map[string]interface{}{
			"id":   "msg_proxy_1",
			"type": "message",
			"role": "assistant",
			"content": []map[string]interface{}{
				{"type": "output_text", "text": text},
			},
		})
	}
	output = append(output, chatChoicesToResponsesFunctionCalls(payload["choices"])...)
	status := "completed"
	if len(output) == 0 {
		output = append(output, map[string]interface{}{
			"id":      "msg_proxy_1",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]interface{}{{"type": "output_text", "text": ""}},
		})
	}
	response := map[string]interface{}{
		"id":     stringValue(payload["id"], "resp-proxy"),
		"object": "response",
		"model":  model,
		"status": status,
		"output": output,
	}
	if usage, ok := payload["usage"]; ok {
		response["usage"] = mapUsageToResponses(usage)
	}
	return json.Marshal(response)
}

func translateAnthropicToChatResponse(body []byte, model string) ([]byte, error) {
	responsesBody, err := translateAnthropicToResponsesResponse(body, model)
	if err != nil {
		return nil, err
	}
	return translateResponsesToChatResponse(responsesBody, model)
}

func translateChatToAnthropicResponse(body []byte, model string) ([]byte, error) {
	responsesBody, err := translateChatToResponsesResponse(body, model)
	if err != nil {
		return nil, err
	}
	return translateResponsesToAnthropicResponse(responsesBody, model)
}

func translateResponsesToAnthropicResponse(body []byte, model string) ([]byte, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid upstream json body")
	}

	text := extractResponsesText(payload)
	content := make([]map[string]interface{}, 0)
	if text != "" {
		content = append(content, map[string]interface{}{
			"type": "text",
			"text": text,
		})
	}
	content = append(content, responsesOutputToAnthropicToolUse(payload["output"])...)
	stopReason := mapResponsesStatusToAnthropicStopReason(payload)
	if len(content) == 0 {
		content = append(content, map[string]interface{}{"type": "text", "text": ""})
	}
	if hasAnthropicToolUse(content) {
		stopReason = "tool_use"
	}
	response := map[string]interface{}{
		"id":            stringValue(payload["id"], "msg_proxy"),
		"type":          "message",
		"role":          "assistant",
		"model":         model,
		"stop_reason":   stopReason,
		"stop_sequence": nil,
		"content":       content,
	}
	if usage, ok := payload["usage"]; ok {
		response["usage"] = mapUsageToAnthropic(usage)
	}
	return json.Marshal(response)
}

func translateAnthropicToResponsesResponse(body []byte, model string) ([]byte, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid upstream json body")
	}

	text := extractAnthropicText(payload["content"])
	output := make([]map[string]interface{}, 0)
	if text != "" {
		output = append(output, map[string]interface{}{
			"id":      "msg_proxy_1",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]interface{}{{"type": "output_text", "text": text}},
		})
	}
	output = append(output, anthropicContentToResponsesFunctionCalls(payload["content"])...)
	if len(output) == 0 {
		output = append(output, map[string]interface{}{
			"id":      "msg_proxy_1",
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]interface{}{{"type": "output_text", "text": ""}},
		})
	}
	response := map[string]interface{}{
		"id":     stringValue(payload["id"], "resp-proxy"),
		"object": "response",
		"model":  model,
		"status": "completed",
		"output": output,
	}
	if usage, ok := payload["usage"]; ok {
		response["usage"] = mapAnthropicUsageToResponses(usage)
	}
	return json.Marshal(response)
}

func collectSystemMessages(messages []interface{}) string {
	parts := make([]string, 0)
	for _, raw := range messages {
		msg, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := msg["role"].(string)
		if !strings.EqualFold(role, "system") {
			continue
		}
		parts = append(parts, contentToText(msg["content"]))
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func normalizeResponsesInput(raw interface{}) []map[string]interface{} {
	switch value := raw.(type) {
	case string:
		text := strings.TrimSpace(value)
		if text == "" {
			return nil
		}
		return []map[string]interface{}{{
			"role":    "user",
			"content": text,
		}}
	case []interface{}:
		items := make([]map[string]interface{}, 0, len(value))
		for _, rawItem := range value {
			item, ok := rawItem.(map[string]interface{})
			if !ok {
				continue
			}
			itemType, _ := item["type"].(string)
			switch itemType {
			case "function_call_output":
				items = append(items, map[string]interface{}{
					"role": "user",
					"content": []map[string]interface{}{
						{
							"type":        "tool_result",
							"tool_use_id": stringValue(item["call_id"], ""),
							"content":     responseToolOutputToText(item["output"]),
						},
					},
				})
				continue
			case "function_call":
				items = append(items, map[string]interface{}{
					"role": "assistant",
					"content": []map[string]interface{}{
						{
							"type":  "tool_use",
							"id":    stringValue(item["call_id"], ""),
							"name":  stringValue(item["name"], ""),
							"input": parseJSONStringObject(stringValue(item["arguments"], "{}")),
						},
					},
				})
				continue
			}
			role, _ := item["role"].(string)
			if role == "" {
				role = "user"
			}
			items = append(items, map[string]interface{}{
				"role":    role,
				"content": normalizeMessageContent(item["content"]),
			})
		}
		return items
	default:
		return nil
	}
}

func normalizeResponsesInputToChatMessages(raw interface{}) []map[string]interface{} {
	switch value := raw.(type) {
	case string:
		text := strings.TrimSpace(value)
		if text == "" {
			return nil
		}
		return []map[string]interface{}{{"role": "user", "content": text}}
	case []interface{}:
		items := make([]map[string]interface{}, 0, len(value))
		for _, rawItem := range value {
			item, ok := rawItem.(map[string]interface{})
			if !ok {
				continue
			}
			itemType, _ := item["type"].(string)
			switch itemType {
			case "function_call_output":
				items = append(items, map[string]interface{}{
					"role":         "tool",
					"tool_call_id": stringValue(item["call_id"], ""),
					"content":      responseToolOutputToText(item["output"]),
				})
			case "function_call":
				items = append(items, map[string]interface{}{
					"role":    "assistant",
					"content": nil,
					"tool_calls": []map[string]interface{}{
						{
							"id":   stringValue(item["call_id"], ""),
							"type": "function",
							"function": map[string]interface{}{
								"name":      stringValue(item["name"], ""),
								"arguments": stringValue(item["arguments"], "{}"),
							},
						},
					},
				})
			default:
				role, _ := item["role"].(string)
				if role == "" {
					role = "user"
				}
				items = append(items, map[string]interface{}{
					"role":    role,
					"content": normalizeMessageContent(item["content"]),
				})
			}
		}
		return items
	default:
		return nil
	}
}

func chatMessageToResponsesInput(msg map[string]interface{}) []map[string]interface{} {
	role, _ := msg["role"].(string)
	if role == "" {
		role = "user"
	}
	if role == "tool" {
		return []map[string]interface{}{{
			"type":    "function_call_output",
			"call_id": stringValue(msg["tool_call_id"], ""),
			"output":  contentToText(msg["content"]),
		}}
	}

	items := make([]map[string]interface{}, 0)
	if content := contentToText(msg["content"]); content != "" {
		items = append(items, map[string]interface{}{
			"role":    role,
			"content": content,
		})
	}
	if role == "assistant" {
		toolCalls, _ := msg["tool_calls"].([]interface{})
		for _, rawToolCall := range toolCalls {
			toolCall, ok := rawToolCall.(map[string]interface{})
			if !ok {
				continue
			}
			function, _ := toolCall["function"].(map[string]interface{})
			items = append(items, map[string]interface{}{
				"type":      "function_call",
				"call_id":   stringValue(toolCall["id"], ""),
				"name":      stringValue(function["name"], ""),
				"arguments": stringValue(function["arguments"], "{}"),
			})
		}
	}
	if len(items) == 0 {
		items = append(items, map[string]interface{}{
			"role":    role,
			"content": normalizeMessageContent(msg["content"]),
		})
	}
	return items
}

func normalizeAnthropicMessages(raw []interface{}) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(raw))
	for _, rawItem := range raw {
		item, ok := rawItem.(map[string]interface{})
		if !ok {
			continue
		}
		role, _ := item["role"].(string)
		if role == "" {
			role = "user"
		}
		content, _ := item["content"].([]interface{})
		if role == "assistant" {
			items = append(items, anthropicAssistantContentToResponsesInput(content)...)
			continue
		}
		items = append(items, anthropicUserContentToResponsesInput(content, role)...)
	}
	return items
}

func anthropicSystemToInstructions(raw interface{}) string {
	switch value := raw.(type) {
	case string:
		return strings.TrimSpace(value)
	case []interface{}:
		parts := make([]string, 0, len(value))
		for _, rawPart := range value {
			part, ok := rawPart.(map[string]interface{})
			if !ok {
				continue
			}
			if text, ok := part["text"].(string); ok && strings.TrimSpace(text) != "" {
				parts = append(parts, text)
			}
		}
		return strings.TrimSpace(strings.Join(parts, "\n"))
	default:
		return ""
	}
}

func normalizeMessageContent(raw interface{}) interface{} {
	switch value := raw.(type) {
	case string:
		return value
	case []interface{}:
		textParts := make([]string, 0, len(value))
		for _, part := range value {
			partMap, ok := part.(map[string]interface{})
			if !ok {
				continue
			}
			partType, _ := partMap["type"].(string)
			switch partType {
			case "text", "input_text", "output_text":
				if text, ok := partMap["text"].(string); ok && text != "" {
					textParts = append(textParts, text)
				}
			}
		}
		if len(textParts) == 1 {
			return textParts[0]
		}
		if len(textParts) > 1 {
			return strings.Join(textParts, "\n")
		}
	}
	return raw
}

func anthropicAssistantContentToResponsesInput(content []interface{}) []map[string]interface{} {
	items := make([]map[string]interface{}, 0)
	textParts := make([]string, 0)
	for _, rawPart := range content {
		part, ok := rawPart.(map[string]interface{})
		if !ok {
			continue
		}
		partType, _ := part["type"].(string)
		switch partType {
		case "text":
			if text, ok := part["text"].(string); ok && text != "" {
				textParts = append(textParts, text)
			}
		case "tool_use":
			items = append(items, map[string]interface{}{
				"type":      "function_call",
				"call_id":   stringValue(part["id"], ""),
				"name":      stringValue(part["name"], ""),
				"arguments": marshalToJSONString(part["input"]),
			})
		}
	}
	if len(textParts) > 0 {
		items = append([]map[string]interface{}{{
			"role":    "assistant",
			"content": strings.Join(textParts, "\n"),
		}}, items...)
	}
	return items
}

func anthropicUserContentToResponsesInput(content []interface{}, role string) []map[string]interface{} {
	items := make([]map[string]interface{}, 0)
	textParts := make([]string, 0)
	for _, rawPart := range content {
		part, ok := rawPart.(map[string]interface{})
		if !ok {
			continue
		}
		partType, _ := part["type"].(string)
		switch partType {
		case "text":
			if text, ok := part["text"].(string); ok && text != "" {
				textParts = append(textParts, text)
			}
		case "tool_result":
			items = append(items, map[string]interface{}{
				"type":    "function_call_output",
				"call_id": stringValue(part["tool_use_id"], ""),
				"output":  anthropicToolResultContentToOutput(part["content"]),
			})
		}
	}
	if len(textParts) > 0 {
		items = append([]map[string]interface{}{{
			"role":    role,
			"content": strings.Join(textParts, "\n"),
		}}, items...)
	}
	return items
}

func contentToText(raw interface{}) string {
	switch value := normalizeMessageContent(raw).(type) {
	case string:
		return value
	default:
		return ""
	}
}

func extractResponsesOutputText(raw interface{}) string {
	items, _ := raw.([]interface{})
	parts := make([]string, 0)
	for _, item := range items {
		msg, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		content, _ := msg["content"].([]interface{})
		for _, rawPart := range content {
			part, ok := rawPart.(map[string]interface{})
			if !ok {
				continue
			}
			partType, _ := part["type"].(string)
			if partType == "output_text" || partType == "text" {
				if text, ok := part["text"].(string); ok && text != "" {
					parts = append(parts, text)
				}
			}
		}
	}
	return strings.Join(parts, "\n")
}

func applyDefaultResponsesReasoning(payload map[string]interface{}) {
	if payload == nil {
		return
	}
	raw, ok := payload["reasoning"]
	if !ok || raw == nil {
		payload["reasoning"] = map[string]interface{}{"effort": defaultResponsesReasoningEffort}
		return
	}
	reasoning, ok := raw.(map[string]interface{})
	if !ok {
		return
	}
	if strings.TrimSpace(stringValue(reasoning["effort"], "")) == "" {
		reasoning["effort"] = defaultResponsesReasoningEffort
	}
	payload["reasoning"] = reasoning
}

func extractResponsesText(payload map[string]interface{}) string {
	if text := strings.TrimSpace(extractResponsesOutputText(payload["output"])); text != "" {
		return text
	}
	return strings.TrimSpace(stringValue(payload["output_text"], ""))
}

func extractResponsesFunctionCalls(raw interface{}) []map[string]interface{} {
	items, _ := raw.([]interface{})
	result := make([]map[string]interface{}, 0)
	for _, rawItem := range items {
		item, ok := rawItem.(map[string]interface{})
		if !ok {
			continue
		}
		if stringValue(item["type"], "") != "function_call" {
			continue
		}
		result = append(result, map[string]interface{}{
			"id":   stringValue(item["call_id"], ""),
			"type": "function",
			"function": map[string]interface{}{
				"name":      stringValue(item["name"], ""),
				"arguments": stringValue(item["arguments"], "{}"),
			},
		})
	}
	return result
}

func responsesOutputToAnthropicToolUse(raw interface{}) []map[string]interface{} {
	items, _ := raw.([]interface{})
	result := make([]map[string]interface{}, 0)
	for _, rawItem := range items {
		item, ok := rawItem.(map[string]interface{})
		if !ok {
			continue
		}
		if stringValue(item["type"], "") != "function_call" {
			continue
		}
		result = append(result, map[string]interface{}{
			"type":  "tool_use",
			"id":    stringValue(item["call_id"], ""),
			"name":  stringValue(item["name"], ""),
			"input": parseJSONStringObject(stringValue(item["arguments"], "{}")),
		})
	}
	return result
}

func anthropicContentToResponsesFunctionCalls(raw interface{}) []map[string]interface{} {
	parts, _ := raw.([]interface{})
	result := make([]map[string]interface{}, 0)
	for _, rawPart := range parts {
		part, ok := rawPart.(map[string]interface{})
		if !ok {
			continue
		}
		if stringValue(part["type"], "") != "tool_use" {
			continue
		}
		result = append(result, map[string]interface{}{
			"type":      "function_call",
			"call_id":   stringValue(part["id"], ""),
			"name":      stringValue(part["name"], ""),
			"arguments": marshalToJSONString(part["input"]),
		})
	}
	return result
}

func extractAnthropicText(raw interface{}) string {
	parts, _ := raw.([]interface{})
	texts := make([]string, 0, len(parts))
	for _, rawPart := range parts {
		part, ok := rawPart.(map[string]interface{})
		if !ok {
			continue
		}
		partType, _ := part["type"].(string)
		if partType != "text" {
			continue
		}
		if text, ok := part["text"].(string); ok && text != "" {
			texts = append(texts, text)
		}
	}
	return strings.Join(texts, "\n")
}

func extractChatAssistantText(raw interface{}) string {
	choices, _ := raw.([]interface{})
	for _, choice := range choices {
		item, ok := choice.(map[string]interface{})
		if !ok {
			continue
		}
		message, _ := item["message"].(map[string]interface{})
		if message == nil {
			continue
		}
		return contentToText(message["content"])
	}
	return ""
}

func chatChoicesToResponsesFunctionCalls(raw interface{}) []map[string]interface{} {
	choices, _ := raw.([]interface{})
	result := make([]map[string]interface{}, 0)
	for _, choice := range choices {
		item, ok := choice.(map[string]interface{})
		if !ok {
			continue
		}
		message, _ := item["message"].(map[string]interface{})
		if message == nil {
			continue
		}
		toolCalls, _ := message["tool_calls"].([]interface{})
		for _, rawToolCall := range toolCalls {
			toolCall, ok := rawToolCall.(map[string]interface{})
			if !ok {
				continue
			}
			function, _ := toolCall["function"].(map[string]interface{})
			result = append(result, map[string]interface{}{
				"type":      "function_call",
				"call_id":   stringValue(toolCall["id"], ""),
				"name":      stringValue(function["name"], ""),
				"arguments": stringValue(function["arguments"], "{}"),
			})
		}
	}
	return result
}

func mapResponsesStatusToFinishReason(payload map[string]interface{}) string {
	if status, _ := payload["status"].(string); status == "completed" {
		return "stop"
	}
	return "length"
}

func mapResponsesStatusToAnthropicStopReason(payload map[string]interface{}) string {
	if status, _ := payload["status"].(string); status == "completed" {
		return "end_turn"
	}
	return "max_tokens"
}

func mapUsageToChat(raw interface{}) map[string]interface{} {
	usageMap, _ := raw.(map[string]interface{})
	return map[string]interface{}{
		"prompt_tokens":     intValue(usageMap["input_tokens"]),
		"completion_tokens": intValue(usageMap["output_tokens"]),
		"total_tokens":      intValue(usageMap["total_tokens"]),
	}
}

func mapUsageToResponses(raw interface{}) map[string]interface{} {
	usageMap, _ := raw.(map[string]interface{})
	prompt := intValue(usageMap["prompt_tokens"])
	completion := intValue(usageMap["completion_tokens"])
	total := intValue(usageMap["total_tokens"])
	if total == 0 {
		total = prompt + completion
	}
	return map[string]interface{}{
		"input_tokens":  prompt,
		"output_tokens": completion,
		"total_tokens":  total,
	}
}

func mapUsageToAnthropic(raw interface{}) map[string]interface{} {
	usageMap, _ := raw.(map[string]interface{})
	return map[string]interface{}{
		"input_tokens":  intValue(usageMap["input_tokens"]),
		"output_tokens": intValue(usageMap["output_tokens"]),
	}
}

func mapAnthropicUsageToResponses(raw interface{}) map[string]interface{} {
	usageMap, _ := raw.(map[string]interface{})
	input := intValue(usageMap["input_tokens"])
	output := intValue(usageMap["output_tokens"])
	return map[string]interface{}{
		"input_tokens":  input,
		"output_tokens": output,
		"total_tokens":  input + output,
	}
}

func intValue(raw interface{}) int {
	switch value := raw.(type) {
	case float64:
		return int(value)
	case int:
		return value
	default:
		return 0
	}
}

func stringValue(raw interface{}, fallback string) string {
	if value, ok := raw.(string); ok && value != "" {
		return value
	}
	return fallback
}

func copyIfExists(from, to map[string]interface{}, keys ...string) {
	for _, key := range keys {
		if value, ok := from[key]; ok {
			to[key] = value
		}
	}
}

func chatToolsToResponsesTools(raw interface{}) []map[string]interface{} {
	tools, _ := raw.([]interface{})
	items := make([]map[string]interface{}, 0, len(tools))
	for _, rawTool := range tools {
		tool, ok := rawTool.(map[string]interface{})
		if !ok {
			continue
		}
		if fn, ok := tool["function"].(map[string]interface{}); ok {
			items = append(items, map[string]interface{}{
				"type":        "function",
				"name":        stringValue(fn["name"], ""),
				"description": stringValue(fn["description"], ""),
				"parameters":  objectValue(fn["parameters"]),
			})
			continue
		}
		if stringValue(tool["type"], "") == "function" {
			items = append(items, map[string]interface{}{
				"type":        "function",
				"name":        stringValue(tool["name"], ""),
				"description": stringValue(tool["description"], ""),
				"parameters":  objectValue(tool["parameters"]),
			})
		}
	}
	return items
}

func responsesToolsToChatTools(raw interface{}) []map[string]interface{} {
	tools, _ := raw.([]interface{})
	items := make([]map[string]interface{}, 0, len(tools))
	for _, rawTool := range tools {
		tool, ok := rawTool.(map[string]interface{})
		if !ok {
			continue
		}
		if stringValue(tool["type"], "") != "function" {
			continue
		}
		items = append(items, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        stringValue(tool["name"], ""),
				"description": stringValue(tool["description"], ""),
				"parameters":  objectValue(firstNonNil(tool["parameters"], tool["input_schema"])),
			},
		})
	}
	return items
}

func anthropicToolsToResponsesTools(raw interface{}) []map[string]interface{} {
	tools, _ := raw.([]interface{})
	items := make([]map[string]interface{}, 0, len(tools))
	for _, rawTool := range tools {
		tool, ok := rawTool.(map[string]interface{})
		if !ok {
			continue
		}
		items = append(items, map[string]interface{}{
			"type":        "function",
			"name":        stringValue(tool["name"], ""),
			"description": stringValue(tool["description"], ""),
			"parameters":  objectValue(tool["input_schema"]),
		})
	}
	return items
}

func responsesToolsToAnthropicTools(raw interface{}) []map[string]interface{} {
	tools, _ := raw.([]interface{})
	items := make([]map[string]interface{}, 0, len(tools))
	for _, rawTool := range tools {
		tool, ok := rawTool.(map[string]interface{})
		if !ok {
			continue
		}
		if stringValue(tool["type"], "") != "function" {
			continue
		}
		items = append(items, map[string]interface{}{
			"name":         stringValue(tool["name"], ""),
			"description":  stringValue(tool["description"], ""),
			"input_schema": objectValue(firstNonNil(tool["parameters"], tool["input_schema"])),
		})
	}
	return items
}

func responseToolOutputToText(raw interface{}) string {
	switch value := raw.(type) {
	case string:
		return value
	case []interface{}:
		parts := make([]string, 0, len(value))
		for _, rawPart := range value {
			part, ok := rawPart.(map[string]interface{})
			if !ok {
				continue
			}
			if text, ok := part["text"].(string); ok && text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "\n")
	default:
		return ""
	}
}

func anthropicToolResultContentToOutput(raw interface{}) string {
	switch value := raw.(type) {
	case string:
		return value
	case []interface{}:
		parts := make([]string, 0, len(value))
		for _, rawPart := range value {
			part, ok := rawPart.(map[string]interface{})
			if !ok {
				continue
			}
			if text, ok := part["text"].(string); ok && text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, "\n")
	default:
		return ""
	}
}

func marshalToJSONString(raw interface{}) string {
	if raw == nil {
		return "{}"
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func parseJSONStringObject(raw string) map[string]interface{} {
	if strings.TrimSpace(raw) == "" {
		return map[string]interface{}{}
	}
	var value map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return map[string]interface{}{}
	}
	return value
}

func objectValue(raw interface{}) map[string]interface{} {
	if value, ok := raw.(map[string]interface{}); ok {
		return value
	}
	return map[string]interface{}{}
}

func firstNonNil(values ...interface{}) interface{} {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func hasAnthropicToolUse(content []map[string]interface{}) bool {
	for _, part := range content {
		if stringValue(part["type"], "") == "tool_use" {
			return true
		}
	}
	return false
}
