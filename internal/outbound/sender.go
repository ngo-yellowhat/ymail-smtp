package outbound

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"
	"time"

	"ysmtp/internal/mail"
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

	_, err = w.Write(msg.Bytes())
	if err != nil {
		return fmt.Errorf("[  ERROR  ] Write: %v", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("[  ERROR  ] Close: %v", err)
	}

	sc.Quit()
	log.Println("[  CLIENT  ] Email send successfully!")

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
