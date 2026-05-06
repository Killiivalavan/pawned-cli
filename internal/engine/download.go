package engine

import (
	"archive/tar"
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	repoAPI       = "https://api.github.com/repos/official-stockfish/Stockfish/releases/latest"
	appName       = "chesshell"
	legacyAppName = "pawned"
	engineDirName = "engine"
)

// GetEnginePath returns the path to the Stockfish executable.
// It first checks the system PATH, then checks the local app engine directory.
// If not found in either, it attempts to download it.
func GetEnginePath() (string, error) {
	// 1. Check system PATH
	binName := "stockfish"
	if runtime.GOOS == "windows" {
		binName = "stockfish.exe"
	}

	path, err := exec.LookPath(binName)
	if err == nil {
		return path, nil
	}

	// 2. Check local engine directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not get config dir: %w", err)
	}
	engineDir := filepath.Join(configDir, appName, engineDirName)
	localPath := filepath.Join(engineDir, binName)

	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	legacyEngineDir := filepath.Join(configDir, legacyAppName, engineDirName)
	legacyLocalPath := filepath.Join(legacyEngineDir, binName)
	if _, err := os.Stat(legacyLocalPath); err == nil {
		return legacyLocalPath, nil
	}

	// 3. Download if not found
	fmt.Println("First time AI setup: Downloading Stockfish chess engine (~50MB)...")
	if err := downloadEngine(engineDir, localPath); err != nil {
		return "", fmt.Errorf("failed to download engine: %w. Try installing stockfish manually via your package manager", err)
	}

	return localPath, nil
}

func downloadEngine(engineDir, localPath string) error {
	assetName, isZip := getAssetName()
	if assetName == "" {
		return fmt.Errorf("no precompiled binary available for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Fetch release info
	resp, err := http.Get(repoAPI)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var release struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return err
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("could not find asset %s in latest release", assetName)
	}

	// Download archive
	fmt.Printf("Downloading %s...\n", assetName)
	archiveResp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer archiveResp.Body.Close()

	if err := os.MkdirAll(engineDir, 0750); err != nil {
		return err
	}

	// Extract directly
	if isZip {
		return extractZip(archiveResp.Body, archiveResp.ContentLength, localPath)
	}
	return extractTar(archiveResp.Body, localPath)
}

func getAssetName() (name string, isZip bool) {
	osStr := runtime.GOOS
	archStr := runtime.GOARCH

	// Using the sf_18 naming conventions
	if osStr == "linux" {
		if archStr == "amd64" {
			// standard x86-64 is the safest for compatibility
			return "stockfish-ubuntu-x86-64.tar", false
		}
	} else if osStr == "darwin" {
		if archStr == "arm64" {
			return "stockfish-macos-m1-apple-silicon.tar", false
		} else if archStr == "amd64" {
			return "stockfish-macos-x86-64-avx2.tar", false
		}
	} else if osStr == "windows" {
		if archStr == "amd64" {
			return "stockfish-windows-x86-64-avx2.zip", true
		}
	}
	return "", false
}

func extractTar(r io.Reader, dest string) error {
	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// We only want the actual executable, ignoring folders or docs
		// Stockfish tarballs usually have a folder, then the binary inside it.
		// The binary usually has no extension, or ends in .exe
		name := filepath.Base(header.Name)
		if !header.FileInfo().IsDir() && strings.HasPrefix(name, "stockfish") && !strings.HasSuffix(name, ".txt") && !strings.HasSuffix(name, ".md") {
			f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755) // Ensure executable
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()

			if runtime.GOOS == "darwin" {
				exec.Command("xattr", "-d", "com.apple.quarantine", dest).Run() // Ignore errors if not quarantined
			}

			return nil // Found and extracted the binary
		}
	}
	return fmt.Errorf("binary not found in tar archive")
}

func extractZip(r io.Reader, size int64, dest string) error {
	// Zip reading in Go requires an io.ReaderAt. We must read the whole file into memory or a temp file.
	// Stockfish zips are small enough to read into memory.
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	zr, err := zip.NewReader(strings.NewReader(string(body)), int64(len(body)))
	if err != nil {
		return err
	}

	for _, file := range zr.File {
		name := filepath.Base(file.Name)
		if !file.FileInfo().IsDir() && strings.HasPrefix(name, "stockfish") && strings.HasSuffix(name, ".exe") {
			rc, err := file.Open()
			if err != nil {
				return err
			}

			f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				rc.Close()
				return err
			}
			_, copyErr := io.Copy(f, rc)
			f.Close()
			rc.Close()

			if copyErr != nil {
				return copyErr
			}
			return nil
		}
	}
	return fmt.Errorf("binary not found in zip archive")
}
