package trading

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
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

// GeminiExecutor handles real trade execution on Gemini exchange
type GeminiExecutor struct {
	APIKey    string
	APISecret string
	BaseURL   string
	Client    *http.Client
}

func NewGeminiExecutor() *GeminiExecutor {
	return &GeminiExecutor{
		APIKey:    os.Getenv("GEMINI_API_KEY"),
		APISecret: os.Getenv("GEMINI_API_SECRET"),
		BaseURL:   "https://api.gemini.com",
		Client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *GeminiExecutor) GetPrice(symbol string) (float64, error) {
	ticker := strings.ToLower(symbol)
	if !strings.HasSuffix(ticker, "usd") {
		ticker += "usd"
	}

	url := fmt.Sprintf("%s/v1/pubticker/%s", g.BaseURL, ticker)
	resp, err := g.Client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Last string `json:"last"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return strconv.ParseFloat(result.Last, 64)
}

func (g *GeminiExecutor) signAndSend(method, path string, payload map[string]interface{}) ([]byte, error) {
	nonce := time.Now().UnixNano()
	payload["request"] = path
	payload["nonce"] = nonce

	payloadBytes, _ := json.Marshal(payload)
	b64Payload := base64.StdEncoding.EncodeToString(payloadBytes)

	mac := hmac.New(sha512.New384, []byte(g.APISecret))
	mac.Write([]byte(b64Payload))
	signature := hex.EncodeToString(mac.Sum(nil))

	req, _ := http.NewRequest(method, g.BaseURL+path, nil)
	req.Header.Set("X-GEMINI-APIKEY", g.APIKey)
	req.Header.Set("X-GEMINI-PAYLOAD", b64Payload)
	req.Header.Set("X-GEMINI-SIGNATURE", signature)
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (g *GeminiExecutor) ExecuteOrder(symbol, side, orderType string, quantity float64) error {
	if g.APIKey == "" || g.APISecret == "" {
		return fmt.Errorf("Gemini API keys not set")
	}

	ticker := strings.ToLower(symbol)
	if !strings.HasSuffix(ticker, "usd") {
		ticker += "usd"
	}

	path := "/v1/order/new"
	payload := map[string]interface{}{
		"symbol":  ticker,
		"amount":  strconv.FormatFloat(quantity, 'f', 8, 64),
		"price":   "1", // Market orders still require a dummy price in Gemini sandbox sometimes
		"side":    strings.ToLower(side),
		"type":    "exchange market", // Use market order
		"options": []string{"fill-or-kill"},
	}

	body, err := g.signAndSend("POST", path, payload)
	if err != nil {
		return err
	}

	fmt.Printf("[Gemini] API Response: %s\n", string(body))
	return nil
}
