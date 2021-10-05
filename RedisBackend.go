package gospam

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/go-redis/redis/v8"
)

type RedisBackend struct {
	client          *redis.Client
	acceptedDomains []string
	ctx             context.Context
	expiration      time.Duration
}

func NewRedisBackend(addr string, password string, db int, acceptedDomains []string, retentionHours int) Backend {
	backend := &RedisBackend{
		acceptedDomains: acceptedDomains,
		ctx:             context.Background(),
		expiration:      time.Duration(retentionHours) * time.Hour,
	}

	backend.client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return backend
}

func (backend *RedisBackend) AnonymousLogin(c *smtp.ConnectionState) (smtp.Session, error) {
	log.Printf("Anonymous login from %s\n", c.RemoteAddr.String())
	return &Session{
		remote:  c.RemoteAddr,
		backend: backend,
	}, nil
}

func (backend *RedisBackend) NewSession(c smtp.ConnectionState, _ string) (smtp.Session, error) {
	log.Printf("New Session with: %s\n", c.RemoteAddr.String())
	return &Session{
		remote:  c.RemoteAddr,
		backend: backend,
	}, nil
}

func (b *RedisBackend) GetEmailById(id int) *EMail {
	scanComplete := false
	for cursor := uint64(0); !scanComplete; {
		var keys []string
		var err error
		keys, cursor, err = b.client.Scan(b.ctx, cursor, fmt.Sprintf("*:%d", id), 100).Result()
		if err != nil {
			break
		}

		for _, key := range keys {
			mailData := b.client.Get(b.ctx, key).Val()
			email := &EMail{}
			err := json.Unmarshal([]byte(mailData), email)
			if err != nil {
				log.Printf("Error unmarshalling email: %s\n", err)
				continue
			}
			return email
		}
		if cursor == 0 {
			scanComplete = true
		}
	}
	return nil
}

func (b *RedisBackend) GetEmailsByAlias(alias string) []*EMail {
	emails := make([]*EMail, 0)

	scanComplete := false
	for cursor := uint64(0); !scanComplete; {
		var keys []string
		var err error
		keys, cursor, err = b.client.Scan(b.ctx, cursor, fmt.Sprintf("%s:*", alias), 100).Result()
		if err != nil {
			log.Printf("Error GetEmailsByAlias(): %s\n", err)
			break
		}

		for _, key := range keys {
			mailData := b.client.Get(b.ctx, key).Val()
			email := &EMail{}
			err := json.Unmarshal([]byte(mailData), email)
			if err != nil {
				log.Printf("Error unmarshalling email: %s\n", err)
				continue
			}
			emails = append(emails, email)
		}
		if cursor == 0 {
			scanComplete = true
		}
	}
	return emails
}

func (b *RedisBackend) GetProcessedEmails() int {
	currentId, err := b.client.Get(b.ctx, "email_id").Int()
	if err != nil {
		return 0
	}
	return currentId
}

func (b *RedisBackend) IsAcceptedDomain(email string) bool {
	if len(b.acceptedDomains) == 0 {
		return true
	}

	emailParts := strings.Split(email, "@")
	domain := emailParts[len(emailParts)-1]

	for _, d := range b.acceptedDomains {
		if strings.EqualFold(d, domain) {
			return true
		}
	}

	return false
}

func (backend *RedisBackend) Login(_ *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	return nil, smtp.ErrAuthUnsupported
}

func (b *RedisBackend) SaveEmail(email *EMail) {
	emailId, err := b.client.Incr(b.ctx, "email_id").Result()
	if err != nil {
		log.Printf("RedisBackend error: %s\n", err)
		return
	}
	email.ID = int(emailId)

	mailData, err := json.Marshal(email)
	if err != nil {
		log.Printf("Error marshalling email: %s\n", err)
		return
	}

	for _, to := range email.To {
		alias := getAlias(to)
		err = b.client.Set(b.ctx, fmt.Sprintf("%s:%d", alias, email.ID), string(mailData), b.expiration).Err()
		if err != nil {
			log.Printf("RedisBackend error: %s\n", err)
			return
		}
	}
}

func getAlias(email string) string {
	return strings.Split(email, "@")[0]
}

func (b *RedisBackend) Cleanup(deadline time.Time) {
	// Nothing todo redis took care of this
}
