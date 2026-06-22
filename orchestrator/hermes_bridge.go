package orchestrator

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// HermesBridge connects the Hustle Machine to Hermes Agent via MCP or CLI
type HermesBridge struct {
	client    *MCPClient
	cli       *HermesCLI
	tools     []HermesTool
	connected bool
	mode      string // "mcp" or "cli"
}

// HermesTool represents a tool available from Hermes
type HermesTool struct {
	Name        string
	Description string
	Category    string
}

// NewHermesBridge creates a new bridge to Hermes Agent
func NewHermesBridge() *HermesBridge {
	return &HermesBridge{}
}

// Connect starts the Hermes MCP server and performs the handshake
// Falls back to CLI mode if MCP is unavailable
func (h *HermesBridge) Connect() error {
	// Try MCP first
	err := h.connectMCP()
	if err == nil {
		h.mode = "mcp"
		return nil
	}
	
	fmt.Printf("[Hermes] MCP connection failed: %v\n", err)
	fmt.Println("[Hermes] Falling back to CLI mode...")
	
	// Fall back to CLI
	cli := NewHermesCLI()
	if !cli.Available {
		return fmt.Errorf("Hermes not available via MCP or CLI")
	}
	
	h.cli = cli
	h.connected = true
	h.mode = "cli"
	
	fmt.Printf("[Hermes] ✅ Connected to Hermes Agent via CLI (path: %s)\n", cli.Path)
	fmt.Println("[Hermes] Available capabilities: web search, scraping, image gen, messaging")
	
	return nil
}

// connectMCP attempts to connect via MCP protocol
func (h *HermesBridge) connectMCP() error {
	// Find hermes executable
	hermesPath := os.Getenv("HERMES_PATH")
	if hermesPath == "" {
		locations := []string{"hermes", "/usr/local/bin/hermes", "/home/hyper/.local/bin/hermes"}
		for _, loc := range locations {
			if _, err := os.Stat(loc); err == nil {
				hermesPath = loc
				break
			}
		}
		if hermesPath == "" {
			hermesPath = "wsl"
		}
	}

	var client *MCPClient
	var err error

	if hermesPath == "wsl" {
		client, err = NewMCPClient("wsl", "-e", "bash", "-lc", "hermes mcp serve --accept-hooks")
	} else {
		client, err = NewMCPClient(hermesPath, "mcp", "serve", "--accept-hooks")
	}

	if err != nil {
		return fmt.Errorf("failed to start Hermes MCP server: %v", err)
	}

	h.client = client

	// Perform MCP handshake with timeout
	if err := client.Initialize(); err != nil {
		client.Close()
		return fmt.Errorf("MCP handshake failed: %v", err)
	}

	// Get available tools
	tools, err := client.ListTools()
	if err != nil {
		client.Close()
		return fmt.Errorf("failed to list Hermes tools: %v", err)
	}

	// Categorize tools
	for _, t := range tools {
		category := categorizeTool(t.Name)
		h.tools = append(h.tools, HermesTool{
			Name:        t.Name,
			Description: t.Description,
			Category:    category,
		})
	}

	h.connected = true
	fmt.Printf("[Hermes] ✅ Connected to Hermes Agent via MCP (%d tools available)\n", len(h.tools))

	categories := make(map[string]int)
	for _, t := range h.tools {
		categories[t.Category]++
	}
	for cat, count := range categories {
		fmt.Printf("[Hermes]   %s: %d tools\n", cat, count)
	}

	return nil
}

// categorizeTool assigns a category based on tool name
func categorizeTool(name string) string {
	name = strings.ToLower(name)
	switch {
	case strings.Contains(name, "web") || strings.Contains(name, "search") || strings.Contains(name, "scrape"):
		return "research"
	case strings.Contains(name, "browser") || strings.Contains(name, "navigate"):
		return "browser"
	case strings.Contains(name, "social") || strings.Contains(name, "twitter") || strings.Contains(name, "x_"):
		return "social"
	case strings.Contains(name, "image") || strings.Contains(name, "video") || strings.Contains(name, "media"):
		return "media"
	case strings.Contains(name, "email") || strings.Contains(name, "mail"):
		return "email"
	case strings.Contains(name, "file") || strings.Contains(name, "read") || strings.Contains(name, "write"):
		return "filesystem"
	case strings.Contains(name, "terminal") || strings.Contains(name, "exec") || strings.Contains(name, "shell"):
		return "terminal"
	case strings.Contains(name, "memory") || strings.Contains(name, "session"):
		return "memory"
	case strings.Contains(name, "cron") || strings.Contains(name, "schedule"):
		return "scheduling"
	case strings.Contains(name, "message") || strings.Contains(name, "chat"):
		return "messaging"
	default:
		return "other"
	}
}

