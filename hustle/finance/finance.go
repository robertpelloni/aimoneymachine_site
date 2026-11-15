package finance

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type Transaction struct {
	Date        string  `json:"date"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
}

type FinanceModule struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewFinanceModule(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *FinanceModule {
	return &FinanceModule{
		Orch:   orch,
		Broker: broker,
	}
}

// ClassifyTransactions uses LLM to categorize financial data for bookkeeping
func (f *FinanceModule) ClassifyTransactions(rawCSV string) ([]Transaction, error) {
	fmt.Printf("[Finance] Classifying transactions from raw input\n")

	prompt := fmt.Sprintf(`Act as a professional bookkeeper. Categorize the following raw transaction data into standard business categories (e.g. Software, Marketing, Inventory, Travel).

DATA:
%s

Respond with a JSON array of objects:
[
  {
    "date": "YYYY-MM-DD",
    "description": "Original desc",
    "amount": 100.0,
    "category": "Category Name"
  }
]

Respond ONLY with valid JSON.`, rawCSV)

	var transactions []Transaction
	if err := f.Orch.LLM.GenerateJSON(prompt, &transactions); err != nil {
		return nil, err
	}

	// Store in memory
	f.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("finance-%d", time.Now().Unix()),
		Content:   fmt.Sprintf("Bookkeeping Task Completed: %d transactions categorized.", len(transactions)),
		Timestamp: time.Now(),
		Tags:      []string{"finance", "bookkeeping", "service"},
	})

	return transactions, nil
}

// GenerateTaxSummary provides a summary of categorized transactions for tax purposes
func (f *FinanceModule) GenerateTaxSummary(transactions []Transaction) (string, error) {
	var combined string
	for _, t := range transactions {
		combined += fmt.Sprintf("- %s: $%.2f (%s)\n", t.Description, t.Amount, t.Category)
	}

	prompt := fmt.Sprintf(`Analyze these business transactions and generate a high-level tax deduction summary.

TRANSACTIONS:
%s

Requirements:
- Total spend by category
- Identify potential tax-deductible business expenses
- Brief advice on record keeping

Format: Professional Markdown.`, combined)

	return f.Orch.LLM.Generate(prompt)
}
