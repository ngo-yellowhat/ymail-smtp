package outbound

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"

	"ysmtp/internal/mail"

	"github.com/emersion/go-msgauth/dkim"
)

func Send(msg mail.Message, to string) error {
	domain, err := GetDomain(to)
	if err != nil {
		return err
	}
	mailServerHost, err := LookupMailServer(domain)
	if err != nil {
		return err
	}
	smtpAddr := mailServerHost + ":25"

	conn, err := net.DialTimeout("tcp", smtpAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("[  ERROR  ] dial: %v", err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return fmt.Errorf("[  ERROR  ] set deadline: %v", err)
	}

	sc, err := smtp.NewClient(conn, mailServerHost)
	if err != nil {
		return fmt.Errorf("[  ERROR  ] SMTP client: %v", err)
	}

	if err := sc.Hello("mail.yellowhat.cz"); err != nil {
		return fmt.Errorf("[  ERROR  ] HELO/EHLO: %v", err)
	}

	if ok, _ := sc.Extension("STARTTLS"); ok {
		config := &tls.Config{InsecureSkipVerify: true}
		if err = sc.StartTLS(config); err != nil {
			return fmt.Errorf("[  ERROR  ] StartTLS: %v", err)
		}
	}

	if err := sc.Mail(msg.From); err != nil {
		return fmt.Errorf("[  ERROR  ] MAIL: %v", err)
	}

	if err := sc.Rcpt(to); err != nil {
		return fmt.Errorf("[  ERROR  ] RCPT: %v", err)
	}

	w, err := sc.Data()
	if err != nil {
		return fmt.Errorf("[  ERROR  ] DATA: %v", err)
	}
	defer w.Close()

	keyBytes, err := os.ReadFile("dkim_private.pem")
	if err != nil {
		return fmt.Errorf("[  ERROR  ] read dkim private key: %v", err)
	}

	pemBlock, _ := pem.Decode(keyBytes)
	if pemBlock == nil {
		return fmt.Errorf("[  ERROR  ] failed to decode PEM block containing private key")
	}

	var parsedKey crypto.Signer
	rsaKey, errPKCS1 := x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
	if errPKCS1 == nil {
		parsedKey = rsaKey
	} else {
		pkcs8Key, errPKCS8 := x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
		if errPKCS8 != nil {
			return fmt.Errorf("[  ERROR  ] parse private key: PKCS1 (%v) | PKCS8 (%v)", errPKCS1, errPKCS8)
		}
		signerKey, ok := pkcs8Key.(crypto.Signer)
		if !ok {
			return fmt.Errorf("[  ERROR  ] parsed PKCS8 does not implement crypto.Signer")
		}
		parsedKey = signerKey
	}

	dkimOptions := &dkim.SignOptions{
		Domain:   "yellowhat.cz",
		Selector: "default",
		Signer:   parsedKey,
	}

	var fullMsg strings.Builder
	fullMsg.WriteString(fmt.Sprintf("From: %s\r\n", msg.From))
	fullMsg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	fullMsg.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	fullMsg.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	fullMsg.WriteString(fmt.Sprintf("Message-ID: <%s>\r\n", msg.MessageID))
	fullMsg.WriteString("MIME-Version: 1.0\r\n")
	fullMsg.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	fullMsg.WriteString("\r\n")
	fullMsg.WriteString(msg.Body)

	dkimSigner, err := dkim.NewSigner(dkimOptions)
	if err != nil {
		return fmt.Errorf("[  ERROR  ] init dkim signer: %v", err)
	}

	_, err = dkimSigner.Write([]byte(fullMsg.String()))
	if err != nil {
		dkimSigner.Close()
		return fmt.Errorf("[  ERROR  ] write to DKIM: %v", err)
	}

	if err := dkimSigner.Close(); err != nil {
		return fmt.Errorf("[  ERROR  ] DKIM close: %v", err)
	}

	dkimHeader := dkimSigner.Signature()

	_, err = w.Write([]byte(dkimHeader))
	if err != nil {
		return fmt.Errorf("[  ERROR  ] write DKIM header to SMTP: %v", err)
	}

	_, err = w.Write([]byte(fullMsg.String()))
	if err != nil {
		return fmt.Errorf("[  ERROR  ] write mail bytes to SMTP: %v", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("[  ERROR  ] SMTP close: %v", err)
	}

	sc.Quit()
	log.Println("[  CLIENT  ] Email sent successfully with DKIM!")

	return nil
}

func GetDomain(mail string) (string, error) {
	i := strings.LastIndex(mail, "@")
	if i == -1 {
		return "", fmt.Errorf("[  ERROR  ] Email is invalid")
	}
	return mail[i+1:], nil
}

func LookupMailServer(domain string) (string, error) {
	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		return "", fmt.Errorf(" [  ERROR  ] LookupMX: %v", err)
	}
	for _, mx := range mxRecords {
		fmt.Println(mx.Host, mx.Pref)
	}
	if len(mxRecords) == 0 {
		return "", fmt.Errorf("[  ERROR  ] Not found MX records for domain %s", domain)
	}
	best := mxRecords[0]
	return strings.TrimSuffix(best.Host, "."), nil
}
