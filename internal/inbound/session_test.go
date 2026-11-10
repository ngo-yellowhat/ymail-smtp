package inbound

import (
	"strings"
	"testing"
)

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
