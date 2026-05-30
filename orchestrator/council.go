package orchestrator

import (
	"fmt"
	"strings"
)

type AgentRole string

const (
	Bull   AgentRole = "Bull"
	Bear   AgentRole = "Bear"
	Critic AgentRole = "Critic"
)

type Council struct {
	Agents map[AgentRole]LLMProvider
}

func NewCouncil(orch *Orchestrator) *Council {
	return &Council{
		Agents: map[AgentRole]LLMProvider{
			Bull:   &MockLLM{Response: "This is an incredible opportunity with high ROI potential!"},
			Bear:   &MockLLM{Response: "The risks are too high and the market is saturated."},
			Critic: &MockLLM{Response: "A balanced approach is needed. Focus on the niche segments identified."},
		},
	}
}

type DebateResult struct {
	Consensus string
	Points    []string
}

func (c *Council) Debate(topic string) (DebateResult, error) {
	fmt.Printf("--- COUNCIL DEBATE: %s ---\n", topic)

	bullOpinion, _ := c.Agents[Bull].Generate("Argue FOR: " + topic)
	bearOpinion, _ := c.Agents[Bear].Generate("Argue AGAINST: " + topic)
	criticSummary, _ := c.Agents[Critic].Generate(fmt.Sprintf("Synthesize these opinions on %s:\nBull: %s\nBear: %s", topic, bullOpinion, bearOpinion))

	fmt.Printf("Bull says: %s\n", bullOpinion)
	fmt.Printf("Bear says: %s\n", bearOpinion)
	fmt.Printf("Critic concludes: %s\n", criticSummary)

	return DebateResult{
		Consensus: criticSummary,
		Points:    []string{bullOpinion, bearOpinion},
	}, nil
}

func (r DebateResult) String() string {
	return fmt.Sprintf("Consensus: %s\nKey points: %s", r.Consensus, strings.Join(r.Points, " | "))
}
