package inbound

import (
	"io"
	"log"
	"mime"
	netmail "net/mail"
	"strings"
	"ysmtp/internal/mail"
	"ysmtp/internal/outbound"

	"github.com/emersion/go-smtp"
)

type Session struct {
	From string
	To   []string

	smtpAddr string
	sender   *outbound.Sender
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
		msgID = mail.GenerateMsgID("yellowhat.cz")
	}

	bodyBuf := new(strings.Builder)
	io.Copy(bodyBuf, parsedMsg.Body)

	msg := mail.Message{From: s.From, To: s.To, Subject: subject, Body: bodyBuf.String(), MessageID: msgID}
	log.Printf("[  NEW EMAIL  ] From: %s | To: %v", s.From, s.To)
	for _, rcpt := range s.To {
		if err := s.sender.Send(msg, rcpt); err != nil {
			log.Printf("[  ERROR  ] Failed to send to %s: %v\n", rcpt, err)
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

func decodeBody(encoding string, r io.Reader) io.Reader {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "base64":
		return base64.NewDecoder(base64.StdEncoding, r)
	case "quoted-printable":
		return quotedprintable.NewReader(r)
	default:
		return r
	}
}
