package config

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/azure/spin-aks-plugin/pkg/azure"
	"github.com/azure/spin-aks-plugin/pkg/logger"
	"github.com/azure/spin-aks-plugin/pkg/prompt"
	"github.com/azure/spin-aks-plugin/pkg/spin"
	"github.com/azure/spin-aks-plugin/pkg/state"
)

const (
	subscriptionKey      = "subscription"
	resourceGroupKey     = "resourceGroup"
	clusterKey           = "cluster"
	containerRegistryKey = "containerRegistry"
	spinManifestKey      = "spinManifest"
	keyVaultKey          = "keyVault"
)

var (
	alphanumUnderscoreParenHyphenPeriodRegex = regexp.MustCompile("^[a-zA-Z0-9_()\\-.]+$")
	alphanumUnderscoreHyphenRegex            = regexp.MustCompile("^[a-zA-Z0-9_\\-]+$")
	alphanumHyphenRegex                      = regexp.MustCompile("^[a-zA-Z0-9\\-]+$")
	alphanumRegex                            = regexp.MustCompile("^[a-zA-Z0-9]+$")
)

// EnsureValid prompts users for all required fields
func EnsureValid(ctx context.Context) error {
	if err := ensureCluster(ctx); err != nil {
		return fmt.Errorf("ensuring cluster: %w", err)
	}

	if err := ensureAcr(ctx); err != nil {
		return fmt.Errorf("ensuring acr: %w", err)
	}

	m, err := ensureSpinManifest(ctx)
	if err != nil {
		return fmt.Errorf("ensuring spin manifest: %w", err)
	}

	if err := ensureKeyVault(ctx, m); err != nil {
		return fmt.Errorf("ensuring keyvault: %w", err)
	}

	return nil
}

func ensureCluster(ctx context.Context) error {
	lgr := logger.FromContext(ctx)
	lgr.Debug("starting to ensure cluster config")

	if c.Cluster.Subscription == "" {
		subs, err := azure.ListSubscriptions(ctx)
		if err != nil {
			return fmt.Errorf("listing subscriptions: %w", err)
		}

		def, err := state.Get(ctx, subscriptionKey)
		if err != nil && !errors.Is(err, state.KeyNotFoundErr) {
			// failing to get subscription from state is not worth failing
			lgr.Debug("failed to get subscription from state: " + err.Error())
			def = ""
		}

		lgr.Debug("prompting for cluster subscription")
		sub, err := prompt.Select("Select your Cluster's Subscription", subs, &prompt.SelectOpt[armsubscription.Subscription]{
			Field: func(t armsubscription.Subscription) string {
				return *t.DisplayName
			},
			Default: def,
		})
		if err != nil {
			return fmt.Errorf("selecting subscription: %w", err)
		}
		c.Cluster.Subscription = *sub.SubscriptionID

		if err := state.Set(ctx, subscriptionKey, *sub.DisplayName); err != nil {
			// failing to set subscription in state is not worth failing
			lgr.Debug("failed to set subscription in state: " + err.Error())
		}

		lgr.Debug("finished prompting for cluster subscription")
	}

	if c.Cluster.ResourceGroup == "" {
		rg, err := getResourceGroup(ctx, c.Cluster.Subscription, "Cluster's")
		if err != nil {
			return fmt.Errorf("getting cluster resource group: %w", err)
		}

		c.Cluster.ResourceGroup = rg
	}

	if c.Cluster.Name == "" {
		cluster, err := GetClusterName(ctx, c.Cluster.Subscription, c.Cluster.ResourceGroup)
		if err != nil {
			return fmt.Errorf("getting cluster name: %w", err)
		}

		c.Cluster.Name = cluster
	}

	lgr.Debug("done ensuring cluster config")
	return nil
}

