package azure

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/azure/spin-aks-plugin/pkg/logger"
)

func acrFactory(subscriptionId string) (*armcontainerregistry.ClientFactory, error) {
	cred, err := getCred()
	if err != nil {
		return nil, fmt.Errorf("getting credential: %w", err)
	}

	factory, err := armcontainerregistry.NewClientFactory(subscriptionId, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating factory: %w", err)
	}

	return factory, nil
}

func ListContainerRegistries(ctx context.Context, subscriptionId, resourceGroup string) ([]armcontainerregistry.Registry, error) {
	lgr := logger.FromContext(ctx).With("subscription", subscriptionId, "resource group", resourceGroup)
	ctx = logger.WithContext(ctx, lgr)
	lgr.Debug("listing ACRs")

	client, err := acrFactory(subscriptionId)
	if err != nil {
		return nil, fmt.Errorf("getting container registry client: %w", err)
	}

	var acrs []armcontainerregistry.Registry
	pager := client.NewRegistriesClient().NewListByResourceGroupPager(resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing container registry page: %w", err)
		}

		for _, acr := range page.Value {
			if acr == nil {
				return nil, errors.New("nil acr")
			}

			acrs = append(acrs, *acr)
		}
	}

	lgr.Debug("finished listing ACRs")
	return acrs, nil
}

func NewContainerRegistry(ctx context.Context, subscriptionId, resourceGroup, name, location string) error {
	lgr := logger.FromContext(ctx).With("subscription", subscriptionId, "resourceGroup", resourceGroup)
	ctx = logger.WithContext(ctx, lgr)
	lgr.Debug("creating new container registry")

	factory, err := acrFactory(subscriptionId)
	if err != nil {
		return fmt.Errorf("getting acr factory: %w", err)
	}

	lgr.Info("creating new Container Registry")
	poll, err := factory.NewRegistriesClient().BeginCreate(ctx, resourceGroup, name, armcontainerregistry.Registry{
		Name:     &name,
		Location: &location,
		SKU: &armcontainerregistry.SKU{
			Name: to.Ptr(armcontainerregistry.SKUNameBasic),
		},
	}, nil)
	if err != nil {
		return fmt.Errorf("starting to create container registry: %w", err)
	}

	if _, err := poll.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("creating container registry: %w", err)
	}

	return nil
}

func EnableKeyvaultCSIDriver(ctx context.Context, subscriptionId, resourceGroup string, mc *armcontainerservice.ManagedCluster) error {
	lgr := logger.FromContext(ctx).With("cluster", mc)
	ctx = logger.WithContext(ctx, lgr)
	lgr.Debug("enabling keyvault CSI driver for cluster")

	if mc == nil {
		return errors.New("managed cluster cannot be nil to install keyvault csi driver")
	}

	if mc.Properties == nil {
		mc.Properties = &armcontainerservice.ManagedClusterProperties{}
	}

	if mc.Properties.AddonProfiles != nil {
		mc.Properties.AddonProfiles["azureKeyvaultSecretsProvider"] = &armcontainerservice.ManagedClusterAddonProfile{
			Enabled: to.Ptr(true),
			Config: map[string]*string{
				"enableSecretRotation": to.Ptr("true"),
			},
		}
	} else {
		mc.Properties.AddonProfiles = map[string]*armcontainerservice.ManagedClusterAddonProfile{
			"azureKeyvaultSecretsProvider": {
				Enabled: to.Ptr(true),
				Config: map[string]*string{
					"enableSecretRotation": to.Ptr("true"),
				},
			},
		}
	}

	err := PutCluster(ctx, subscriptionId, resourceGroup, mc)
	if err != nil {
		return fmt.Errorf("putting cluster: %w", err)
	}

	return nil
}
