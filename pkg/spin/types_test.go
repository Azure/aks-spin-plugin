package spin

import (
	"reflect"
	"testing"
)

func TestSpinToml(t *testing.T) {

	m := Manifest{}
	testToml := `
	[[component]]
	id = "hello"
	source = {url="hello.wasm",digest="test-digest"}
	[[component]]
	id = "world"
	source = "source-string"
	`

	expectedManifest := Manifest{
		Components: []Component{
			Component{
				Id: "hello",
				Source: ComponentSourceURL{
					Url:    "hello.wasm",
					Digest: "test-digest",
				}},
			{
				Id:     "world",
				Source: ComponentSourceString("source-string")},
		},
	}

	m, err := load([]byte(testToml))
	if err != nil {
		t.Errorf("failed to load toml: %s", err.Error())
	}
	if !reflect.DeepEqual(m, expectedManifest) {
		t.Errorf("unmarshaled spin manifest does not match expected manifest")
	}
}
