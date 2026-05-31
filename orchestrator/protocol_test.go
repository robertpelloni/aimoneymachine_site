package orchestrator

import (
	"net/url"
	"testing"
)

func TestHustleProtocol(t *testing.T) {
	proto := NewHustleProtocol()

	proto.Register("research", func(params url.Values) error {
		return nil
	})

	validURIs := []string{
		"hustle://research?query=AI+Trends",
	}

	for _, uri := range validURIs {
		err := proto.HandleURI(uri)
		if err != nil {
			t.Errorf("Failed to handle valid URI %s: %v", uri, err)
		}
	}

	invalidURIs := []string{
		"http://research",
		"hustle://unknown",
		"invalid-uri",
	}

	for _, uri := range invalidURIs {
		err := proto.HandleURI(uri)
		if err == nil {
			t.Errorf("Expected error for invalid URI %s, but got nil", uri)
		}
	}
}
