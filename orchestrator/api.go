package orchestrator

import (
	"crypto/subtle"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// dashboardHTML embeds the Consolidated Single-Page Command Center directly into the
// Go binary so :8082 serves the entire frontend with zero external file dependencies.
//
//go:embed dashboard.html
var dashboardHTML []byte

// API exposes the Orchestrator, Protocol Router, Mesh Broker, Workflow Chains,
// SQLite Store, Memory Tiers, and Companion Service Logs over HTTP.
//
// Architectural note: TradingSnapshot, CalendarSnapshot, and SchedulerSnapshot are
// injected as function callbacks from main.go rather than importing hustle/* packages
// here, preventing circular module dependencies (hustle/* already imports orchestrator).
type API struct {
	Orchestrator      *Orchestrator
	Protocol          *HustleProtocol
	Broker            *A2ABroker
	ChainManager      *ChainManager
	Discoverer        *ChainDiscoverer
	MultiAgent        *MultiAgentOrchestrator
	TradingSnapshot   func() interface{}
	CalendarSnapshot  func() interface{}
	SchedulerSnapshot func() interface{}
	PriceSource       string
	// Token, when non-empty, guards all state-changing POST endpoints via
	// Authorization: Bearer <token> or X-API-Key (constant-time comparison).
	Token string
}

func NewAPI(orch *Orchestrator, protocol *HustleProtocol, broker *A2ABroker, chainMgr *ChainManager, disc *ChainDiscoverer) *API {
	return &API{
		Orchestrator: orch,
		Protocol:     protocol,
		Broker:       broker,
		ChainManager: chainMgr,
		Discoverer:   disc,
		Token:        os.Getenv("AIMM_API_TOKEN"),
	}
}

func (a *API) requireToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.Token != "" {
			presented := r.Header.Get("X-API-Key")
			if presented == "" {
				if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
					presented = strings.TrimPrefix(auth, "Bearer ")
				}
			}
			if subtle.ConstantTimeCompare([]byte(presented), []byte(a.Token)) != 1 {
				http.Error(w, "unauthorized: valid API token required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	}
}

// RegisterRoutes attaches all API and single-page dashboard endpoints to mux.
func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/dispatch", a.requireToken(a.handleDispatch))
	mux.HandleFunc("/status", a.handleStatus)
	mux.HandleFunc("/message", a.requireToken(a.handleMessage))
	mux.HandleFunc("/register", a.requireToken(a.handleRegister))
	mux.HandleFunc("/sync", a.requireToken(a.handleSync))
	mux.HandleFunc("/chains", a.handleChains)
	mux.HandleFunc("/agents", a.handleAgents)
	mux.HandleFunc("/content", a.handleContent)
	mux.HandleFunc("/strategy", a.handleStrategy)
	mux.HandleFunc("/healer", a.requireToken(a.handleHealer))

	// Consolidated single-page dashboard + legacy /bot redirect
	mux.HandleFunc("/", a.handleDashboard)
	mux.HandleFunc("/dashboard", a.handleDashboard)
	mux.HandleFunc("/bot", a.handleBotRedirect)

	// /api/* surface used by the Consolidated Dashboard
	mux.HandleFunc("/api/dispatch", a.requireToken(a.handleDispatch))
	mux.HandleFunc("/api/status", a.handleStatus)
	mux.HandleFunc("/api/chains", a.handleChains)
	mux.HandleFunc("/api/agents", a.handleAgents)
	mux.HandleFunc("/api/content", a.handleContent)
	mux.HandleFunc("/api/strategy", a.handleStrategy)
	mux.HandleFunc("/api/log", a.handleLog)
	mux.HandleFunc("/api/service-logs", a.handleServiceLogs)
	mux.HandleFunc("/api/progress", a.handleProgress)
	mux.HandleFunc("/api/ledger", a.handleLedger)
	mux.HandleFunc("/api/tasks", a.handleTasks)
	mux.HandleFunc("/api/memories", a.handleMemories)
	mux.HandleFunc("/api/outputs", a.handleOutputs)
	mux.HandleFunc("/api/hymns", a.handleHymns)
}

func (a *API) Start(port string) error {
	mux := http.NewServeMux()
	a.RegisterRoutes(mux)
	fmt.Printf("[API] Orchestrator listening on port %s (dashboard: /)\n", port)
	return http.ListenAndServe(":"+port, mux)
}

func (a *API) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/dashboard" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(dashboardHTML)
}

