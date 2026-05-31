package orchestrator

import (
	"fmt"
	"net/url"
	"strings"
)

type HustleHandler func(params url.Values) error

// HustleProtocol handles parsing and routing of hustle:// URIs
type HustleProtocol struct {
	Handlers map[string]HustleHandler
}

func NewHustleProtocol() *HustleProtocol {
	return &HustleProtocol{
		Handlers: make(map[string]HustleHandler),
	}
}

func (p *HustleProtocol) Register(module string, handler HustleHandler) {
	p.Handlers[strings.ToLower(module)] = handler
}

// HandleURI parses a URI like hustle://research?query=AI+Trends and routes it
func (p *HustleProtocol) HandleURI(rawURI string) error {
	u, err := url.Parse(rawURI)
	if err != nil {
		return fmt.Errorf("invalid hustle URI: %v", err)
	}

	if u.Scheme != "hustle" {
		return fmt.Errorf("invalid scheme: %s (expected hustle)", u.Scheme)
	}

	module := strings.ToLower(u.Host)
	params := u.Query()

	fmt.Printf("Protocol routing to module: %s with params: %v\n", module, params)

	handler, ok := p.Handlers[module]
	if !ok {
		return fmt.Errorf("unknown module in protocol: %s", module)
	}

	return handler(params)
}
