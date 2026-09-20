package inbound

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"mime"
	netmail "net/mail"
	"strings"
	"time"
	"ysmtp/internal/mail"
	"ysmtp/internal/outbound"

	"github.com/emersion/go-smtp"
)

type Session struct {
	From string
	To   []string

	smtpAddr string
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	log.Println("Mail from: ", from)
	s.From = from
	return nil
}

func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	log.Println("Mail to: ", to)
	s.To = append(s.To, to)
	return nil
}

func (s *Session) Data(r io.Reader) error {
	parsedMsg, err := netmail.ReadMessage(r)
	if err != nil {
		return err
	}

	// декодирование кириллицы
	rawSubject := parsedMsg.Header.Get("Subject")
	dec := new(mime.WordDecoder)
	subject, err := dec.DecodeHeader(rawSubject)
	if err != nil {
		subject = rawSubject
		// если это обычный текст
	}

	msgID := parsedMsg.Header.Get("Message-ID")
	if msgID == "" {
		msgID = generateMsgID("yellowhat.cz")
	}

	bodyBuf := new(strings.Builder)
	io.Copy(bodyBuf, parsedMsg.Body)

	msg := mail.Message{From: s.From, To: s.To, Subject: subject, Body: bodyBuf.String(), MessageID: msgID}
	log.Printf("[  NEW EMAIL  ] From: %s | To: %v", s.From, s.To)
	for _, rcpt := range s.To {
		if err := outbound.Send(msg, rcpt); err != nil {
			fmt.Printf("[  ERROR  ] Failed to send to %s: %v", rcpt, err)
			continue
		}
	}
	return nil
}

func (s *Session) Reset() {
	s.From = ""
	s.To = nil
}
func (s *Session) Logout() error { return nil }

func generateMsgID(domain string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("<%x-%d@%s>", b, time.Now().UnixNano(), domain)
}