// handleBotRedirect folds historical /bot?sym=BTC links into the single-page
// dashboard's inline Bot Inspector so zero subpages exist.
func (a *API) handleBotRedirect(w http.ResponseWriter, r *http.Request) {
	sym := r.URL.Query().Get("sym")
	if sym == "" {
		sym = r.URL.Query().Get("symbol")
	}
	if sym != "" {
		http.Redirect(w, r, "/?bot="+sym+"#trading", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/#trading", http.StatusFound)
}

func firstExisting(names ...string) string {
	for _, n := range names {
		if n == "" {
			continue
		}
		if _, err := os.Stat(n); err == nil {
			return n
		}
	}
	return ""
}

func tailString(s string, maxBytes int) string {
	if len(s) <= maxBytes {
		return s
	}
	cut := s[len(s)-maxBytes:]
	if idx := strings.IndexByte(cut, '\n'); idx >= 0 && idx+1 < len(cut) {
		return cut[idx+1:]
	}
	return cut
}

// handleLog serves live orchestrator output from three layered sources:
//  1. GlobalLogBuffer (in-process stdout ring buffer — works identically on Windows & Linux systemd)
//  2. orchestrator_live.log / orchestrator.log on disk if present
//  3. journalctl -u aimm-orchestrator fallback on Linux when the ring buffer just booted
func (a *API) handleLog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	ring := GlobalLogBuffer.String()
	if len(strings.TrimSpace(ring)) > 120 {
		_, _ = w.Write([]byte(tailString(ring, 48000)))
		return
	}

	if p := firstExisting("orchestrator_live.log", "orchestrator.log", "../orchestrator_live.log", "/opt/aimoneymachine/orchestrator_live.log"); p != "" {
		if data, err := os.ReadFile(p); err == nil && len(data) > 0 {
			_, _ = w.Write([]byte(tailString(string(data), 48000)))
			return
		}
	}

	if runtime.GOOS == "linux" {
		if out, err := exec.Command("journalctl", "-u", "aimm-orchestrator", "-n", "220", "--no-pager").Output(); err == nil && len(out) > 0 {
			_, _ = w.Write(out)
			return
		}
	}

	_, _ = w.Write([]byte(ring))
}

// handleServiceLogs tails strictly whitelisted companion service logs so operators
// can inspect publisher, 100K expander, content audit, trader, freellm, and watchdog
// directly inside the Consolidated Dashboard without SSHing in.
func (a *API) handleServiceLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	svc := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("service")))
	if svc == "" || svc == "orchestrator" {
		a.handleLog(w, r)
		return
	}

	type logSpec struct {
		files []string
		unit  string
	}
	specs := map[string]logSpec{
		"content":    {files: []string{"expand.log", "/opt/aimoneymachine/expand.log", "/opt/aimoneymachine/expand2.log"}, unit: "aimm-content"},
		"publisher":  {files: []string{"publisher.log", "/opt/aimoneymachine/publisher.log"}, unit: "aimm-publisher"},
		"audit":      {files: []string{"audit.log", "/opt/aimoneymachine/audit.log"}, unit: "aimm-audit"},
		"trader":     {files: []string{"trader.log", "/opt/aimoneymachine/trader.log"}, unit: "aimm-python-trader"},
		"hustle-gen": {files: []string{"hustle_gen.log", "/opt/aimoneymachine/hustle_gen.log"}, unit: "aimm-hustle-gen"},
		"watchdog":   {files: []string{"watchdog.log", "/opt/aimoneymachine/watchdog.log"}, unit: "aimm-watchdog"},
		"freellm":    {files: []string{"freellm.log", "/opt/aimoneymachine/freellm.log"}, unit: "aimm-freellm"},
		"bluesky":    {files: []string{"bluesky.log", "/opt/aimoneymachine/bluesky.log"}, unit: "aimm-bluesky"},
		"twitter":    {files: []string{"twitter.log", "/opt/aimoneymachine/twitter.log"}, unit: "aimm-twitter"},
	}

	spec, ok := specs[svc]
	if !ok {
		http.Error(w, "unknown service name", http.StatusBadRequest)
		return
	}

	if p := firstExisting(spec.files...); p != "" {
		if data, err := os.ReadFile(p); err == nil && len(data) > 0 {
			_, _ = w.Write([]byte(tailString(string(data), 40000)))
			return
		}
	}
	if runtime.GOOS == "linux" && spec.unit != "" {
		if out, err := exec.Command("journalctl", "-u", spec.unit, "-n", "180", "--no-pager").Output(); err == nil && len(out) > 0 {
			_, _ = w.Write(out)
			return
		}
	}
	fmt.Fprintf(w, "[%s] No local log entries found on this host.", svc)
}

