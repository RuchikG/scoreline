//go:build ignore

// test_commentary.go - Debug script to inspect FotMob commentary data for a single match
//
// This script fetches raw match data and looks for commentary-related fields.
// Use this to investigate what commentary data FotMob provides.
//
// Usage:
//   go run scripts/test_commentary.go <match_id>
//
// Examples:
//   go run scripts/test_commentary.go 4513498    # Specific match ID
//
// To find match IDs, use the api_results_raw.go script first to list matches.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://www.fotmob.com/api"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	matchID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("Error: Invalid match_id: %s (must be a number)\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║          FotMob Commentary Inspector                         ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("Match ID: %d\n", matchID)
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch raw match details
	rawResponse, err := fetchRawMatchDetails(ctx, matchID)
	if err != nil {
		fmt.Printf("❌ Error fetching match details: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Successfully fetched match details")
	fmt.Println()

	// Display match info
	displayMatchInfo(rawResponse)

	// Look for commentary-related fields
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("SEARCHING FOR COMMENTARY FIELDS")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println()

	commentaryFields := findCommentaryFields(rawResponse, "")
	if len(commentaryFields) == 0 {
		fmt.Println("❌ No commentary-related fields found in the response.")
		fmt.Println()
		fmt.Println("The following field names were searched:")
		fmt.Println("  - commentary, comment, comments")
		fmt.Println("  - liveTicker, ticker, updates")
		fmt.Println("  - timeline, incidents")
		fmt.Println("  - text, description, narrative")
	} else {
		fmt.Printf("✓ Found %d commentary-related fields:\n\n", len(commentaryFields))
		for _, field := range commentaryFields {
			fmt.Printf("  📝 %s\n", field)
		}
	}

	// Display all top-level keys for investigation
	fmt.Println()
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("TOP-LEVEL RESPONSE STRUCTURE")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println()
	displayStructure(rawResponse, "", 0, 3)

	// Specifically look at content.matchFacts for any commentary
	fmt.Println()
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("CONTENT.MATCHFACTS STRUCTURE (detailed)")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println()
	if content, ok := rawResponse["content"].(map[string]interface{}); ok {
		if matchFacts, ok := content["matchFacts"].(map[string]interface{}); ok {
			displayStructure(matchFacts, "matchFacts", 0, 4)
		} else {
			fmt.Println("No matchFacts found in content")
		}
	} else {
		fmt.Println("No content field found")
	}

	// Extract and display liveticker data
	fmt.Println()
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("LIVETICKER / COMMENTARY DATA (full)")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println()
	displayLiveticker(rawResponse)
}

// displayLiveticker extracts and displays the liveticker/commentary data.
func displayLiveticker(data map[string]interface{}) {
	content, ok := data["content"].(map[string]interface{})
	if !ok {
		fmt.Println("❌ No content field found")
		return
	}

	liveticker, ok := content["liveticker"].(map[string]interface{})
	if !ok {
		fmt.Println("❌ No liveticker field found in content")
		fmt.Println()
		fmt.Println("Available content keys:")
		for key := range content {
			fmt.Printf("  - %s\n", key)
		}
		return
	}

	fmt.Println("✓ Liveticker data found!")
	fmt.Println()

	// Print raw liveticker JSON for full inspection
	livetickerJSON, err := json.MarshalIndent(liveticker, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling liveticker: %v\n", err)
		return
	}

	fmt.Println("Raw liveticker JSON:")
	fmt.Println(string(livetickerJSON))
}

// fetchRawMatchDetails fetches the raw JSON response for a match.
func fetchRawMatchDetails(ctx context.Context, matchID int) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/matchDetails?matchId=%d", baseURL, matchID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var rawResponse map[string]interface{}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}

	return rawResponse, nil
}

