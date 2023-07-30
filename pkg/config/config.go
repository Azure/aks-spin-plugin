package config

import (
	"fmt"
	"os"
	"path"

	"github.com/BurntSushi/toml"
)

var (
	c    config
	opts *Opts
)

func (o *Opts) Default() {
	if o == nil {
		o = &Opts{}
	}

	if o.path == "" {
		o.path = "./aks-spin.toml"
	}
}

func Load(o Opts) error {
	opts = &o
	opts.Default()

	// TODO: how to better handle some things being in env or flags?
	if _, err := toml.DecodeFile(opts.path, &c); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("decoding aks spin config toml file %s: %w", opts.path, err)
	}

	return nil
}

func Write() error {
	// create directories if they don't exist
	dirs := path.Dir(opts.path)
	if _, err := os.Stat(dirs); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("validating directories %s: %w", dirs, err)
		}

		if err := os.MkdirAll(dirs, os.ModeDir|0755); err != nil {
			return fmt.Errorf("making directories %s: %w", dirs, err)
		}
	}

	// open file handles creating the file if it doesn't exist
	f, err := os.OpenFile(opts.path, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return fmt.Errorf("opening config file: %w", err)
	}
	defer f.Close()

	if err := toml.NewEncoder(f).Encode(c); err != nil {
		return fmt.Errorf("encoding aks spin config toml file %s: %w", opts.path, err)
	}

	return nil
}
