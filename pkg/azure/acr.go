package azure

import (
	"context"
	"errors"
	"fmt"

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
