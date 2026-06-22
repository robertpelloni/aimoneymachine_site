package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/robertpelloni/freellm/internal/config"
	"github.com/robertpelloni/freellm/internal/db"
	"github.com/robertpelloni/freellm/internal/engine"
	"github.com/robertpelloni/freellm/internal/proxy"
)

func main() {
	log.Println("[FreeLLM] Starting headless server on :4000")

	database, err := db.InitDB()
	if err != nil {
		log.Printf("[DB] Warning: %v (running without DB)\n", err)
	}

	cfg, err := config.LoadConfig("freellm-config.yaml")
	if err != nil {
		log.Printf("Warning: freellm-config.yaml not found, using defaults: %v", err)
		cfg = &config.Config{Port: 4000}
	}

	eventLogger := engine.NewEventLogger(100, database)

	apiKeys := map[string]string{}
	providers := []string{"openai", "anthropic", "google", "groq", "deepseek", "together", "nvidia_nim",
		"perplexity", "replicate", "hyperbolic", "cloudflare", "codestral", "siliconflow",
		"novita", "nebius", "ai21", "minimax", "moonshot", "stepfun", "zhipu", "internlm", "arcee", "dashscope"}
	envMap := map[string]string{
		"openai": "OPENAI_API_KEY", "anthropic": "ANTHROPIC_API_KEY", "google": "GOOGLE_API_KEY",
		"groq": "GROQ_API_KEY", "deepseek": "DEEPSEEK_API_KEY", "together": "TOGETHER_API_KEY",
		"nvidia_nim": "NVIDIA_API_KEY", "perplexity": "PERPLEXITY_API_KEY", "replicate": "REPLICATE_API_TOKEN",
		"hyperbolic": "HYPERBOLIC_API_KEY", "cloudflare": "CLOUDFLARE_API_KEY", "codestral": "CODESTRAL_API_KEY",
		"siliconflow": "SILICONFLOW_API_KEY", "novita": "NOVITA_API_KEY", "nebius": "NEBIUS_API_KEY",
		"ai21": "AI21_API_KEY", "minimax": "MINIMAX_API_KEY", "moonshot": "MOONSHOT_API_KEY",
		"stepfun": "STEPFUN_API_KEY", "zhipu": "ZHIPU_API_KEY", "internlm": "INTERNLM_API_KEY",
		"arcee": "ARCEE_API_KEY", "dashscope": "DASHSCOPE_API_KEY",
	}
	for _, p := range providers {
		if key := os.Getenv(envMap[p]); key != "" {
			apiKeys[p] = key
		}
	}
	log.Printf("API keys configured: %d/%d providers\n", len(apiKeys), len(providers))

	benchmarker := engine.NewBenchmarker(apiKeys, 100, eventLogger)
	go benchmarker.Run()

	port := cfg.Port
	if port == 0 {
		port = 4000
	}
	gateway := proxy.NewGateway(100, database, port)
	gateway.UpdateModels(benchmarker.Rankings())

	log.Printf("[FreeLLM] Listening on :%d\n", port)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("[FreeLLM] Shutting down...")
}