func (a *API) handleProgress(w http.ResponseWriter, r *http.Request) {
	p := firstExisting("expansion_progress.json", "../expansion_progress.json", "bin/expansion_progress.json", "/opt/aimoneymachine/bin/expansion_progress.json")
	w.Header().Set("Content-Type", "application/json")
	if p == "" {
		fmt.Fprint(w, `{"done":[]}`)
		return
	}
	data, err := os.ReadFile(p)
	if err != nil || len(data) == 0 {
		fmt.Fprint(w, `{"done":[]}`)
		return
	}
	_, _ = w.Write(data)
}

// handleDispatch executes any hustle:// protocol URI and appends any newly generated
// L1 memory entry to the response so interactive dashboard modules show immediate output,
// while preserving the "URI ... dispatched successfully" contract required by smoke_prod.py.
func (a *API) handleDispatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URI string `json:"uri"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	fmt.Printf("[API] Received dispatch request for URI: %s\n", req.URI)

	var beforeL1, beforeL2 int
	if a.Orchestrator != nil {
		beforeL1 = a.Orchestrator.L1.Len()
		beforeL2 = a.Orchestrator.L2.Len()
	}

	start := time.Now()
	err := a.Protocol.HandleURI(req.URI)
	dur := time.Since(start)

	if a.Orchestrator != nil && a.Orchestrator.DB != nil {
		st, msg := "success", ""
		if err != nil {
			st, msg = "failed", err.Error()
		}
		_ = a.Orchestrator.DB.LogTaskExecution(req.URI, dur, st, msg)
	}

	if err != nil {
		fmt.Printf("[API] Dispatch failed for URI %s: %v\n", req.URI, err)
		http.Error(w, fmt.Sprintf("Protocol error: %v", err), http.StatusInternalServerError)
		return
	}

	var latestSummary string
	if a.Orchestrator != nil {
		if snap := a.Orchestrator.L1.EntriesSnapshot(); len(snap) > beforeL1 {
			latestSummary = snap[len(snap)-1].Content
		} else if snap2 := a.Orchestrator.L2.EntriesSnapshot(); len(snap2) > beforeL2 {
			latestSummary = snap2[len(snap2)-1].Content
		}
	}

	if r.URL.Query().Get("format") == "json" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":          true,
			"message":     fmt.Sprintf("URI %s dispatched successfully", req.URI),
			"duration_ms": dur.Milliseconds(),
			"result":      latestSummary,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	if latestSummary != "" {
		fmt.Fprintf(w, "URI %s dispatched successfully\n%s", req.URI, latestSummary)
	} else {
		fmt.Fprintf(w, "URI %s dispatched successfully", req.URI)
	}
}

func (a *API) handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Invalid Message JSON", http.StatusBadRequest)
		return
	}

	if msg.Topic != "" {
		fmt.Printf("[API] Received incoming A2A Topic message: %s\n", msg.Topic)
		a.Broker.Publish(msg)
	} else if msg.Target != "" {
		fmt.Printf("[API] Received incoming A2A direct message for target: %s\n", msg.Target)
		if err := a.Broker.Route(msg); err != nil {
			http.Error(w, fmt.Sprintf("Routing error: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Message must have Target or Topic", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	a.Broker.RegisterPeer(req.ID, req.URL)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Peer %s registered successfully", req.ID)
}

func (a *API) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if a.Orchestrator == nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "Initializing"})
		return
	}

	peers := 0
	peerMap := map[string]string{}
	if a.Broker != nil {
		peers = len(a.Broker.Peers)
		for k, v := range a.Broker.Peers {
			peerMap[k] = v
		}
	}

	did := "did:hustle:local"
	if a.Orchestrator.Identity != nil {
		did = a.Orchestrator.Identity.GetDID()
	}

	feeds := a.Orchestrator.RSSFeeds
	if len(feeds) == 0 {
		feeds = []string{"https://news.ycombinator.com/rss"}
	}

	resp := map[string]interface{}{
		"status":             "Active",
		"version":            a.Orchestrator.Version,
		"profit":             a.Orchestrator.Ledger.Profit(),
		"revenue":            a.Orchestrator.Ledger.TotalRevenue(),
		"expenses":           a.Orchestrator.Ledger.TotalExpenses(),
		"wealth_goal":        a.Orchestrator.WealthGoal,
		"dry_run":            a.Orchestrator.DryRun,
		"did":                did,
		"memories":           a.Orchestrator.L1.Len(),
		"memories_l1":        a.Orchestrator.L1.Len(),
		"memories_l2":        a.Orchestrator.L2.Len(),
		"memories_l3":        a.Orchestrator.L3.Len(),
		"peers":              peers,
		"peer_map":           peerMap,
		"rss_feeds":          feeds,
		"task_queue":         a.Orchestrator.TaskQueue,
		"price_source":       a.PriceSource,
		"auth_required":      a.Token != "",
		"registered_modules": a.Protocol.RegisteredModules(),
		"integrations": map[string]bool{
			"wordpress":  os.Getenv("WORDPRESS_URL") != "" && os.Getenv("WORDPRESS_PASSWORD") != "",
			"newsletter": os.Getenv("BUTTONDOWN_API_KEY") != "" || os.Getenv("SUBSTACK_API_KEY") != "" || os.Getenv("CONVERTKIT_API_SECRET") != "",
			"shopify":    os.Getenv("SHOPIFY_ACCESS_TOKEN") != "" && os.Getenv("SHOPIFY_SHOP_NAME") != "",
			"gumroad":    os.Getenv("GUMROAD_ACCESS_TOKEN") != "",
			"bluesky":    os.Getenv("BLUESKY_HANDLE") != "" && os.Getenv("BLUESKY_APP_PASSWORD") != "",
			"twitter":    os.Getenv("TWITTER_API_KEY") != "" && os.Getenv("TWITTER_ACCESS_TOKEN") != "",
			"linkedin":   os.Getenv("LINKEDIN_ACCESS_TOKEN") != "" && os.Getenv("LINKEDIN_AUTHOR_URN") != "",
			"smtp":       os.Getenv("SMTP_HOST") != "",
			"alpaca":     os.Getenv("ALPACA_API_KEY") != "" && os.Getenv("ALPACA_API_SECRET") != "",
			"binance":    os.Getenv("BINANCE_API_KEY") != "",
			"coinbase":   os.Getenv("COINBASE_API_KEY") != "",
			"coingecko":  os.Getenv("COINGECKO_API_KEY") != "",
			"elevenlabs": os.Getenv("ELEVENLABS_API_KEY") != "",
			"youtube":    os.Getenv("YOUTUBE_CLIENT_ID") != "" && os.Getenv("YOUTUBE_REFRESH_TOKEN") != "",
			"llm":        os.Getenv("LLM_BASE_URL") != "",
		},
	}

	if a.Orchestrator.Cache != nil {
		resp["cache_stats"] = a.Orchestrator.Cache.Stats()
	}
	if a.Orchestrator.Optimizer != nil {
		resp["bandit_stats"] = a.Orchestrator.Optimizer.Snapshot()
	}
	if a.TradingSnapshot != nil {
		resp["trading_bots"] = a.TradingSnapshot()
	}
	if pct, freeGB, ok := DiskUsage("."); ok {
		resp["disk_used_pct"] = float64(int(pct*10)) / 10
		resp["disk_free_gb"] = float64(int(freeGB*10)) / 10
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (a *API) handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("[API] Remote trigger: Repository Synchronization Protocol")
	if err := a.Protocol.HandleURI("hustle://sync"); err != nil {
		http.Error(w, fmt.Sprintf("Sync failed: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Sync protocol initiated successfully")
}

func (a *API) handleChains(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if a.ChainManager == nil {
		http.Error(w, "Chain manager not initialized", http.StatusServiceUnavailable)
		return
	}
	_ = json.NewEncoder(w).Encode(a.ChainManager.Snapshot())
}

func (a *API) handleAgents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if a.MultiAgent != nil {
		states := a.MultiAgent.StatusReport()
		if len(states) > 0 {
			_ = json.NewEncoder(w).Encode(states)
			return
		}
	}
	if a.Orchestrator == nil {
		_ = json.NewEncoder(w).Encode([]MemoryEntry{})
		return
	}
	agentEntries := a.Orchestrator.L1.Search("agent_loop")
	if agentEntries == nil {
		agentEntries = []MemoryEntry{}
	}
	_ = json.NewEncoder(w).Encode(agentEntries)
}

func (a *API) handleContent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	files, err := os.ReadDir("output/content")
	if err != nil {
		_ = json.NewEncoder(w).Encode([]string{})
		return
	}
	fileList := make([]string, 0, len(files))
	for _, f := range files {
		if !f.IsDir() {
			fileList = append(fileList, f.Name())
		}
	}
	_ = json.NewEncoder(w).Encode(fileList)
}

func (a *API) handleStrategy(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if a.Orchestrator == nil {
		_ = json.NewEncoder(w).Encode([]MemoryEntry{})
		return
	}
	strategies := a.Orchestrator.L1.Search("collective_strategy")
	if strategies == nil {
		strategies = []MemoryEntry{}
	}
	_ = json.NewEncoder(w).Encode(strategies)
}

func (a *API) handleHealer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Issue string `json:"issue"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	h := NewHealer(a.Orchestrator)
	go h.Loop(req.Issue)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Healer loop initiated")
}

