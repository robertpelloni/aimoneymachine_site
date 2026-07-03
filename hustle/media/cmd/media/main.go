package main

import (
	"flag"
	"fmt"
	"github.com/robertpelloni/hustle/hustle/media"
	"github.com/robertpelloni/hustle/orchestrator"
)

func main() {
	topic := flag.String("topic", "AI side hustles", "Topic for YouTube Shorts")
	flag.Parse()

	orch := orchestrator.NewOrchestrator()

	factory := media.NewShortsFactory(orch)
	if err := factory.Run(*topic); err != nil {
		fmt.Printf("Media pipeline failed: %v\n", err)
	} else {
		fmt.Println("Media pipeline complete.")
	}
}
