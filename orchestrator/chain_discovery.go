package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// ValidateChainSteps rejects steps that are not hustle:// URIs for
// registered modules. The discovery prompt once templated the literal
// placeholders hustle://step1 / hustle://step2, and the LLM parroted them
// into persisted chains that could only ever fail at execution time.
func ValidateChainSteps(steps []string, p *HustleProtocol) error {
	if len(steps) == 0 {
		return fmt.Errorf("chain has no steps")
	}
	for _, s := range steps {
		u, err := url.Parse(s)
		if err != nil || u.Scheme != "hustle" || u.Host == "" {
			return fmt.Errorf("invalid chain step %q: not a hustle:// URI", s)
		}
		if p != nil && !p.Has(u.Host) {
			return fmt.Errorf("invalid chain step %q: unknown module %q", s, u.Host)
		}
		// LLMs also invent parameter values, not just modules
		// (observed live: social?platform=YouTube+Shorts, which can
		// never post — Shorts belong to the media module).
		if strings.EqualFold(u.Host, "social") {
			if platform := u.Query().Get("platform"); platform != "" {
				switch strings.ToLower(platform) {
				case "twitter", "linkedin", "bluesky":
				default:
					return fmt.Errorf("invalid chain step %q: unsupported social platform %q", s, platform)
				}
			}
		}
		// Closed action sets for the other handlers a chain can drive.
		// Free-text modules (research, content, curation...) accept
		// anything and are deliberately not policed here.
		if action := u.Query().Get("action"); action != "" {
			var allowed []string
			switch strings.ToLower(u.Host) {
			case "calendar":
				allowed = []string{"status", "process", "add", "schedule_week", "schedule"}
			case "publisher":
				allowed = []string{"status", "wordpress", "wordpress_markdown", "newsletter"}
			case "affiliate":
				allowed = []string{"run", "batch_insert"}
			case "swarm":
				allowed = []string{"sync", "aggregate", "set_goal"}
			case "media":
				allowed = []string{"shorts"}
			}
			if allowed != nil && !containsFold(allowed, action) {
				return fmt.Errorf("invalid chain step %q: unsupported %s action %q", s, u.Host, action)
			}
		}
	}
	return nil
}

func containsFold(list []string, v string) bool {
	for _, item := range list {
		if strings.EqualFold(item, v) {
			return true
		}
	}
	return false
}

// DiscoveredChain includes scheduling information
type DiscoveredChain struct {
	Chain
	IntervalMinutes int `json:"interval_minutes"`
}

// ChainDiscoverer utilizes the LLM to evolve system workflows based on data
type ChainDiscoverer struct {
	Orchestrator *Orchestrator
	Manager      *ChainManager
}

func NewChainDiscoverer(orch *Orchestrator, mgr *ChainManager) *ChainDiscoverer {
	return &ChainDiscoverer{
		Orchestrator: orch,
		Manager:      mgr,
	}
}

