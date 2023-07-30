package spin

import (
	"fmt"
	"os"
	"reflect"

	"github.com/BurntSushi/toml"
)

func Load() (Manifest, error) {
	m := Manifest{}
	spinTomlContents, err := os.ReadFile("spin.toml")
	if err != nil {
		return m, fmt.Errorf("unable to open spin.toml file: %w", err)
	}

	m, err = load(spinTomlContents)
	return m, nil
}

func load(spinTomlContents []byte) (Manifest, error) {
	m := Manifest{}

	_, err := toml.Decode(string(spinTomlContents), &m)
	if err != nil {
		return m, fmt.Errorf("failed to decode spin.toml decode: %w", err)
	}

	rm := rawManifest{}
	_, err = toml.Decode(string(spinTomlContents), &rm)
	if err != nil {
		return m, fmt.Errorf("failed to decode raw spin.toml decode: %w", err)
	}

	for i, c := range rm.Components {
		var sumTypeSource ComponentSource

		rawSource := c.Source
		v := reflect.ValueOf(rawSource)
		switch v.Kind() {
		case reflect.String:
			sumTypeSource = ComponentSourceString(v.String())
		case reflect.Map:
			sumTypeSource = ComponentSourceURL{
				Url:    v.MapIndex(reflect.ValueOf("url")).String(),
				Digest: v.MapIndex(reflect.ValueOf("digest")).String(),
			}
		default:
			return m, fmt.Errorf("unknown component source type: %v", v.Kind())
		}
		m.Components[i].Source = sumTypeSource
	}

	return m, nil
}
