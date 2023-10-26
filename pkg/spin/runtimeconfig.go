package spin

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type RuntimeVariable struct {
	Default  string `toml:"default"`
	Required bool   `toml:"required"`
}

type RuntimeComponent struct {
	Config map[string]RuntimeVariable `toml:"config"`
}

type RuntimeConfigType string

const (
	RuntimeConfigTypeCosmos RuntimeConfigType = "azure_cosmos"
)

type KeyValueStore struct {
	Type      RuntimeConfigType `toml:"type"`
	Key       string            `toml:"key"`
	Url       string            `toml:"url"`
	Account   string            `toml:"account"`
	Database  string            `toml:"database"`
	Container string            `toml:"container"`
}

type RuntimeConfig struct {
	Variables     map[string]RuntimeVariable `toml:"variables"`
	KeyValueStore map[string]KeyValueStore   `toml:"key_value_store"`
}

func LoadRuntimeConfig(runtimeConfigLocation string) (RuntimeConfig, error) {
	rc := RuntimeConfig{}
	runtimeConfigContents, err := os.ReadFile(runtimeConfigLocation)
	if err != nil {
		return rc, fmt.Errorf("unable to open runtime config file: %w", err)
	}

	rc, err = loadRuntimeConfig(runtimeConfigContents)
	if err != nil {
		return rc, fmt.Errorf("unable to load runtime config: %w", err)
	}

	return rc, nil
}

func loadRuntimeConfig(runtimeConfigContents []byte) (RuntimeConfig, error) {
	rc := RuntimeConfig{}

	_, err := toml.Decode(string(runtimeConfigContents), &rc)
	if err != nil {
		return rc, fmt.Errorf("failed to decode runtime config: %w", err)
	}

	err = validateRuntimeConfig(rc)
	if err != nil {
		return rc, fmt.Errorf("failed to validate runtime config: %w", err)
	}

	return rc, nil
}

func validateRuntimeConfig(rc RuntimeConfig) error {
	if len(rc.KeyValueStore) > 0 {
		for k, store := range rc.KeyValueStore {
			// TODO: support more store names, but for now we match Fermyon Cloud supporting only "default" store
			// https://developer.fermyon.com/spin/dynamic-configuration#key-value-store-runtime-configuration
			//
			switch store.Type {
			case RuntimeConfigTypeCosmos:
				if store.Key == "" {
					return fmt.Errorf("key value store %s is missing key", k)
				}
				if store.Account == "" {
					return fmt.Errorf("key value store %s is missing account", k)
				}
				if store.Database == "" {
					return fmt.Errorf("key value store %s is missing database", k)
				}
				if store.Container == "" {
					return fmt.Errorf("key value store %s is missing container", k)
				}
			default:
				return fmt.Errorf("key value store %s has unsupported type %s", k, store.Type)
			}
		}
	}
	return nil
}

func NewRuntimeConfig(manifest Manifest) RuntimeConfig {
	rc := RuntimeConfig{
		Variables: map[string]RuntimeVariable{},
	}
	for k, v := range manifest.Variables {
		rc.Variables[k] = RuntimeVariable{
			Default:  v.Def,
			Required: v.Required,
		}
	}
	return rc
}
