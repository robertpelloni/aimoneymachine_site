package research

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

// Lead represents a potential business lead
type Lead struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Needs       string `json:"needs"`
	ContactURL  string `json:"contact_url"`
	Email       string `json:"email,omitempty"`
}

// DiscoverLeads uses the LLM to identify potential business leads from research
func DiscoverLeads(orch *orchestrator.Orchestrator, topic string) ([]Lead, error) {
	fmt.Printf("[LeadGen] Hunting for leads related to: %s\n", topic)

	// 1. Perform research to get context
	searcher := NewResearchSearch(Tavily, orch, nil)
	results, err := searcher.Query(topic + " companies needs automation")
	if err != nil {
		return nil, fmt.Errorf("research failed: %v", err)
	}

	// 2. Build prompt for lead extraction
	var context string
	for _, r := range results {
		context += fmt.Sprintf("- %s: %s (%s)\n", r.Title, r.Snippet, r.URL)
	}

	prompt := fmt.Sprintf(`Analyze the following research results and identify 3 SPECIFIC companies or niches that would benefit from AI automation services.

RESEARCH CONTEXT:
%s

Respond with a JSON array of 3 objects:
[
  {
    "name": "Company or Niche Name",
    "description": "Short description of what they do",
    "needs": "Specific AI/automation need you identified",
    "contact_url": "URL to their contact page or relevant info",
    "email": "Email address if found in context, otherwise leave empty"
  }
]

Respond ONLY with valid JSON.`, context)

	var leads []Lead
	if err := orch.LLM.GenerateJSON(prompt, &leads); err != nil {
		return nil, fmt.Errorf("lead extraction failed: %v", err)
	}

	// 3. Record in memory
	for _, lead := range leads {
		orch.L2.Add(orchestrator.MemoryEntry{
			ID:        fmt.Sprintf("lead-%d", time.Now().UnixNano()),
			Content:   fmt.Sprintf("DISCOVERED LEAD: %s - %s. Needs: %s. Contact: %s. Email: %s", lead.Name, lead.Description, lead.Needs, lead.ContactURL, lead.Email),
			BaseScore: 75.0,
			Timestamp: time.Now(),
			Tags:      []string{"leadgen", topic, lead.Name},
		})
	}

	return leads, nil
}
