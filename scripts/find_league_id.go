//go:build ignore

// find_league_id.go - Find FotMob league IDs by name or country
//
// Usage:
//   go run scripts/find_league_id.go <search_term>
//   go run scripts/find_league_id.go --south-america  (searches for first division leagues in Peru, Ecuador, Chile, Uruguay)
//
// Examples:
//   go run scripts/find_league_id.go "Premier League"
//   go run scripts/find_league_id.go Japan
//   go run scripts/find_league_id.go --south-america

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/find_league_id.go <search_term>")
		fmt.Println("       go run scripts/find_league_id.go --south-america")
		fmt.Println("Example: go run scripts/find_league_id.go \"Premier League\"")
		fmt.Println("         go run scripts/find_league_id.go --south-america")
		os.Exit(1)
	}

	if os.Args[1] == "--south-america" {
		searchSouthAmericaLeagues()
		return
	}

	term := strings.Join(os.Args[1:], " ")
	results := search(term)

	if len(results) == 0 {
		fmt.Println("No leagues found")
		os.Exit(0)
	}

	fmt.Printf("%-6s %-35s %s\n", "ID", "Name", "Country")
	fmt.Println(strings.Repeat("-", 55))
	for _, r := range results {
		fmt.Printf("%-6d %-35s %s\n", r.ID, truncate(r.Name, 35), r.Country)
	}
}

type league struct {
	ID      int
	Name    string
	Country string
}

func search(term string) []league {
	client := &http.Client{Timeout: 10 * time.Second}
	url := "https://www.fotmob.com/api/search/suggest?term=" + url.QueryEscape(term)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var data []struct {
		Suggestions []struct {
			Type    string `json:"type"`
			ID      string `json:"id"`
			Name    string `json:"name"`
			Country string `json:"ccode"`
		} `json:"suggestions"`
	}

	json.NewDecoder(resp.Body).Decode(&data)

	seen := make(map[int]bool)
	var results []league

	for _, group := range data {
		for _, s := range group.Suggestions {
			if s.Type != "league" {
				continue
			}
			var id int
			fmt.Sscanf(s.ID, "%d", &id)
			if seen[id] {
				continue
			}
			seen[id] = true
			results = append(results, league{ID: id, Name: s.Name, Country: s.Country})
		}
	}
	return results
}

func searchSouthAmericaLeagues() {
	countries := []string{"Brazil", "Argentina", "Uruguay", "Colombia", "Chile", "Peru", "Ecuador"}

	fmt.Printf("%-6s %-35s %s\n", "ID", "Name", "Country")
	fmt.Println(strings.Repeat("-", 55))

	for _, country := range countries {
		results := search(country + " primera division")

		// Also try some alternative search terms
		if len(results) == 0 {
			results = search(country + " primera")
		}
		if len(results) == 0 {
			results = search(country + " liga 1")
		}

		// Filter for leagues that seem to be first division
		var firstDivisionLeagues []league
		for _, r := range results {
			name := strings.ToLower(r.Name)
			if strings.Contains(name, "primera") ||
				strings.Contains(name, "liga 1") ||
				strings.Contains(name, "serie a") ||
				strings.Contains(name, "division") ||
				(strings.Contains(name, "liga") && !strings.Contains(name, "liga 2") && !strings.Contains(name, "liga 3")) {
				firstDivisionLeagues = append(firstDivisionLeagues, r)
			}
		}

		// If we found potential first division leagues, show the first one
		if len(firstDivisionLeagues) > 0 {
			r := firstDivisionLeagues[0]
			fmt.Printf("%-6d %-35s %s\n", r.ID, truncate(r.Name, 35), r.Country)
		} else if len(results) > 0 {
			// Fallback: show first result
			r := results[0]
			fmt.Printf("%-6d %-35s %s\n", r.ID, truncate(r.Name, 35), r.Country)
		} else {
			fmt.Printf("N/A    %-35s No league found\n", country)
		}

		// Small delay to be respectful to the server
		time.Sleep(200 * time.Millisecond)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
