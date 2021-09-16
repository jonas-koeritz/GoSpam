package gospam

import "github.com/emersion/go-smtp"

type Backend interface {
	smtp.Backend
	SaveEmail(*EMail)
	IsAcceptedDomain(email string) bool
}
