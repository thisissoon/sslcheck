package slack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAttachment(t *testing.T) {
	type args struct {
		text   string
		status int
	}
	txt := "test abc"
	tests := []struct {
		name string
		args args
		want Attachment
	}{
		{
			name: "create an attachment with a status of 0",
			args: args{
				text:   txt,
				status: 0,
			},
			want: Attachment{
				Text:  txt,
				Color: "good",
			},
		},
		{
			name: "create an attachment with a status of 1",
			args: args{
				text:   txt,
				status: 1,
			},
			want: Attachment{
				Text:  txt,
				Color: "warning",
			},
		},
		{
			name: "create an attachment with a status of 2",
			args: args{
				text:   txt,
				status: 2,
			},
			want: Attachment{
				Text:  txt,
				Color: "danger",
			},
		},
		{
			name: "create an attachment with a status of 3",
			args: args{
				text:   txt,
				status: 3,
			},
			want: Attachment{
				Text:  txt,
				Color: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewAttachment(tt.args.text, tt.args.status)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewMsg(t *testing.T) {
	type args struct {
		a []Attachment
	}
	tests := []struct {
		name string
		args args
		want Msg
	}{
		{
			name: "create a message",
			args: args{
				a: []Attachment{
					Attachment{"test txt", "danger"},
				},
			},
			want: Msg{
				Text: "SSL certificate status",
				Attachments: []Attachment{
					Attachment{"test txt", "danger"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMsg(tt.args.a)
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
				Attachments: []Attachment{
					Attachment{
						Text:  "test attachment",
						Color: "warning",
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
				Attachments: []Attachment{
					Attachment{
						Text:  "test attachment",
						Color: "warning",
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
