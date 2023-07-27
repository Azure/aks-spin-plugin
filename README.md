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
- Last used spin aks config path

State will primarily be used to autofill prompts as intelligently as possible. This will allow most users to simply press "enter" through the prompts but also allows more advanced cases to customize usage.

### Config

A config file is where this plugin will read values from.

This file by default will be called aks-spin.toml but can be another toml file that's referenced by the global `-c` of `--config` flag. This allows for users to have different "environments".

### Feature selection

Users will be prompted for various Azure resources when using the cli. The CLI will provide an additional option for creating resources as well.

For example, if a user is asked for an AKS cluster we give them the option to select one but also give them the option to create a new default one.

Any user input should be possible through a CLI flag so that this can be completely consumed by CI/CD.

### Versions

This plugin will be compatible with all spin versions 1.x.x

### Commands

#### spin aks init

Walks users through creating a new spin aks config.

Asks for

- config filename + location (will be autofilled as the default of `./spin-aks.toml` so most users just have to press enter)
- cluster name (subscription, rg, name)
- acr name (subscription, rg, name)
- spin.toml file location (tries to detect automatically)

#### spin aks build

Functions like spin build but also ensures that current Spin application will work for AKS (not all Spin versions are compatible, Spin version should be 1.x.x). Builds the .wasm files needed for the docker image.

#### spin aks scaffold dockerfile

Creates the dockerfile. See the other scaffold command below for more information. Dockerfile is by default output to `./Dockerfile` (with root being where the command is run).

Flags

- `--dockerfile-dest` changes the destination of the Dockerfile. Filename can be included here but defaults to Dockerfile.
- `--override` overrides existing files without prompts. By default the cli prompts.
- `-y` accepts all defaults meaning the cli won't prompt.
- `-c` or `--config` specifies the aks spin toml file location. Defaults to ./aks-spin.toml.
- `--cosmos` indicates that user wants to use cosmos. Can also specify this in the aks spin config.

Ensure that the redis address is a secret. See more under scaffold command for more info.

If cosmos is needed user selects a cosmos instance (or is prompted to make one). https://developer.fermyon.com/spin/dynamic-configuration#key-value-store-runtime-configuration runtime configuration will be created for the user. A secret variable will be added to the spin.toml indicating the cosmos key. TODO: we need to verify that spin variables can be injected into this from the environment. If they can't we should add this to upstream.

#### spin aks push

User selects they are pushing to acr or another registry.

Builds and pushes docker container to registry.

Stores the image tag in the aks spin toml config.

#### spin aks scaffold k8s

Creates manifests. Should give option of Helm, Kustomize, Kube. Helm output to `./charts/`, Kustomize output to `./base` and `./overlays`, and normal K8s files to `./manifests/`.

Flags

- `--k8s-dest` changes the destination of Kustomize, Kube, or Helm files.
- `-t or --type` chooses the type of k8s files. Options are Helm, Kustomize, and Kube.
- `--override` overrides existing files without prompts. By default the cli prompts.
- `-y` accepts all defaults meaning the cli won't prompt.
- `-c` or `--config` specifies the aks spin toml file location. Defaults to ./aks-spin.toml.

If there's already helm files or kustomize files we merge our additions with the existing files.

If we have already created these files, the cli handles updating them to the "latest versions".

Included in the manifests is a KWASM deployment along with deployments + services for the spin application.

Checks the spin.toml variables https://developer.fermyon.com/spin/manifest-reference#the-variables-table. If it's a secret, the user is prompted to select a keyvault secret for this (or is given the option to create a kv secret). Secrets will use the aks kv csi driver to load secrets into the spin application pod. These need to be mounted by the pod according to the csi driver spec (even though we are only using them as env variables in the pod). This will be represented in generated manifests. Secret locations will be stored in the spin aks toml config.

If a spin variable isn't a secret it's configured directly through plaintext env variables on the deployment.

If the trigger is redis, ensure the address is a spin variable secret. If it's not we should prompt user and update to one. When we update to one, make the default their current value. https://developer.fermyon.com/spin/redis-trigger#specifying-an-application-as-redis. We ask the user for the select an Azure Cache for Redis instance and ensure all channels are properly set up on that redis instance. We set the address secret to the redis instance and create a kv secret for it (following the normal secret workflow).

The Dockerfile and k8s file locations are stored in the aks spin toml.

#### spin aks variable put

If variable isn't a secret updates the plaintext env variable representing the secret in the k8s files.

Updates a variable in the keyvault if the variable is a secret. TODO: need to figure out the secret autorotation strategy in a future iteration.

#### spin aks deploy

Applies manifests to the k8s cluster. Also ensures cluster has permission to access acr, if not it prompts to attach.

If secrets are used by the application then we prompt them to install the keyvault csi driver addon. Also prompt to attach the keyvault to the cluster addon identity so we can pull the secrets.

Doesn't support private clusters for now. We can in the future pretty easily thanks to az aks command invoke.

https://helm.sh/docs/topics/advanced/#go-sdk

https://pkg.go.dev/sigs.k8s.io/kustomize/api/krusty#Kustomizer

#### spin aks up

Goes through all steps to ensure your application is running. This is one command for all the things mentioned above.

All steps are idempotent and these commands can be used to update what's running in the cluster to a new application version.
