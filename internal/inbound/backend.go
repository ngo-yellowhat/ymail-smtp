package inbound

import (
	"ysmtp/internal/outbound"

	"github.com/emersion/go-smtp"
)

type Backend struct {
	sender *outbound.Sender
}

func NewBackend(sender *outbound.Sender) *Backend {
	return &Backend{
		sender: sender,
	}
}

func (bkd *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{
		sender: bkd.sender,
	}, nil
}
