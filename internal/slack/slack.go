package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func statusToEmoji(status int) string {
	switch status {
	case 0:
		return ":large_green_circle:"
	case 1:
		return ":warning:"
	case 2:
		return ":red_circle:"
	default:
		return ""
	}
}

// A TextBlock represents text content within a block
type TextBlock struct {
	Type  string `json:"type,omitempty"`
	Text  string `json:"text,omitempty"`
	Emoji bool   `json:"emoji,omitempty"`
}

// A Field within a block
type Field struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Block represents a slack block (see https://app.slack.com/block-kit-builder)
type Block struct {
	Type   string     `json:"type"`
	Text   *TextBlock `json:"text,omitempty"`
	Fields []Field    `json:"fields,omitempty"`
}

// Msg a slack message including blocks
type Msg struct {
	Text   string  `json:"text"`
	Blocks []Block `json:"blocks"`
}

// NewStatusBlock create a new block with text and status code, which gets translated into an emoji
func NewStatusBlock(host string, msg string, status int) Block {
	return Block{
		Type: "section",
		Fields: []Field{
			{
				Type: "mrkdwn",
				Text: "*Host:*\n" + host,
			},
			{
				Type: "mrkdwn",
				Text: fmt.Sprintf("*Status:*\n%s %s", statusToEmoji(status), msg),
			},
		},
	}
}

// NewMsg create a new message for slack with blocks
func NewMsg(b []Block) Msg {
	blocks := []Block{
		{
			Type: "header",
			Text: &TextBlock{
				Type: "plain_text",
				Text: "SSL Certificate Status",
			},
		},
	}
	blocks = append(blocks, b...)
	return Msg{
		Text:   "SSL Certificate Status",
		Blocks: blocks,
	}
}

// SendMsg sends a message to slack webhook url endpoint
func SendMsg(msg Msg, hookUrl string) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshalling json for slack message: %v", err)
	}
	rsp, err := http.Post(hookUrl, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error posting message to slack: %v", err)
	}
	if rsp.StatusCode >= 400 {
		return fmt.Errorf("error sending message to slack. Got status: %v", rsp.Status)
	}
	return nil
}
