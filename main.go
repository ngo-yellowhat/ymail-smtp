package main

import (
	"flag"
	"log"
	"net"
	"time"

	"yellowsmtp/internal/inbound"
	"yellowsmtp/internal/outbound"

	"github.com/emersion/go-smtp"
)

func main() {
	flagSend := flag.Bool("send", false, "automaticly send a test email after startup")
	flag.Parse()

	be := &inbound.Backend{}
	s := smtp.NewServer(be)

	s.Addr = ":2525"
	s.Domain = "localhost"
	s.ReadTimeout = 5 * time.Minute
	s.WriteTimeout = 10 * time.Second
	s.AllowInsecureAuth = true

	if *flagSend {
		go func() {
			log.Println("[  SERVER  ] SMTP start on :2525")
			if err := s.ListenAndServe(); err != nil {
				log.Fatalf("[  ERROR  ] Server: %v", err)
			}
		}()
		for {
			conn, err := net.Dial("tcp", "localhost:2525")
			if err == nil {
				conn.Close()
				break
			}
			time.Sleep(60 * time.Millisecond)
		}
		log.Println("[  TEST  ] Email sending test..")
		outbound.Send()
		time.Sleep(500 * time.Millisecond)
		return
	}
	log.Println("[  SERVER  ] SMTP start on :2525")
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("[  ERROR  ] Server: %v", err)
	}
}
