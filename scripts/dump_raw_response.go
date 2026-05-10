//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("📄 Raw FotMob API Response Dumper")
		fmt.Println("")
		fmt.Println("Usage: go run scripts/dump_raw_response.go <match_id>")
		fmt.Println("Example: go run scripts/dump_raw_response.go 4813581")
		fmt.Println("")
		fmt.Println("This tool dumps the complete raw JSON response from FotMob")
		fmt.Println("Use this to inspect the exact API structure")
		os.Exit(1)
	}

	matchIDStr := os.Args[1]
	fmt.Printf("📄 Dumping raw FotMob API response for match ID: %s\n\n", matchIDStr)

	url := fmt.Sprintf("https://www.fotmob.com/api/matchDetails?matchId=%s", matchIDStr)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("❌ Error creating request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Scoreline/1.0)")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ HTTP %d: %s\n", resp.StatusCode, resp.Status)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Pretty print the JSON
	var raw interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		fmt.Printf("❌ Error parsing JSON: %v\n", err)
		fmt.Println("Raw response:")
		fmt.Println(string(body))
		os.Exit(1)
	}

	pretty, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error pretty-printing JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🔍 Complete FotMob API Response:")
	fmt.Println("=====================================")
	fmt.Println(string(pretty))

	// Extract and highlight highlights section
	fmt.Println("\n🎯 Highlights Section Analysis:")
	fmt.Println("=====================================")

	// Navigate to highlights
	rawMap, ok := raw.(map[string]interface{})
	if !ok {
		fmt.Println("❌ Response root is not an object")
		return
	}

	content, ok := rawMap["content"]
	if !ok {
		fmt.Println("❌ No 'content' field")
		return
	}

	contentMap, ok := content.(map[string]interface{})
	if !ok {
		fmt.Println("❌ 'content' is not an object")
		return
	}

	matchFacts, ok := contentMap["matchFacts"]
	if !ok {
		fmt.Println("❌ No 'matchFacts' field in content")
		return
	}

	matchFactsMap, ok := matchFacts.(map[string]interface{})
	if !ok {
		fmt.Println("❌ 'matchFacts' is not an object")
		return
	}

	highlights, exists := matchFactsMap["highlights"]
	if !exists {
		fmt.Println("❌ No 'highlights' field in matchFacts")
		fmt.Println("💡 FotMob API does not provide highlights for this match")
		return
	}

	if highlights == nil {
		fmt.Println("⚠️  'highlights' field exists but is null")
		return
	}

	fmt.Println("✅ Found 'highlights' field with data")

	highlightsMap, ok := highlights.(map[string]interface{})
	if !ok {
		fmt.Printf("❌ 'highlights' is not an object (type: %T)\n", highlights)
		return
	}

	fmt.Println("\n📋 Highlights Data:")
	for key, value := range highlightsMap {
		fmt.Printf("   %s: %v\n", key, value)
	}

	// Pretty print just the highlights section
	fmt.Println("\n📄 Highlights JSON:")
	highlightsJSON, _ := json.MarshalIndent(highlightsMap, "", "  ")
	fmt.Printf("%s\n", string(highlightsJSON))
}
