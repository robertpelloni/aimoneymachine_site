package orchestrator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// MCPClient connects to an MCP server via stdio transport
type MCPClient struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  *bufio.Scanner
	mu      sync.Mutex
	nextID  int
	tools   []MCPTool
	running bool
}

// MCPTool represents a tool exposed by an MCP server
type MCPTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// JSON-RPC message types
type jsonrpcRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type initializeParams struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    capabilities `json:"capabilities"`
	ClientInfo      clientInfo   `json:"clientInfo"`
}

type capabilities struct {
	Tools *toolsCapability `json:"tools,omitempty"`
}

type toolsCapability struct{}

type clientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type initializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	ServerInfo      serverInfo   `json:"serverInfo"`
	Capabilities    capabilities `json:"capabilities"`
}

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type toolsListResult struct {
	Tools []MCPTool `json:"tools"`
}

type toolCallParams struct {
	Name      string      `json:"name"`
	Arguments interface{} `json:"arguments,omitempty"`
}

type toolCallResult struct {
	Content []toolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// NewMCPClient creates a new MCP client that connects to a server via stdio
func NewMCPClient(command string, args ...string) (*MCPClient, error) {
	cmd := exec.Command(command, args...)
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %v", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start MCP server: %v", err)
	}
	
	// Wait a moment to see if the process starts successfully
	time.Sleep(500 * time.Millisecond)
	
	// Check if process is still running
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		return nil, fmt.Errorf("MCP server exited immediately")
	}
	
	client := &MCPClient{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  bufio.NewScanner(stdout),
		nextID:  1,
		running: true,
	}
	
	// Set a large buffer for the scanner
	client.stdout.Buffer(make([]byte, 1024*1024), 1024*1024)
	
	return client, nil
}

// send sends a JSON-RPC request and returns the response
func (c *MCPClient) send(method string, params interface{}) (*jsonrpcResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	id := c.nextID
	c.nextID++
	
	req := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
	
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}
	
	// MCP uses newline-delimited JSON
	_, err = fmt.Fprintf(c.stdin, "%s\n", data)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	
	// Read response (with timeout)
	type result struct {
		resp *jsonrpcResponse
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		if c.stdout.Scan() {
			var resp jsonrpcResponse
			if err := json.Unmarshal(c.stdout.Bytes(), &resp); err != nil {
				ch <- result{nil, fmt.Errorf("failed to parse response: %v", err)}
				return
			}
			ch <- result{&resp, nil}
		} else {
			ch <- result{nil, fmt.Errorf("no response from MCP server")}
		}
	}()
	
	select {
	case r := <-ch:
		return r.resp, r.err
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout waiting for MCP server response")
	}
}

// Initialize performs the MCP handshake
func (c *MCPClient) Initialize() error {
	params := initializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities: capabilities{
			Tools: &toolsCapability{},
		},
		ClientInfo: clientInfo{
			Name:    "hustle-machine",
			Version: "1.0.0",
		},
	}
	
	resp, err := c.send("initialize", params)
	if err != nil {
		return fmt.Errorf("initialize failed: %v", err)
	}
	
	if resp.Error != nil {
		return fmt.Errorf("initialize error: %s", resp.Error.Message)
	}
	
	var result initializeResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return fmt.Errorf("failed to parse initialize result: %v", err)
	}
	
	fmt.Printf("[MCP] Connected to %s v%s (protocol %s)\n", 
		result.ServerInfo.Name, result.ServerInfo.Version, result.ProtocolVersion)
	
	// Send initialized notification
	_, err = c.send("notifications/initialized", nil)
	if err != nil {
		// Notification might not get a response, that's ok
	}
	
	return nil
}

// ListTools retrieves available tools from the MCP server
func (c *MCPClient) ListTools() ([]MCPTool, error) {
	resp, err := c.send("tools/list", nil)
	if err != nil {
		return nil, fmt.Errorf("tools/list failed: %v", err)
	}
	
	if resp.Error != nil {
		return nil, fmt.Errorf("tools/list error: %s", resp.Error.Message)
	}
	
	var result toolsListResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tools list: %v", err)
	}
	
	c.tools = result.Tools
	return result.Tools, nil
}

// CallTool calls a tool on the MCP server
func (c *MCPClient) CallTool(name string, args map[string]interface{}) (string, error) {
	params := toolCallParams{
		Name:      name,
		Arguments: args,
	}
	
	resp, err := c.send("tools/call", params)
	if err != nil {
		return "", fmt.Errorf("tool call failed: %v", err)
	}
	
	if resp.Error != nil {
		return "", fmt.Errorf("tool call error: %s", resp.Error.Message)
	}
	
	var result toolCallResult
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return "", fmt.Errorf("failed to parse tool result: %v", err)
	}
	
	if result.IsError {
		var texts []string
		for _, c := range result.Content {
			texts = append(texts, c.Text)
		}
		return "", fmt.Errorf("tool error: %s", strings.Join(texts, "\n"))
	}
	
	var texts []string
	for _, c := range result.Content {
		texts = append(texts, c.Text)
	}
	return strings.Join(texts, "\n"), nil
}

// GetTools returns the cached tools list
func (c *MCPClient) GetTools() []MCPTool {
	return c.tools
}

// Close shuts down the MCP client
func (c *MCPClient) Close() error {
	c.running = false
	c.stdin.Close()
	return c.cmd.Process.Kill()
}