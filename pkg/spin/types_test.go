package spin

import (
	"fmt"
	"testing"
)

func TestSpinToml(t *testing.T) {

	m := Manifest{}
	testToml := `
	[[component]]
	id = "hello"
	source = {url="hello.wasm",digest="ligma"}
	[[component]]
	id = "world"
	source = "string"
	`

	m, err := load([]byte(testToml))
	if err != nil {
		t.Errorf("failed to load toml: %s", err.Error())
	}
	fmt.Println(m)

	for _, c := range m.Components {
		fmt.Println(c)

		switch c.Source.(type) {
		case ComponentSourceURL:
			fmt.Println("url")
		case ComponentSourceString:
			fmt.Println("string")
		default:
			t.Errorf("unknown component source type: %T", c.Source)
		}
	}
}
