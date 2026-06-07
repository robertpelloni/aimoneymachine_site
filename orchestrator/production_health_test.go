package orchestrator

import (
	"testing"
	"net/url"
)

func TestProductionHealthCheck(t *testing.T) {
	protocol := NewHustleProtocol()

	// Define expected modules
	modules := []string{"research", "curation", "social", "trading", "swarm", "chain", "healer", "sync"}

	for _, mod := range modules {
		protocol.Register(mod, func(p url.Values) error { return nil })
	}

	for _, mod := range modules {
		err := protocol.HandleURI("hustle://" + mod)
		if err != nil {
			t.Errorf("Production module %s failed registration check: %v", mod, err)
		}
	}
}
