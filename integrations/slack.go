package integrations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// SlackIntegration provides Slack bot functionality
type SlackIntegration struct {
	webhookURL string
	botToken   string
}

// NewSlackIntegration creates a new Slack integration
func NewSlackIntegration(webhookURL, botToken string) *SlackIntegration {
	return &SlackIntegration{
		webhookURL: webhookURL,
		botToken:   botToken,
	}
}

// SendMessage sends a message to Slack
func (si *SlackIntegration) SendMessage(channel, text string) error {
	message := map[string]interface{}{
		"channel": channel,
		"text":    text,
	}
	
	jsonData, _ := json.Marshal(message)
	resp, err := http.Post(si.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("slack API error: %d", resp.StatusCode)
	}
	
	return nil
}

// SendCommandResult sends command execution result to Slack
func (si *SlackIntegration) SendCommandResult(channel, command string, success bool, output string) error {
	emoji := "✅"
	if !success {
		emoji = "❌"
	}
	
	text := fmt.Sprintf("%s Command: `%s`\n```%s```", emoji, command, output)
	return si.SendMessage(channel, text)
}

// SendApprovalRequest sends approval request to Slack
func (si *SlackIntegration) SendApprovalRequest(channel, command, requester string) error {
	text := fmt.Sprintf("🔔 Approval Required\n"+
		"Command: `%s`\n"+
		"Requested by: %s\n"+
		"React with ✅ to approve or ❌ to reject", command, requester)
	
	return si.SendMessage(channel, text)
}