// handleLedger exposes recent transactions, per-hustle ROI breakdown, and financial recommendations.
func (a *API) handleLedger(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if a.Orchestrator == nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"transactions": []Transaction{}})
		return
	}

	a.Orchestrator.Ledger.mu.RLock()
	txs := make([]Transaction, len(a.Orchestrator.Ledger.Transactions))
	copy(txs, a.Orchestrator.Ledger.Transactions)
	a.Orchestrator.Ledger.mu.RUnlock()

	byHustle := make(map[string]float64)
	for _, t := range txs {
		if t.Type == Revenue {
			byHustle[t.Hustle] += t.Amount
		} else {
			byHustle[t.Hustle] -= t.Amount
		}
	}

	// Return newest 60 transactions first
	recent := txs
	if len(recent) > 60 {
		recent = recent[len(recent)-60:]
	}
	reversed := make([]Transaction, len(recent))
	for i := range recent {
		reversed[i] = recent[len(recent)-1-i]
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"revenue":      a.Orchestrator.Ledger.TotalRevenue(),
		"expenses":     a.Orchestrator.Ledger.TotalExpenses(),
		"profit":       a.Orchestrator.Ledger.Profit(),
		"wealth_goal":  a.Orchestrator.WealthGoal,
		"by_hustle":    byHustle,
		"analysis":     a.Orchestrator.Ledger.AnalyzeProfitability(),
		"transactions": reversed,
	})
}

