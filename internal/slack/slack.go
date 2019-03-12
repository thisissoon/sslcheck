package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	Danger  = "danger"
	Warning = "warning"
	Good    = "good"
	Default = ""
)

func statusToColor(status int) string {
	switch status {
	case 0:
		return Good
	case 1:
		return Warning
	case 2:
		return Danger
	default:
		return Default
	}
}

// Attachment a slack message attachment
type Attachment struct {
	Text  string `json:"text"`
	Color string `json:"color"`
}

// Msg a slack message including attachments
type Msg struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}

// NewAttachment create a new attachment by providing the text and the status code, which gets converted into a color
func NewAttachment(text string, status int) Attachment {
	return Attachment{
		Text:  text,
		Color: statusToColor(status),
	}
}

// NewMsg create a new message for slack providing attachments
func NewMsg(a []Attachment) Msg {
	return Msg{
		Text:        "SSL certificate status",
		Attachments: a,
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
