package config

// Opts are options for configuring the location of the config
type Opts struct {
	// Path is the path to the spin aks config
	Path string
}

type config struct {
	Cluster           Cluster           `toml:"cluster"`
	ContainerRegistry ContainerRegistry `toml:"container_registry"`
	// SpinManifest is the path to the Spin manifest file (spin.toml)
	SpinManifest string `toml:"spin_manifest"`
	// Dockerfile is the path to the Dockerfile
	Dockerfile string `toml:"dockerfile"`
	// K8sResources is the path to the Kubernetes resource files
	K8sResources string `toml:"kubernetes_resources"`
	Store        Store  `toml:"store,omitempty"`
}

type ResourceId struct {
	Subscription  string `toml:"subscription"`
	ResourceGroup string `toml:"resource_group"`
	Name          string `toml:"name"`
}

type Cluster struct {
	ResourceId
}

type ContainerRegistry struct {
	ResourceId
}

type storeKind string

var (
	Redis  storeKind = "redis"
	Cosmos storeKind = "cosmos"
)

type Store struct {
	Kind storeKind `toml:"kind"`
	ResourceId
}

// newish represents a type that can be either instantiated or
// a new instance is requested. If the IsNew field is true then
// a new instance is requested. Otherwise, the Data field should be
// populated with the instance.
type newish[T any] struct {
	Data  *T
	IsNew bool
}
