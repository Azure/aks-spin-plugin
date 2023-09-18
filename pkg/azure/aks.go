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
	"github.com/google/uuid"
)

const (
	acrResourceIdTemplate = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ContainerRegistry/registries/%s"
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

func LinkAcr(ctx context.Context, subscriptionId, clusterResourceGroup, clusterName, acrResourceGroup, acrName string) error {
	lgr := logger.FromContext(ctx).With("subscription", subscriptionId, "resource group", clusterResourceGroup, "cluster name", clusterName,
		"acr name", acrName)
	ctx = logger.WithContext(ctx, lgr)
	lgr.Debug("linking ACR")

	// add validation for acr?

	client, err := aksFactory(subscriptionId)
	if err != nil {
		return fmt.Errorf("getting aks client: %w", err)
	}

	lgr.Debug("getting cluster information")
	cluster, err := client.NewManagedClustersClient().Get(ctx, clusterResourceGroup, clusterName, nil)

	if err != nil {
		return fmt.Errorf("getting cluster information: %w", err)
	}

	var assigneeId *string

	if cluster.Identity == nil {
		return fmt.Errorf("serviceprincipal clusters are not supported at this time")
		//lgr.Debug("detected service principal cluster")
		//clientId := cluster.ManagedCluster.Properties.ServicePrincipalProfile.ClientID
		//
		//if clientId == nil {
		//	return fmt.Errorf("client id for sp is nil")
		//}
		//
		//tenant, err := ListTenants(ctx)
		//if err != nil {
		//	return fmt.Errorf("listing tenants: %w", err)
		//}
		//tenantId := *tenant[0].TenantID
		//
		//aadClient, err := NewAadClient(ctx, tenantId)
		//if err != nil {
		//	return fmt.Errorf("creating aad client: %w", err)
		//}
		//
		//spObjId, err := aadClient.getObjectIdFromClientId(ctx, *clientId)
		//if err != nil {
		//	return fmt.Errorf("getting object id from client id: %w", err)
		//}
		//
		//assigneeId = spObjId

	} else {
		switch *cluster.Identity.Type {
		case armcontainerservice.ResourceIdentityTypeSystemAssigned:
			lgr.Debug("detected system-assigned identity cluster")
			assigneeId = cluster.Identity.PrincipalID
		case armcontainerservice.ResourceIdentityTypeUserAssigned:
			lgr.Debug("detected user-assigned identity cluster")
			// https://github.com/Azure/azure-cli/blob/8f91d71e8c3af9ab10024e12c51a0dab573df9f2/src/azure-cli/azure/cli/command_modules/acs/managed_cluster_decorator.py#L6177
			msiInfo, ok := cluster.Properties.IdentityProfile["kubeletidentity"]
			fmt.Println(cluster.Properties.IdentityProfile)
			if !ok {
				return errors.New("missing kubeletidentity on User Assigned Identity cluster")
			}
			assigneeId = msiInfo.ObjectID
		default:
			return fmt.Errorf("unknown cluster identity type")
		}
	}
	if assigneeId == nil {
		return errors.New("missing principal id for cluster")
	}

	raClient, err := createRoleAssignmentClient(subscriptionId)
	if err != nil {
		return fmt.Errorf("creating role assignment client: %w", err)
	}

	scope := fmt.Sprintf(acrResourceIdTemplate, subscriptionId, acrResourceGroup, acrName)

	raUid := uuid.New().String()
	err = raClient.createRoleAssignment(ctx, *assigneeId, acrPullRoleDefinition, scope, raUid)
	return err
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

func GetCluster(ctx context.Context, subscriptionId, resourceGroup, clusterName string) (*armcontainerservice.ManagedCluster, error) {
	lgr := logger.FromContext(ctx).With("subscription", subscriptionId, "resource group", resourceGroup, "cluster name", clusterName)
	ctx = logger.WithContext(ctx, lgr)
	lgr.Debug("getting AKS cluster")

	client, err := aksFactory(subscriptionId)
	if err != nil {
		return nil, fmt.Errorf("getting aks client: %w", err)
	}

	lgr.Info("getting managed cluster")
	res, err := client.NewManagedClustersClient().Get(ctx, resourceGroup, clusterName, nil)
	if err != nil {
		return nil, fmt.Errorf("getting managed cluster: %w", err)
	}

	return &res.ManagedCluster, nil
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
