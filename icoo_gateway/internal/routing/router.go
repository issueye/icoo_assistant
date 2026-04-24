package routing

import (
	"fmt"
	"strings"

	"icoo_gateway/internal/conversation"
	"icoo_gateway/internal/team"
)

type TeamLookup interface {
	Get(id string) (team.Team, bool)
	HasMember(teamID, agentID string) bool
}

type ConversationReader interface {
	Get(id string) (conversation.Conversation, bool)
}

type ConversationWriter interface {
	AddMessage(conversationID string, input conversation.AddMessageInput) (conversation.Message, error)
}

type Router struct {
	Teams         TeamLookup
	Conversations ConversationReader
	Writer        ConversationWriter
}

func (r Router) RouteExternalMessage(conversationID string, message conversation.Message) error {
	if strings.TrimSpace(message.Scope) != "external" {
		return nil
	}
	if r.Conversations == nil || r.Writer == nil {
		return fmt.Errorf("router dependencies not configured")
	}
	record, ok := r.Conversations.Get(conversationID)
	if !ok {
		return fmt.Errorf("conversation not found")
	}
	if record.Mode != "team" {
		return nil
	}
	if r.Teams == nil {
		return fmt.Errorf("team lookup not configured")
	}
	targetTeam, ok := r.Teams.Get(record.TargetTeamID)
	if !ok {
		_, err := r.Writer.AddMessage(conversationID, conversation.AddMessageInput{
			Scope:   "system",
			Role:    "system",
			Content: "routing_warning: target team not found; dispatch skipped",
		})
		return err
	}
	entryAgentID := strings.TrimSpace(targetTeam.EntryAgentID)
	if entryAgentID == "" {
		_, err := r.Writer.AddMessage(conversationID, conversation.AddMessageInput{
			Scope:   "system",
			Role:    "system",
			Content: "routing_warning: team has no entry_agent_id; dispatch skipped",
		})
		return err
	}
	if !r.Teams.HasMember(targetTeam.ID, entryAgentID) {
		_, err := r.Writer.AddMessage(conversationID, conversation.AddMessageInput{
			Scope:   "system",
			Role:    "system",
			Content: "routing_warning: entry_agent_id is not an active team member; dispatch skipped",
		})
		return err
	}
	if _, err := r.Writer.AddMessage(conversationID, conversation.AddMessageInput{
		Scope:        "internal",
		Role:         "assistant",
		SenderType:   "system",
		SenderID:     "router",
		ReceiverType: "agent",
		ReceiverID:   entryAgentID,
		Content:      fmt.Sprintf("dispatch_external_message: conversation=%s source_message=%s route_to=%s", conversationID, message.ID, entryAgentID),
	}); err != nil {
		return err
	}
	_, err := r.Writer.AddMessage(conversationID, conversation.AddMessageInput{
		Scope:   "system",
		Role:    "system",
		Content: fmt.Sprintf("routing_placeholder: awaiting summary from entry agent %s", entryAgentID),
	})
	return err
}
