package spin

import (
	"fmt"

	"github.com/spf13/viper"
)

func Parse() (*manifest, error) {
	parser := viper.New()
	parser.SetConfigType("toml")
	parser.SetConfigName("spin")

	var m manifest
	if err := viper.Unmarshal(&m); err != nil {
		return nil, fmt.Errorf("unmarshalling spin manifest: %w", err)
	}

	return &m, nil
}
