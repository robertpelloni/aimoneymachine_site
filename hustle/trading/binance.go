package trading

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// BinanceExecutor handles real trade execution on Binance
type BinanceExecutor struct {
	APIKey    string
	APISecret string
	BaseURL   string
	Client    *http.Client
}

func (b *BinanceExecutor) GetPrice(symbol string) (float64, error) {
	url := fmt.Sprintf("%s/api/v3/ticker/price?symbol=%s", b.BaseURL, symbol)
	resp, err := b.Client.Get(url)
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

func NewBinanceExecutor() *BinanceExecutor {
	return &BinanceExecutor{
		APIKey:    os.Getenv("BINANCE_API_KEY"),
		APISecret: os.Getenv("BINANCE_API_SECRET"),
		BaseURL:   "https://api.binance.com",
		Client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// sign generates the HMAC SHA256 signature for Binance API
func (b *BinanceExecutor) sign(payload string) string {
	mac := hmac.New(sha256.New, []byte(b.APISecret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// ExecuteOrder sends a real order to Binance
func (b *BinanceExecutor) ExecuteOrder(symbol, side, orderType string, quantity float64) error {
	if b.APIKey == "" || b.APISecret == "" {
		return fmt.Errorf("Binance API keys not set")
	}

	endpoint := "/api/v3/order"
	timestamp := time.Now().UnixMilli()

	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("side", side)
	params.Set("type", orderType)
	params.Set("quantity", strconv.FormatFloat(quantity, 'f', 8, 64))
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))

	signature := b.sign(params.Encode())
	params.Set("signature", signature)

	fullURL := fmt.Sprintf("%s%s?%s", b.BaseURL, endpoint, params.Encode())

	req, _ := http.NewRequest("POST", fullURL, nil)
	req.Header.Set("X-MBX-APIKEY", b.APIKey)

	resp, err := b.Client.Do(req)
	if err != nil {
		return fmt.Errorf("Binance API request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Binance API error %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("[Binance] Successfully executed %s %s for %s\n", side, symbol, strconv.FormatFloat(quantity, 'f', 2, 64))
	return nil
}

// GetBalance returns the account balance for a specific asset
func (b *BinanceExecutor) GetBalance(asset string) (float64, error) {
	endpoint := "/api/v3/account"
	timestamp := time.Now().UnixMilli()

	params := url.Values{}
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))
	signature := b.sign(params.Encode())
	params.Set("signature", signature)

	fullURL := fmt.Sprintf("%s%s?%s", b.BaseURL, endpoint, params.Encode())

	req, _ := http.NewRequest("GET", fullURL, nil)
	req.Header.Set("X-MBX-APIKEY", b.APIKey)

	resp, err := b.Client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var account struct {
		Balances []struct {
			Asset  string `json:"asset"`
			Free   string `json:"free"`
			Locked string `json:"locked"`
		} `json:"balances"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return 0, err
	}

	for _, b := range account.Balances {
		if b.Asset == asset {
			val, _ := strconv.ParseFloat(b.Free, 64)
			return val, nil
		}
	}

	return 0, nil
}
