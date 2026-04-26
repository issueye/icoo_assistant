package proxy

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

type responsesStreamAggregate struct {
	response       map[string]interface{}
	outputItems    map[int]map[string]interface{}
	outputTexts    map[int]string
	outputItemSeen map[int]bool
}

func aggregateResponsesStreamToJSON(body io.Reader) ([]byte, error) {
	aggregate := &responsesStreamAggregate{
		outputItems:    make(map[int]map[string]interface{}),
		outputTexts:    make(map[int]string),
		outputItemSeen: make(map[int]bool),
	}
	reader := bufio.NewReader(body)
	for {
		event, err := readSSEEvent(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		data := strings.TrimSpace(event.Data)
		if data == "" || data == "[DONE]" {
			continue
		}
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(data), &payload); err != nil {
			return nil, fmt.Errorf("decode upstream stream event: %w", err)
		}
		eventType := strings.TrimSpace(event.Name)
		if eventType == "" {
			eventType = stringValue(payload["type"], "")
		}
		switch eventType {
		case "response.created", "response.in_progress", "response.completed", "response.failed":
			response := objectValue(payload["response"])
			if response != nil {
				aggregate.response = cloneMap(response)
			}
		case "response.output_item.done":
			item := objectValue(payload["item"])
			if item == nil {
				continue
			}
			outputIndex := intValue(payload["output_index"])
			aggregate.outputItems[outputIndex] = cloneMap(item)
			aggregate.outputItemSeen[outputIndex] = true
		case "response.output_text.done":
			outputIndex := intValue(payload["output_index"])
			text := stringValue(payload["text"], "")
			if strings.TrimSpace(text) != "" || text != "" {
				aggregate.outputTexts[outputIndex] = text
			}
		case "error":
			errObj := objectValue(payload["error"])
			if errObj != nil {
				return nil, fmt.Errorf(stringValue(errObj["message"], "upstream stream returned error"))
			}
			return nil, fmt.Errorf("upstream stream returned error")
		}
	}

	response := aggregate.response
	if response == nil {
		response = map[string]interface{}{
			"id":     "resp_proxy_stream",
			"object": "response",
			"status": "completed",
		}
	} else {
		response = cloneMap(response)
	}

	if len(aggregate.outputItemSeen) > 0 {
		response["output"] = aggregate.orderedOutputItems()
	} else if len(aggregate.outputTexts) > 0 {
		response["output"] = aggregate.syntheticMessageOutput()
	}

	if text := strings.TrimSpace(extractResponsesOutputText(response["output"])); text != "" {
		response["output_text"] = text
	} else if text := aggregate.combinedOutputText(); text != "" {
		response["output_text"] = text
	}
	return json.Marshal(response)
}

func (a *responsesStreamAggregate) orderedOutputItems() []map[string]interface{} {
	indexes := make([]int, 0, len(a.outputItemSeen))
	for index := range a.outputItemSeen {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)
	items := make([]map[string]interface{}, 0, len(indexes))
	for _, index := range indexes {
		item := cloneMap(a.outputItems[index])
		if item == nil {
			item = map[string]interface{}{}
		}
		content, _ := item["content"].([]interface{})
		if stringValue(item["type"], "") == "message" && len(content) == 0 {
			if text := a.outputTexts[index]; text != "" {
				item["content"] = []interface{}{
					map[string]interface{}{
						"type": "output_text",
						"text": text,
					},
				}
			}
		}
		items = append(items, item)
	}
	return items
}

func (a *responsesStreamAggregate) syntheticMessageOutput() []map[string]interface{} {
	indexes := make([]int, 0, len(a.outputTexts))
	for index := range a.outputTexts {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)
	items := make([]map[string]interface{}, 0, len(indexes))
	for _, index := range indexes {
		items = append(items, map[string]interface{}{
			"id":   fmt.Sprintf("msg_proxy_%d", index),
			"type": "message",
			"role": "assistant",
			"content": []interface{}{
				map[string]interface{}{
					"type": "output_text",
					"text": a.outputTexts[index],
				},
			},
		})
	}
	return items
}

func (a *responsesStreamAggregate) combinedOutputText() string {
	if len(a.outputTexts) == 0 {
		return ""
	}
	indexes := make([]int, 0, len(a.outputTexts))
	for index := range a.outputTexts {
		indexes = append(indexes, index)
	}
	sort.Ints(indexes)
	parts := make([]string, 0, len(indexes))
	for _, index := range indexes {
		if text := strings.TrimSpace(a.outputTexts[index]); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n")
}

func cloneMap(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return nil
	}
	output := make(map[string]interface{}, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
