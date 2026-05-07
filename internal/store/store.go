package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const (
	dataFileName  = "data.json"
	appName       = "chesshell"
	legacyAppName = "pawned"
	fileVersion   = 1
)

var (
	ErrCorruptedFile = errors.New("data file is corrupted")
)

// GetPath determines the full path to the data file based on the OS.
func GetPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, appName)
	return filepath.Join(appDir, dataFileName), nil
}

// GetDataDir returns the directory where chesshell data is stored.
func GetDataDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, appName), nil
}

func getLegacyPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	appDir := filepath.Join(configDir, legacyAppName)
	return filepath.Join(appDir, dataFileName), nil
}

// Load reads the data file from disk.
// If the file doesn't exist, it returns a new, empty Data object.
// If the file is corrupted, it backs it up and returns an empty Data object.
func Load() (*Data, error) {
	path, err := GetPath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			legacyPath, legacyErr := getLegacyPath()
			if legacyErr != nil {
				return &Data{Version: fileVersion, History: []HistoryItem{}}, nil
			}

			legacyFile, legacyOpenErr := os.Open(legacyPath)
			if legacyOpenErr != nil {
				return &Data{Version: fileVersion, History: []HistoryItem{}}, nil
			}
			defer legacyFile.Close()

			var legacyData Data
			if decodeErr := json.NewDecoder(legacyFile).Decode(&legacyData); decodeErr != nil {
				return &Data{Version: fileVersion, History: []HistoryItem{}}, ErrCorruptedFile
			}
			if legacyData.History == nil {
				legacyData.History = []HistoryItem{}
			}

			return &legacyData, nil
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

	if data.History == nil {
		data.History = []HistoryItem{}
	}

	return &data, nil
}

// Save writes the data to the disk.
// It ensures the directory exists before writing.
func Save(data *Data) error {
	path, err := GetPath()
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
