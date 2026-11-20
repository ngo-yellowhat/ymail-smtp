package outbound

import (
	"bytes"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"sort"
	"time"

	"ysmtp/internal/mail"

	"github.com/emersion/go-msgauth/dkim"
)

type Sender struct {
	heloDomain   string
	dkimDomain   string
	dkimSelector string
	dkimKey      crypto.Signer
}

/*
func (s *Sender) DKIMdomain() string {
	return s.dkimDomain
}
понадобится позже
*/

func NewSender(heloDomain, dkimDomain, dkimSelector, dkimKeyPath string) (*Sender, error) {
	s := &Sender{
		heloDomain:   heloDomain,
		dkimDomain:   dkimDomain,
		dkimSelector: dkimSelector,
	}
	if dkimKeyPath == "" || dkimKeyPath == "." {
		return s, nil
	} else {
		signer, err := loadPrivateKey(dkimKeyPath)
		if err != nil {
			return nil, fmt.Errorf("[  ERROR  ] load dkim key: %w", err)
		}
		s.dkimKey = signer
	}

	return s, nil
}

func (s *Sender) Send(msg mail.Message, to string) error {
	raw := msg.Bytes()
	payload := raw
	reader := bytes.NewReader(raw)

	if s.dkimKey != nil {
		opts := &dkim.SignOptions{
			Domain:   s.dkimDomain,
			Selector: s.dkimSelector,
			Signer:   s.dkimKey,
			Hash:     crypto.SHA256,
		}
		var signed bytes.Buffer
		if err := dkim.Sign(&signed, reader, opts); err != nil {
			return fmt.Errorf("[  ERROR  ] DKIM sign: %v", err)
		}
		payload = signed.Bytes()
	}

	domain, err := mail.GetDomain(to)
	if err != nil {
		return err
	}

	mxRecords, err := lookupMailServer(domain)
	if err != nil {
		return fmt.Errorf("[  ERROR  ] MX lookup: %v", err)
	}
	mxHost := mxRecords[0].Host

	connTimeout := net.Dialer{
		Timeout: 5 * time.Second,
	}
	conn, err := connTimeout.Dial("tcp", mxHost+":25")
	if err != nil {
		return fmt.Errorf("[  ERROR  ] dial host: %v", err)
	}
	defer func() { _ = conn.Close() }()

	sc, err := smtp.NewClient(conn, mxHost)
	if err != nil {
		return fmt.Errorf("[  ERROR  ] failed SMTP: %v", err)
	}
	if err := sc.Hello(s.heloDomain); err != nil {
		return fmt.Errorf("[  ERROR  ] HELO: %v", err)
	}

	tlsConfig := tls.Config{
		ServerName: mxHost,
		MinVersion: tls.VersionTLS12,
	}
	if starttls, _ := sc.Extension("STARTTLS"); starttls {
		if err := sc.StartTLS(&tlsConfig); err != nil {
			return fmt.Errorf("[  ERROR  ] STARTTLS: %v", err)
		}
	}

	if err := sc.Mail(msg.From); err != nil {
		return fmt.Errorf("[  ERROR  ] mail from: %v", err)
	}
	if err := sc.Rcpt(to); err != nil {
		return fmt.Errorf("[ ERROR  ] mail to: %v", err)
	}

	w, err := sc.Data()
	if err != nil {
		return fmt.Errorf("[  ERROR  ] data :%v", err)
	}
	if _, err := w.Write(payload); err != nil {
		return fmt.Errorf("[  ERROR  ] write data: %v", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("[  ERROR  ] close data: %v", err)
	}

	return sc.Quit()
}

func loadPrivateKey(path string) (crypto.Signer, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	pemBlock, _ := pem.Decode(keyBytes)
	if pemBlock == nil {
		return nil, errors.New("[  ERROR  ] failed to decode PEM block")
	}

	if rsaKey, err := x509.ParsePKCS1PrivateKey(pemBlock.Bytes); err == nil {
		return rsaKey, nil
	}

	pkcs8Key, err := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("[  ERROR  ] parse key failed: %v", err)
	}

	signerKey, ok := pkcs8Key.(crypto.Signer)
	if !ok {
		return nil, errors.New("[  ERROR  ] key does not impelement crypto.Signer")
	}

	return signerKey, nil
}

func lookupMailServer(domain string) ([]*net.MX, error) {
	mxRecords, err := net.LookupMX(domain)
	if err != nil || len(mxRecords) == 0 {
		return nil, fmt.Errorf(" [  ERROR  ] LookupMX: %v", err)
	}
	sort.Slice(mxRecords, func(i, j int) bool {
		return mxRecords[i].Pref < mxRecords[j].Pref
	})
	return mxRecords, nil
}
