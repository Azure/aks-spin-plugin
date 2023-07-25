# spin-aks-plugin

A Spin Azure Kubernetes Service plugin with cloud provider support

## Plan

### Auth

We use the [azidentity DefaultAzureCredential](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity#NewDefaultAzureCredential) to authenticate to the Azure SDKs. This will leave users with different options for configuration that match other Azure tooling. Some of the options users will ahve include `az login` (az cli) and env variables.

### State

Stores a "state" in `$(DATA_DIR)/spin/plugins/aks/state`. This follows the Spin model described [here](https://developer.fermyon.com/spin/cache).

State that's stored includes the following
- Subscription, resource group, cluster name of most recently selected cluster
- Subscription, resource group, cluster name of most recently selected container registry
- Registry, image name, tag of most recently deployed image per application

State will primarily be used to autofill prompts as intelligently as possible. This will allow most users to simply press "enter" through the prompts but also allows more advanced cases to customzie usage.

### Commands

#### spin aks scaffold

TODO: is this name good?

Creates Dockerfile and manifests. Should give option of Helm, Kustomize, Kube.

Optional `-o`` or `--output` flag to specify output directory. 

Optional `-y` flag to accept all defaults.

TODO: how to handle wanting Dockerfile and manifests to be in different directories?

TODO: can we leverage Draft to handle this?

TODO: this should also handle updating existing manifests to new forms.

#### spin aks build

Functions like spin build but also ensures that current Spin application will work for AKS (not all Spin versions are compatible). Builds the .wasm files needed for the docker image.

https://docs.docker.com/engine/api/sdk/

#### spin aks push

User selects they are pushing to acr or another registry.

Pushes docker container to registry.

https://docs.docker.com/engine/api/sdk/

TODO: how to configure? support more than acr?

#### spin aks apply

Applies manifests k8s cluster. Maybe this should be deploy? We give option of Helm, Kube, Kustomize.

TODO: how does this work for a private cluster? Could use command invoke https://learn.microsoft.com/en-us/azure/aks/command-invoke.

https://helm.sh/docs/topics/advanced/#go-sdk

#### spin aks up

todo: is this the right name?

Goes through all steps to ensure your application is running



## TODO: cluster provisioning, more detailed information about cluster / spin matrix, ensuring spin is okay for cluster

## TODO: handle spin variables https://github.com/fermyon/spin/blob/a3d97b1aefb912ff02313875ab4f1b3c0364dac1/docs/content/sips/002-app-config.md and secrets. use annotations to seperate acr spin variables from others? Must support azure keyvault because k8s secrets are not secure.

## TODO: how to handle azure kv and acr permissions? Do we prompt to attach?

## TODO: key value store with cosmos db https://developer.fermyon.com/spin/dynamic-configuration#key-value-store-runtime-configuration

## TODO: what do we have to do for redis applications?