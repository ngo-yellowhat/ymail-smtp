package inbound

import (
	"strings"
	"testing"

	"ysmtp/internal/mail"
)

type fakeSender struct {
	Sent []struct {
		Msg mail.Message
		To  string
	}
}

func (f *fakeSender) Send(msg mail.Message, to string) error {
	f.Sent = append(f.Sent, struct {
		Msg mail.Message
		To  string
	}{Msg: msg, To: to})
	return nil
}

func TestSession_Data_SendsMessage(t *testing.T) {
	sender := &fakeSender{}

	s := &Session{
		From:        "alice@yellowhat.cz",
		To:          []string{"bob@example.com"},
		sender:      sender,
		localDomain: "yellowhat.cz",
	}

	raw := "Subject: Hello\r\n" + "\r\n" + "TEST BODY.\r\n" // минимальное письмо по стандартам rfc5321

	err := s.Data(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("Data returned error: %v", err)
	}

	if len(sender.Sent) != 1 {
		t.Fatalf("expected 1 sent message, got %d", err)
	}

	sent := sender.Sent[0]
	if sent.To != "bob@example.com" {
		t.Errorf("To = %q, want = %q", sent.To, "bob@example.com")
	}
	if sent.Msg.From != "alice@yellowhat.cz" {
		t.Errorf("From = %q, want = %q", sent.Msg.From, "alice@yellowhat.cz")
	}
	if sent.Msg.Subject != "Hello" {
		t.Errorf("Subject = %q, want = %q", sent.Msg.Subject, "Hello")
	}
	if sent.Msg.Body != "TEST BODY.\r\n" {
		t.Errorf("Body = %q, want = %q", sent.Msg.Body, "TEST BODY.\r\n")
	}
}

func TestExtractBody(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		encoding    string
		input       string
		want        Body
		wantErr     bool
	}{
		{
			name:        "plain text",
			contentType: "text/plain",
			input:       "hello",
			want:        Body{Text: "hello"},
		},
		{
			name:        "alternative",
			contentType: `multipart/alternative; boundary="b1"`,
			input: "--b1\r\n" +
				"Content-Type: text/plain\r\n\r\n" +
				"hello\r\n" +
				"--b1\r\n" +
				"Content-Type: text/html\r\n\r\n" +
				"<p>hello</p>\r\n" +
				"--b1--\r\n",
			want: Body{Text: "hello", HTML: "<p>hello</p>"},
		},
		{
			name:        "base64",
			contentType: "text/plain",
			encoding:    "base64",
			input:       "aGVsbG8=", // "aGVsbG8K" - hello\n
			want:        Body{Text: "hello"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractBody(tt.contentType, tt.encoding, strings.NewReader(tt.input), 0)
			if (err != nil) != tt.wantErr {
				t.Fatalf("extractBody(%q) error = %v, want error = %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("extractBody(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
