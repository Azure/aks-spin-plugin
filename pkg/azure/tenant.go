package azure

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
	"github.com/azure/spin-aks-plugin/pkg/logger"
)

func ListTenants(ctx context.Context) ([]armsubscription.TenantIDDescription, error) {
	lgr := logger.FromContext(ctx)
	lgr.Debug("listing Azure subscriptions")

	cred, err := getCred()
	if err != nil {
		return nil, fmt.Errorf("getting credentials: %w", err)
	}

	client, err := armsubscription.NewTenantsClient(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating tenants client: %w", err)
	}

	var tenants []armsubscription.TenantIDDescription
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing tenants page: %w", err)
		}

		for _, t := range page.Value {
			if t == nil {
				return nil, errors.New("nil tenant") // this should never happen but it's good to check just in case
			}

			tenants = append(tenants, *t)
		}
	}

	lgr.Debug("finished listing Azure tenants")
	return tenants, nil
}
