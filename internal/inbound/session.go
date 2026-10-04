package inbound

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/quotedprintable"
	netmail "net/mail"
	"strings"

	"ysmtp/internal/mail"
	"ysmtp/internal/outbound"

	"github.com/emersion/go-smtp"
)

type Session struct {
	From string
	To   []string

	smtpAddr    string
	sender      *outbound.Sender
	localDomain string
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	log.Println("Mail from: ", from)
	s.From = from
	return nil
}

func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	log.Println("Mail to: ", to)
	domainTo, err := mail.GetDomain(to)
	if err != nil {
		return &smtp.SMTPError{
			Code:         553,
			EnhancedCode: smtp.EnhancedCode{5, 1, 3},
			Message:      "The recipients address is not correct",
		}
	}

	if strings.EqualFold(domainTo, s.localDomain) {
		return &smtp.SMTPError{
			Code:         550,
			EnhancedCode: smtp.EnhancedCode{5, 1, 1},
			Message:      "Message not delivered",
		}
	}
	s.To = append(s.To, to)
	return nil
}

func (s *Session) Data(r io.Reader) error {
	parsedMsg, err := netmail.ReadMessage(r)
	if err != nil {
		return err
	}

	// декодирование письма
	// subject
	rawSubject := parsedMsg.Header.Get("Subject")
	decodeSubject := new(mime.WordDecoder)
	subject, err := decodeSubject.DecodeHeader(rawSubject)
	if err != nil {
		subject = rawSubject
	}
	// body
	encode := parsedMsg.Header.Get("Content-Transfer-Encoding")
	decoded := decodeBody(encode, parsedMsg.Body)
	bodyBuf := new(strings.Builder)
	if _, err := io.Copy(bodyBuf, decoded); err != nil {
		return fmt.Errorf("[  ERROR  ] failed to decode body: %w", err)
	}

	msgID := parsedMsg.Header.Get("Message-ID")
	if msgID == "" {
		msgID = mail.GenerateMsgID(s.sender.DKIMdomain)
	}

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
