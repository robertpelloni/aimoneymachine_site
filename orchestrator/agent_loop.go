package orchestrator

import (
	"fmt"
	"strings"
	"time"
)

// AgentLoopState tracks the state of an autonomous agent loop
type AgentLoopState struct {
	AgentID    string    `json:"agent_id"`
	HustleType string    `json:"hustle_type"`
	Iterations int       `json:"iterations"`
	MaxIter    int       `json:"max_iter"`
	LastAction string    `json:"last_action"`
	LastResult string    `json:"last_result"`
	StartTime  time.Time `json:"start_time"`
	Status     string    `json:"status"` // running, paused, completed, failed
	Errors     int       `json:"errors"`
	Successes  int       `json:"successes"`
}

// AgentLoop runs a continuous LLM-driven decision loop for a hustle type.
// The LLM acts as the "brain" — it decides what to do next based on memory context,
// executes via the protocol, observes the result, and plans the next step.
type AgentLoop struct {
	Orch    *Orchestrator
	Proto   *HustleProtocol
	Broker  *A2ABroker
	State   AgentLoopState
	Verbose bool
}

// NewAgentLoop creates a new autonomous agent loop
func NewAgentLoop(orch *Orchestrator, proto *HustleProtocol, broker *A2ABroker, hustleType string) *AgentLoop {
	return &AgentLoop{
		Orch:   orch,
		Proto:  proto,
		Broker: broker,
		State: AgentLoopState{
			AgentID:    fmt.Sprintf("agent-%s-%d", hustleType, time.Now().Unix()),
			HustleType: hustleType,
			MaxIter:    100, // reasonable default
			Status:     "initialized",
			StartTime:  time.Now(),
		},
	}
}

// Run executes the agent loop continuously until max iterations or a stop signal
func (a *AgentLoop) Run() error {
	a.State.Status = "running"
	fmt.Printf("[AgentLoop] 🚀 Starting autonomous agent for: %s (max %d iterations)\n", a.State.HustleType, a.State.MaxIter)

	for a.State.Iterations < a.State.MaxIter {
		a.State.Iterations++
		fmt.Printf("\n[AgentLoop] ━━━ Iteration %d/%d ━━━\n", a.State.Iterations, a.State.MaxIter)

		// 1. OBSERVE: Gather context from memory
		context := a.observe()

		// 2. THINK: Ask the LLM what to do next
		action, err := a.think(context)
		if err != nil {
			a.State.Errors++
			fmt.Printf("[AgentLoop] ❌ Thinking failed: %v\n", err)
			// Fall back to heuristic action
			action = a.defaultAction()
			fmt.Printf("[AgentLoop] Falling back to heuristic: %s\n", action)
		}

		// 3. ACT: Execute the decided action via protocol
		start := time.Now()
		result := a.act(action)
		duration := time.Since(start)

		status := "success"
		msg := ""
		if strings.HasPrefix(result, "ERROR") {
			status = "failed"
			msg = result
		}

		if a.Orch.DB != nil {
			a.Orch.DB.LogTaskExecution(action, duration, status, msg)
		}

		// 4. LEARN: Store the outcome in memory
		a.learn(action, result)

		// 5. EVALUATE: Check if we should continue
		if !a.evaluate() {
			fmt.Printf("[AgentLoop] ⏹ Agent loop terminating after %d iterations\n", a.State.Iterations)
			break
		}

		// Small pause to avoid hammering the LLM
		time.Sleep(2 * time.Second)
	}

	if a.State.Iterations >= a.State.MaxIter {
		a.State.Status = "completed"
		fmt.Printf("[AgentLoop] ✅ Completed %d iterations. Successes: %d, Errors: %d\n",
			a.State.Iterations, a.State.Successes, a.State.Errors)
	} else {
		a.State.Status = "stopped"
	}

	return nil
}

// observe gathers context from memory for the LLM to reason about
func (a *AgentLoop) observe() string {
	var sb strings.Builder

	// Recent L1 events
	recent := a.Orch.L1.Search(a.State.HustleType)
	if len(recent) > 5 {
		recent = recent[len(recent)-5:]
	}
	sb.WriteString("RECENT ACTIVITY:\n")
	for _, e := range recent {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", e.Content, e.Timestamp.Format("15:04:05")))
	}

	// Financial context
	sb.WriteString(fmt.Sprintf("\nFINANCIAL: Revenue $%.2f, Expenses $%.2f, Profit $%.2f\n",
		a.Orch.Ledger.TotalRevenue(), a.Orch.Ledger.TotalExpenses(), a.Orch.Ledger.Profit()))

	// L2 successes
	successes := a.Orch.L2.Search(a.State.HustleType)
	if len(successes) > 3 {
		successes = successes[len(successes)-3:]
	}
	if len(successes) > 0 {
		sb.WriteString("PAST SUCCESSES:\n")
		for _, e := range successes {
			sb.WriteString(fmt.Sprintf("- %s\n", e.Content))
		}
	}

	// Previous iteration
	if a.State.LastAction != "" {
		sb.WriteString(fmt.Sprintf("\nPREVIOUS ACTION: %s\nPREVIOUS RESULT: %s\n", a.State.LastAction, a.State.LastResult))
	}

	return sb.String()
}

