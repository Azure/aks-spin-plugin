package azure

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/azure/spin-aks-plugin/pkg/logger"
)

var (
	// we maintain a cache of locations because locations are something that rarely change
	// and basically will never change within the time a user is running the command.
	// A user should never run this command and expect to see locations that were just added
	// during the period in which their cli command is running.
	cachedLocations []armsubscriptions.Location = nil
)

func ListLocations(ctx context.Context, subscriptionId string) ([]armsubscriptions.Location, error) {
	lgr := logger.FromContext(ctx).With("subscriptionId", subscriptionId)
	ctx = logger.WithContext(ctx, lgr)
	lgr.Debug("listing locations")

	if cachedLocations != nil {
		lgr.Debug("returning locations from cache")
		return cachedLocations, nil
	}

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
	cachedLocations = locations
	return locations, nil
}
