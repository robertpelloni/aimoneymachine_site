package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: cluster <node_count>")
		os.Exit(1)
	}

	nodeCount, _ := strconv.Atoi(os.Args[1])
	basePort := 8080

	fmt.Printf("🚀 Launching %d-node AI Hustle Machine Cluster...\n", nodeCount)

	for i := 0; i < nodeCount; i++ {
		port := basePort + i
		nodeID := fmt.Sprintf("node-%d", i)
		dbFile := fmt.Sprintf("hustle-%d.db", i)
		memFile := fmt.Sprintf("memory-%d.json", i)

		// Seed logic: First node is the seed
		seedFlag := ""
		if i > 0 {
			seedFlag = fmt.Sprintf("-seed http://localhost:%d", basePort)
		}

		fmt.Printf("[Cluster] Starting %s on port %d...\n", nodeID, port)

		// Run orchestrator in API mode
		cmd := exec.Command("./bin/orchestrator",
			"-api", strconv.Itoa(port),
			"-dry-run", // Safety first for clusters
		)

		// Set environment variables for isolation
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("NODE_ID=%s", nodeID),
			fmt.Sprintf("DB_FILE=%s", dbFile),
			fmt.Sprintf("MEMORY_FILE=%s", memFile),
		)

		if seedFlag != "" {
			cmd.Args = append(cmd.Args, "-seed", fmt.Sprintf("http://localhost:%d", basePort))
		}

		err := cmd.Start()
		if err != nil {
			fmt.Printf("[Cluster] ❌ Failed to start %s: %v\n", nodeID, err)
			continue
		}

		fmt.Printf("[Cluster] Node %s (PID %d) is UP.\n", nodeID, cmd.Process.Pid)
		time.Sleep(1 * time.Second) // Staggered startup
	}

	fmt.Println("\n[Cluster] All nodes launched. Press Ctrl+C to terminate all (manual cleanup needed for now).")
	select {}
}