func ensureAcr(ctx context.Context) error {
	lgr := logger.FromContext(ctx)
	lgr.Debug("starting to ensure acr config")

	if c.ContainerRegistry.Subscription == "" {
		subs, err := azure.ListSubscriptions(ctx)
		if err != nil {
			return fmt.Errorf("listing subscriptions: %w", err)
		}

		def, err := state.Get(ctx, subscriptionKey)
		if err != nil && !errors.Is(err, state.KeyNotFoundErr) {
			// failing to get subscription from state is not worth failing
			lgr.Debug("failed to get subscription from state: " + err.Error())
			def = ""
		}

		lgr.Debug("prompting for acr subscription")
		sub, err := prompt.Select("Select your Container Registry's Subscription", subs, &prompt.SelectOpt[armsubscription.Subscription]{
			Field: func(t armsubscription.Subscription) string {
				return *t.DisplayName
			},
			Default: def,
		})
		if err != nil {
			return fmt.Errorf("selecting subscription: %w", err)
		}
		c.ContainerRegistry.Subscription = *sub.SubscriptionID

		if err := state.Set(ctx, subscriptionKey, *sub.DisplayName); err != nil {
			// failing to set subscription in state is not worth failing
			lgr.Debug("failed to set subscription in state: " + err.Error())
		}

		lgr.Debug("finished prompting for acr subscription")
	}

	if c.ContainerRegistry.ResourceGroup == "" {
		rg, err := getResourceGroup(ctx, c.ContainerRegistry.Subscription, "Container Registry's")
		if err != nil {
			return fmt.Errorf("getting container registry's resource group: %w", err)
		}

		c.ContainerRegistry.ResourceGroup = rg
	}

	if c.ContainerRegistry.Name == "" {
		acr, err := getContainerRegistry(ctx, c.ContainerRegistry.Subscription, c.ContainerRegistry.ResourceGroup)
		if err != nil {
			return fmt.Errorf("getting container registry's name: %w", err)
		}

		c.ContainerRegistry.Name = acr
	}

	lgr.Debug("done ensuring acr config")
	return nil
}

func ensureSpinManifest(ctx context.Context) (spin.Manifest, error) {
	lgr := logger.FromContext(ctx)
	lgr.Debug("starting to ensure spin manifest")
	m := spin.Manifest{}

	if c.SpinManifest == "" {
		lgr.Debug("prompting for spin manifest")

		guess, err := searchFile("spin.toml")
		if err != nil {
			// don't want to fail on attempt to guess
			lgr.Debug("failed to guess spin manifest: " + err.Error())
		}

		if guess == "" {
			def, err := state.Get(ctx, spinManifestKey)
			if err == nil {
				guess = def
			}
		}

		manifest, err := prompt.Input("Input your spin manifest location", &prompt.InputOpt{
			Validate: prompt.FileExists,
			Default:  guess,
		})
		if err != nil {
			return m, fmt.Errorf("inputting spin manifest: %w", err)
		}

		c.SpinManifest = manifest

		if err := state.Set(ctx, spinManifestKey, manifest); err != nil {
			// failing to set spin manifest in state is not worth failing
			lgr.Debug("failed to set spin manifest in state: " + err.Error())
		}

		lgr.Debug("finished prompting for spin manifest")
	}

	m, err := spin.Load(c.SpinManifest)
	if err != nil {
		return m, fmt.Errorf("loading spin manifest: %w", err)
	}

	lgr.Debug("done ensuring spin manifest")
	return m, nil
}

