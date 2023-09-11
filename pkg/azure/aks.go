package azure

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v2"
	"github.com/azure/spin-aks-plugin/pkg/logger"
)

func aksFactory(subscriptionId string) (*armcontainerservice.ClientFactory, error) {
	cred, err := getCred()
	if err != nil {
		return nil, fmt.Errorf("getting credential: %w", err)
	}

	factory, err := armcontainerservice.NewClientFactory(subscriptionId, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating factory: %w", err)
	}

	return factory, nil
}

func ListClusters(ctx context.Context, subscriptionId, resourceGroup string) ([]armcontainerservice.ManagedCluster, error) {
	lgr := logger.FromContext(ctx).With("subscription", subscriptionId, "resource group", resourceGroup)
	ctx = logger.WithContext(ctx, lgr)
	lgr.Debug("listing AKS clusters")

	client, err := aksFactory(subscriptionId)
	if err != nil {
		return nil, fmt.Errorf("getting aks client: %w", err)
	}

	var clusters []armcontainerservice.ManagedCluster
	pager := client.NewManagedClustersClient().NewListByResourceGroupPager(resourceGroup, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing managed clusters page: %w", err)
		}

		for _, cluster := range page.Value {
			if cluster == nil {
				return nil, errors.New("nil cluster")
			}

			clusters = append(clusters, *cluster)
		}
	}

	lgr.Debug("finished listing AKS clusters")
	return clusters, nil
}
