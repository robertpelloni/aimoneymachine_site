package trading

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// KrakenExecutor handles real trade execution on Kraken
type KrakenExecutor struct {
	APIKey    string
	APISecret string
	BaseURL   string
	Client    *http.Client
}

func (k *KrakenExecutor) GetPrice(symbol string) (float64, error) {
	url := fmt.Sprintf("%s/0/public/Ticker?pair=%s", k.BaseURL, symbol)
	resp, err := k.Client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Result map[string]struct {
			C []string `json:"c"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	for _, data := range result.Result {
		return strconv.ParseFloat(data.C[0], 64)
	}
	return 0, fmt.Errorf("no price data for %s", symbol)
}

func NewKrakenExecutor() *KrakenExecutor {
	return &KrakenExecutor{
		APIKey:    os.Getenv("KRAKEN_API_KEY"),
		APISecret: os.Getenv("KRAKEN_API_SECRET"),
		BaseURL:   "https://api.kraken.com",
		Client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// sign generates the Kraken-specific API signature
func (k *KrakenExecutor) sign(urlPath string, values url.Values) string {
	nonce := values.Get("nonce")
	postData := values.Encode()

	sha := sha256.New()
	sha.Write([]byte(nonce + postData))
	shaSum := sha.Sum(nil)

	secret, _ := base64.StdEncoding.DecodeString(k.APISecret)
	mac := hmac.New(sha512.New, secret)
	mac.Write(append([]byte(urlPath), shaSum...))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// ExecuteOrder sends a real order to Kraken
func (k *KrakenExecutor) ExecuteOrder(symbol, side, orderType string, quantity float64) error {
	if k.APIKey == "" || k.APISecret == "" {
		return fmt.Errorf("Kraken API keys not set")
	}

	urlPath := "/0/private/AddOrder"
	nonce := fmt.Sprintf("%d", time.Now().UnixNano())

	values := url.Values{}
	values.Set("nonce", nonce)
	values.Set("pair", symbol)
	values.Set("type", strings.ToLower(side))
	values.Set("ordertype", strings.ToLower(orderType))
	values.Set("volume", strconv.FormatFloat(quantity, 'f', 8, 64))

	signature := k.sign(urlPath, values)

	req, _ := http.NewRequest("POST", k.BaseURL+urlPath, strings.NewReader(values.Encode()))
	req.Header.Set("API-Key", k.APIKey)
	req.Header.Set("API-Sign", signature)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := k.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Kraken API error %d: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("[Kraken] Successfully submitted %s order for %s\n", side, symbol)
	return nil
}
