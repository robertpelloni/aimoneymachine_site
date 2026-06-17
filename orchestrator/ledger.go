package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type TransactionType string

const (
	Revenue TransactionType = "Revenue"
	Expense TransactionType = "Expense"
)

type Transaction struct {
	ID        string          `json:"id"`
	Amount    float64         `json:"amount"`
	Type      TransactionType `json:"type"`
	Hustle    string          `json:"hustle"`
	Note      string          `json:"note"`
	Timestamp time.Time       `json:"timestamp"`
}

type Ledger struct {
	Transactions []Transaction `json:"transactions"`
}

func (l *Ledger) Add(t Transaction) {
	if t.ID == "" {
		t.ID = fmt.Sprintf("txn-%d", time.Now().UnixNano())
	}
	if t.Timestamp.IsZero() {
		t.Timestamp = time.Now()
	}
	l.Transactions = append(l.Transactions, t)
}

func (l *Ledger) TotalRevenue() float64 {
	total := 0.0
	for _, t := range l.Transactions {
		if t.Type == Revenue {
			total += t.Amount
		}
	}
	return total
}

func (l *Ledger) TotalExpenses() float64 {
	total := 0.0
	for _, t := range l.Transactions {
		if t.Type == Expense {
			total += t.Amount
		}
	}
	return total
}

func (l *Ledger) Profit() float64 {
	return l.TotalRevenue() - l.TotalExpenses()
}

func (l *Ledger) AnalyzeProfitability() string {
	if len(l.Transactions) == 0 {
		return "Insufficient data for financial analysis."
	}

	hustleProfits := make(map[string]float64)
	for _, t := range l.Transactions {
		if t.Type == Revenue {
			hustleProfits[t.Hustle] += t.Amount
		} else {
			hustleProfits[t.Hustle] -= t.Amount
		}
	}

	bestHustle := ""
	maxProfit := -1000000.0
	worstHustle := ""
	minProfit := 1000000.0

	for h, p := range hustleProfits {
		if p > maxProfit {
			maxProfit = p
			bestHustle = h
		}
		if p < minProfit {
			minProfit = p
			worstHustle = h
		}
	}

	if minProfit < 0 {
		return fmt.Sprintf("CRITICAL ROI WARNING: Underperforming hustle detected: %s (Deficit: $%.2f). Action: Terminate immediately to preserve wealth. Best Hustle: %s (Profit: $%.2f).", worstHustle, -minProfit, bestHustle, maxProfit)
	}

	return fmt.Sprintf("Best Performing Hustle: %s (Profit: $%.2f). Recommendation: Increase agent allocation to %s.", bestHustle, maxProfit, bestHustle)
}

func (l *Ledger) Save(filepath string) error {
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0644)
}

func (l *Ledger) Load(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, l)
}

// TransferCapital handles mesh-wide capital reallocation (Self-Funding Swarm)
func (l *Ledger) TransferCapital(amount float64, targetPeerID string, note string) {
	l.Add(Transaction{
		Amount:    amount,
		Type:      Expense,
		Hustle:    "swarm_funding",
		Note:      fmt.Sprintf("Capital transfer to %s: %s", targetPeerID, note),
		Timestamp: time.Now(),
	})
	fmt.Printf("[Ledger] Authorized $%.2f capital transfer to peer %s\n", amount, targetPeerID)
}

// ReceiveCapital handles inbound mesh funding
func (l *Ledger) ReceiveCapital(amount float64, sourcePeerID string, note string) {
	l.Add(Transaction{
		Amount:    amount,
		Type:      Revenue,
		Hustle:    "swarm_funding",
		Note:      fmt.Sprintf("Inbound funding from %s: %s", sourcePeerID, note),
		Timestamp: time.Now(),
	})
	fmt.Printf("[Ledger] Received $%.2f capital injection from peer %s\n", amount, sourcePeerID)
}
