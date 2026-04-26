package proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type sseEvent struct {
	Name string
	Data string
}

type anthropicStreamBlock struct {
	Index        int
	Key          string
	Type         string
	ItemID       string
	OutputIndex  int
	ContentIndex int
	ToolCallID   string
	ToolName     string
	HasDelta     bool
	Stopped      bool
}

type anthropicStreamState struct {
	w              http.ResponseWriter
	flusher        http.Flusher
	service        *Service
	requestID      string
	model          string
	messageID      string
	messageStarted bool
	messageStopped bool
	nextBlockIndex int
	sawToolUse     bool
	inputTokens    int
	outputTokens   int
	textBlocks     map[string]*anthropicStreamBlock
	toolBlocks     map[string]*anthropicStreamBlock
}

func (s *Service) translateResponsesStreamToAnthropic(w http.ResponseWriter, body io.Reader, model, requestID string) error {
	state := &anthropicStreamState{
		w:          w,
		service:    s,
		requestID:  requestID,
		model:      model,
		textBlocks: make(map[string]*anthropicStreamBlock),
		toolBlocks: make(map[string]*anthropicStreamBlock),
	}
	if flusher, ok := w.(http.Flusher); ok {
		state.flusher = flusher
	}

	reader := bufio.NewReader(body)
	for {
		event, err := readSSEEvent(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			_ = state.emitErrorEvent(err.Error())
			return err
		}
		if strings.TrimSpace(event.Data) == "" {
			continue
		}
		state.logUpstreamEvent(event)
		if strings.TrimSpace(event.Data) == "[DONE]" {
			if err := state.finish(nil); err != nil {
				return err
			}
			return nil
		}

		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(event.Data), &payload); err != nil {
			_ = state.emitErrorEvent("failed to decode upstream stream event")
			return fmt.Errorf("decode upstream stream event: %w", err)
		}
		eventType := strings.TrimSpace(event.Name)
		if eventType == "" {
			eventType = stringValue(payload["type"], "")
		}
		if eventType == "" {
			continue
		}

		if err := state.handleResponsesEvent(eventType, payload); err != nil {
			_ = state.emitErrorEvent(err.Error())
			return err
		}
		if state.messageStopped {
			return nil
		}
	}

	if err := state.finish(nil); err != nil {
		return err
	}
	return nil
}

func readSSEEvent(reader *bufio.Reader) (sseEvent, error) {
	var (
		name      string
		dataLines []string
	)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return sseEvent{}, err
		}
		if len(line) == 0 && err == io.EOF {
			if len(dataLines) == 0 && name == "" {
				return sseEvent{}, io.EOF
			}
			return sseEvent{Name: name, Data: strings.Join(dataLines, "\n")}, nil
		}

		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "" {
			if len(dataLines) == 0 && name == "" {
				if err == io.EOF {
					return sseEvent{}, io.EOF
				}
				continue
			}
			return sseEvent{Name: name, Data: strings.Join(dataLines, "\n")}, nil
		}
		if strings.HasPrefix(trimmed, ":") {
			if err == io.EOF {
				return sseEvent{Name: name, Data: strings.Join(dataLines, "\n")}, nil
			}
			continue
		}
		if strings.HasPrefix(trimmed, "event:") {
			name = strings.TrimSpace(trimmed[len("event:"):])
		}
		if strings.HasPrefix(trimmed, "data:") {
			value := trimmed[len("data:"):]
			if strings.HasPrefix(value, " ") {
				value = value[1:]
			}
			dataLines = append(dataLines, value)
		}
		if err == io.EOF {
			return sseEvent{Name: name, Data: strings.Join(dataLines, "\n")}, nil
		}
	}
}