// handleTasks combines live daemon scheduler countdowns with SQLite task_history executions.
func (a *API) handleTasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var scheduled interface{} = []struct{}{}
	if a.SchedulerSnapshot != nil {
		scheduled = a.SchedulerSnapshot()
	}
	var history []TaskHistoryEntry
	if a.Orchestrator != nil && a.Orchestrator.DB != nil {
		if rows, err := a.Orchestrator.DB.GetTaskHistory(80); err == nil {
			history = rows
		}
	}
	if history == nil {
		history = []TaskHistoryEntry{}
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"scheduled": scheduled,
		"history":   history,
	})
}

// handleMemories supports tier filtering (L1/L2/L3/all), keyword filtering, and SQLite vector similarity search.
func (a *API) handleMemories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if a.Orchestrator == nil {
		_ = json.NewEncoder(w).Encode([]MemoryEntry{})
		return
	}

	tier := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("tier")))
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	mode := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("mode")))
	limit := 50
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 200 {
		limit = l
	}

	if mode == "vector" && q != "" && a.Orchestrator.DB != nil && a.Orchestrator.Embedder != nil {
		t := tier
		if t == "" || t == "ALL" {
			t = "L1"
		}
		if vec, err := a.Orchestrator.Embedder.Embed(q); err == nil {
			if hits, verr := a.Orchestrator.DB.VectorSearch(t, vec, limit); verr == nil && len(hits) > 0 {
				_ = json.NewEncoder(w).Encode(hits)
				return
			}
		}
	}

	var combined []MemoryEntry
	collect := func(entries []MemoryEntry) {
		for _, e := range entries {
			if q == "" || strings.Contains(strings.ToLower(e.Content), strings.ToLower(q)) || tagsContain(e.Tags, q) {
				combined = append(combined, e)
			}
		}
	}

	switch tier {
	case "L1":
		collect(a.Orchestrator.L1.EntriesSnapshot())
	case "L2":
		collect(a.Orchestrator.L2.EntriesSnapshot())
	case "L3":
		collect(a.Orchestrator.L3.EntriesSnapshot())
	default:
		collect(a.Orchestrator.L1.EntriesSnapshot())
		collect(a.Orchestrator.L2.EntriesSnapshot())
		collect(a.Orchestrator.L3.EntriesSnapshot())
	}

	sort.Slice(combined, func(i, j int) bool {
		return combined[i].Timestamp.After(combined[j].Timestamp)
	})
	if len(combined) > limit {
		combined = combined[:limit]
	}
	if combined == nil {
		combined = []MemoryEntry{}
	}
	_ = json.NewEncoder(w).Encode(combined)
}

