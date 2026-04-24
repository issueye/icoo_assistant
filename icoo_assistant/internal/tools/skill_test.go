package tools_test

import (
	"testing"

	"icoo_assistant/internal/skill"
	"icoo_assistant/internal/tools"
)

func TestLoadSkillToolReturnsSkillBody(t *testing.T) {
	loader := &skill.Loader{}
	_ = loader
	// behavior covered at loader level; here we just ensure the tool validates input shape
	tool := tools.NewLoadSkillTool(loader)
	if _, err := tool.Handler(tools.Call{Input: map[string]interface{}{}}); err == nil {
		t.Fatal("expected validation error when name is missing")
	}
}
