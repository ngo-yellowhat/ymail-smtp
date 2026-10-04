package inbound

import (
	"ysmtp/internal/outbound"

	"github.com/emersion/go-smtp"
)

type Backend struct {
	sender      *outbound.Sender
	localDomain string
}

func NewBackend(sender *outbound.Sender, localDomain string) *Backend {
	return &Backend{
		sender:      sender,
		localDomain: localDomain,
	}
}

func (bkd *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{
		sender:      bkd.sender,
		localDomain: bkd.localDomain,
	}, nil
}
