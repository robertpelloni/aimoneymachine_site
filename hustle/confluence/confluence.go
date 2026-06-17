package confluence

import (
	"fmt"
	"github.com/robertpelloni/hustle/orchestrator"
	"time"
)

type Recipe string

const (
	SaaSGrowth    Recipe = "saas_growth"
	DropshipLoop  Recipe = "dropship_loop"
	TradingAlpha  Recipe = "trading_alpha"
)

type ConfluenceEngine struct {
	Orch   *orchestrator.Orchestrator
	Broker *orchestrator.A2ABroker
}

func NewConfluenceEngine(orch *orchestrator.Orchestrator, broker *orchestrator.A2ABroker) *ConfluenceEngine {
	return &ConfluenceEngine{
		Orch:   orch,
		Broker: broker,
	}
}

// ExecuteRecipe runs a synergistic sequence of hustle modules
func (e *ConfluenceEngine) ExecuteRecipe(recipe Recipe) error {
	fmt.Printf("[Confluence] Executing Strategic Recipe: %s\n", recipe)

	switch recipe {
	case SaaSGrowth:
		return e.runSaaSGrowth()
	case DropshipLoop:
		return e.runDropshipLoop()
	case TradingAlpha:
		return e.runTradingAlpha()
	default:
		return fmt.Errorf("unknown recipe: %s", recipe)
	}
}

func (e *ConfluenceEngine) runSaaSGrowth() error {
	fmt.Println("[Confluence] Step 1: Researching SaaS market gaps...")
	e.Orch.Protocol.HandleURI("hustle://research?query=profitable+micro+saas+niche+2026")

	fmt.Println("[Confluence] Step 2: Ideating Product...")
	e.Orch.Protocol.HandleURI("hustle://saas?action=ideate&niche=productivity")

	fmt.Println("[Confluence] Step 3: Generating Marketing Content...")
	e.Orch.Protocol.HandleURI("hustle://content?topic=Why+productivity+tools+are+essential&type=blog")

	e.logConfluence(SaaSGrowth, "Successfully orchestrated Research -> SaaS -> Marketing pipeline.")
	return nil
}

func (e *ConfluenceEngine) runDropshipLoop() error {
	fmt.Println("[Confluence] Step 1: Discovering trending products...")
	e.Orch.Protocol.HandleURI("hustle://ecommerce?action=discover&niche=smart+home")

	fmt.Println("[Confluence] Step 2: Localizing for global markets...")
	e.Orch.Protocol.HandleURI("hustle://localization?action=translate&lang=Spanish&text=Smart+Home+Essentials")

	fmt.Println("[Confluence] Step 3: Launching Ads...")
	e.Orch.Protocol.HandleURI("hustle://ecommerce?action=ads&platform=Instagram")

	e.logConfluence(DropshipLoop, "Successfully orchestrated Ecom -> Localization -> Ads pipeline.")
	return nil
}

func (e *ConfluenceEngine) runTradingAlpha() error {
	fmt.Println("[Confluence] Step 1: Scanning for Market Alpha...")
	e.Orch.Protocol.HandleURI("hustle://research?query=crypto+sentiment+btc+eth")

	fmt.Println("[Confluence] Step 2: Executing Sentiment-Driven Trade...")
	e.Orch.Protocol.HandleURI("hustle://trading?action=arbitrage&symbol=BTC")

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
