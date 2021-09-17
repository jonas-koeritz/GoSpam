package gospam

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/mail"
	"time"

	"github.com/emersion/go-smtp"
)

type Session struct {
	backend Backend
	from    string
	to      []string
}

func (s *Session) AuthPlain(username, password string) error {
	return smtp.ErrAuthUnsupported
}

func (s *Session) Mail(from string, opts smtp.MailOptions) error {
	log.Printf("Mail from: %s\n", from)
	s.from = from
	return nil
}

func (s *Session) Rcpt(to string) error {
	log.Printf("Mail to: %s\n", to)
	if !s.backend.IsAcceptedDomain(to) {
		log.Printf("not in AcceptedDomains\n")
		return &smtp.SMTPError{
			Code:         550,
			EnhancedCode: smtp.EnhancedCode{5, 1, 1},
			Message:      "Invalid recipient",
		}
	}
	if s.to == nil {
		s.to = make([]string, 0)
	}
	s.to = append(s.to, to)

	return nil
}

func (s *Session) Data(r io.Reader) error {
	if m, err := ioutil.ReadAll(r); err != nil {
		return err
	} else {
		parsedMail, err := mail.ReadMessage(bytes.NewReader(m))
		if err != nil {
			return err
		}
		b, err := ioutil.ReadAll(parsedMail.Body)
		if err != nil {
			return err
		}

		mimeDecoder := new(mime.WordDecoder)
		subject, err := mimeDecoder.Decode(parsedMail.Header.Get("Subject"))
		if err != nil {
			subject = parsedMail.Header.Get("Subject")
		}

		s.backend.SaveEmail(&EMail{
			Time:    time.Now(),
			From:    s.from,
			To:      s.to,
			Header:  parsedMail.Header,
			Body:    b,
			Data:    m,
			Subject: subject,
		})
		log.Printf("Saved E-Mail Message with %d bytes of body data\n", len(b))
	}

	return nil
}

func (s *Session) Reset() {
	s.from = ""
	s.to = make([]string, 0)
}

func (s *Session) Logout() error {
	return nil
}
