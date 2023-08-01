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

	if m, err = load(spinTomlContents); err != nil {
		return m, fmt.Errorf("unable to load spin.toml: %w", err)
	}

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
			s := ComponentSourceString(rawSource.(string))
			sumTypeSource = s
		case reflect.Map:
			stringMap, ok := rawSource.(map[string]interface{})
			if !ok {
				return m, fmt.Errorf("casting component %d to map[string]interface{}", i)
			}
			url, ok := stringMap["url"].(string)
			if !ok {
				return m, fmt.Errorf("casting url on component %d", i)
			}
			digest, ok := stringMap["digest"].(string)
			if !ok {
				return m, fmt.Errorf("casting digest on component %d", i)
			}
			sumTypeSource = ComponentSourceURL{
				Url:    url,
				Digest: digest,
			}
		default:
			return m, fmt.Errorf("unknown component source type: %v", v.Kind())
		}
		m.Components[i].Source = sumTypeSource
	}

	return m, nil
}
