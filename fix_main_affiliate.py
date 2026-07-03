import re

with open("orchestrator/cmd/orchestrator/main.go", "r") as f:
    content = f.read()

# First, modify the protocol.Register("social", ...)
# Find the social.SchedulePost call
search_str = """
		if contentStr != "" {
			fmt.Printf("[Social] Posting explicit content to %s: %s\\n", platform, contentStr)
			return provider.Post(orch, platform, contentStr)
		}

		social.SchedulePost(orch, provider, platform, topic)
		return nil
"""
replace_str = """
		if contentStr != "" {
			fmt.Printf("[Social] Posting explicit content to %s: %s\\n", platform, contentStr)
			return provider.Post(orch, platform, contentStr)
		}

		generatedContent := social.GenerateContent(orch, topic)

		// Inject affiliate link
		affModule := affiliate.NewModule(orch)
		generatedContent = affModule.InjectAffiliateLink(generatedContent, topic)

		generatedContent = social.FormatForPlatform(generatedContent, platform)
		fmt.Printf("Scheduling post for %s: %s\\n", platform, generatedContent)

		err := provider.Post(orch, platform, generatedContent)
		if err == nil {
			fmt.Printf("[Social] ✅ Successfully posted to %s\\n", platform)
			orch.L1.Add(orchestrator.MemoryEntry{
				ID:        fmt.Sprintf("social-%s-%d", platform, time.Now().Unix()),
				Content:   fmt.Sprintf("Posted to %s: %s", platform, generatedContent),
				Timestamp: time.Now(),
				Tags:      []string{"social", platform},
			})

			orch.Ledger.Add(orchestrator.Transaction{
				Amount: 0.01,
				Type:   orchestrator.Expense,
				Hustle: "SocialMedia",
				Note:   fmt.Sprintf("API post to %s", platform),
			})
		} else {
			fmt.Printf("[Social] ❌ Failed to post to %s: %v\\n", platform, err)
		}

		return err
"""
content = content.replace(search_str, replace_str)

# Ensure affiliate import is added
if '"github.com/robertpelloni/hustle/hustle/affiliate"' not in content:
    content = content.replace('"github.com/robertpelloni/hustle/hustle/social"', '"github.com/robertpelloni/hustle/hustle/affiliate"\n\t"github.com/robertpelloni/hustle/hustle/social"')

with open("orchestrator/cmd/orchestrator/main.go", "w") as f:
    f.write(content)
