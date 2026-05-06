package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const (
	dataFileName = "data.json"
	appName      = "pawned"
	fileVersion  = 1
)

var (
	ErrCorruptedFile = errors.New("data file is corrupted")
)

// getPath determines the full path to the data file based on the OS.
func getPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, appName)
	return filepath.Join(appDir, dataFileName), nil
}

// Load reads the data file from disk.
// If the file doesn't exist, it returns a new, empty Data object.
// If the file is corrupted, it backs it up and returns an empty Data object.
func Load() (*Data, error) {
	path, err := getPath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return a fresh Data object
			return &Data{Version: fileVersion, History: []HistoryItem{}}, nil
		}
		return nil, err // Other file-opening error
	}
	defer file.Close()

	var data Data
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		// File is corrupted
		file.Close() // Close before renaming
		backupPath := path + ".bak"
		if renameErr := os.Rename(path, backupPath); renameErr != nil {
			return nil, renameErr
		}
		// Return a fresh Data object along with the corruption error
		return &Data{Version: fileVersion, History: []HistoryItem{}}, ErrCorruptedFile
	}

	return &data, nil
}

// Save writes the data to the disk.
// It ensures the directory exists before writing.
func Save(data *Data) error {
	path, err := getPath()
	if err != nil {
		return err
	}

	// Ensure the parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	// Create or truncate the file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Use an encoder for pretty-printing the JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