func (s *anthropicStreamState) handleResponsesEvent(eventType string, payload map[string]interface{}) error {
	switch eventType {
	case "response.created":
		response := objectValue(payload["response"])
		s.messageID = firstNonEmpty(
			stringValue(response["id"], ""),
			stringValue(payload["response_id"], ""),
			s.messageID,
			"msg_proxy_stream",
		)
		usage := mapUsageToAnthropic(response["usage"])
		s.inputTokens = intValue(usage["input_tokens"])
		if s.inputTokens == 0 {
			s.inputTokens = intValue(objectValue(response["usage"])["input_tokens"])
		}
		return s.ensureMessageStart()
	case "response.output_text.delta":
		if err := s.ensureMessageStart(); err != nil {
			return err
		}
		block, err := s.ensureTextBlock(payload)
		if err != nil {
			return err
		}
		text := stringValue(payload["delta"], "")
		if text == "" {
			return nil
		}
		block.HasDelta = true
		return s.emitEvent("content_block_delta", map[string]interface{}{
			"type":  "content_block_delta",
			"index": block.Index,
			"delta": map[string]interface{}{
				"type": "text_delta",
				"text": text,
			},
		})
	case "response.output_text.done":
		if err := s.ensureMessageStart(); err != nil {
			return err
		}
		block, err := s.ensureTextBlock(payload)
		if err != nil {
			return err
		}
		text := stringValue(payload["text"], "")
		if text != "" && !block.HasDelta {
			block.HasDelta = true
			if err := s.emitEvent("content_block_delta", map[string]interface{}{
				"type":  "content_block_delta",
				"index": block.Index,
				"delta": map[string]interface{}{
					"type": "text_delta",
					"text": text,
				},
			}); err != nil {
				return err
			}
		}
		return s.stopBlock(block)
	case "response.output_item.added":
		item := objectValue(payload["item"])
		if stringValue(item["type"], "") != "function_call" {
			return nil
		}
		if err := s.ensureMessageStart(); err != nil {
			return err
		}
		_, err := s.ensureToolBlock(payload, item)
		return err
	case "response.function_call_arguments.delta":
		if err := s.ensureMessageStart(); err != nil {
			return err
		}
		block, err := s.ensureToolBlock(payload, nil)
		if err != nil {
			return err
		}
		delta := stringValue(payload["delta"], "")
		if delta == "" {
			return nil
		}
		block.HasDelta = true
		s.sawToolUse = true
		return s.emitEvent("content_block_delta", map[string]interface{}{
			"type":  "content_block_delta",
			"index": block.Index,
			"delta": map[string]interface{}{
				"type":         "input_json_delta",
				"partial_json": delta,
			},
		})
	case "response.function_call_arguments.done":
		if err := s.ensureMessageStart(); err != nil {
			return err
		}
		block, err := s.ensureToolBlock(payload, map[string]interface{}{
			"type":    "function_call",
			"id":      firstNonEmpty(stringValue(payload["item_id"], ""), stringValue(payload["id"], "")),
			"call_id": stringValue(payload["call_id"], ""),
			"name":    stringValue(payload["name"], ""),
		})
		if err != nil {
			return err
		}
		arguments := stringValue(payload["arguments"], "")
		if arguments != "" && !block.HasDelta {
			block.HasDelta = true
			if err := s.emitEvent("content_block_delta", map[string]interface{}{
				"type":  "content_block_delta",
				"index": block.Index,
				"delta": map[string]interface{}{
					"type":         "input_json_delta",
					"partial_json": arguments,
				},
			}); err != nil {
				return err
			}
		}
		return s.stopBlock(block)
	case "response.output_item.done":
		item := objectValue(payload["item"])
		if stringValue(item["type"], "") != "function_call" {
			return nil
		}
		if err := s.ensureMessageStart(); err != nil {
			return err
		}
		block, err := s.ensureToolBlock(payload, item)
		if err != nil {
			return err
		}
		arguments := stringValue(item["arguments"], "")
		if arguments != "" && !block.HasDelta {
			block.HasDelta = true
			if err := s.emitEvent("content_block_delta", map[string]interface{}{
				"type":  "content_block_delta",
				"index": block.Index,
				"delta": map[string]interface{}{
					"type":         "input_json_delta",
					"partial_json": arguments,
				},
			}); err != nil {
				return err
			}
		}
		return s.stopBlock(block)
	case "response.completed":
		response := objectValue(payload["response"])
		return s.finish(response)
	case "response.failed":
		response := objectValue(payload["response"])
		message := "upstream response failed"
		if errObj := objectValue(response["error"]); errObj != nil {
			message = stringValue(errObj["message"], message)
		}
		if err := s.emitErrorEvent(message); err != nil {
			return err
		}
		s.messageStopped = true
		return nil
	case "error":
		errObj := objectValue(payload["error"])
		message := stringValue(errObj["message"], "upstream stream returned error")
		if err := s.emitEvent("error", map[string]interface{}{
			"type": "error",
			"error": map[string]interface{}{
				"type":    firstNonEmpty(stringValue(errObj["type"], ""), "api_error"),
				"message": message,
			},
		}); err != nil {
			return err
		}
		s.messageStopped = true
		return nil
	default:
		return nil
	}
}

