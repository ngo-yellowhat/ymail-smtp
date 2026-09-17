package inbound

import (
	"io"
	"log"
	"strings"

	"github.com/emersion/go-smtp"
)

type Session struct {
	From string
	To   []string
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	log.Println("Mail from: ", from)
	s.From = from
	return nil
}

func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	log.Println("Mail for: ", to)
	s.To = append(s.To, to)
	return nil
}

func (s *Session) Data(r io.Reader) error {
	buf := new(strings.Builder)
	n, err := io.Copy(buf, r)
	log.Printf("[  DEBUG  ] Data: copy %d byte; err=%v", n, err)
	if err != nil {
		return err
	}
	log.Printf("[NEW EMAIL] From: %s | To: %v", s.From, s.To)
	log.Println("Content:\n", buf.String())
	return nil
}

func (s *Session) Reset() {
	s.From = ""
	s.To = nil
}
func (s *Session) Logout() error { return nil }
