package outbound

import (
	"crypto/tls"
	"log"
	"net"
	"net/smtp"
	"time"

	"yellowsmtp/internal/mail"
)

func Send() {
	// flagPort := flag.Int("port", 587, "smtp port for send")
	// flagMail := flag.String("from-mail", "", "email address from")

	msg := mail.Message{
		From: "amnes00a@gmail.com",
		To:   []string{"amnes00a@gmail.com"}, Subject: "TEST",
		Body: "DEMO DEMO DEMO",
	}

	conn, err := net.DialTimeout("tcp", "smtp.gmail.com:587", 10*time.Second)
	if err != nil {
		log.Fatalf("[  ERROR  ] dial: %v", err)
	}
	if err := conn.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		conn.Close()
		log.Fatalf("[  ERROR  ] set deadline: %v", err)
	}

	sc, err := smtp.NewClient(conn, "smtp.gmail.com")
	if err != nil {
		sc.Close()
		log.Fatalf("[  ERROR  ] stmp client: %v", err)
	}

	if ok, _ := sc.Extension("STARTTLS"); ok {
		config := &tls.Config{InsecureSkipVerify: true}
		if err = sc.StartTLS(config); err != nil {
			log.Fatalf("[  ERROR  ] StartTLS: %v", err)
		}
	}

	auth := smtp.PlainAuth("", "amnes00a@gmail.com", "qmmu ywji nnvz fbmy", "smtp.gmail.com")
	if err := sc.Auth(auth); err != nil {
		log.Fatalf("[  ERROR  ] Auth: %v", err)
	}
	if err := sc.Mail(msg.From); err != nil {
		log.Fatalf("[  ERROR  ] MAIL: %v", err)
	}

	for _, addr := range msg.To {
		if err := sc.Rcpt(addr); err != nil {
			log.Fatalf("[  ERROR  ] RCPT: %v", err)
		}
	}

	w, err := sc.Data()
	if err != nil {
		log.Fatalf("[  ERROR  ] DATA: %v", err)
	}

	_, err = w.Write(msg.Bytes())
	if err != nil {
		log.Fatalf("[  ERROR  ] Write: %v", err)
	}

	if err := w.Close(); err != nil {
		log.Fatalf("[  ERROR  ] Close: %v", err)
	}

	sc.Quit()
	log.Println("[  CLIENT  ] Email send successfully!")
}
