package azure

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/azure/spin-aks-plugin/pkg/logger"
)

type Role struct {
	Name string
	ID   string
}

const (
	acrPullRoleName       = "AcrPull"
	acrPullRoleDefinition = "7f951dda-4ed3-4680-a7ca-43fe172d538d"
	acrResourceIdTemplate = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ContainerRegistry/registries/%s"
)

var (
	AcrPullRole = Role{
		Name: acrPullRoleName,
		ID:   fmt.Sprintf("/providers/Microsoft.Authorization/roleDefinitions/%s", acrPullRoleDefinition),
	}
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

func CheckACRPullAccess(ctx context.Context, subscriptionId, resourceGroup, registryName, clusterName string) error {
	lgr := logger.FromContext(ctx).With("subscription", subscriptionId, "resource group", resourceGroup, "registry", registryName)
	ctx = logger.WithContext(ctx, lgr)
	lgr.Debug("checking cluster's acr pull access")

	roles, err := ListRoleAssignments(ctx, subscriptionId, resourceGroup)
	if err != nil {
		return fmt.Errorf("listing role assignments: %w", err)
	}

	// retrieve specific registry by name
	client, err := acrFactory(subscriptionId)
	acr, err := client.NewRegistriesClient().Get(ctx, resourceGroup, registryName, nil)
	if err != nil {
		return fmt.Errorf("get acr by name: %w", err)
	}

	// retrieve specific cluster by name
	mc, err := GetCluster(ctx, subscriptionId, resourceGroup, clusterName)
	if err != nil {
		return fmt.Errorf("get mc by name: %w", err)
	}

	scope := acr.ID
	kubeletId := mc.Identity.PrincipalID
	for _, role := range roles {
		if (*role.Name == AcrPullRole.Name) && (*role.ID == AcrPullRole.ID) && (*scope == *role.Properties.Scope) && (*kubeletId == *role.Properties.PrincipalID) {
			// tbarnes94: success case
			// checking that cluster has permissions to pull from acr
			// matching up the mc's kubelet id (principalId) to the role's principalId (role.Properties.PrincipalID)
			// matching up the scope from the role (role.Properties.Scope) to the scope in the acr (registry id)

			return nil
		}
	}

	return errors.New("cluster does not have AcrPull permission")
}

func ListRoleAssignments(ctx context.Context, subscriptionId, resourceGroup string) ([]armauthorization.RoleAssignment, error) {
	lgr := logger.FromContext(ctx).With("subscription", subscriptionId, "resource group", resourceGroup)
	ctx = logger.WithContext(ctx, lgr)
	lgr.Debug("listing role assignments")

	cred, err := getCred()
	if err != nil {
		return nil, fmt.Errorf("getting az credentials: %w", err)
	}

	client, err := armauthorization.NewRoleAssignmentsClient(subscriptionId, cred, nil)
	pager := client.NewListForResourceGroupPager(resourceGroup, nil)

	var roles []armauthorization.RoleAssignment
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing role assignments page: %w", err)
		}
		for _, role := range page.Value {
			if role == nil {
				return nil, errors.New("nil role")
			}

			roles = append(roles, *role)
		}
	}

	lgr.Debug("finished listing role assignments")

	return roles, nil
}