func (s *anthropicStreamState) ensureMessageStart() error {
	if s.messageStarted {
		return nil
	}
	if s.messageID == "" {
		s.messageID = "msg_proxy_stream"
	}
	s.messageStarted = true
	return s.emitEvent("message_start", map[string]interface{}{
		"type": "message_start",
		"message": map[string]interface{}{
			"id":            s.messageID,
			"type":          "message",
			"role":          "assistant",
			"content":       []interface{}{},
			"model":         s.model,
			"stop_reason":   nil,
			"stop_sequence": nil,
			"usage": map[string]interface{}{
				"input_tokens":  s.inputTokens,
				"output_tokens": s.outputTokens,
			},
		},
	})
}

func (s *anthropicStreamState) ensureTextBlock(payload map[string]interface{}) (*anthropicStreamBlock, error) {
	itemID := firstNonEmpty(stringValue(payload["item_id"], ""), stringValue(payload["id"], ""))
	outputIndex := intValue(payload["output_index"])
	contentIndex := intValue(payload["content_index"])
	key := fmt.Sprintf("%s:%d:%d", itemID, outputIndex, contentIndex)
	if block, ok := s.textBlocks[key]; ok {
		return block, nil
	}
	block := &anthropicStreamBlock{
		Index:        s.nextBlockIndex,
		Key:          key,
		Type:         "text",
		ItemID:       itemID,
		OutputIndex:  outputIndex,
		ContentIndex: contentIndex,
	}
	s.nextBlockIndex++
	s.textBlocks[key] = block
	if err := s.emitEvent("content_block_start", map[string]interface{}{
		"type":  "content_block_start",
		"index": block.Index,
		"content_block": map[string]interface{}{
			"type": "text",
			"text": "",
		},
	}); err != nil {
		return nil, err
	}
	return block, nil
}

func (s *anthropicStreamState) ensureToolBlock(payload map[string]interface{}, item map[string]interface{}) (*anthropicStreamBlock, error) {
	itemID := firstNonEmpty(
		stringValue(payload["item_id"], ""),
		stringValue(objectValue(item)["id"], ""),
		stringValue(payload["id"], ""),
	)
	outputIndex := intValue(payload["output_index"])
	key := firstNonEmpty(itemID, strconv.Itoa(outputIndex))
	if block, ok := s.toolBlocks[key]; ok {
		if block.ToolName == "" {
			block.ToolName = stringValue(objectValue(item)["name"], "")
		}
		if block.ToolCallID == "" {
			block.ToolCallID = firstNonEmpty(
				stringValue(objectValue(item)["call_id"], ""),
				stringValue(payload["call_id"], ""),
				itemID,
			)
		}
		return block, nil
	}
	callID := firstNonEmpty(
		stringValue(objectValue(item)["call_id"], ""),
		stringValue(payload["call_id"], ""),
		itemID,
	)
	name := firstNonEmpty(
		stringValue(objectValue(item)["name"], ""),
		stringValue(payload["name"], ""),
	)
	block := &anthropicStreamBlock{
		Index:       s.nextBlockIndex,
		Key:         key,
		Type:        "tool_use",
		ItemID:      itemID,
		OutputIndex: outputIndex,
		ToolCallID:  callID,
		ToolName:    name,
	}
	s.nextBlockIndex++
	s.toolBlocks[key] = block
	s.sawToolUse = true
	if err := s.emitEvent("content_block_start", map[string]interface{}{
		"type":  "content_block_start",
		"index": block.Index,
		"content_block": map[string]interface{}{
			"type":  "tool_use",
			"id":    firstNonEmpty(block.ToolCallID, block.ItemID, fmt.Sprintf("tool_%d", block.Index)),
			"name":  block.ToolName,
			"input": map[string]interface{}{},
		},
	}); err != nil {
		return nil, err
	}
	return block, nil
}