// think asks the LLM to decide the next action based on context
func (a *AgentLoop) think(context string) (string, error) {
	prompt := fmt.Sprintf(`You are an autonomous AI hustle agent managing: %s

Your job is to decide the NEXT ACTION to take. You must respond with a SINGLE hustle:// URI.

Available modules and their parameters:
- hustle://research?query=TOPIC — Search the web for intelligence on a topic
- hustle://curation?topic=TOPIC — Curate and summarize content on a topic
- hustle://social?platform=PLATFORM&topic=TOPIC — Generate and post content (platform: Twitter, LinkedIn, Bluesky)
- hustle://trading?symbol=SYMBOL — Execute trading strategy for a crypto symbol
- hustle://content?topic=TOPIC&type=TYPE — Generate monetizable content (type: blog, newsletter, seo, thread, tool)
- hustle://content?topic=TOPIC&type=TYPE&publish=true — Generate AND publish content to WordPress/Newsletter
- hustle://leadgen?topic=TOPIC — Discover business leads for a topic
- hustle://calendar?action=process — Process and publish scheduled content
- hustle://chain?name=CHAIN_NAME — Execute a multi-step workflow chain
- hustle://chain?action=discover — Discover and create new high-ROI workflow chains
- hustle://healer?issue=DESCRIPTION — Diagnose and fix a problem
- hustle://ecommerce?action=ads&platform=TikTok — Generate social ads for products
- hustle://ecommerce?action=fulfill — Fulfill pending dropshipping orders
- hustle://saas?action=ideate — Generate new micro-SaaS MVPs
- hustle://automation?action=build — Architect business process workflows
- hustle://finance?action=classify — Automate client bookkeeping
- hustle://bi?action=report — Generate strategic business intelligence
- hustle://legal?action=generate — Generate compliance documents
- hustle://crm?action=update — Manage lead lifecycles
- hustle://pod?action=plan — Plan Print-on-Demand designs
- hustle://media?action=plan — Plan multimedia video production
- hustle://careers?action=search — Discover job/contract opportunities
- hustle://kdp?action=plan — Plan book interiors for Amazon KDP
- hustle://retail?buy=PRICE&sell=PRICE — Scan Amazon FBA arbitrage ROI
- hustle://domains?name=DOMAIN — Valuate domain names for flipping
- hustle://realestate_manage — Manage guest comms for STR/Airbnb
- hustle://stocks?action=analyze&symbol=TICKER — Quantitative stock analysis
- hustle://swarm?action=sync — Synchronize memory with mesh peers
- hustle://swarm?action=aggregate — Aggregate mesh-wide status

CURRENT CONTEXT:
%s

Iteration: %d | Successes: %d | Errors: %d

STRATEGY GUIDELINES:
1. Content generation has the HIGHEST ROI. Use hustle://content?type=blog&publish=true to maximize search traffic.
2. High-ROI Chain: research -> leadgen -> outreach -> content?type=newsletter&publish=true -> social.
3. Use hustle://calendar?action=process daily to maintain a consistent publishing schedule across platforms.
4. Research trending topics, discover leads, prepare outreach pitches, write about them in a newsletter, and post to social media to grow your audience.
4. If errors are high, run healer. If profitable, double down on what works.
5. Vary your actions — do NOT repeat the same action twice in a row.
6. Use specific, targeted queries rather than generic ones.
7. Mix content types: blog posts for SEO, newsletters for subscribers, threads for viral reach, and utility tools for utility SEO traffic.

Respond with ONLY the hustle:// URI, nothing else. Example: hustle://content?topic=AI+agent+market+trends+2026&type=blog`,
		a.State.HustleType, context, a.State.Iterations, a.State.Successes, a.State.Errors)

	response, err := a.Orch.LLM.Generate(prompt)
	if err != nil {
		return "", fmt.Errorf("LLM thinking failed: %v", err)
	}

	// Extract the URI from the response
	uri := strings.TrimSpace(response)
	// Handle cases where LLM adds extra text
	if !strings.HasPrefix(uri, "hustle://") {
		start := strings.Index(uri, "hustle://")
		if start >= 0 {
			uri = uri[start:]
		} else {
			return "", fmt.Errorf("LLM did not produce a valid hustle:// URI: %s", uri)
		}
	}
	// Trim anything after the URI (newlines, extra text)
	if nlIdx := strings.Index(uri, "\n"); nlIdx >= 0 {
		uri = uri[:nlIdx]
	}
	if spaceIdx := strings.Index(uri, " "); spaceIdx >= 0 {
		uri = uri[:spaceIdx]
	}

	return strings.TrimSpace(uri), nil
}

