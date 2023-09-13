// Package state provides a simple key-value store for persisting state between plugin runs.
// This should never be relied on but should be used to improve the user experience. Common uses
// will be remembering the latest customer inputs and using them as defaults for the next run.
package state

import (
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/azure/spin-aks-plugin/pkg/logger"
)

var (
	// KeyNotFoundErr is returned when a key is not found in the state
	KeyNotFoundErr = errors.New("key not found in state")

	mu sync.Mutex // needed to ensure no state is lost if multiple goroutines try to save  or load state concurrently
)

// Set sets a key-value pair in the state
func Set(ctx context.Context, key string, value string) error {
	mu.Lock()
	defer mu.Unlock()

	state, err := loadState(ctx)
	if err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	state[key] = value
	if err := saveState(ctx, state); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	return nil
}

// Get gets a value from the state by key
func Get(ctx context.Context, key string) (string, error) {
	state, err := loadState(ctx)
	if err != nil {
		return "", fmt.Errorf("loading state: %w", err)
	}

	value, ok := state[key]
	if !ok {
		return "", KeyNotFoundErr
	}

	return value, nil
}

func loadState(ctx context.Context) (map[string]string, error) {
	path := statePath()
	lgr := logger.FromContext(ctx).With("path", path)
	lgr.Debug("loading state")

	lgr.Debug("opening state file")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening state file: %w", err)
	}
	defer f.Close()

	lgr.Debug("decoding current state")
	var decoded map[string]string
	d := gob.NewDecoder(f)
	if err := d.Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decoding state: %w", err)
	}

	lgr.Debug("state loaded")
	return decoded, nil
}

func saveState(ctx context.Context, state map[string]string) error {
	path := statePath()
	lgr := logger.FromContext(ctx).With("path", path)
	lgr.Debug("saving state")

	lgr.Debug("opening state file")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening state file: %w", err)
	}
	defer f.Close()

	lgr.Debug("encoding state")
	e := gob.NewEncoder(f)
	if err := e.Encode(state); err != nil {
		return fmt.Errorf("encoding state: %w", err)
	}

	lgr.Debug("state saved")
	return nil
}

func statePath() string {
	return filepath.Join(dataDir(), "spin", "plugins", "aks", "state")
}

// dataDir returns the path to the data directory according to the Spin spec
// https://developer.fermyon.com/spin/cache#base-directories
func dataDir() string {
	fallbackUnix := filepath.Join(os.Getenv("HOME"), "/.spin")

	switch system := runtime.GOOS; system {
	case "darwin": // macOS
		if xdgHome := os.Getenv("XDG_DATA_HOME"); xdgHome != "" {
			if exists(xdgHome) {
				return xdgHome
			}
		}

		if share := filepath.Join(os.Getenv("HOME"), "/.local/share"); exists(share) {
			return share
		}

		return fallbackUnix
	case "linux":
		if homebrewPrefix := os.Getenv("HOMEBREW_PREFIX"); homebrewPrefix != "" {
			if homebrewPath := filepath.Join(homebrewPrefix, "/etc/ferymon-spin"); exists(homebrewPath) {
				return homebrewPath
			}
		}

		if applicationSupport := filepath.Join(os.Getenv("HOME"), "/Library/Application Support"); exists(applicationSupport) {
			return applicationSupport
		}

		return fallbackUnix
	case "windows":
		if localApp := os.Getenv("LOCALAPPDATA"); localApp != "" {
			if exists(localApp) {
				return localApp
			}
		}

		if local := filepath.Join(os.Getenv("USERPROFILE"), "/AppData/Local"); exists(local) {
			return local
		}

		return filepath.Join(os.Getenv("USERPROFILE"), "./spin")
	default:
		return fallbackUnix
	}
}

// exists checks if a file or directory exists
func exists(path string) bool {
	if _, err := os.Stat(path); err == nil { // not perfect error handling but it's good enough
		return true
	}

	return false
}
