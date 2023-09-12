package azure

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/azure/spin-aks-plugin/pkg/logger"
)

type Role struct {
	Name string
	ID   string
}

const (
<<<<<<< HEAD
=======
	// tbarnes94: placeholder for now
	subscriptionId = "8ecadfc9-d1a3-4ea4-b844-0d9f87e4d7c8"

>>>>>>> bbffb72b84a8e2f03fbe48435d2bed7023e97454
	acrPullRoleName       = "AcrPull"
	acrPullRoleDefinition = "7f951dda-4ed3-4680-a7ca-43fe172d538d"
)

var (
	AcrPullRole = Role{
		Name: acrPullRoleName,
<<<<<<< HEAD
		ID:   fmt.Sprintf("/providers/Microsoft.Authorization/roleDefinitions/%s", acrPullRoleDefinition),
=======
		ID:   fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Authorization/roleDefinitions/%s", subscriptionId, acrPullRoleDefinition),
>>>>>>> bbffb72b84a8e2f03fbe48435d2bed7023e97454
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

func CheckACRPullAccess(ctx context.Context, subscriptionId, resourceGroup string) error {
	roles, err := ListRoleAssignments(ctx, subscriptionId, resourceGroup)
	if err != nil {
		return fmt.Errorf("listing role assignments: %w", err)
	}
	for _, role := range roles {
		if (*role.Name == AcrPullRole.Name) && (*role.ID == AcrPullRole.ID) {
			// tbarnes94: success case
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
