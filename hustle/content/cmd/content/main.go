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
	siteDir := flag.String("site-dir", "output/content", "Source directory for site generation")
	siteOutput := flag.String("site-output", "output/site", "Output directory for static site")
	deploySite := flag.Bool("deploy", false, "Generate static HTML site from markdown content")
	publishTarget := flag.String("publish-target", "", "Deploy target: wordpress or github (requires env vars)")
	servePort := flag.Int("serve", 0, "Start local HTTP server on this port after generating site")
	flag.Parse()

	orch := orchestrator.NewOrchestrator()
	mod := content.NewContentModule(orch, *outputDir)

	// Deploy mode: generate static site
	if *deploySite {
		config := content.DefaultSiteConfig()
		config.OutputDir = *siteOutput

		fmt.Printf("\n🏗️  Generating static site...\n")
		fmt.Printf("   Source: %s\n", *siteDir)
		fmt.Printf("   Output: %s\n", *siteOutput)

		if err := content.GenerateSite(*siteDir, config); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Optional deployment after site generation
		if *publishTarget != "" {
			fmt.Printf("\n🚀 Deploying to %s...\n", *publishTarget)
			if err := content.DeploySite(*siteOutput, *publishTarget); err != nil {
				fmt.Printf("Deployment error: %v\n", err)
				return
			}
			fmt.Println("✅ Deployment complete!")
		}

		if *servePort > 0 {
			fmt.Printf("\n🌐 Starting server on port %d...\n", *servePort)
			if err := content.ServeSite(*siteOutput, *servePort); err != nil {
				fmt.Printf("Server error: %v\n", err)
			}
		}
		return
	}

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