func ensureKeyVault(ctx context.Context, m spin.Manifest) error {
	lgr := logger.FromContext(ctx)
	lgr.Debug("starting to ensure keyvault config")

	lgr.Debug(fmt.Sprintf("found %d variables", len(m.Variables)))
	hasSecretVariable := false
	for _, v := range m.Variables {
		hasSecretVariable = hasSecretVariable || v.Secret
	}

	if hasSecretVariable {
		lgr.Debug("found at least one secret variable, prompting for keyvault")

		subs, err := azure.ListSubscriptions(ctx)
		if err != nil {
			return fmt.Errorf("listing subscriptions: %w", err)
		}

		def, err := state.Get(ctx, subscriptionKey)
		if err != nil && !errors.Is(err, state.KeyNotFoundErr) {
			// failing to get subscription from state is not worth failing
			lgr.Debug("failed to get subscription from state: " + err.Error())
			def = ""
		}

		lgr.Debug("prompting for keyvault subscription")
		sub, err := prompt.Select("Select your KeyVault's Subscription", subs, &prompt.SelectOpt[armsubscription.Subscription]{
			Field: func(t armsubscription.Subscription) string {
				return *t.DisplayName
			},
			Default: def,
		})
		if err != nil {
			return fmt.Errorf("selecting subscription: %w", err)
		}
		c.KeyVault.Subscription = *sub.SubscriptionID

		if err := state.Set(ctx, subscriptionKey, *sub.DisplayName); err != nil {
			// failing to set subscription in state is not worth failing
			lgr.Debug("failed to set subscription in state: " + err.Error())
		}

		if c.KeyVault.ResourceGroup == "" {
			rg, err := getResourceGroup(ctx, c.KeyVault.Subscription, "KeyVault's")
			if err != nil {
				return fmt.Errorf("getting keyvault's resource group: %w", err)
			}

			c.KeyVault.ResourceGroup = rg
		}
		if c.KeyVault.Name == "" {
			kv, err := getKeyVault(ctx, c.KeyVault.Subscription, c.KeyVault.ResourceGroup)
			if err != nil {
				return fmt.Errorf("getting keyvault's name: %w", err)
			}

			c.KeyVault.Name = kv
		}

		// TODO check keyvault access policy for existing keyvaults to see if we need to add the cluster's identity or user put/get permissions

		akv, err := azure.GetKeyVault(ctx, c.KeyVault.Subscription, c.KeyVault.ResourceGroup, c.KeyVault.Name)
		c.TenantID = akv.TenantId
		if err != nil {
			return fmt.Errorf("getting keyvault: %w", err)
		}

		cluster, err := azure.GetManagedCluster(ctx, c.Cluster.Subscription, c.Cluster.ResourceGroup, c.Cluster.Name)
		clusterId := *cluster.Identity.PrincipalID
		err = akv.AddAccessPolicy(ctx, clusterId, armkeyvault.Permissions{
			Secrets: []*armkeyvault.SecretPermissions{to.Ptr(armkeyvault.SecretPermissionsGet)},
		})
		if err != nil {
			return fmt.Errorf("adding keyvault access policy for cluster: %w", err)
		}

		err = akv.AddUserAccessPolicy(ctx, armkeyvault.Permissions{
			Secrets: []*armkeyvault.SecretPermissions{
				to.Ptr(armkeyvault.SecretPermissionsGet),
				to.Ptr(armkeyvault.SecretPermissionsSet),
			},
		})
		if err != nil {
			return fmt.Errorf("adding keyvault access policy for user: %w", err)
		}
	} else {
		lgr.Debug("no secret variables found, skipping keyvault")
	}

	return nil

}

// getResourceGroup goes through steps of prompting user for a resource group. Possessive is the possessive
// form of what the resource group would be used for. For example "Cluster's" would be passed in as possessive
// if we are getting the resource group for the cluster
func getResourceGroup(ctx context.Context, subscriptionId, possessive string) (string, error) {
	lgr := logger.FromContext(ctx)
	lgr.Debug(fmt.Sprintf("starting to get %s resource group", possessive))

	if subscriptionId == "" {
		return "", errors.New("subscriptionId is empty")
	}

	rgs, err := azure.ListResourceGroups(ctx, subscriptionId)
	if err != nil {
		return "", fmt.Errorf("listing resource groups: %w", err)
	}

	def, err := state.Get(ctx, resourceGroupKey)
	if err != nil && !errors.Is(err, state.KeyNotFoundErr) {
		// failing to get resource group from state is not worth failing
		lgr.Debug("failed to get resource group from state: " + err.Error())
		def = ""
	}

	rgsWithNew := withNew(rgs)
	lgr.Debug(fmt.Sprintf("prompting for %s resource group", possessive))
	selection, err := prompt.Select(fmt.Sprintf("Select your %s Resource Group", possessive), rgsWithNew, &prompt.SelectOpt[newish[armresources.ResourceGroup]]{
		Field: func(t newish[armresources.ResourceGroup]) string {
			if t.IsNew {
				return "New Resource Group"
			}

			return *t.Data.Name
		},
		Default: def,
	})
	if err != nil {
		return "", fmt.Errorf("prompting for resource group: %w", err)
	}

	if !selection.IsNew {
		if err := state.Set(ctx, resourceGroupKey, *selection.Data.Name); err != nil {
			// failing to set resource group in state is not worth failing
			lgr.Debug("failed to set resource group in state: " + err.Error())
		}

		lgr.Debug(fmt.Sprintf("finished getting %s resource group", possessive))
		return *selection.Data.Name, nil
	}

	name, err := prompt.Input("Input your new Resource Group name", &prompt.InputOpt{
		Validate: validateResourceGroup,
	})
	if err != nil {
		return "", fmt.Errorf("inputting new resource group name: %w", err)
	}

	locations, err := azure.ListLocations(ctx, subscriptionId)
	if err != nil {
		return "", fmt.Errorf("listing locations: %w", err)
	}

	location, err := prompt.Select("Input your new Resource Group location", locations, &prompt.SelectOpt[armsubscriptions.Location]{
		Field: func(t armsubscriptions.Location) string {
			return *t.DisplayName
		},
	})
	if err != nil {
		return "", fmt.Errorf("selecting new resource group location: %w", err)
	}

	if err := azure.NewResourceGroup(ctx, subscriptionId, name, *location.Name); err != nil {
		return "", fmt.Errorf("creating new resource group: %w", err)
	}
	lgr.Info("created Resource Group " + name)

	if err := state.Set(ctx, resourceGroupKey, name); err != nil {
		// failing to set resource group in state is not worth failing
		lgr.Debug("failed to set resource group in state: " + err.Error())
	}

	lgr.Debug(fmt.Sprintf("finished getting %s resource group", possessive))
	return name, nil
}

