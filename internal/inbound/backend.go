package inbound

import (
	"github.com/emersion/go-smtp"
)

type Backend struct {
	smtpAddr string
}

func (bkd *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{
		smtpAddr: bkd.smtpAddr,
	}, nil
}
