package model

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolveDBPath(t *testing.T) {
	dbName := "test_monitor.db"
	path := ResolveDBPath(dbName)

	if runtime.GOOS == "darwin" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			expected := filepath.Join(homeDir, "Library", "Application Support", "IPMonitor", dbName)
			if path != expected {
				t.Fatalf("Esperava caminho %s no macOS, mas obteve: %s", expected, path)
			}
		}
	} else {
		exePath, err := os.Executable()
		if err == nil {
			expected := filepath.Join(filepath.Dir(exePath), dbName)
			if path != expected {
				t.Fatalf("Esperava caminho %s, mas obteve: %s", expected, path)
			}
		}
	}
}
