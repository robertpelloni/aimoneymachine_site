package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/robertpelloni/hustle/hustle/content"
	"github.com/robertpelloni/hustle/orchestrator"
)

func main() {
	topic := flag.String("topic", "AI agents and automation", "Content topic")
	niche := flag.String("niche", "AI & automation", "Target niche")
	contentType := flag.String("type", "blog", "Content type: blog, newsletter, seo, thread")
	keywords := flag.String("keywords", "AI,automation,agents,2026", "Comma-separated SEO keywords")
	outputDir := flag.String("output", "output/content", "Output directory for generated content")
	brainstorm := flag.Bool("ideas", false, "Just generate topic ideas instead of content")
	flag.Parse()

	orch := orchestrator.NewOrchestrator()
	mod := content.NewContentModule(orch, *outputDir)

	if *brainstorm {
		fmt.Printf("Generating topic ideas for niche: %s\n", *niche)
		ideas, err := mod.GenerateTopicIdeas(*niche, 10)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("\n📊 Topic Ideas:")
		for i, idea := range ideas {
			fmt.Printf("  %d. %s\n", i+1, idea)
		}
		return
	}

	kwList := strings.Split(*keywords, ",")
	for i := range kwList {
		kwList[i] = strings.TrimSpace(kwList[i])
	}

	ct := content.ContentType(*contentType)
	req := content.ContentRequest{
		Topic:       *topic,
		Type:        ct,
		Keywords:    kwList,
		TargetWords: 800,
		Niche:       *niche,
	}

	result, err := mod.Generate(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\n📝 Generated: %s\n", result.Title)
	fmt.Printf("   Type: %s | Words: ~%d\n", result.Type, len(strings.Fields(result.Body)))
	if result.Filepath != "" {
		fmt.Printf("   Saved: %s\n", result.Filepath)
	}
}
