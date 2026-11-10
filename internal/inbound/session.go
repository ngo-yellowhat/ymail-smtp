package inbound

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	netmail "net/mail"
	"os"
	"strings"

	"ysmtp/internal/mail"
	"ysmtp/internal/outbound"

	"github.com/emersion/go-smtp"
)

const (
	maxDepth    = 10
	maxPartSize = 5 << 20
)

type Session struct {
	From        string
	To          []string
	sender      *outbound.Sender
	localDomain string
}
type Body struct {
	Text string // text/plain
	HTML string // text/html
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	log.Println("Mail from: ", from)
	s.From = from
	return nil
}

func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	log.Println("Mail to: ", to)
	domainTo, err := mail.GetDomain(to)
	if err != nil {
		return &smtp.SMTPError{
			Code:         553,
			EnhancedCode: smtp.EnhancedCode{5, 1, 3},
			Message:      "The recipients address is not correct",
		}
	}

	if strings.EqualFold(domainTo, s.localDomain) {
		return &smtp.SMTPError{
			Code:         550,
			EnhancedCode: smtp.EnhancedCode{5, 1, 1},
			Message:      "Message not delivered",
		}
	}
	s.To = append(s.To, to)
	return nil
}

func (s *Session) Data(r io.Reader) error {
	rawEmailData, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("[  ERROR  ] failed to read raw email data: %w", err)
	}

	parsedMsg, err := netmail.ReadMessage(bytes.NewReader(rawEmailData))
	if err != nil {
		return err
	}

	rawSubject := parsedMsg.Header.Get("Subject")
	decodeSubject := new(mime.WordDecoder)
	subject, err := decodeSubject.DecodeHeader(rawSubject)
	if err != nil {
		subject = rawSubject
	}

	msgID := parsedMsg.Header.Get("Message-ID")
	if msgID == "" {
		msgID = mail.GenerateMsgID(s.sender.DKIMdomain)
	}

	safeMsgID := strings.NewReplacer("<", "", ">", "", "/", "_", "@", "_").Replace(msgID)
	_ = os.Mkdir("./stored_emails", 0o755)
	fileName := fmt.Sprintf("./stored_emails/%s.eml", safeMsgID)
	if err := os.WriteFile(fileName, rawEmailData, 0o600); err != nil {
		log.Printf("[  ERROR  ] failed to save .eml file: %v\n", err)
	} else {
		log.Printf("[  SESSION  ] raw email saved to %s\n", fileName)
	}

	body, err := extractBody(
		parsedMsg.Header.Get("Content-Type"),
		parsedMsg.Header.Get("Content-Transfer-Encoding"),
		parsedMsg.Body,
		0,
	)
	if err != nil {
		log.Printf("[  ERROR  ] extract body %v\n", err)
		return &smtp.SMTPError{
			Code:         554,
			EnhancedCode: smtp.EnhancedCode{5, 6, 0},
			Message:      "The message is corrupted or too deeply nested",
		}
	}

	text := body.Text
	if text == "" { // html уходят с "text/plain". вложения теряются
		text = body.HTML // я это потом перепишу на более нормальный вариант
	}
	msg := mail.Message{From: s.From, To: s.To, Subject: subject, Body: text, MessageID: msgID}
	for _, rcpt := range s.To {
		if err := s.sender.Send(msg, rcpt); err != nil {
			log.Printf("[  ERROR  ] Failed to send to %s: %v\n", rcpt, err)
			continue
		}
	}

	return nil
}

func (s *Session) Reset() {
	s.From = ""
	s.To = nil
}
func (s *Session) Logout() error { return nil }

func decodeBody(encoding string, r io.Reader) io.Reader {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "base64":
		return base64.NewDecoder(base64.StdEncoding, r)
	case "quoted-printable":
		return quotedprintable.NewReader(r)
	default:
		return r
	}
}

func extractBody(contentType, encoding string, r io.Reader, depth int) (Body, error) {
	if depth > maxDepth {
		return Body{}, fmt.Errorf("multipart nesting too deep (max %d)", maxDepth)
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = "text/plain"
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(r, params["boundary"])
		var body Body
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return Body{}, fmt.Errorf("failed to read multipart part: %w", err)
			}
			sub, err := extractBody(
				part.Header.Get("Content-Type"),
				part.Header.Get("Content-Transfer-Encoding"),
				part,
				depth+1,
			)
			if err != nil {
				return Body{}, err
			}
			if body.Text == "" {
				body.Text = sub.Text
			}
			if body.HTML == "" {
				body.HTML = sub.HTML
			}
		}
		return body, nil

	}
	if mediaType != "text/plain" && mediaType != "text/html" {
		return Body{}, nil
	}

	decoded := decodeBody(encoding, r)
	data, err := io.ReadAll(io.LimitReader(decoded, maxPartSize))
	if err != nil {
		return Body{}, fmt.Errorf("read part: %w", err)
	}

	switch mediaType {
	case "text/plain":
		return Body{Text: string(data)}, nil
	case "text/html":
		return Body{HTML: string(data)}, nil
	}
	return Body{}, nil
}
