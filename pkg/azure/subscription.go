package azure

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

func ListSubscriptions(ctx context.Context) ([]armsubscription.Subscription, error) {
	cred, err := getCred()
	if err != nil {
		return nil, fmt.Errorf("getting credentials: %w", err)
	}

	client, err := armsubscription.NewSubscriptionsClient(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating subscriptions client: %w", err)
	}

	var subs []armsubscription.Subscription
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing subscription page: %w", err)
		}

		for _, sub := range page.Value {
			if sub == nil {
				return nil, errors.New("nil sub") // this should never happen but it's good to check just in case
			}

			subs = append(subs, *sub)
		}
	}

	return subs, nil
}
