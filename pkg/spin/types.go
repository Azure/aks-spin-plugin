package spin

// https://developer.fermyon.com/spin/manifest-reference

type Manifest struct {
	SpinVersion         string
	SpinManifestVersion string
	Name                string
	Version             string
	Description         string
	Authors             []string
	Trigger             manifestTrigger
	Variables           variables
	Components          []Component `toml:"component"`
}

type manifestTrigger struct {
	// t type of trigger
	T    string `mapstructure:"type"`
	Base string
}

type variables map[string]struct {
	// def default value of variable
	Def      string `toml:"default"`
	Required bool
	Secret   bool
}

// go-sumtype:decl ComponentSource
type ComponentSource interface {
	isComponentSource()
}

type ComponentSourceURL struct {
	Url    string `toml:"url"`
	Digest string `toml:"digest"`
}

func (ComponentSourceURL) isComponentSource() {}

type ComponentSourceString string

func (ComponentSourceString) isComponentSource() {}

type rawManifest struct {
	Components []rawComponent `toml:"component"`
}

type rawComponent struct {
	Source interface{} `toml:"source"`
}

type Component struct {
	Id               string          `toml:"id"`
	Description      string          `toml:"description"`
	Source           ComponentSource `toml:"never"`
	Files            []struct{}      // this is a sum type and must be handled in a special way
	ExcludeFiles     []string        `toml:"exclude_files"`
	AllowedHttpHosts []string        `toml:"allowed_http_hosts"`
	KeyValueStores   []string        `toml:"key_value_stores"`
	Environment      map[string]string
	Trigger          componentTrigger
	Build            build
	Config           map[string]string
}

type componentTrigger struct {
	route    string
	executor executor
	channel  string
}

type executor struct {
	// t type of executor
	t          string `toml:"type"`
	argv       string
	entrypoint string
}

type build struct {
	command string
	workdir string
}