// Discover analyzes the ledger and memory to propose and register a new workflow
func (d *ChainDiscoverer) Discover() (*DiscoveredChain, error) {
	fmt.Println("[ChainDiscoverer] Analyzing system state for workflow evolution...")

	// 1. Get Financial Context
	profitAnalysis := d.Orchestrator.Ledger.AnalyzeProfitability()

	// 2. Get Recent High-Value Memories
	recentSuccesses := d.Orchestrator.L1.Search("success")
	successContext := ""
	for _, e := range recentSuccesses {
		successContext += fmt.Sprintf("- %s (Tags: %v)\n", e.Content, e.Tags)
	}

	// 3. Construct Prompt for LLM
	prompt := fmt.Sprintf("As the Swarm Evolution Engine, analyze the following system state and propose a NEW sequential workflow (Chain) of protocol URIs. Focus on LUXURY HUSTLES: high-ROI, low-maintenance, and passive income generation.\nFinancial Context: %s\nRecent Successes:\n%s\n\nExisting Protocol Modules (use ONLY these, with real query parameters): research (query), curation (topic), social (platform: Twitter|LinkedIn|Bluesky ONLY, topic), trading (symbol), stocks (symbol), content (topic, type, action), affiliate (action, niche), products (type, topic), ecommerce (action, niche), saas (action, niche), domains (action, domain), kdp (action, niche), pod (action, niche), retail (buy, sell, category), realestate (rent, rate, location), media (action, topic), calendar (action, topic, platform), publisher (action), leadgen (niche), outreach (action, platform, niche), crm (action, email, status), confluence (recipe), council (topic), automation (problem), bi (action), finance (action), legal (type, business), design (name, niche), devagency (action), marketing (action), localization (lang, text), support (action), careers (action, title), qa (action), chain (name), swarm (action), synergy_leadgen (topic, niche), healer (issue)\n\nOutput a JSON object for the NEW Chain including a suggested execution interval in minutes. Every step MUST be a real hustle:// URI for one of the modules above — never invent placeholder steps:\n{\n  \"name\": \"short_snake_case_name\",\n  \"description\": \"brief luxury-focused description\",\n  \"steps\": [\"hustle://research?query=trending+AI+products\", \"hustle://affiliate?action=run&niche=AI+products\", \"hustle://social?platform=Twitter&topic=AI+products\"],\n  \"interval_minutes\": 60\n}", profitAnalysis, successContext)

	response, err := d.Orchestrator.LLM.Generate(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate discovery: %v", err)
	}

	// 4. Parse JSON from response. Reasoning-style models sometimes
	// narrate ("The user wants me to act as...") with no JSON at all;
	// retry once through GenerateJSON, which re-asks for strict JSON.
	discovered, err := parseDiscoveredChain(response)
	if err == nil && len(discovered.Steps) == 0 {
		// Well-formed JSON with no steps (observed live: the model
		// returned a named chain with an empty workflow) is just as
		// unusable — give it the same strict-JSON retry.
		err = fmt.Errorf("chain %q has no steps", discovered.Name)
	}
	if err != nil {
		fmt.Printf("[ChainDiscoverer] First response unusable (%v); retrying with strict JSON\n", err)
		var retry DiscoveredChain
		if jerr := d.Orchestrator.LLM.GenerateJSON(prompt, &retry); jerr != nil {
			return nil, fmt.Errorf("invalid LLM response format: %s", response)
		}
		discovered = &retry
	}

	// 5. Validate steps against the real protocol before persisting:
	// a chain that can only fail at execution time must never register.
	if err := ValidateChainSteps(discovered.Steps, d.Manager.Protocol); err != nil {
		return nil, fmt.Errorf("discovered chain %q rejected: %w", discovered.Name, err)
	}
	// Steps referencing other chains by name must point at chains that
	// already exist (this needs the manager, so it lives here rather
	// than in ValidateChainSteps).
	for _, step := range discovered.Steps {
		if u, err := url.Parse(step); err == nil && strings.EqualFold(u.Host, "chain") {
			if name := u.Query().Get("name"); name != "" && !d.Manager.Has(name) {
				return nil, fmt.Errorf("discovered chain %q rejected: step %q references unknown chain %q", discovered.Name, step, name)
			}
		}
	}

	// 6. Register the discovery (skip if already exists)
	if d.Manager.Has(discovered.Name) {
		fmt.Printf("[ChainDiscoverer] Chain '%s' already exists, skipping duplicate\n", discovered.Name)
		return nil, fmt.Errorf("chain '%s' already exists", discovered.Name)
	}
	d.Manager.Register(&discovered.Chain)

	// Log to L2 Vault
	d.Orchestrator.L2.Add(MemoryEntry{
		ID:        fmt.Sprintf("chain-disc-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Discovered and registered NEW LUXURY workflow: %s (%s) with interval %d min", discovered.Name, discovered.Description, discovered.IntervalMinutes),
		Timestamp: time.Now(),
		Tags:      []string{"discovery", "chain", "evolution", "luxury"},
	})

	return discovered, nil
}

// parseDiscoveredChain extracts the first {...} object from a raw LLM
// response and unmarshals it.
func parseDiscoveredChain(response string) (*DiscoveredChain, error) {
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON object in response")
	}
	var discovered DiscoveredChain
	if err := json.Unmarshal([]byte(response[start:end+1]), &discovered); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return &discovered, nil
}
