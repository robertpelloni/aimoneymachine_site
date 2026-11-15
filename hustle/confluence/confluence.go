package confluence

import (
	"fmt"
	"time"

	"github.com/robertpelloni/hustle/orchestrator"
)

type Recipe string

const (
	SaaSGrowth   Recipe = "saas_growth"
	DropshipLoop Recipe = "dropship_loop"
	TradingAlpha Recipe = "trading_alpha"
)

// ConfluenceEngine executes multi-module synergistic recipes via HustleProtocol.
type ConfluenceEngine struct {
	Orch     *orchestrator.Orchestrator
	Protocol *orchestrator.HustleProtocol
	Broker   *orchestrator.A2ABroker
}

func NewConfluenceEngine(orch *orchestrator.Orchestrator, protocol *orchestrator.HustleProtocol, broker *orchestrator.A2ABroker) *ConfluenceEngine {
	return &ConfluenceEngine{
		Orch:     orch,
		Protocol: protocol,
		Broker:   broker,
	}
}

func (e *ConfluenceEngine) ExecuteRecipe(recipe Recipe) error {
	fmt.Printf("[Confluence] Executing Strategic Recipe: %s\n", recipe)
	if e.Protocol == nil {
		return fmt.Errorf("confluence protocol router not initialized")
	}

	switch recipe {
	case SaaSGrowth:
		return e.runSaaSGrowth()
	case DropshipLoop:
		return e.runDropshipLoop()
	case TradingAlpha:
		return e.runTradingAlpha()
	default:
		return fmt.Errorf("unknown confluence recipe: %s (expected saas_growth, dropship_loop, or trading_alpha)", recipe)
	}
}

func (e *ConfluenceEngine) runSaaSGrowth() error {
	steps := []string{
		"hustle://research?query=profitable+micro+saas+niche+2026",
		"hustle://saas?action=ideate&niche=productivity",
		"hustle://content?topic=Why+AI+productivity+tools+are+essential&type=blog",
	}
	for _, s := range steps {
		if err := e.Protocol.HandleURI(s); err != nil {
			return err
		}
	}
	e.logConfluence(SaaSGrowth, "Successfully orchestrated Research -> SaaS -> Content pipeline.")
	return nil
}

func (e *ConfluenceEngine) runDropshipLoop() error {
	steps := []string{
		"hustle://ecommerce?action=discover&niche=smart+home",
		"hustle://localization?action=translate&lang=Spanish&text=Smart+Home+Essentials",
		"hustle://ecommerce?action=ads&platform=Instagram&niche=smart+home",
	}
	for _, s := range steps {
		if err := e.Protocol.HandleURI(s); err != nil {
			return err
		}
	}
	e.logConfluence(DropshipLoop, "Successfully orchestrated Ecommerce -> Localization -> Ads pipeline.")
	return nil
}

func (e *ConfluenceEngine) runTradingAlpha() error {
	steps := []string{
		"hustle://research?query=crypto+sentiment+btc+eth",
		"hustle://trading?symbol=BTC",
	}
	for _, s := range steps {
		if err := e.Protocol.HandleURI(s); err != nil {
			return err
		}
	}
	e.logConfluence(TradingAlpha, "Successfully orchestrated Research -> Trading pipeline.")
	return nil
}

func (e *ConfluenceEngine) logConfluence(recipe Recipe, result string) {
	e.Orch.L1.Add(orchestrator.MemoryEntry{
		ID:        fmt.Sprintf("confluence-%s-%d", recipe, time.Now().Unix()),
		Content:   fmt.Sprintf("Confluence Recipe [%s] Result: %s", recipe, result),
		Timestamp: time.Now(),
		Tags:      []string{"confluence", string(recipe), "strategy"},
	})
}
