package spin

// after attempting to use the go-sumptypes library to decode the spin manifest toml
// I've decided to make a union type of the arrays each variant sum type to decode the manifest
// it's more overhead to check, but much easier to handle from within go compared to always casting and checking types

type ComponentSourceURL struct {
	Url    string `toml:"url"`
	Digest string `toml:"digest"`
}
type ComponentSourceString string
type ComponentSource struct {
	URLSource    ComponentSourceURL
	StringSource ComponentSourceString
}

type ComponentFileString string
type ComponentFileMap struct {
	Source      string `toml:"source"`
	Destination string `toml:"destination"`
}
type ComponentFiles struct {
	StringFiles []ComponentFileString
	MapFiles    []ComponentFileMap
}

// rawComponent is used for decoding the sum types of the spin manifest toml
type rawComponent struct {
	Source interface{} `toml:"source"`
	Files  interface{} `toml:"files"`
}

// rawManifest is used for decoding the sum types in  spin manifest toml
type rawManifest struct {
	Components []rawComponent `toml:"component"`
}
