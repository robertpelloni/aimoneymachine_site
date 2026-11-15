package stocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/robertpelloni/hustle/orchestrator"
)

// AlpacaExecutor handles paper/live equity order execution on Alpaca Markets.
type AlpacaExecutor struct {
	APIKey    string
	APISecret string
	BaseURL   string
	Client    *http.Client
}

func NewAlpacaExecutor() *AlpacaExecutor {
	baseURL := os.Getenv("ALPACA_BASE_URL")
	if baseURL == "" {
		baseURL = "https://paper-api.alpaca.markets"
	}
	return &AlpacaExecutor{
		APIKey:    os.Getenv("ALPACA_API_KEY"),
		APISecret: os.Getenv("ALPACA_API_SECRET"),
		BaseURL:   baseURL,
		Client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (a *AlpacaExecutor) ExecuteOrder(symbol, side string, quantity float64) error {
	if a.APIKey == "" || a.APISecret == "" {
		return fmt.Errorf("Alpaca API keys not set (configure ALPACA_API_KEY and ALPACA_API_SECRET)")
	}

	url := fmt.Sprintf("%s/v2/orders", strings.TrimRight(a.BaseURL, "/"))
	order := map[string]interface{}{
		"symbol":        strings.ToUpper(symbol),
		"qty":           fmt.Sprintf("%f", quantity),
		"side":          strings.ToLower(side),
		"type":          "market",
		"time_in_force": "gtc",
	}
	body, _ := json.Marshal(order)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("APCA-API-KEY-ID", a.APIKey)
	req.Header.Set("APCA-API-SECRET-KEY", a.APISecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		data, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Alpaca API error %d: %s", resp.StatusCode, string(data))
	}

	fmt.Printf("[Alpaca] Successfully executed %s %s for %f shares\n", side, symbol, quantity)
	return nil
}

type StockModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewStockModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *StockModule {
	return &StockModule{
		Orch:   orch,
		Broker: broker,
	}
}

func stockHeadline(s string) string {
	line := strings.TrimSpace(s)
	if idx := strings.IndexByte(line, '\n'); idx >= 0 {
		line = strings.TrimSpace(line[:idx])
	}
	r := []rune(line)
	if len(r) > 100 {
		return string(r[:100]) + "..."
	}
	return line
}

// AnalyzeStock performs a quantitative sentiment & risk assessment for an equity ticker.
func (s *StockModule) AnalyzeStock(symbol string) (string, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		symbol = "NVDA"
	}
	fmt.Printf("[Stocks] Analyzing stock: %s\n", symbol)

	prompt := fmt.Sprintf(`Act as a quantitative stock analyst. Analyze the stock: "%s".

Requirements:
- Recent price trends
- Market sentiment (Bullish/Bearish)
- Key risk factors
- Recommendation (BUY, SELL, HOLD)

Format: Professional Markdown.`, symbol)

	analysis, err := s.Orch.LLM.Generate(prompt)
	if err != nil {
		return "", err
	}

	s.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("stock-analysis-%s-%d", symbol, time.Now().Unix()),
		Content:   fmt.Sprintf("Stock Analysis for %s: %s\n\n%s", symbol, stockHeadline(analysis), analysis),
		Timestamp: time.Now(),
		Tags:      []string{"stocks", "trading", symbol},
	})

	return analysis, nil
}
