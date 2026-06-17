package ecommerce

import (
	"strings"
	"github.com/robertpelloni/hustle/orchestrator"
)

// ProductFromMemory reconstructs a Product struct from a memory entry
func ProductFromMemory(e orchestrator.MemoryEntry) Product {
	p := Product{
		Name: "Unknown Product",
	}
	// Memory content looks like: "Trending Product Discovery: Name (Niche: Niche, Demand: 9.5)"
	if strings.Contains(e.Content, "Product Discovery:") {
		parts := strings.Split(e.Content, ": ")
		if len(parts) > 1 {
			nameParts := strings.Split(parts[1], " (")
			p.Name = nameParts[0]
		}
	}
	return p
}
