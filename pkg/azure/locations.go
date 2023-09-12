package azure

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/azure/spin-aks-plugin/pkg/logger"
)

func ListLocations(ctx context.Context, subscriptionId string) ([]armsubscriptions.Location, error) {
	lgr := logger.FromContext(ctx).With("subscriptionId", subscriptionId)
	ctx = logger.WithContext(ctx, lgr)
	lgr.Debug("listing locations")

	cred, err := getCred()
	if err != nil {
		return nil, fmt.Errorf("getting credentials: %w", err)
	}

	client, err := armsubscriptions.NewClient(cred, nil)

	var locations []armsubscriptions.Location
	pager := client.NewListLocationsPager(subscriptionId, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing locations page: %w", err)
		}

		for _, location := range page.Value {
			if location == nil {
				return nil, errors.New("nil location")
			}

			locations = append(locations, *location)
		}
	}

	lgr.Debug("finished listing locations")
	return locations, nil
}
