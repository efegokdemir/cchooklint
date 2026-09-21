package discover

import (
	"errors"
	"os"
	"path/filepath"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func homeSettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "settings.json"), nil
}

func Find() ([]string, error) {
	var result []string

	projectPath := filepath.Join(".claude", "settings.json")
	if fileExists(projectPath) {
		result = append(result, projectPath)
	}

	localPath := filepath.Join(".claude", "settings.local.json")
	if fileExists(localPath) {
		result = append(result, localPath)
	}

	homePath, err := homeSettingsPath()
	if err == nil && fileExists(homePath) {
		result = append(result, homePath)
	}

	return result, nil
}