// IsConnected returns whether the bridge is active
func (h *HermesBridge) IsConnected() bool {
	return h.connected
}

// GetMode returns the connection mode (mcp or cli)
func (h *HermesBridge) GetMode() string {
	return h.mode
}

// GetTools returns available Hermes tools
func (h *HermesBridge) GetTools() []HermesTool {
	return h.tools
}

// GetToolsByCategory returns tools filtered by category
func (h *HermesBridge) GetToolsByCategory(category string) []HermesTool {
	var filtered []HermesTool
	for _, t := range h.tools {
		if t.Category == category {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// CallTool calls a Hermes tool by name
func (h *HermesBridge) CallTool(name string, args map[string]interface{}) (string, error) {
	if !h.connected {
		return "", fmt.Errorf("Hermes bridge not connected")
	}
	
	if h.mode == "mcp" && h.client != nil {
		return h.client.CallTool(name, args)
	}
	
	// CLI mode — use hermes -z for one-shot execution
	if h.cli != nil {
		// Build a prompt that describes the tool call
		var argStrs []string
		for k, v := range args {
			argStrs = append(argStrs, fmt.Sprintf("%s=%v", k, v))
		}
		prompt := fmt.Sprintf("Use the %s tool with these parameters: %s", name, strings.Join(argStrs, ", "))
		return h.cli.runCommand("-z", prompt)
	}
	
	return "", fmt.Errorf("no Hermes connection available")
}

// WebSearch performs a web search using Hermes
func (h *HermesBridge) WebSearch(query string) (string, error) {
	if !h.connected {
		return "", fmt.Errorf("Hermes bridge not connected")
	}
	
	if h.mode == "mcp" {
		return h.CallTool("web_search", map[string]interface{}{"query": query})
	}
	
	// CLI mode
	return h.cli.WebSearch(query)
}

// ScrapeURL scrapes a URL using Hermes
func (h *HermesBridge) ScrapeURL(url string) (string, error) {
	if !h.connected {
		return "", fmt.Errorf("Hermes bridge not connected")
	}
	
	if h.mode == "mcp" {
		return h.CallTool("scrape_url", map[string]interface{}{"url": url})
	}
	
	return h.cli.ScrapeURL(url)
}

// SearchTwitter searches Twitter/X using Hermes
func (h *HermesBridge) SearchTwitter(query string) (string, error) {
	if !h.connected {
		return "", fmt.Errorf("Hermes bridge not connected")
	}
	
	if h.mode == "mcp" {
		return h.CallTool("x_search", map[string]interface{}{"query": query})
	}
	
	return h.cli.SearchTwitter(query)
}

// GenerateImage generates an image using Hermes
func (h *HermesBridge) GenerateImage(prompt string) (string, error) {
	if !h.connected {
		return "", fmt.Errorf("Hermes bridge not connected")
	}
	
	if h.mode == "mcp" {
		return h.CallTool("image_gen", map[string]interface{}{"prompt": prompt})
	}
	
	return h.cli.GenerateImage(prompt)
}

// SendMessage sends a message via Hermes messaging
func (h *HermesBridge) SendMessage(platform, message string) (string, error) {
	if !h.connected {
		return "", fmt.Errorf("Hermes bridge not connected")
	}
	
	if h.mode == "mcp" {
		return h.CallTool("messaging_send", map[string]interface{}{
			"platform": platform,
			"message":  message,
		})
	}
	
	return h.cli.SendMessage(platform, message)
}

// GetCapabilities returns a description of Hermes capabilities
func (h *HermesBridge) GetCapabilities() string {
	if !h.connected {
		return "Hermes Agent: Not connected"
	}
	
	if h.mode == "mcp" {
		var sb strings.Builder
		sb.WriteString("HERMES AGENT TOOLS (via MCP):\n")
		categories := make(map[string][]HermesTool)
		for _, t := range h.tools {
			categories[t.Category] = append(categories[t.Category], t)
		}
		for cat, tools := range categories {
			sb.WriteString(fmt.Sprintf("\n%s:\n", strings.ToUpper(cat)))
			for _, t := range tools {
				sb.WriteString(fmt.Sprintf("  - %s: %s\n", t.Name, t.Description))
			}
		}
		return sb.String()
	}
	
	return h.cli.GetCapabilities()
}

// Close shuts down the Hermes bridge
func (h *HermesBridge) Close() error {
	if h.client != nil {
		return h.client.Close()
	}
	return nil
}

// HermesCLI provides direct access to Hermes Agent via CLI commands
type HermesCLI struct {
	Path      string
	Available bool
}

// NewHermesCLI creates a new Hermes CLI client
func NewHermesCLI() *HermesCLI {
	h := &HermesCLI{}
	
	// Quick check via WSL
	cmd := exec.Command("wsl", "-e", "bash", "-lc", "which hermes")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		h.Path = "wsl"
		h.Available = true
		return h
	}
	
	// Try direct hermes
	if _, err := exec.LookPath("hermes"); err == nil {
		h.Path = "hermes"
		h.Available = true
	}
	
	return h
}

// runCommand executes a hermes command and returns the output
func (h *HermesCLI) runCommand(args ...string) (string, error) {
	if !h.Available {
		return "", fmt.Errorf("hermes not available")
	}
	
	var cmd *exec.Cmd
	if h.Path == "wsl" {
		cmd = exec.Command("wsl", "-e", "bash", "-lc", "hermes "+strings.Join(args, " "))
	} else {
		cmd = exec.Command(h.Path, args...)
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("hermes command failed: %v", err)
	}
	
	return string(output), nil
}

// Status returns the status of Hermes
func (h *HermesCLI) Status() (string, error) {
	return h.runCommand("status")
}

// WebSearch performs a web search using Hermes
func (h *HermesCLI) WebSearch(query string) (string, error) {
	return h.runCommand("-z", fmt.Sprintf("Search the web for: %s. Provide a brief summary.", query))
}

// ScrapeURL scrapes a URL using Hermes
func (h *HermesCLI) ScrapeURL(url string) (string, error) {
	return h.runCommand("-z", fmt.Sprintf("Scrape and summarize: %s", url))
}

// SearchTwitter searches Twitter/X using Hermes
func (h *HermesCLI) SearchTwitter(query string) (string, error) {
	return h.runCommand("-z", fmt.Sprintf("Search Twitter/X for: %s", query))
}

// GenerateImage generates an image using Hermes
func (h *HermesCLI) GenerateImage(prompt string) (string, error) {
	return h.runCommand("-z", fmt.Sprintf("Generate an image: %s", prompt))
}

// SendMessage sends a message via Hermes messaging
func (h *HermesCLI) SendMessage(platform, message string) (string, error) {
	return h.runCommand("send", "--platform", platform, "--message", message)
}

// GetCapabilities returns a description of Hermes capabilities
func (h *HermesCLI) GetCapabilities() string {
	if !h.Available {
		return "Hermes Agent: Not available"
	}
	
	return `HERMES AGENT CAPABILITIES (CLI mode):
- Web Search: Search the web for information
- Web Scraping: Scrape and extract content from URLs
- Twitter/X Search: Search Twitter/X for posts and trends
- Image Generation: Generate images from text prompts
- Video Generation: Generate videos from text prompts
- Text-to-Speech: Convert text to speech
- Browser Automation: Automate browser interactions
- Email: Send and manage emails
- Cross-Platform Messaging: Send messages via Telegram, Discord, WhatsApp, Slack
- Cron Jobs: Schedule recurring tasks
- Memory: Store and retrieve information
- Skills: Execute specialized skills`
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}