// act executes the decided action via the protocol
func (a *AgentLoop) act(uri string) string {
	fmt.Printf("[AgentLoop] 🎯 Executing: %s\n", uri)
	err := a.Proto.HandleURI(uri)
	if err != nil {
		a.State.Errors++
		a.State.LastResult = fmt.Sprintf("ERROR: %v", err)
		fmt.Printf("[AgentLoop] ❌ Action failed: %v\n", err)
		return a.State.LastResult
	}

	a.State.Successes++
	a.State.LastAction = uri
	a.State.LastResult = "SUCCESS"
	fmt.Printf("[AgentLoop] ✅ Action succeeded\n")
	return a.State.LastResult
}

// learn stores the outcome of this iteration in memory
func (a *AgentLoop) learn(action, result string) {
	tags := []string{"agent_loop", a.State.HustleType}
	if result == "SUCCESS" {
		tags = append(tags, "success")
	} else {
		tags = append(tags, "failure")
	}

	entry := MemoryEntry{
		ID:        fmt.Sprintf("agent-%s-iter-%d", a.State.HustleType, a.State.Iterations),
		Content:   fmt.Sprintf("Agent iteration %d: Action=%s Result=%s", a.State.Iterations, action, result),
		BaseScore: 50.0,
		Timestamp: time.Now(),
		Tags:      tags,
	}

	a.Orch.L1.Add(entry)

	// Promote successes to L2 vault
	if result == "SUCCESS" && a.State.Successes%5 == 0 {
		a.Orch.L2.Add(MemoryEntry{
			ID:        fmt.Sprintf("agent-success-%d", time.Now().Unix()),
			Content:   fmt.Sprintf("Agent %s completed %d successful iterations. Last action: %s", a.State.HustleType, a.State.Successes, action),
			BaseScore: 80.0,
			Timestamp: time.Now(),
			Tags:      []string{"agent_loop", a.State.HustleType, "milestone"},
		})
	}
}

// evaluate decides whether to continue the loop
func (a *AgentLoop) evaluate() bool {
	// Stop if too many consecutive errors
	if a.State.Errors > 5 && a.State.Successes == 0 {
		fmt.Printf("[AgentLoop] ⛔ Too many errors (%d) with no successes. Stopping.\n", a.State.Errors)
		a.State.Status = "failed"
		return false
	}

	// Stop if profit is very negative
	if a.Orch.Ledger.Profit() < -1000 {
		fmt.Printf("[AgentLoop] ⛔ Wealth preservation triggered. Profit: $%.2f. Stopping.\n", a.Orch.Ledger.Profit())
		a.State.Status = "wealth_preservation"
		return false
	}

	return true
}

// defaultAction returns a heuristic action when LLM thinking fails
func (a *AgentLoop) defaultAction() string {
	actions := []string{
		"hustle://content?topic=AI+automation+trends+2026&type=blog",
		"hustle://research?query=AI+automation+trends+2026",
		"hustle://curation?topic=AI+agents",
		"hustle://trading?symbol=BTC",
		"hustle://content?topic=crypto+market+analysis&type=newsletter",
		"hustle://social?platform=Twitter&topic=AI+agents",
		"hustle://chain?action=discover",
		"hustle://swarm?action=sync",
	}
	// Pick based on iteration to vary actions
	idx := a.State.Iterations % len(actions)
	return actions[idx]
}

// AgentLoopStatus returns the current state for dashboard/API
func (a *AgentLoop) AgentLoopStatus() AgentLoopState {
	return a.State
}

// MultiAgentOrchestrator manages multiple concurrent agent loops
type MultiAgentOrchestrator struct {
	Orch    *Orchestrator
	Proto   *HustleProtocol
	Broker  *A2ABroker
	Agents  map[string]*AgentLoop
	Verbose bool
}

// NewMultiAgentOrchestrator creates a manager for multiple concurrent agents
func NewMultiAgentOrchestrator(orch *Orchestrator, proto *HustleProtocol, broker *A2ABroker) *MultiAgentOrchestrator {
	return &MultiAgentOrchestrator{
		Orch:   orch,
		Proto:  proto,
		Broker: broker,
		Agents: make(map[string]*AgentLoop),
	}
}

