package config

// Opts are options for configuring the location of the config
type Opts struct {
	// Path is the path to the spin aks config
	Path string
}

type config struct {
	Cluster           Cluster           `toml:"cluster"`
	ContainerRegistry ContainerRegistry `toml:"container_registry"`
	// spinManifest is the path to the Spin manifest file (spin.toml)
	SpinManifest string `toml:"spin_manifest"`
	// dockerfile is the path to the Dockerfile
	Dockerfile string `toml:"dockerfile"`
	// k8sResources is the path to the Kubernetes resource files
	K8sResources string `toml:"kubernetes_resources"`
	Store        Store  `toml:"store,omitempty"`
}

type ResourceId struct {
	subscription  string `toml:"subscription"`
	resourceGroup string `toml:"resource_group"`
	name          string `toml:"name"`
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
