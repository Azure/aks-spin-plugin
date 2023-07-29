package spin

import (
	"fmt"

	"github.com/spf13/viper"
)

func Load() (*manifest, error) {
	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigName("spin")
	v.AddConfigPath("./")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("loading spin manifest config: %w", err)
	}

	m := &manifest{}
	if err := v.Unmarshal(m); err != nil {
		return nil, fmt.Errorf("unmarshalling spin manifest: %w", err)
	}

	return m, nil
}
