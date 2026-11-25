package inbound

import (
	"ysmtp/internal/mail"

	"github.com/emersion/go-smtp"
)

type Sender interface {
	Send(msg mail.Message, to string) error
}

type Backend struct {
	sender      Sender
	localDomain string
	saveDir     string
}

func NewBackend(sender Sender, localDomain, saveDir string) *Backend {
	return &Backend{
		sender:      sender,
		localDomain: localDomain,
		saveDir:     saveDir,
	}
}

func (bkd *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{
		sender:      bkd.sender,
		localDomain: bkd.localDomain,
		saveDir:     bkd.saveDir,
	}, nil
}
