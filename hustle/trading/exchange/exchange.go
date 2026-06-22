// Package exchange defines a standard interface for cryptocurrency exchange integrations.
package exchange

import (
	"time"
)

// OrderSide indicates buy or sell.
type OrderSide string

const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"
)

// OrderType indicates market or limit order.
type OrderType string

const (
	Market OrderType = "MARKET"
	Limit  OrderType = "LIMIT"
)

// OrderStatus represents the current state of an order.
type OrderStatus string

const (
	OrderNew           OrderStatus = "NEW"
	OrderPartiallyFiel OrderStatus = "PARTIALLY_FILLED"
	OrderFilled        OrderStatus = "FILLED"
	OrderCanceled      OrderStatus = "CANCELED"
	OrderRejected      OrderStatus = "REJECTED"
	OrderExpired       OrderStatus = "EXPIRED"
)

// Ticker represents the current price snapshot for a symbol.
type Ticker struct {
	Symbol      string    `json:"symbol"`
	BidPrice    float64   `json:"bid_price"`
	AskPrice    float64   `json:"ask_price"`
	LastPrice   float64   `json:"last_price"`
	Volume      float64   `json:"volume"`
	HighPrice   float64   `json:"high_price"`
	LowPrice    float64   `json:"low_price"`
	PriceChange float64   `json:"price_change"`
	Timestamp   time.Time `json:"timestamp"`
}

// OrderBookEntry represents a single level in the order book.
type OrderBookEntry struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

// OrderBook represents the full order book snapshot.
type OrderBook struct {
	Symbol    string            `json:"symbol"`
	Bids      []OrderBookEntry  `json:"bids"`
	Asks      []OrderBookEntry  `json:"asks"`
	Timestamp time.Time         `json:"timestamp"`
}

// Order represents a placed order.
type Order struct {
	ID            string      `json:"id"`
	ClientOrderID string      `json:"client_order_id"`
	Symbol        string      `json:"symbol"`
	Side          OrderSide   `json:"side"`
	Type          OrderType   `json:"type"`
	Status        OrderStatus `json:"status"`
	Price         float64     `json:"price"`
	Quantity      float64     `json:"quantity"`
	ExecutedQty   float64     `json:"executed_qty"`
	QuoteOrderQty float64     `json:"quote_order_qty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// Balance represents a single asset balance.
type Balance struct {
	Asset      string  `json:"asset"`
	Free       float64 `json:"free"`
	Locked     float64 `json:"locked"`
	Total      float64 `json:"total"` // convenience: free + locked
}

// Account represents the full account snapshot.
type Account struct {
	Balances        []Balance `json:"balances"`
	CanTrade        bool      `json:"can_trade"`
	CanWithdraw     bool      `json:"can_withdraw"`
	CanDeposit      bool      `json:"can_deposit"`
}

// ExchangeProvider defines the interface for interacting with a cryptocurrency exchange.
type ExchangeProvider interface {
	// Name returns the exchange name (e.g. "binance", "kraken").
	Name() string

	// GetTicker fetches the current ticker for a symbol (e.g. "BTCUSDT").
	GetTicker(symbol string) (*Ticker, error)

	// GetOrderBook fetches the order book for a symbol with a given depth limit.
	GetOrderBook(symbol string, limit int) (*OrderBook, error)

	// PlaceOrder places an order on the exchange.
	PlaceOrder(symbol string, side OrderSide, orderType OrderType, quantity, price float64) (*Order, error)

	// CancelOrder cancels an open order by ID.
	CancelOrder(symbol, orderID string) error

	// GetOrderStatus retrieves the status of an order.
	GetOrderStatus(symbol, orderID string) (*Order, error)

	// GetAccount fetches the account balances and permissions.
	GetAccount() (*Account, error)

	// GetBalance fetches the balance for a single asset.
	GetBalance(asset string) (*Balance, error)

	// SubscribeTrades streams real-time trade data for symbols.
	// Returns a channel of tickers and a close function.
	// Call the close function to stop the stream.
	SubscribeTickers(symbols ...string) (<-chan *Ticker, func(), error)

	// DryRun returns whether the provider is in dry-run mode.
	DryRun() bool

	// SetDryRun enables or disables dry-run mode.
	SetDryRun(dry bool)
}
