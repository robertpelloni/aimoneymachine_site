package stocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// AlpacaExecutor handles real trade execution on Alpaca for stocks
type AlpacaExecutor struct {
	APIKey    string
	APISecret string
	BaseURL   string
	Client    *http.Client
}

func NewAlpacaExecutor() *AlpacaExecutor {
	return &AlpacaExecutor{
		APIKey:    os.Getenv("ALPACA_API_KEY"),
		APISecret: os.Getenv("ALPACA_API_SECRET"),
		BaseURL:   "https://paper-api.alpaca.markets", // Default to paper trading
		Client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (a *AlpacaExecutor) ExecuteOrder(symbol, side string, quantity float64) error {
	if a.APIKey == "" || a.APISecret == "" {
		return fmt.Errorf("Alpaca API keys not set")
	}

	url := fmt.Sprintf("%s/v2/orders", a.BaseURL)

	order := map[string]interface{}{
		"symbol":        symbol,
		"qty":           fmt.Sprintf("%f", quantity),
		"side":          side,
		"type":          "market",
		"time_in_force": "gtc",
	}
	body, _ := json.Marshal(order)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("APCA-API-KEY-ID", a.APIKey)
	req.Header.Set("APCA-API-SECRET-KEY", a.APISecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.Client.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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

// AnalyzeStock performs a sentiment and technical assessment of a stock
func (s *StockModule) AnalyzeStock(symbol string) (string, error) {
	fmt.Printf("[Stocks] Analyzing stock: %s\n", symbol)

	prompt := fmt.Sprintf(`Act as a quantitative stock analyst. Analyze the stock: "%s".

Requirements:
- Recent price trends
- Market sentiment (Bullish/Bearish)
- Key risk factors
- Recommendation (BUY, SELL, HOLD)

Format: Professional Markdown.`, symbol)

	analysis, err := s.Orch.LLM.Generate(prompt)
	if err != nil { return "", err }

	// Store in memory
	s.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("stock-analysis-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Stock Analysis for %s: %s", symbol, analysis[:strings.Index(analysis, "\n")]),
		Timestamp: time.Now(),
		Tags:      []string{"stocks", "trading", symbol},
	})

	return analysis, nil
}
