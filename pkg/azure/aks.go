package azure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
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

func NewCluster(ctx context.Context, subscriptionId, resourceGroup, name, location string) error {
	lgr := logger.FromContext(ctx).With("subscription", subscriptionId, "resource group", resourceGroup, "name", name)
	ctx = logger.WithContext(ctx, lgr)
	lgr.Debug("creating AKS cluster")

	client, err := aksFactory(subscriptionId)
	if err != nil {
		return fmt.Errorf("getting aks client: %w", err)
	}

	lgr.Info("creating new Managed Cluster")
	poll, err := client.NewManagedClustersClient().BeginCreateOrUpdate(ctx, resourceGroup, name, armcontainerservice.ManagedCluster{
		// matches dev/test preset cluster configuration
		Name:     &name,
		Location: &location,
		Identity: &armcontainerservice.ManagedClusterIdentity{
			Type: to.Ptr(armcontainerservice.ResourceIdentityTypeSystemAssigned),
		},
		Properties: &armcontainerservice.ManagedClusterProperties{
			DNSPrefix: to.Ptr(name),
			AgentPoolProfiles: []*armcontainerservice.ManagedClusterAgentPoolProfile{
				{
					Name:              to.Ptr("default"),
					VMSize:            to.Ptr("Standard_DS2_v2"),
					Count:             to.Ptr(int32(2)),
					MinCount:          to.Ptr(int32(2)),
					MaxCount:          to.Ptr(int32(10)),
					EnableAutoScaling: to.Ptr(true),
					Mode:              to.Ptr(armcontainerservice.AgentPoolModeSystem),
				},
			},
		},
	}, nil)
	if err != nil {
		return fmt.Errorf("starting to create cluster: %w", err)
	}

	if _, err := pollWithLog(ctx, poll, "still creating new managed cluster"); err != nil {
		return fmt.Errorf("creating cluster: %w", err)
	}

	return nil
}

func PutCluster(ctx context.Context, subscriptionId, resourceGroup string, mc *armcontainerservice.ManagedCluster) error {
	lgr := logger.FromContext(ctx).With("subscription", subscriptionId, "resource group", resourceGroup, "cluster", mc)
	ctx = logger.WithContext(ctx, lgr)
	lgr.Debug("putting AKS cluster")

	client, err := aksFactory(subscriptionId)
	if err != nil {
		return fmt.Errorf("getting aks client: %w", err)
	}

	poll, err := client.NewManagedClustersClient().BeginCreateOrUpdate(ctx, resourceGroup, *mc.Name, *mc, nil)
	if err != nil {
		return fmt.Errorf("starting to pull cluster: %w", err)
	}

	if _, err := pollWithLog(ctx, poll, "still putting new managed cluster"); err != nil {
		return fmt.Errorf("putting cluster: %w", err)
	}

	return nil
}

func pollWithLog[T any](ctx context.Context, p *runtime.Poller[T], msg string) (T, error) {
	lgr := logger.FromContext(ctx)

	resCh := make(chan errorCh[T], 1)
	go func() {
		result, err := p.PollUntilDone(ctx, nil)
		resCh <- errorCh[T]{
			err:  err,
			data: result,
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return *new(T), ctx.Err()
		case res := <-resCh:
			return res.data, res.err
		case <-time.After(10 * time.Second):
			lgr.Info(msg)
		}
	}
}

type errorCh[T any] struct {
	err  error
	data T
}
