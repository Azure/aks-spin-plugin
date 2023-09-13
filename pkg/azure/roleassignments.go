package azure

import (
	"context"
	"fmt"
	"github.com/azure/spin-aks-plugin/pkg/logger"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
)

type roleAssignmentClient struct {
	client *armauthorization.RoleAssignmentsClient
}

func createRoleAssignmentClient(subscriptionId string) (*roleAssignmentClient, error) {
	credential, err := getCred()
	if err != nil {
		return nil, fmt.Errorf("getting credential: %w", err)
	}

	client, err := armauthorization.NewRoleAssignmentsClient(subscriptionId, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("creating role assignments client: %w", err)
	}

	return &roleAssignmentClient{client: client}, nil
}

func (r *roleAssignmentClient) createRoleAssignment(ctx context.Context, objectId, roleId, scope, assignmentName string) error {
	lgr := logger.FromContext(ctx).With("objectId", objectId, "assignmentName", assignmentName, "scope", scope)

	ctx = logger.WithContext(ctx, lgr)
	lgr.Debug("linking ACR")

	fullAssignmentId := fmt.Sprintf("/%s/providers/Microsoft.Authorization/roleAssignments/%s", scope, assignmentName)
	fulLDefinitionId := fmt.Sprintf("/providers/Microsoft.Authorization/roleDefinitions/%s", roleId)

	params := armauthorization.RoleAssignmentCreateParameters{
		Properties: &armauthorization.RoleAssignmentProperties{
			PrincipalID:      &objectId,
			RoleDefinitionID: &fulLDefinitionId,
		},
	}

	resp, err := r.client.CreateByID(ctx, fullAssignmentId, params, nil)
	lgr.Debug("response from create role assignment", "resp", resp)

	if err != nil {
		return fmt.Errorf("creating role assignment: %w", err)
	}

	return nil
}
