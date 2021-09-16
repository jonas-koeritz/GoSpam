package gospam

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-smtp"
)

type Backend struct {
	currentID        uint64
	emails           []*EMail
	MaxStoredMessage int
	mailMutex        sync.Mutex
}

func (backend *Backend) NewSession(_ smtp.ConnectionState, _ string) (smtp.Session, error) {
	return &Session{
		backend: backend,
	}, nil
}

func (backend *Backend) AnonymousLogin(_ *smtp.ConnectionState) (smtp.Session, error) {
	return &Session{
		backend: backend,
	}, nil
}

func (backend *Backend) Login(_ *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	return nil, smtp.ErrAuthUnsupported
}

func (b *Backend) Email(email *EMail) {
	b.mailMutex.Lock()
	if b.emails == nil {
		b.emails = make([]*EMail, 0)
	}

	if len(b.emails) >= b.MaxStoredMessage {
		b.emails = b.emails[1:]
	}

	email.ID = b.currentID
	b.currentID++

	b.emails = append(b.emails, email)
	b.mailMutex.Unlock()
}

func (b *Backend) GetEmailsByAlias(alias string) []*EMail {
	emails := make([]*EMail, 0)
	for _, e := range b.emails {
		for _, recipient := range e.To {
			if strings.HasPrefix(recipient, alias+"@") {
				emails = append(emails, e)
				break
			}
		}
	}
	return emails
}

func (b *Backend) GetEmailById(id uint64) *EMail {
	for _, e := range b.emails {
		if e.ID == id {
			return e
		}
	}
	return nil
}

func (b *Backend) Cleanup(retentionHours int) {
	deadline := time.Now().Add(time.Duration(-retentionHours) * time.Hour)
	log.Printf("Deleting all messages received before %s\n", deadline)
	unexpiredMails := make([]*EMail, 0)
	for _, e := range b.emails {
		if !e.Time.Before(deadline) {
			unexpiredMails = append(unexpiredMails, e)
		}
	}
	b.mailMutex.Lock()
	b.emails = unexpiredMails
	b.mailMutex.Unlock()
}
