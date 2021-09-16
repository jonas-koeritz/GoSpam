package gospam

import (
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-smtp"
)

type InMemoryBackend struct {
	currentID        uint64
	emails           []*EMail
	MaxStoredMessage int
	mailMutex        sync.Mutex
}

func (backend *InMemoryBackend) NewSession(_ smtp.ConnectionState, _ string) (smtp.Session, error) {
	return &Session{
		backend: backend,
	}, nil
}

func (backend *InMemoryBackend) AnonymousLogin(_ *smtp.ConnectionState) (smtp.Session, error) {
	return &Session{
		backend: backend,
	}, nil
}

func (backend *InMemoryBackend) Login(_ *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	return nil, smtp.ErrAuthUnsupported
}

func (b *InMemoryBackend) SaveEmail(email *EMail) {
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

func (b *InMemoryBackend) GetEmailsByAlias(alias string) []*EMail {
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

func (b *InMemoryBackend) GetEmailById(id uint64) *EMail {
	for _, e := range b.emails {
		if e.ID == id {
			return e
		}
	}
	return nil
}

func (b *InMemoryBackend) Cleanup(deadline time.Time) {
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