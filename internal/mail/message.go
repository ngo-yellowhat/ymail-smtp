package mail

import (
	"fmt"
	"strings"
)

type Message struct {
	From      string
	To        []string
	Subject   string
	Body      string
	MessageID string
}

func (m Message) Bytes() []byte {
	headers := make(map[string]string)

	headers["From"] = m.From
	headers["To"] = strings.Join(m.To, ", ")
	headers["Subject"] = m.Subject
	headers["Message-ID"] = m.MessageID
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""

	var msg strings.Builder
	for k, v := range headers { // k - заголовок ; v - значение заголовка
		if v != "" {
			msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
	}
	msg.WriteString("\r\n")
	msg.WriteString(m.Body)

	return []byte(msg.String())
}