// displayMatchInfo shows basic match information.
func displayMatchInfo(data map[string]interface{}) {
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println("MATCH INFORMATION")
	fmt.Println(strings.Repeat("═", 70))
	fmt.Println()

	// Try to extract basic match info
	if general, ok := data["general"].(map[string]interface{}); ok {
		if homeTeam, ok := general["homeTeam"].(map[string]interface{}); ok {
			if name, ok := homeTeam["name"].(string); ok {
				fmt.Printf("Home Team: %s\n", name)
			}
		}
		if awayTeam, ok := general["awayTeam"].(map[string]interface{}); ok {
			if name, ok := awayTeam["name"].(string); ok {
				fmt.Printf("Away Team: %s\n", name)
			}
		}
		if leagueName, ok := general["leagueName"].(string); ok {
			fmt.Printf("League: %s\n", leagueName)
		}
	}

	// Score from header
	if header, ok := data["header"].(map[string]interface{}); ok {
		if teams, ok := header["teams"].([]interface{}); ok && len(teams) >= 2 {
			if home, ok := teams[0].(map[string]interface{}); ok {
				if away, ok := teams[1].(map[string]interface{}); ok {
					homeScore, _ := home["score"].(float64)
					awayScore, _ := away["score"].(float64)
					fmt.Printf("Score: %.0f - %.0f\n", homeScore, awayScore)
				}
			}
		}
		if status, ok := header["status"].(map[string]interface{}); ok {
			if liveTime, ok := status["liveTime"].(map[string]interface{}); ok {
				if short, ok := liveTime["short"].(string); ok {
					fmt.Printf("Status: %s\n", short)
				}
			}
		}
	}
	fmt.Println()
}

// findCommentaryFields recursively searches for commentary-related fields.
func findCommentaryFields(data interface{}, path string) []string {
	var results []string

	commentaryKeywords := []string{
		"commentary", "comment", "comments",
		"liveticker", "ticker", "updates", "update",
		"timeline", "incidents", "incident",
		"text", "description", "narrative",
		"highlight", "highlights",
		"pulse", "feed",
	}

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			currentPath := key
			if path != "" {
				currentPath = path + "." + key
			}

			// Check if key matches any commentary keyword
			keyLower := strings.ToLower(key)
			for _, keyword := range commentaryKeywords {
				if strings.Contains(keyLower, keyword) {
					// Found a potential commentary field
					valueType := fmt.Sprintf("%T", value)
					if arr, ok := value.([]interface{}); ok {
						valueType = fmt.Sprintf("[]interface{} (len=%d)", len(arr))
					}
					results = append(results, fmt.Sprintf("%s [%s]", currentPath, valueType))
					break
				}
			}

			// Recurse into nested objects
			results = append(results, findCommentaryFields(value, currentPath)...)
		}
	case []interface{}:
		// Only check first element of arrays to avoid duplicates
		if len(v) > 0 {
			results = append(results, findCommentaryFields(v[0], path+"[0]")...)
		}
	}

	return results
}

// displayStructure shows the structure of the JSON response.
func displayStructure(data interface{}, path string, depth int, maxDepth int) {
	if depth > maxDepth {
		fmt.Printf("%s... (max depth reached)\n", strings.Repeat("  ", depth))
		return
	}

	indent := strings.Repeat("  ", depth)

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			currentPath := key
			if path != "" {
				currentPath = path + "." + key
			}

			switch val := value.(type) {
			case map[string]interface{}:
				fmt.Printf("%s📁 %s (object, %d keys)\n", indent, key, len(val))
				displayStructure(val, currentPath, depth+1, maxDepth)
			case []interface{}:
				fmt.Printf("%s📋 %s (array, %d items)\n", indent, key, len(val))
				if len(val) > 0 && depth < maxDepth {
					fmt.Printf("%s  └─ First item:\n", indent)
					displayStructure(val[0], currentPath+"[0]", depth+2, maxDepth)
				}
			case string:
				// Truncate long strings
				display := val
				if len(display) > 50 {
					display = display[:50] + "..."
				}
				fmt.Printf("%s📄 %s: \"%s\"\n", indent, key, display)
			case float64:
				fmt.Printf("%s🔢 %s: %.0f\n", indent, key, val)
			case bool:
				fmt.Printf("%s✓ %s: %v\n", indent, key, val)
			case nil:
				fmt.Printf("%s⊘ %s: null\n", indent, key)
			default:
				fmt.Printf("%s? %s: %T\n", indent, key, value)
			}
		}
	}
}

func printUsage() {
	fmt.Println("Usage: go run scripts/test_commentary.go <match_id>")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run scripts/test_commentary.go 4513498")
	fmt.Println()
	fmt.Println("To find match IDs, first run:")
	fmt.Println("  go run scripts/api_results_raw.go 47 today")
	fmt.Println()
	fmt.Println("This will show Premier League matches for today with their IDs.")
}