func tagsContain(tags []string, q string) bool {
	ql := strings.ToLower(q)
	for _, t := range tags {
		if strings.Contains(strings.ToLower(t), ql) {
			return true
		}
	}
	return false
}

// handleOutputs inspects generated artifacts across output/products, output/shorts, output/site, and content_calendar.json.
func (a *API) handleOutputs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	readManifests := func(base string) []json.RawMessage {
		out := []json.RawMessage{}
		entries, err := os.ReadDir(base)
		if err != nil {
			return out
		}
		for i := len(entries) - 1; i >= 0 && len(out) < 25; i-- {
			e := entries[i]
			if !e.IsDir() {
				continue
			}
			mpath := filepath.Join(base, e.Name(), "manifest.json")
			if data, err := os.ReadFile(mpath); err == nil && json.Valid(data) {
				out = append(out, json.RawMessage(data))
			}
		}
		return out
	}

	sitePages := []string{}
	if sfiles, err := os.ReadDir("output/site"); err == nil {
		for _, f := range sfiles {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".html") {
				sitePages = append(sitePages, f.Name())
			}
		}
	}

	var calendar interface{} = []struct{}{}
	if a.CalendarSnapshot != nil {
		calendar = a.CalendarSnapshot()
	} else if cdata, err := os.ReadFile("content_calendar.json"); err == nil && json.Valid(cdata) {
		calendar = json.RawMessage(cdata)
	}

	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"products":   readManifests("output/products"),
		"shorts":     readManifests("output/shorts"),
		"site_pages": sitePages,
		"calendar":   calendar,
	})
}

// handleHymns (reconciled from stash@{0}) lists generated Hymn Remixes (*_projectm.mp4)
// queued in /opt/hymnmania/youtube_ready (or HYMNMANIA_DIR) for YouTube distribution.
func (a *API) handleHymns(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	dir := os.Getenv("HYMNMANIA_DIR")
	if dir == "" {
		dir = firstExisting("/opt/hymnmania/youtube_ready", "output/hymns")
	}
	type hymnItem struct {
		Name    string  `json:"name"`
		Title   string  `json:"title"`
		SizeMB  float64 `json:"size_mb"`
		ModTime string  `json:"mod_time"`
	}
	items := []hymnItem{}
	if dir != "" {
		if entries, err := os.ReadDir(dir); err == nil {
			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".mp4") {
					continue
				}
				var sizeMB float64
				var modStr string
				if info, ierr := e.Info(); ierr == nil {
					sizeMB = float64(int((float64(info.Size())/1048576.0)*10)) / 10
					modStr = info.ModTime().Format("2006-01-02 15:04")
				}
				title := strings.TrimSuffix(e.Name(), "_projectm.mp4")
				title = strings.TrimSuffix(title, ".mp4")
				title = strings.ReplaceAll(title, "_", " ")
				items = append(items, hymnItem{
					Name:    e.Name(),
					Title:   title,
					SizeMB:  sizeMB,
					ModTime: modStr,
				})
			}
		}
	}
	_ = json.NewEncoder(w).Encode(items)
}
