package main

import (
	"flag"
	"fmt"
	"github.com/robertpelloni/hustle/hustle/affiliate"
	"github.com/robertpelloni/hustle/orchestrator"
)

func main() {
	niche := flag.String("niche", "AI automation tools", "Niche to run affiliate pipeline for")
	flag.Parse()

	orch := orchestrator.NewOrchestrator()
	// orch.LoadConfig(".env") // Assuming basic config load

	module := affiliate.NewAffiliateModule(orch)
	if err := module.Run(*niche); err != nil {
		fmt.Printf("Affiliate pipeline failed: %v\n", err)
	} else {
		fmt.Println("Affiliate pipeline complete.")
	}
}
