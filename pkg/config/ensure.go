package config

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/azure/spin-aks-plugin/pkg/azure"
	"github.com/azure/spin-aks-plugin/pkg/logger"
	"github.com/azure/spin-aks-plugin/pkg/prompt"
)

// EnsureValid prompts users for all required fields
func EnsureValid(ctx context.Context) error {
	if err := ensureCluster(ctx); err != nil {
		return fmt.Errorf("ensuring cluster: %w", err)
	}

	if err := ensureAcr(ctx); err != nil {
		return fmt.Errorf("ensuring acr: %w", err)
	}

	if err := ensureSpinManifest(ctx); err != nil {
		return fmt.Errorf("ensuring spin manifest: %w", err)
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

		lgr.Debug("prompting for cluster subscription")
		sub, err := prompt.Select("Select your Cluster's Subscription", subs, &prompt.SelectOpt[armsubscription.Subscription]{
			Field: func(t armsubscription.Subscription) string {
				return *t.DisplayName
			},
		})
		if err != nil {
			return fmt.Errorf("selecting subscription: %w", err)
		}
		c.Cluster.Subscription = *sub.SubscriptionID
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
		clusters, err := azure.ListClusters(ctx, c.Cluster.Subscription, c.Cluster.ResourceGroup)
		if err != nil {
			return fmt.Errorf("listing clusters: %w", err)
		}

		lgr.Debug("prompting for cluster name")
		cluster, err := prompt.Select("Select your Cluster", clusters, &prompt.SelectOpt[armcontainerservice.ManagedCluster]{
			Field: func(t armcontainerservice.ManagedCluster) string {
				return *t.Name
			},
		})
		if err != nil {
			return fmt.Errorf("selecting cluster: %w", err)
		}

		c.Cluster.Name = *cluster.Name
		lgr.Debug("finished prompting for cluster name")
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

		lgr.Debug("prompting for acr subscription")
		sub, err := prompt.Select("Select your Container Registry's Subscription", subs, &prompt.SelectOpt[armsubscription.Subscription]{
			Field: func(t armsubscription.Subscription) string {
				return *t.DisplayName
			},
		})
		if err != nil {
			return fmt.Errorf("selecting subscription: %w", err)
		}

		c.ContainerRegistry.Subscription = *sub.SubscriptionID
		lgr.Debug("finished prompting for acr subscription")
	}

	if c.ContainerRegistry.ResourceGroup == "" {
		rg, err := getResourceGroup(ctx, c.ContainerRegistry.Subscription, "Container Registry's")
		if err != nil {
			return fmt.Errorf("getting container registry's resource group")
		}

		c.ContainerRegistry.ResourceGroup = rg
	}

	if c.ContainerRegistry.Name == "" {
		acrs, err := azure.ListContainerRegistries(ctx, c.ContainerRegistry.Subscription, c.ContainerRegistry.ResourceGroup)
		if err != nil {
			return fmt.Errorf("listing acrs: %w", err)
		}

		lgr.Debug("prompting for acr name")
		acr, err := prompt.Select("Select your Container Registry", acrs, &prompt.SelectOpt[armcontainerregistry.Registry]{
			Field: func(t armcontainerregistry.Registry) string {
				return *t.Name
			},
		})
		if err != nil {
			return fmt.Errorf("selecting acr: %w")
		}

		c.ContainerRegistry.Name = *acr.Name
		lgr.Debug("finished prompting for acr name")
	}

	lgr.Debug("done ensuring acr config")
	return nil
}

func ensureSpinManifest(ctx context.Context) error {
	lgr := logger.FromContext(ctx)
	lgr.Debug("starting to ensure spin manifest")

	if c.SpinManifest == "" {
		lgr.Debug("prompting for spin manifest")

		guess, err := searchFile("spin.toml")
		if err != nil {
			// don't want to fail on attempt to guess
			lgr.Debug("failed to guess spin manifest: " + err.Error())
		}

		manifest, err := prompt.Input("Input your spin manifest location", &prompt.InputOpt{
			Validate: prompt.FileExists,
			Default:  guess,
		})
		if err != nil {
			return fmt.Errorf("inputting spin manifest: %w", err)
		}

		c.SpinManifest = manifest
		lgr.Debug("finished prompting for spin manifest")
	}

	lgr.Debug("done ensuring spin manifest")
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

	rgsWithNew := withNew(rgs)
	lgr.Debug(fmt.Sprintf("prompting for %s resource group"), possessive)
	selection, err := prompt.Select(fmt.Sprintf("Select your %s Resource Group", possessive), rgsWithNew, &prompt.SelectOpt[newish[armresources.ResourceGroup]]{
		Field: func(t newish[armresources.ResourceGroup]) string {
			if t.IsNew {
				return "New Resource Group"
			}

			return *t.Data.Name
		},
	})
	if err != nil {
		return "", fmt.Errorf("prompting for resource group: %w", err)
	}

	if !selection.IsNew {
		lgr.Debug(fmt.Sprintf("finished getting %s resource group", possessive))
		return *selection.Data.Name, nil
	}

	name, err := prompt.Input("Input your new Resource Group name", &prompt.InputOpt{})
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

	lgr.Debug(fmt.Sprintf("finished getting %s resource group", possessive))
	return name, nil
}

func withNew[T any](instantiated []T) []newish[T] {
	ret := make([]newish[T], 0, len(instantiated)+1)

	ret = append(ret, newish[T]{IsNew: true})
	for _, inst := range instantiated {
		ret = append(ret, newish[T]{Data: &inst})
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
