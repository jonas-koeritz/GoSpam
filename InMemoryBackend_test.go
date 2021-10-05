package gospam_test

import (
	"testing"

	"github.com/jonas-koeritz/gospam"
)

func TestSubdomains(t *testing.T) {
	backend := &gospam.InMemoryBackend{
		MaxStoredMessage: 100,
		AcceptedDomains:  []string{"example.com"},
		AcceptSubdomains: true,
	}

	if backend.IsAcceptedDomain("sub.example.com") == false {
		t.Errorf("subdomain not accepted when it should've been")
	}

	if backend.IsAcceptedDomain("example.com") == false {
		t.Errorf("domain not accepted when it should've been")
	}

	backend = &gospam.InMemoryBackend{
		MaxStoredMessage: 100,
		AcceptedDomains:  []string{"example.com"},
		AcceptSubdomains: false,
	}

	if backend.IsAcceptedDomain("sub.example.com") == true {
		t.Errorf("subdomain accepted when it shouldn't have been")
	}

	if backend.IsAcceptedDomain("example.com") == false {
		t.Errorf("domain not accepted when it should've been")
	}
}
