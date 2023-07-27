package config

import (
	"fmt"
	"path"

	"github.com/spf13/viper"
)

func (o *Opts) Default() {
	if o == nil {
		o = &Opts{}
	}

	if o.path == "" {
		o.path = "./aks-spin"
	}
}

func getViper(opts *Opts) *viper.Viper {
	opts.Default()
	filepath, filename := path.Split(opts.path)

	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigName(filename)
	v.AddConfigPath(filepath)
	return v
}

func Parse(opts *Opts) (*config, error) {
	v := getViper(opts)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("loading aks-spin-plugin config: %w", err)
	}

	c := &config{}
	if err := v.Unmarshal(c); err != nil {
		return nil, fmt.Errorf("unmarshalling aks-spin-plugin config: %w", err)
	}

	return c, nil
}

func (c *config) Write(opts *Opts) error {
	v := getViper(opts)

	if err := v.ReadInConfig(); err != nil {
		if _, configNotFound := err.(viper.ConfigFileNotFoundError); !configNotFound {
			return fmt.Errorf("reading current aks-spin-plugin config: %w", err)
		}
	}

	// TODO: this won't work with viper. Viper has no support for marshalling struct back into viper config

	return nil
}