// AddAgent creates and registers a new agent loop
func (m *MultiAgentOrchestrator) AddAgent(hustleType string, maxIter int) *AgentLoop {
	agent := NewAgentLoop(m.Orch, m.Proto, m.Broker, hustleType)
	agent.State.MaxIter = maxIter
	m.Agents[agent.State.AgentID] = agent
	fmt.Printf("[MultiAgent] Registered agent: %s for hustle type: %s\n", agent.State.AgentID, hustleType)
	return agent
}

// RunAll starts all registered agents concurrently
func (m *MultiAgentOrchestrator) RunAll() {
	fmt.Printf("[MultiAgent] 🚀 Launching %d agents concurrently\n", len(m.Agents))

	for id, agent := range m.Agents {
		go func(a *AgentLoop, agentID string) {
			if err := a.Run(); err != nil {
				fmt.Printf("[MultiAgent] Agent %s crashed: %v\n", agentID, err)
			}
		}(agent, id)
	}

	// Monitor loop
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		allDone := true
		for id, agent := range m.Agents {
			status := agent.State.Status
			if status == "running" {
				allDone = false
			}
			fmt.Printf("[MultiAgent] %s: iter=%d success=%d errors=%d status=%s\n",
				id, agent.State.Iterations, agent.State.Successes, agent.State.Errors, status)
		}
		if allDone {
			fmt.Println("[MultiAgent] All agents finished.")
			break
		}
	}
}

// StatusReport returns a summary of all agents
func (m *MultiAgentOrchestrator) StatusReport() []AgentLoopState {
	var states []AgentLoopState
	for _, agent := range m.Agents {
		states = append(states, agent.State)
	}
	return states
}

// HustlePlan represents an LLM-generated plan for a hustle
type HustlePlan struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Steps       []string `json:"steps"`
	IntervalMin int      `json:"interval_min"`
	Priority    string   `json:"priority"` // high, medium, low
}

// PlanHustles asks the LLM to analyze current state and generate a set of actionable hustle plans
func PlanHustles(orch *Orchestrator) ([]HustlePlan, error) {
	profitAnalysis := orch.Ledger.AnalyzeProfitability()

	// High-signal inputs: Discovered Alpha and Leads
	alphaEntries := orch.L2.Search("alpha")
	leadEntries := orch.L2.Search("leadgen")

	var signalContext string
	if len(alphaEntries) > 0 {
		signalContext += "\nDISCOVERED ALPHA (HIGH-SIGNAL):\n"
		for _, e := range alphaEntries {
			signalContext += fmt.Sprintf("- %s\n", e.Content)
		}
	}
	if len(leadEntries) > 0 {
		signalContext += "\nDISCOVERED LEADS (ACQUISITION TARGETS):\n"
		for _, e := range leadEntries {
			signalContext += fmt.Sprintf("- %s\n", e.Content)
		}
	}

	recentActivity := orch.L1.Search("")
	if len(recentActivity) > 10 {
		recentActivity = recentActivity[len(recentActivity)-10:]
	}

	var activityStr string
	for _, e := range recentActivity {
		activityStr += fmt.Sprintf("- %s\n", e.Content)
	}

	prompt := fmt.Sprintf(`You are a strategic AI hustle planner. Analyze the system state and generate EXACTLY 5 actionable hustle plans.

FINANCIAL STATE: %s
RECENT ACTIVITY:
%s
%s

Generate plans from these categories:
1. Research — Web intelligence gathering on trending topics
2. LeadGen — Automated lead discovery for business opportunities
3. Curation — Content aggregation and newsletter generation
4. Social — Automated content posting for audience growth
5. Trading — Crypto and Stock trading with real execution
6. Content — Blog/SEO/Newsletter/Script/Tool generation
7. Ecommerce — Dropshipping, Ads, and Fulfillment
8. Services — DevAgency, Support, Automation, Finance, Legal, CRM
9. Physical/RealEstate — POD, KDP, Retail Arb, STR Management, Domains

Respond with a JSON array of exactly 5 objects:
[
  {
    "name": "short_name",
    "description": "what this hustle does and why it will be profitable",
    "category": "Research|LeadGen|Curation|Social|Trading|Content",
    "steps": ["hustle://step1", "hustle://step2"],
    "interval_min": 60,
    "priority": "high|medium|low"
  }
]

Include "publish=true" in content steps where appropriate to ensure monetization.
PRIORITIZE: Use the DISCOVERED ALPHA or LEADS to seed content topics and keywords. If you found a lead, generate content targeting their niche and prepare an outreach pitch.
Be specific. Use real trending topics. Focus on LUXURY (high-ROI, low-maintenance).`, profitAnalysis, activityStr, signalContext)

	var plans []HustlePlan
	if err := orch.LLM.GenerateJSON(prompt, &plans); err != nil {
		return nil, fmt.Errorf("failed to generate hustle plans: %v", err)
	}

	return plans, nil
}
