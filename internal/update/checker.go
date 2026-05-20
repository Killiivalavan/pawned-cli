package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

// Checker handles checking for and applying updates for the application.
type Checker struct {
	currentVersion string
	repoOwner      string
	repoName       string
	httpClient     *http.Client
}

// NewChecker creates a new Checker instance.
func NewChecker(currentVersion, repoOwner, repoName string) *Checker {
	return &Checker{
		currentVersion: currentVersion,
		repoOwner:      repoOwner,
		repoName:       repoName,
		httpClient: &http.Client{
			Timeout: 10 * time.Second, // Reasonable timeout for update checks
		},
	}
}

// githubRelease represents a minimal structure of the GitHub release API response.
type githubRelease struct {
	TagName string `json:"tag_name"`
}

// CheckLatest checks the GitHub repository for a newer release than the current version.
// It returns the latest version found, a boolean indicating if it's newer, and any error encountered.
func (c *Checker) CheckLatest() (string, bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", c.repoOwner, c.repoName)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", false, fmt.Errorf("creating request: %w", err)
	}
	
	// Set a reasonable user agent
	req.Header.Set("User-Agent", fmt.Sprintf("%s-update-checker/%s", c.repoName, c.currentVersion))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", false, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", false, fmt.Errorf("decoding response: %w", err)
	}

	latestVersion := release.TagName
	// Strip the leading "v" if present for semver comparison
	normalizedLatest := latestVersion
	if strings.HasPrefix(normalizedLatest, "v") {
		// x/mod/semver expects the 'v' prefix, so we shouldn't strip it if we want to use semver.Compare.
		// Wait, the roadmap says:
		// - Parse tag_name from JSON response
		// - Strip leading v from the tag
		// - Compare with currentVersion using golang.org/x/mod/semver
		// Actually, golang.org/x/mod/semver REQUIRES a "v" prefix. Let's make sure both have "v" prefix before comparing.
		// Let's adjust this logic carefully.
	}

	// Make sure both versions have a "v" prefix for semver comparison
	vCurrent := c.currentVersion
	if !strings.HasPrefix(vCurrent, "v") {
		vCurrent = "v" + vCurrent
	}
	
	vLatest := latestVersion
	if !strings.HasPrefix(vLatest, "v") {
		vLatest = "v" + vLatest
	}

	if !semver.IsValid(vCurrent) {
		return "", false, fmt.Errorf("invalid current version semver: %s", c.currentVersion)
	}
	if !semver.IsValid(vLatest) {
		return "", false, fmt.Errorf("invalid latest version semver: %s", latestVersion)
	}

	// semver.Compare returns 1 if vLatest > vCurrent
	isNewer := semver.Compare(vLatest, vCurrent) > 0

	// We return the raw tag name from github (e.g. "v1.2.3")
	return latestVersion, isNewer, nil
}

// GetAssetURL constructs the download URL for the correct asset for the current OS and architecture.
// It uses runtime.GOOS and runtime.GOARCH.
func (c *Checker) GetAssetURL(latestVersion string, os, arch string) string {
	ext := ""
	if os == "windows" {
		ext = ".exe"
	}
	assetName := fmt.Sprintf("%s-%s-%s%s", c.repoName, os, arch, ext)
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", c.repoOwner, c.repoName, latestVersion, assetName)
}

// ApplyUpdate downloads the new binary from the given URL and replaces the current executable.
func (c *Checker) ApplyUpdate(downloadURL string) error {
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("downloading update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code downloading update: %d", resp.StatusCode)
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "chesshell-update-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	
	// Ensure we clean up if we fail before atomic rename
	defer os.Remove(tmpPath)

	// Download the file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing to temp file: %w", err)
	}
	tmpFile.Close()

	// Make executable
	if err := os.Chmod(tmpPath, 0755); err != nil {
		return fmt.Errorf("making temp file executable: %w", err)
	}

	// Get current executable path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("getting executable path: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, exePath); err != nil {
		return fmt.Errorf("replacing binary: %w", err)
	}

	return nil
}
