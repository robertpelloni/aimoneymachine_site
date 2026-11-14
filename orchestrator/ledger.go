package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
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
	// mu guards Transactions: scheduler tasks, API handlers and the
	// agent loop now touch the ledger from multiple goroutines.
	mu sync.RWMutex
}

func (l *Ledger) Add(t Transaction) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if t.ID == "" {
		t.ID = fmt.Sprintf("txn-%d", time.Now().UnixNano())
	}
	if t.Timestamp.IsZero() {
		t.Timestamp = time.Now()
	}
	l.Transactions = append(l.Transactions, t)
}

func (l *Ledger) TotalRevenue() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.totalRevenueLocked()
}

func (l *Ledger) totalRevenueLocked() float64 {
	total := 0.0
	for _, t := range l.Transactions {
		if t.Type == Revenue {
			total += t.Amount
		}
	}
	return total
}

func (l *Ledger) TotalExpenses() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.totalExpensesLocked()
}

func (l *Ledger) totalExpensesLocked() float64 {
	total := 0.0
	for _, t := range l.Transactions {
		if t.Type == Expense {
			total += t.Amount
		}
	}
	return total
}

func (l *Ledger) Profit() float64 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.totalRevenueLocked() - l.totalExpensesLocked()
}

func (l *Ledger) AnalyzeProfitability() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
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
	l.mu.RLock()
	defer l.mu.RUnlock()
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return WriteFileAtomic(filepath, data, 0644)
}

func (l *Ledger) Load(filepath string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, l)
}
