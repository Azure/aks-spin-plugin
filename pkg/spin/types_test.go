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
	files = "hello.txt"
	[[component]]
	id = "world"
	source = "source-string"
	files = {source="source.txt",destination="destination.txt"}
	[[component]]
	id = "string_and_map_files"
	files = ["source-1","source-2",{source="source-3",destination="destination-3"}]
	`

	expectedManifest := Manifest{
		Components: []Component{
			Component{
				Id: "hello",
				Source: ComponentSource{
					URLSource: ComponentSourceURL{
						Url:    "hello.wasm",
						Digest: "test-digest",
					},
				},
				Files: ComponentFiles{
					StringFiles: []ComponentFileString{
						ComponentFileString("hello.txt"),
					},
				},
			},
			{
				Id: "world",
				Source: ComponentSource{
					StringSource: ComponentSourceString("source-string"),
				},
				Files: ComponentFiles{
					MapFiles: []ComponentFileMap{
						{
							Source:      "source.txt",
							Destination: "destination.txt",
						},
					},
				},
			},
			{
				Id: "string_and_map_files",
				Files: ComponentFiles{
					StringFiles: []ComponentFileString{
						ComponentFileString("source-1"),
						ComponentFileString("source-2"),
					},
					MapFiles: []ComponentFileMap{
						{
							Source:      "source-3",
							Destination: "destination-3",
						},
					},
				},
			},
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
