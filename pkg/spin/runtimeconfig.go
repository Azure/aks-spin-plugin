package spin

type RuntimeVariable struct {
	Default  string `toml:"default"`
	Required bool   `toml:"required"`
}

type RuntimeComponent struct {
	Config map[string]RuntimeVariable `toml:"config"`
}

type RuntimeConfig struct {
	Variables map[string]RuntimeVariable `toml:"variables"`
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
