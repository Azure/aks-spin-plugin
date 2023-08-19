package config

// Opts are options for configuring the location of the config
type Opts struct {
	// Path is the path to the spin aks config
	Path string
}

type config struct {
	Cluster           Cluster
	ContainerRegistry ContainerRegistry
	// spinManifest is the path to the Spin manifest file (spin.toml)
	SpinManifest string
	// dockerfile is the path to the Dockerfile
	Dockerfile string
	// k8sResources is the path to the Kubernetes resource files
	K8sResources string
	Store        Store
}

type ResourceId struct {
	subscription  string
	resourceGroup string
	name          string
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
	kind storeKind
	ResourceId
}
