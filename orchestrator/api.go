package orchestrator

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// dashboardHTML is the consolidated single-page dashboard (moved here from
// the repo root in v1.3.0 so the Go binary is self-contained; the Python
// dashboard_server.py fallback reads this same file from orchestrator/).
//
//go:embed dashboard.html
var dashboardHTML []byte

type API struct {
	Orchestrator *Orchestrator
	Protocol     *HustleProtocol
	Broker       *A2ABroker
	ChainManager *ChainManager
	Discoverer   *ChainDiscoverer
	MultiAgent   *MultiAgentOrchestrator
}

func NewAPI(orch *Orchestrator, protocol *HustleProtocol, broker *A2ABroker, chainMgr *ChainManager, disc *ChainDiscoverer) *API {
	return &API{
		Orchestrator: orch,
		Protocol:     protocol,
		Broker:       broker,
		ChainManager: chainMgr,
		Discoverer:   disc,
	}
}

func (a *API) Start(port string) error {
	http.HandleFunc("/dispatch", a.handleDispatch)
	http.HandleFunc("/status", a.handleStatus)
	http.HandleFunc("/message", a.handleMessage)
	http.HandleFunc("/register", a.handleRegister)
	http.HandleFunc("/sync", a.handleSync)
	http.HandleFunc("/chains", a.handleChains)
	http.HandleFunc("/agents", a.handleAgents)
	http.HandleFunc("/content", a.handleContent)
	http.HandleFunc("/strategy", a.handleStrategy)
	http.HandleFunc("/healer", a.handleHealer)
	// Embedded dashboard + the /api/* aliases its JS fetches (same handlers,
	// so the page works served from :8082 with no Python proxy in between).
	http.HandleFunc("/", a.handleDashboard)
	http.HandleFunc("/dashboard", a.handleDashboard)
	http.HandleFunc("/api/dispatch", a.handleDispatch)
	http.HandleFunc("/api/status", a.handleStatus)
	http.HandleFunc("/api/chains", a.handleChains)
	http.HandleFunc("/api/agents", a.handleAgents)
	http.HandleFunc("/api/content", a.handleContent)
	http.HandleFunc("/api/strategy", a.handleStrategy)
	http.HandleFunc("/api/log", a.handleLog)
	http.HandleFunc("/api/progress", a.handleProgress)

	fmt.Printf("[API] Orchestrator listening on port %s (dashboard: /)\n", port)
	return http.ListenAndServe(":"+port, nil)
}

// handleDashboard serves the embedded consolidated dashboard. Unknown paths
// 404 instead of falling through to the page, so typos stay visible.
func (a *API) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/dashboard" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(dashboardHTML)
}

// firstExisting returns the first path that exists (dashboard log/progress
// files live in different places locally vs prod), or "".
func firstExisting(names ...string) string {
	for _, n := range names {
		if _, err := os.Stat(n); err == nil {
			return n
		}
	}
	return ""
}

func (a *API) handleLog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	p := firstExisting("orchestrator_live.log", "orchestrator.log", "../orchestrator_live.log", "/opt/aimoneymachine/orchestrator_live.log")
	if p == "" {
		return // empty log is a normal state; the dashboard says so itself
	}
	data, _ := os.ReadFile(p)
	w.Write(data)
}

func (a *API) handleProgress(w http.ResponseWriter, r *http.Request) {
	p := firstExisting("expansion_progress.json", "../expansion_progress.json", "bin/expansion_progress.json", "/opt/aimoneymachine/bin/expansion_progress.json")
	if p == "" {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"done":[]}`)
		return
	}
	data, _ := os.ReadFile(p)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

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

	if err := a.Protocol.HandleURI(req.URI); err != nil {
		// Log server-side too: a failed dispatch (e.g. an LLM call that
		// hit its timeout under load) was previously invisible in the
		// journal — the client saw a timeout and the server said nothing.
		fmt.Printf("[API] Dispatch failed for URI %s: %v\n", req.URI, err)
		http.Error(w, fmt.Sprintf("Protocol error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "URI %s dispatched successfully", req.URI)
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
		// Process topic message locally
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
	resp := map[string]interface{}{
		"status":   "Active",
		"version":  a.Orchestrator.Version,
		"profit":   a.Orchestrator.Ledger.Profit(),
		"memories": len(a.Orchestrator.L1.Entries),
		"peers":    len(a.Broker.Peers),
	}
	// Disk headroom for the working directory's filesystem (see
	// DiskUsage); omitted on platforms without an implementation.
	if pct, freeGB, ok := DiskUsage("."); ok {
		resp["disk_used_pct"] = float64(int(pct*10)) / 10
		resp["disk_free_gb"] = float64(int(freeGB*10)) / 10
	}
	json.NewEncoder(w).Encode(resp)
}

func (a *API) handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Println("[API] Remote trigger: Repository Synchronization Protocol")
	// Trigger the protocol URI
	if err := a.Protocol.HandleURI("hustle://sync"); err != nil {
		http.Error(w, fmt.Sprintf("Sync failed: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Sync protocol initiated successfully")
}

func (a *API) handleChains(w http.ResponseWriter, r *http.Request) {
	if a.ChainManager == nil {
		http.Error(w, "Chain manager not initialized", http.StatusServiceUnavailable)
		return
	}
	json.NewEncoder(w).Encode(a.ChainManager.Chains)
}

func (a *API) handleAgents(w http.ResponseWriter, r *http.Request) {
	if a.MultiAgent != nil {
		json.NewEncoder(w).Encode(a.MultiAgent.StatusReport())
		return
	}
	agentEntries := a.Orchestrator.L1.Search("agent_loop")
	json.NewEncoder(w).Encode(agentEntries)
}

func (a *API) handleContent(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir("output/content")
	if err != nil {
		// No content generated yet is a normal state, not a server error:
		// the dashboard renders an empty library from [].
		json.NewEncoder(w).Encode([]string{})
		return
	}
	var fileList []string
	for _, f := range files {
		if !f.IsDir() {
			fileList = append(fileList, f.Name())
		}
	}
	json.NewEncoder(w).Encode(fileList)
}

func (a *API) handleStrategy(w http.ResponseWriter, r *http.Request) {
	strategies := a.Orchestrator.L1.Search("collective_strategy")
	json.NewEncoder(w).Encode(strategies)
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