func (s *anthropicStreamState) stopBlock(block *anthropicStreamBlock) error {
	if block == nil || block.Stopped {
		return nil
	}
	block.Stopped = true
	return s.emitEvent("content_block_stop", map[string]interface{}{
		"type":  "content_block_stop",
		"index": block.Index,
	})
}

func (s *anthropicStreamState) finish(response map[string]interface{}) error {
	if s.messageStopped {
		return nil
	}
	if response != nil {
		s.messageID = firstNonEmpty(stringValue(response["id"], ""), s.messageID)
		usage := mapUsageToAnthropic(response["usage"])
		if value := intValue(usage["input_tokens"]); value > 0 {
			s.inputTokens = value
		}
		if value := intValue(usage["output_tokens"]); value > 0 {
			s.outputTokens = value
		}
	}
	if err := s.ensureMessageStart(); err != nil {
		return err
	}
	for _, block := range s.sortedBlocks() {
		if err := s.stopBlock(block); err != nil {
			return err
		}
	}
	stopReason := "end_turn"
	if response != nil {
		stopReason = mapResponsesStatusToAnthropicStopReason(response)
	}
	if s.sawToolUse {
		stopReason = "tool_use"
	}
	if err := s.emitEvent("message_delta", map[string]interface{}{
		"type": "message_delta",
		"delta": map[string]interface{}{
			"stop_reason":   stopReason,
			"stop_sequence": nil,
		},
		"usage": map[string]interface{}{
			"input_tokens":  s.inputTokens,
			"output_tokens": s.outputTokens,
		},
	}); err != nil {
		return err
	}
	if err := s.emitEvent("message_stop", map[string]interface{}{
		"type": "message_stop",
	}); err != nil {
		return err
	}
	s.messageStopped = true
	return nil
}

func (s *anthropicStreamState) sortedBlocks() []*anthropicStreamBlock {
	items := make([]*anthropicStreamBlock, 0, len(s.textBlocks)+len(s.toolBlocks))
	for _, block := range s.textBlocks {
		items = append(items, block)
	}
	for _, block := range s.toolBlocks {
		items = append(items, block)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Index < items[j].Index
	})
	return items
}

func (s *anthropicStreamState) emitErrorEvent(message string) error {
	return s.emitEvent("error", map[string]interface{}{
		"type": "error",
		"error": map[string]interface{}{
			"type":    "api_error",
			"message": message,
		},
	})
}

func (s *anthropicStreamState) emitEvent(eventName string, payload map[string]interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	s.logDownstreamEvent(eventName, data)
	if _, err := fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", eventName, data); err != nil {
		return err
	}
	if s.flusher != nil {
		s.flusher.Flush()
	}
	return nil
}

func (s *anthropicStreamState) logUpstreamEvent(event sseEvent) {
	if s.service == nil {
		return
	}
	s.service.logChain("conversion.stream.upstream_event",
		"request_id", s.requestID,
		"event", firstNonEmpty(event.Name, "<data-only>"),
		"data", s.service.logBody([]byte(event.Data)),
	)
}

func (s *anthropicStreamState) logDownstreamEvent(eventName string, payload []byte) {
	if s.service == nil {
		return
	}
	s.service.logChain("conversion.stream.downstream_event",
		"request_id", s.requestID,
		"event", eventName,
		"data", s.service.logBody(payload),
	)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
