package gospam

import (
	"time"

	"github.com/emersion/go-smtp"
)

type Backend interface {
	smtp.Backend
	SaveEmail(*EMail)
	IsAcceptedDomain(email string) bool
	GetProcessedEmails() int
	GetEmailsByAlias(alias string) []*EMail
	GetEmailById(id int) *EMail
	Cleanup(deadline time.Time)
}
