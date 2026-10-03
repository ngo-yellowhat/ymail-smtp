package mail

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"
)

type Message struct {
	From      string
	To        []string
	Subject   string
	Body      string
	MessageID string
}

func (m Message) Bytes() []byte {
	var msg strings.Builder
	cleanId := strings.Trim(m.MessageID, "<>")

	msg.WriteString(fmt.Sprintf("From: %s\r\n", m.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(m.To, ", ")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", m.Subject))
	msg.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	if cleanId != "" {
		msg.WriteString(fmt.Sprintf("Message-ID: <%s>\r\n", cleanId))
	}
	msg.WriteString(fmt.Sprintf("MIME-Version: 1.0\r\n"))
	msg.WriteString(fmt.Sprintf("Content-Type: text/plain; charset=utf-8\r\n"))
	msg.WriteString("\r\n")
	msg.WriteString(m.Body)

	return []byte(msg.String())
}

func GenerateMsgID(domain string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%d@%s", b, time.Now().UnixNano(), domain)
}
