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
