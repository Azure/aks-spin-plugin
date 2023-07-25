package spin

// https://developer.fermyon.com/spin/manifest-reference
type manifest struct {
	spinManifestVersion string
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
	t    string
	base string
}

type variables map[string]struct {
	// def default value of variable
	def      string
	required bool
	secret   bool
}

type component struct {
	id               string
	description      string
	source           struct{}   // this is a sum type and must be handled in a special way
	files            []struct{} // this is a sum type and must be handled in a special way
	excludeFiles     []string
	allowedHttpHosts []string
	keyValueStores   []string
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
	t          string
	argv       string
	entrypoint string
}

type build struct {
	command string
	workdir string
}
