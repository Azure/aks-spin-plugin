package config

import (
	"fmt"
	"os"
	"path"

	"github.com/BurntSushi/toml"
	"github.com/caarlos0/env/v9"
)

var (
	c    config
	opts *Opts
)

// def sets empty options to their defaults
func (o *Opts) def() {
	if o == nil {
		o = &Opts{}
	}

	if o.Path == "" {
		o.Path = "./aks-spin.toml"
	}
}

// Get returns the current aks spin config
func Get() config {
	return c
}

// Load loads any current aks spin configs from a file or the environment, with precedence towards env variables
func Load(o Opts) error {
	opts = &o
	opts.def()

	if _, err := toml.DecodeFile(opts.Path, &c); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("decoding aks spin config toml file %s: %w", opts.Path, err)
	}

	if err := env.ParseWithOptions(&c, env.Options{
		Prefix:                "AKS_SPIN_",
		UseFieldNameByDefault: true,
	}); err != nil {
		return fmt.Errorf("parsing aks spin config from env variables: %w", err)
	}

	return nil
}

// Write writes the current aks spin config to a file
func Write() error {
	// create directories if they don't exist
	dirs := path.Dir(opts.Path)
	if _, err := os.Stat(dirs); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("validating directories %s: %w", dirs, err)
		}

		if err := os.MkdirAll(dirs, os.ModeDir|0755); err != nil {
			return fmt.Errorf("making directories %s: %w", dirs, err)
		}
	}

	// open file handles creating the file if it doesn't exist
	f, err := os.OpenFile(opts.Path, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return fmt.Errorf("opening config file: %w", err)
	}
	defer f.Close()

	if err := toml.NewEncoder(f).Encode(c); err != nil {
		return fmt.Errorf("encoding aks spin config toml file %s: %w", opts.Path, err)
	}

	return nil
}

func GetKeyVault() KeyVault{
	return c.KeyVault
}
