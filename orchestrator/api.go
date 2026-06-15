package orchestrator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

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

	// Serve the mobile-friendly dashboard
	fs := http.FileServer(http.Dir("orchestrator/static"))
	http.Handle("/", fs)

	fmt.Printf("[API] Orchestrator listening on port %s\n", port)
	return http.ListenAndServe(":"+port, nil)
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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "Active",
		"version":  a.Orchestrator.Version,
		"profit":   a.Orchestrator.Ledger.Profit(),
		"memories": len(a.Orchestrator.L1.Entries),
		"peers":    len(a.Broker.Peers),
	})
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
		http.Error(w, fmt.Sprintf("Error reading content: %v", err), http.StatusInternalServerError)
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
