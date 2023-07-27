package spin

// https://developer.fermyon.com/spin/manifest-reference

type manifest struct {
	SpinVersion         string `mapstructure:"spin_version"`
	SpinManifestVersion string `mapstructure:"spin_manifest_version"`
	Name                string
	Version             string
	Description         string
	Authors             []string
	Trigger             manifestTrigger
	Variables           variables
	Component           []component
}

type manifestTrigger struct {
	// t type of trigger
	T    string `mapstructure:"type"`
	Base string
}

type variables map[string]struct {
	// def default value of variable
	Def      string `mapstructure:"default"`
	Required bool
	Secret   bool
}

type component struct {
	Id               string
	Description      string
	// TODO: this field blocks unmarshalling
	//Source           struct{}   // this is a sum type and must be handled in a special way
	Files            []struct{} // this is a sum type and must be handled in a special way
	ExcludeFiles     []string   `mapstructure:"exclude_files"`
	AllowedHttpHosts []string   `mapstructure:"allowed_http_hosts"`
	KeyValueStores   []string   `mapstructure:"key_value_stores"`
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
	t          string `mapstructure:"type"`
	argv       string
	entrypoint string
}

type build struct {
	command string
	workdir string
}
