package spin

// https://developer.fermyon.com/spin/manifest-reference
type manifest struct {
	spinVersion         string `mapstructure:"spin_version"`
	spinManifestVersion string `mapstructure:"spin_manifest_version"`
	name                string
	version             string
	description         string
	authors             []string
	trigger             manifestTrigger
	variables           variables
	component           []component
}

type manifestTrigger struct {
	// t type of trigger
	t    string `mapstructure:"type"`
	base string
}

type variables map[string]struct {
	// def default value of variable
	def      string `mapstructure:"default"`
	required bool
	secret   bool
}

type component struct {
	id               string
	description      string
	source           struct{}   // this is a sum type and must be handled in a special way
	files            []struct{} // this is a sum type and must be handled in a special way
	excludeFiles     []string   `mapstructure:"exclude_files"`
	allowedHttpHosts []string   `mapstructure:"allowed_http_hosts"`
	keyValueStores   []string   `mapstructure:"key_value_stores"`
	environment      map[string]string
	trigger          componentTrigger
	build            build
	config           map[string]string
}

type componentTrigger struct {
	route    string
	executor executor
	channel  string
}

type executor struct {
	// t type of executor
	t          string `mapstructure:"type"`
	argv       string
	entrypoint string
}

type build struct {
	command string
	workdir string
}
