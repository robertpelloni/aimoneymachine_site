package trading

import (
	"strings"

	"github.com/robertpelloni/hustle/hustle/trading/exchange/coinbase"
)

// CoinbasePriceFetcher adapts coinbase.Provider to trading.PriceFetcher.
type CoinbasePriceFetcher struct {
	provider *coinbase.Provider
}

func NewCoinbasePriceFetcher() *CoinbasePriceFetcher {
	return &CoinbasePriceFetcher{
		provider: coinbase.NewProvider(),
	}
}

func (f *CoinbasePriceFetcher) GetPrice(symbol string) (float64, error) {
	pair := strings.ToUpper(strings.TrimSpace(symbol))
	if !strings.Contains(pair, "-") {
		pair = pair + "-USD"
	}
	ticker, err := f.provider.GetTicker(pair)
	if err != nil {
		return 0, err
	}
	return ticker.LastPrice, nil
}
