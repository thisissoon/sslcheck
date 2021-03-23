package slack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStatusBlock(t *testing.T) {
	type args struct {
		host   string
		msg    string
		status int
	}
	msg := "test"
	tests := []struct {
		name string
		args args
		want Block
	}{
		{
			name: "create a block with a status of 0",
			args: args{
				host:   "thisissoon.com",
				msg:    msg,
				status: 0,
			},
			want: Block{
				Type: "section",
				Fields: []Field{
					{
						Type: "mrkdwn",
						Text: "*Host:*\nthisissoon.com",
					},
					{
						Type: "mrkdwn",
						Text: "*Status:*\n:large_green_circle: test",
					},
				},
			},
		},
		{
			name: "create a block with a status of 1",
			args: args{
				host:   "thisissoon.com",
				msg:    msg,
				status: 1,
			},
			want: Block{
				Type: "section",
				Fields: []Field{
					{
						Type: "mrkdwn",
						Text: "*Host:*\nthisissoon.com",
					},
					{
						Type: "mrkdwn",
						Text: "*Status:*\n:warning: test",
					},
				},
			},
		},
		{
			name: "create a block with a status of 2",
			args: args{
				host:   "thisissoon.com",
				msg:    msg,
				status: 2,
			},
			want: Block{
				Type: "section",
				Fields: []Field{
					{
						Type: "mrkdwn",
						Text: "*Host:*\nthisissoon.com",
					},
					{
						Type: "mrkdwn",
						Text: "*Status:*\n:red_circle: test",
					},
				},
			},
		},
		{
			name: "create a block with a status of 3",
			args: args{
				host:   "thisissoon.com",
				msg:    msg,
				status: 3,
			},
			want: Block{
				Type: "section",
				Fields: []Field{
					{
						Type: "mrkdwn",
						Text: "*Host:*\nthisissoon.com",
					},
					{
						Type: "mrkdwn",
						Text: "*Status:*\n test",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewStatusBlock(tt.args.host, tt.args.msg, tt.args.status)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewMsg(t *testing.T) {
	type args struct {
		blocks []Block
	}
	tests := []struct {
		name string
		args args
		want Msg
	}{
		{
			name: "create a message",
			args: args{
				blocks: []Block{
					{
						Type: "section",
						Fields: []Field{
							{
								Type: "mrkdwn",
								Text: "*Host:*\nthisissoon.com",
							},
							{
								Type: "mrkdwn",
								Text: "*Status:*\n test",
							},
						},
					},
				},
			},
			want: Msg{
				Text: "SSL Certificate Status",
				Blocks: []Block{
					{
						Type: "header",
						Text: &TextBlock{
							Type: "plain_text",
							Text: "SSL Certificate Status",
						},
					},
					{
						Type: "section",
						Fields: []Field{
							{
								Type: "mrkdwn",
								Text: "*Host:*\nthisissoon.com",
							},
							{
								Type: "mrkdwn",
								Text: "*Status:*\n test",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMsg(tt.args.blocks)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSendMsg(t *testing.T) {
	tests := []struct {
		name            string
		msg             Msg
		url             string
		endpointRspCode int
		wantErr         bool
	}{
		{
			name: "should send a message",
			msg: Msg{
				Text: "hello message",
				Blocks: []Block{
					{
						Type: "mrkdwn",
						Text: &TextBlock{
							Type: "plain_text",
							Text: "SSL Certificate Status",
						},
					},
				},
			},
			url:             "/slack-endpoint",
			endpointRspCode: http.StatusOK,
			wantErr:         false,
		},
		{
			name: "should return an error if the service is unavailable",
			msg: Msg{
				Text: "hello message",
				Blocks: []Block{
					{
						Type: "mrkdwn",
						Text: &TextBlock{
							Type: "plain_text",
							Text: "SSL Certificate Status",
						},
					},
				},
			},
			url:             "/slack-endpoint",
			endpointRspCode: http.StatusUnauthorized,
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.endpointRspCode)
				// check for POST method
				if r.Method != "POST" {
					t.Errorf("Expected 'POST' request, got %s", r.Method)
				}
				// check we are calling specified URL
				if r.URL.EscapedPath() != "/slack-endpoint" {
					t.Errorf("Expected request to `/slack-endpoint`, got ‘%s’", r.URL.EscapedPath())
				}
				// decode request body and check it matches
				decoder := json.NewDecoder(r.Body)
				var m Msg
				err := decoder.Decode(&m)
				if err != nil {
					t.Fatalf("error decoding response")
				}
				assert.Equal(t, tt.msg, m)
			}))
			defer ts.Close()
			if err := SendMsg(tt.msg, ts.URL+tt.url); (err != nil) != tt.wantErr {
				t.Errorf("SendMsg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
