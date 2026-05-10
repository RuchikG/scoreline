// Package data provides utilities for loading mock football match data.
package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	appName         = "scoreline"
	legacyAppName   = "golazo"
	configDir       = ".scoreline"
	legacyConfigDir = ".golazo"
)

// ConfigDir returns the path to the scoreline config directory.
// On Linux, follows XDG Base Directory spec (~/.config/scoreline).
// On other systems (macOS, Windows), uses ~/.scoreline.
func ConfigDir() (string, error) {
	return configDirFor(appName, configDir, true)
}

// LegacyConfigDir returns the old golazo config directory without creating it.
func LegacyConfigDir() (string, error) {
	return configDirFor(legacyAppName, legacyConfigDir, false)
}

func configDirFor(xdgName, dotName string, create bool) (string, error) {
	var configPath string

	if runtime.GOOS == "linux" {
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			configPath = filepath.Join(xdgConfig, xdgName)
		} else {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("get home directory: %w", err)
			}
			configPath = filepath.Join(homeDir, ".config", xdgName)
		}
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, dotName)
	}

	if create {
		if err := os.MkdirAll(configPath, 0755); err != nil {
			return "", fmt.Errorf("create config directory: %w", err)
		}
	}

	return configPath, nil
}

// CacheDir returns the path to the scoreline cache directory.
// Uses os.UserCacheDir() which returns:
//   - Linux: ~/.cache/scoreline (or $XDG_CACHE_HOME/scoreline)
//   - macOS: ~/Library/Caches/scoreline
//   - Windows: %LocalAppData%/scoreline
func CacheDir() (string, error) {
	return cacheDirFor(appName, true)
}

// LegacyCacheDir returns the old golazo cache directory without creating it.
func LegacyCacheDir() (string, error) {
	return cacheDirFor(legacyAppName, false)
}

func cacheDirFor(name string, create bool) (string, error) {
	userCache, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("get user cache directory: %w", err)
	}

	cachePath := filepath.Join(userCache, name)
	if create {
		if err := os.MkdirAll(cachePath, 0755); err != nil {
			return "", fmt.Errorf("create cache directory: %w", err)
		}
	}

	return cachePath, nil
}

// LegacyCacheFile returns the matching legacy cache file when it exists.
func LegacyCacheFile(name string) (string, bool) {
	dir, err := LegacyCacheDir()
	if err != nil {
		return "", false
	}
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); err == nil {
		return path, true
	}
	return "", false
}

// ScorelineCacheFile returns a path inside the current cache directory.
func ScorelineCacheFile(name string) (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name), nil
}

// EnsureCacheSubdir returns a named cache subdirectory, creating it if needed.
func EnsureCacheSubdir(name string) (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", fmt.Errorf("create cache subdirectory: %w", err)
	}
	return path, nil
}

// MockDataPath returns the path to the mock data file.
func MockDataPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "matches.json"), nil
}

// LiveUpdate represents a single live update string.
type LiveUpdate struct {
	MatchID int
	Update  string
	Time    time.Time
}

// SaveLiveUpdate appends a live update to the storage.
func SaveLiveUpdate(matchID int, update string) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	updatesFile := filepath.Join(dir, fmt.Sprintf("updates_%d.json", matchID))

	var updates []LiveUpdate
	if data, err := os.ReadFile(updatesFile); err == nil {
		// Best effort to load existing updates; if unmarshal fails, start with empty slice
		if err := json.Unmarshal(data, &updates); err != nil {
			// Invalid JSON in file - start fresh with empty slice
			updates = []LiveUpdate{}
		}
	}

	updates = append(updates, LiveUpdate{
		MatchID: matchID,
		Update:  update,
		Time:    time.Now(),
	})

	data, err := json.Marshal(updates)
	if err != nil {
		return fmt.Errorf("marshal updates: %w", err)
	}

	return os.WriteFile(updatesFile, data, 0644)
}

// LiveUpdates retrieves live updates for a match.
func LiveUpdates(matchID int) ([]string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, err
	}

	updatesFile := filepath.Join(dir, fmt.Sprintf("updates_%d.json", matchID))
	data, err := os.ReadFile(updatesFile)
	if err != nil {
		return []string{}, nil // Return empty if file doesn't exist
	}

	var updates []LiveUpdate
	if err := json.Unmarshal(data, &updates); err != nil {
		return nil, fmt.Errorf("unmarshal updates: %w", err)
	}

	result := make([]string, 0, len(updates))
	for _, update := range updates {
		result = append(result, update.Update)
	}

	return result, nil
}

// LoadLatestVersion reads the latest known version from storage.
// Returns empty string if file doesn't exist or can't be read.
func LoadLatestVersion() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	versionFile := filepath.Join(dir, "latest_version.txt")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		return "", nil // Return empty if file doesn't exist
	}

	return strings.TrimSpace(string(data)), nil
}

// SaveLatestVersion saves the latest version to storage.
func SaveLatestVersion(version string) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}

	versionFile := filepath.Join(dir, "latest_version.txt")
	return os.WriteFile(versionFile, []byte(strings.TrimSpace(version)), 0644)
}

// CheckLatestVersion fetches the latest version from GitHub releases.
// Uses GitHub's redirect URL which is simpler than the API.
// Returns the version tag (e.g., "v1.2.3").
func CheckLatestVersion() (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get("https://github.com/RuchikG/scoreline/releases/latest")
	if err != nil {
		return "", fmt.Errorf("fetch latest release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// GitHub redirects to: https://github.com/RuchikG/scoreline/releases/tag/v1.2.3
	// Extract version from the final URL
	finalURL := resp.Request.URL.String()

	// Look for "/releases/tag/" in the URL
	if idx := strings.LastIndex(finalURL, "/releases/tag/"); idx != -1 {
		version := finalURL[idx+len("/releases/tag/"):]
		if version != "" {
			return version, nil
		}
	}

	// Fallback: try to read from response body (in case redirect doesn't work)
	body, err := io.ReadAll(resp.Body)
	if err == nil && len(body) > 0 {
		// This is a fallback and might not work, but better than nothing
		bodyStr := string(body)
		if idx := strings.Index(bodyStr, "/releases/tag/"); idx != -1 {
			start := idx + len("/releases/tag/")
			end := strings.Index(bodyStr[start:], "\"")
			if end != -1 {
				version := bodyStr[start : start+end]
				if version != "" {
					return version, nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not extract version from GitHub response")
}

// ShouldCheckVersion returns true if we should check for a new version.
// Checks if the latest_version.txt file is older than 24 hours.
func ShouldCheckVersion() bool {
	dir, err := ConfigDir()
	if err != nil {
		return false
	}

	versionFile := filepath.Join(dir, "latest_version.txt")
	info, err := os.Stat(versionFile)
	if err != nil {
		return true // File doesn't exist, should check
	}

	// Check if file is older than 24 hours
	return time.Since(info.ModTime()) > 24*time.Hour
}
