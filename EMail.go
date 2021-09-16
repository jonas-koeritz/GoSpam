package gospam

import (
	"net/mail"
	"time"
)

type EMail struct {
	ID      int
	Time    time.Time
	From    string
	To      []string
	Body    []byte
	Header  mail.Header
	Subject string
	Data    []byte
}
