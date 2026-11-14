package orchestrator

import (
	"encoding/json"
	"fmt"
	"time"
)

type StatusReport struct {
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Revenue   float64   `json:"revenue"`
	Expenses  float64   `json:"expenses"`
	Profit    float64   `json:"profit"`
}

func WriteStatusReport(version, status, message string, ledger Ledger) error {
	report := StatusReport{
		Version:   version,
		Timestamp: time.Now(),
		Status:    status,
		Message:   message,
		Revenue:   ledger.TotalRevenue(),
		Expenses:  ledger.TotalExpenses(),
		Profit:    ledger.Profit(),
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	err = WriteFileAtomic("STATUS.json", data, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("Status report written to STATUS.json: %s (Profit: $%.2f)\n", status, report.Profit)
	return nil
}