func GetClusterName(ctx context.Context, subscriptionId, resourceGroup string) (string, error) {
	lgr := logger.FromContext(ctx)
	lgr.Debug("starting to get cluster")

	if subscriptionId == "" {
		return "", errors.New("subscriptionId is  empty")
	}
	if resourceGroup == "" {
		return "", errors.New("resourceGroup is empty")
	}

	clusters, err := azure.ListClusters(ctx, c.Cluster.Subscription, c.Cluster.ResourceGroup)
	if err != nil {
		return "", fmt.Errorf("listing clusters: %w", err)
	}

	def, err := state.Get(ctx, clusterKey)
	if err != nil && !errors.Is(err, state.KeyNotFoundErr) {
		// failing to get cluster from state is not worth failing
		lgr.Debug("failed to get cluster from state: " + err.Error())
		def = ""
	}

	clustersWithNew := withNew(clusters)
	selection, err := prompt.Select("Select your Cluster", clustersWithNew, &prompt.SelectOpt[newish[armcontainerservice.ManagedCluster]]{
		Field: func(t newish[armcontainerservice.ManagedCluster]) string {
			if t.IsNew {
				return "New Managed Cluster"
			}

			return *t.Data.Name
		},
		Default: def,
	})
	if err != nil {
		return "", fmt.Errorf("selecting cluster: %w", err)
	}

	if !selection.IsNew {
		if err := state.Set(ctx, clusterKey, *selection.Data.Name); err != nil {
			// failing to set cluster in state is not worth failing
			lgr.Debug("failed to set cluster in state: " + err.Error())
		}

		lgr.Debug("finished getting cluster")
		return *selection.Data.Name, nil
	}

	name, err := prompt.Input("Input your new Managed Cluster name", &prompt.InputOpt{
		Validate: validateCluster,
	})
	if err != nil {
		return "", fmt.Errorf("inputting new managed cluster name: %w", err)
	}

	locations, err := azure.ListLocations(ctx, subscriptionId)
	if err != nil {
		return "", fmt.Errorf("listing locations: %w", err)
	}

	location, err := prompt.Select("Input your new Managed Cluster location", locations, &prompt.SelectOpt[armsubscriptions.Location]{
		Field: func(t armsubscriptions.Location) string {
			return *t.DisplayName
		},
	})
	if err != nil {
		return "", fmt.Errorf("selecting new managed cluster location: %w", err)
	}

	if err := azure.NewCluster(ctx, subscriptionId, resourceGroup, name, *location.Name); err != nil {
		return "", fmt.Errorf("creating new managed cluster: %w", err)
	}
	lgr.Info("created Managed Cluster " + name)

	if err := state.Set(ctx, clusterKey, name); err != nil {
		// failing to set cluster in state is not worth failing
		lgr.Debug("failed to set cluster in state: " + err.Error())
	}

	lgr.Debug("finished getting cluster")
	return name, nil
}

