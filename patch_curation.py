import re

with open("orchestrator/cmd/orchestrator/main.go", "r") as f:
    content = f.read()

# Make sure affiliate package is imported
if '"aimoneymachine/orchestrator/internal/hustle/affiliate"' not in content:
    content = content.replace(
        '"github.com/robertpelloni/hustle/hustle/curation"',
        '"github.com/robertpelloni/hustle/hustle/curation"\n\t"github.com/robertpelloni/hustle/hustle/affiliate"'
    )

new_curation = """	protocol.Register("curation", func(p url.Values) error {
		topic := p.Get("topic")
		if topic == "" {
			topic = "AI"
		}
		feeds := orch.RSSFeeds
		if len(feeds) == 0 {
			feeds = []string{"https://news.ycombinator.com/rss"}
		}
		c := &curation.CurationModule{
			Orchestrator: orch,
			Fetcher:      curation.NewRSSFetcher(),
			Feeds:        feeds,
		}

		err := c.Curate(topic)
		if err == nil {
			// Find the summary from L1 memory
			entries := orch.L1.Search("curation")
			if len(entries) > 0 {
				lastEntry := entries[len(entries)-1]

				// Inject Affiliate Links
				affModule := affiliate.NewModule(orch)
				injectedSummary := affModule.InjectAffiliateLink(lastEntry.Content, topic)

				// Re-save the injected summary to L1 so social module can pick it up
				orch.L1.Add(orchestrator.MemoryEntry{
					ID:        fmt.Sprintf("curation-affiliate-%s-%d", topic, time.Now().Unix()),
					Content:   injectedSummary,
					Timestamp: time.Now(),
					Tags:      []string{"curation", topic, "monetized"},
				})
				fmt.Printf("[Curation] Monetized Summary:\\n%s\\n", injectedSummary)
			}
		}
		return err
	})"""

# replace \n with actual string for fmt.Printf
new_curation = new_curation.replace("\\n", "\\\\n")

content = re.sub(r'protocol\.Register\("curation", func\(p url\.Values\) error \{.*?return c\.Curate\(topic\)\n\t\}\)', new_curation, content, flags=re.DOTALL)

with open("orchestrator/cmd/orchestrator/main.go", "w") as f:
    f.write(content)
