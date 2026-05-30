package main

import (
	"fmt"
	"github.com/robertpelloni/hustle/hustle/curation"
	"github.com/robertpelloni/hustle/orchestrator"
)

func main() {
	fmt.Println("=== Content Curation Hustle Module ===")
	orch := orchestrator.NewOrchestrator()

	curator := &curation.CurationModule{Orchestrator: orch}
	err := curator.Curate("Artificial Intelligence")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Println("Module execution complete.")
}
