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

	fmt.Fprintf(&msg, "From: %s\r\n", m.From)
	fmt.Fprintf(&msg, "To: %s\r\n", strings.Join(m.To, ", "))
	fmt.Fprintf(&msg, "Subject: %s\r\n", m.Subject)
	fmt.Fprintf(&msg, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	if cleanId != "" {
		fmt.Fprintf(&msg, "Message-ID: <%s>\r\n", cleanId)
	}
	fmt.Fprint(&msg, "MIME-Version: 1.0\r\n")
	fmt.Fprint(&msg, "Content-Type: text/plain; charset=utf-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(m.Body)

	return []byte(msg.String())
}

func GenerateMsgID(domain string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%d@%s", b, time.Now().UnixNano(), domain)
}

func GetDomain(mail string) (string, error) {
	i := strings.LastIndex(mail, "@")
	if i == -1 {
		return "", fmt.Errorf("[  ERROR  ] Email is invalid")
	}
	return mail[i+1:], nil
}
