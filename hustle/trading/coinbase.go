package trading

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// CoinbaseExecutor handles real trade execution on Coinbase Advanced Trade API
type CoinbaseExecutor struct {
	APIKey    string
	APISecret string
	BaseURL   string
	Client    *http.Client
}

func NewCoinbaseExecutor() *CoinbaseExecutor {
	return &CoinbaseExecutor{
		APIKey:    os.Getenv("COINBASE_API_KEY"),
		APISecret: os.Getenv("COINBASE_API_SECRET"),
		BaseURL:   "https://api.coinbase.com",
		Client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *CoinbaseExecutor) GetPrice(symbol string) (float64, error) {
	// Coinbase Advanced Trade uses product_id like BTC-USD
	productID := symbol
	if !strings.Contains(symbol, "-") {
		productID = symbol + "-USD"
	}

	url := fmt.Sprintf("%s/api/v3/brokerage/products/%s", c.BaseURL, productID)
	req, _ := http.NewRequest("GET", url, nil)
	c.signRequest(req, "GET", "/api/v3/brokerage/products/"+productID, "")

	resp, err := c.Client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Price string `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return strconv.ParseFloat(result.Price, 64)
}

func (c *CoinbaseExecutor) signRequest(req *http.Request, method, path, body string) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	message := timestamp + method + path + body

	h := hmac.New(sha256.New, []byte(c.APISecret))
	h.Write([]byte(message))
	signature := hex.EncodeToString(h.Sum(nil))

	req.Header.Set("CB-ACCESS-KEY", c.APIKey)
	req.Header.Set("CB-ACCESS-SIGN", signature)
	req.Header.Set("CB-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("Content-Type", "application/json")
}

func (c *CoinbaseExecutor) ExecuteOrder(symbol, side, orderType string, quantity float64) error {
	if c.APIKey == "" || c.APISecret == "" {
		return fmt.Errorf("Coinbase API keys not set")
	}

	productID := symbol
	if !strings.Contains(symbol, "-") {
		productID = symbol + "-USD"
	}

	endpoint := "/api/v3/brokerage/orders"

	// Create a client order ID for idempotency
	clientOrderID := fmt.Sprintf("hustle-%d", time.Now().UnixNano())

	orderConfig := map[string]interface{}{
		"client_order_id": clientOrderID,
		"product_id":      productID,
		"side":            side,
		"order_configuration": map[string]interface{}{
			"market_market_ioc": map[string]interface{}{
				"base_size": strconv.FormatFloat(quantity, 'f', 8, 64),
			},
		},
	}

	bodyBytes, _ := json.Marshal(orderConfig)
	bodyStr := string(bodyBytes)

	req, _ := http.NewRequest("POST", c.BaseURL+endpoint, strings.NewReader(bodyStr))
	c.signRequest(req, "POST", endpoint, bodyStr)

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("Coinbase API request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Coinbase API error %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("[Coinbase] Successfully executed %s %s for %s\n", side, symbol, strconv.FormatFloat(quantity, 'f', 8, 64))
	return nil
}