func getContainerRegistry(ctx context.Context, subscriptionId, resourceGroup string) (string, error) {
	lgr := logger.FromContext(ctx)
	lgr.Debug("starting to get container registry")

	if subscriptionId == "" {
		return "", errors.New("subscriptionId is empty")
	}
	if resourceGroup == "" {
		return "", errors.New("resourceGroup is empty")
	}

	acrs, err := azure.ListContainerRegistries(ctx, c.ContainerRegistry.Subscription, c.ContainerRegistry.ResourceGroup)
	if err != nil {
		return "", fmt.Errorf("listing acrs: %w", err)
	}

	def, err := state.Get(ctx, containerRegistryKey)
	if err != nil && !errors.Is(err, state.KeyNotFoundErr) {
		// failing to get container registry from state is not worth failing
		lgr.Debug("failed to get container registry from state: " + err.Error())
		def = ""
	}

	acrsWithNew := withNew(acrs)
	selection, err := prompt.Select("Select your Container Registry", acrsWithNew, &prompt.SelectOpt[newish[armcontainerregistry.Registry]]{
		Field: func(t newish[armcontainerregistry.Registry]) string {
			if t.IsNew {
				return "New Container Registry"
			}

			return *t.Data.Name
		},
		Default: def,
	})
	if err != nil {
		return "", fmt.Errorf("selecting container registry: %w", err)
	}

	if !selection.IsNew {
		if err := state.Set(ctx, containerRegistryKey, *selection.Data.Name); err != nil {
			// failing to set container registry in state is not worth failing
			lgr.Debug("failed to set container registry in state: " + err.Error())
		}

		lgr.Debug("finished getting container registry")
		return *selection.Data.Name, nil
	}

	name, err := prompt.Input("Input your new Container Registry name", &prompt.InputOpt{
		Validate: validateContainerRegistry,
	})
	if err != nil {
		return "", fmt.Errorf("inputting new container registry name: %w", err)
	}

	locations, err := azure.ListLocations(ctx, subscriptionId)
	if err != nil {
		return "", fmt.Errorf("listing locations: %w", err)
	}

	location, err := prompt.Select("Input your new Container Registry location", locations, &prompt.SelectOpt[armsubscriptions.Location]{
		Field: func(t armsubscriptions.Location) string {
			return *t.DisplayName
		},
	})
	if err != nil {
		return "", fmt.Errorf("selecting new container registry location: %w", err)
	}

	if err := azure.NewContainerRegistry(ctx, subscriptionId, resourceGroup, name, *location.Name); err != nil {
		return "", fmt.Errorf("creating new container registry: %w", err)
	}
	lgr.Info("created Container Registry " + name)

	if err := state.Set(ctx, containerRegistryKey, name); err != nil {
		// failing to set container registry in state is not worth failing
		lgr.Debug("failed to set container registry in state: " + err.Error())
	}

	lgr.Debug("finished getting container registry")
	return name, nil
}

func getKeyVault(ctx context.Context, subscriptionId, resourceGroup string) (string, error) {
	lgr := logger.FromContext(ctx)
	lgr.Debug("starting to get keyvault")

	if subscriptionId == "" {
		return "", errors.New("subscriptionId is empty")
	}
	if resourceGroup == "" {
		return "", errors.New("resourceGroup is empty")
	}

	kvs, err := azure.ListKeyVaults(ctx, c.ContainerRegistry.Subscription, c.ContainerRegistry.ResourceGroup)
	if err != nil {
		return "", fmt.Errorf("listing kvs: %w", err)
	}

	def, err := state.Get(ctx, keyVaultKey)
	if err != nil && !errors.Is(err, state.KeyNotFoundErr) {
		// failing to get key vault from state is not worth failing
		lgr.Debug("failed to get key vault from state: " + err.Error())
		def = ""
	}

	kvsWithNew := withNew(kvs)
	selection, err := prompt.Select("Select your KeyVault", kvsWithNew, &prompt.SelectOpt[newish[azure.Akv]]{
		Field: func(t newish[azure.Akv]) string {
			if t.IsNew {
				return "New KeyVault"
			}

			return t.Data.Name
		},
		Default: def,
	})
	if err != nil {
		return "", fmt.Errorf("selecting keyvault: %w", err)
	}

	if !selection.IsNew {
		if err := state.Set(ctx, keyVaultKey, selection.Data.Name); err != nil {
			// failing to set container registry in state is not worth failing
			lgr.Debug("failed to set keyvault in state: " + err.Error())
		}

		lgr.Debug("finished getting keyvault")
		return selection.Data.Name, nil
	}

	name, err := prompt.Input("Input your new KeyVault name", &prompt.InputOpt{
		Validate: validateKeyVault,
	})
	if err != nil {
		return "", fmt.Errorf("inputting new keyvault name: %w", err)
	}

	locations, err := azure.ListLocations(ctx, subscriptionId)
	if err != nil {
		return "", fmt.Errorf("listing locations: %w", err)
	}

	location, err := prompt.Select("Input your new KeyVault location", locations, &prompt.SelectOpt[armsubscriptions.Location]{
		Field: func(t armsubscriptions.Location) string {
			return *t.DisplayName
		},
	})
	if err != nil {
		return "", fmt.Errorf("selecting new keyvault location: %w", err)
	}

	t, err := azure.GetTenant(ctx)
	if err != nil {
		return "", fmt.Errorf("getting tenant: %w", err)
	}
	tenantId := *t.TenantID

	_, err = azure.NewAkv(ctx, tenantId, subscriptionId, resourceGroup, name, *location.Name)
	if err != nil {
		return "", fmt.Errorf("creating new container registry: %w", err)
	}
	lgr.Info("created KeyVault" + name)

	if err := state.Set(ctx, keyVaultKey, name); err != nil {
		// failing to set container registry in state is not worth failing
		lgr.Debug("failed to set keyvault in state: " + err.Error())
	}

	lgr.Debug("enabling KeyVault CSI driver add-on")
	err = azure.EnableKeyvaultCSIDriver(ctx, c.Cluster.Subscription, c.Cluster.ResourceGroup, c.Cluster.Name)
	if err != nil {
		return "", fmt.Errorf("enabling CSI driver add-on: %w", err)
	}
	lgr.Debug("finished enabling KeyVault CSI driver add-on")

	lgr.Debug("finished getting keyvault")
	return name, nil
}
func withNew[T any](instantiated []T) []newish[T] {
	ret := make([]newish[T], 0, len(instantiated)+1)

	ret = append(ret, newish[T]{IsNew: true})
	for _, inst := range instantiated {
		func(t T) { // needed for loop variable capture
			ret = append(ret, newish[T]{Data: &t})
		}(inst)
	}

	return ret
}

// searchFile searches for the file inside the current path and recursively searches directories.
// returns full path to file if found.
func searchFile(filename string) (string, error) {
	ret := ""
	if err := filepath.Walk(".",
		func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && info.Name() == filename {
				ret = path
			}

			return nil
		}); err != nil {
		return "", fmt.Errorf("searching for file %s: %w", filename, err)
	}

	return ret, nil
}

func validateResourceGroup(rg string) error {
	if len(rg) == 0 || len(rg) > 90 {
		return errors.New("must be between 1 and 90 characters long")
	}

	if !alphanumUnderscoreParenHyphenPeriodRegex.MatchString(rg) {
		return errors.New("must contain only alphanumerics, underscores, parentheses, hyphens, and periods")
	}

	if rg[len(rg)-1:] == "." {
		return errors.New("cannot end in a period")
	}

	return nil
}

func validateCluster(cluster string) error {
	if len(cluster) == 0 || len(cluster) > 63 {
		return errors.New("must be between 1 and 63 characters long")
	}

	if !alphanumUnderscoreHyphenRegex.MatchString(cluster) {
		return errors.New("must contain only alphanumerics, underscores, and hyphens")
	}

	if !alphanumRegex.MatchString(cluster[0:1]) {
		return errors.New("must start with an alphanumeric character")
	}

	if !alphanumRegex.MatchString(cluster[len(cluster)-1:]) {
		return errors.New("must end with an alphanumeric character")
	}

	return nil
}

func validateContainerRegistry(cr string) error {
	if len(cr) == 0 || len(cr) > 50 {
		return errors.New("must be between 1 and 50 characters long")
	}

	if !alphanumRegex.MatchString(cr) {
		return errors.New("must contain only alphanumerics")
	}

	return nil
}

func validateKeyVault(kv string) error {
	if len(kv) < 3 || len(kv) > 24 {
		return errors.New("must be between 1 and 90 characters long")
	}

	if !alphanumHyphenRegex.MatchString(kv) {
		return errors.New("must contain only alphanumerics and hyphens")
	}

	if !unicode.IsLetter(rune(kv[0])) {
		return errors.New("must start with a letter")
	}

	if strings.Contains(kv, "--") {
		return errors.New("cannot contain consecutive hyphens")
	}

	lastLetter, _ := utf8.DecodeLastRuneInString(kv)
	if !unicode.IsLetter(lastLetter) && !unicode.IsNumber(lastLetter) {
		return errors.New("must end with a letter or number")
	}

	return nil
}
