package azure

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/go-autorest/autorest"
	"github.com/azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
)

// OAuthWrapper is an interface to satisfy token provider interface
type OAuthWrapper struct {
	token string
}

func (o *OAuthWrapper) OAuthToken() string {
	decoded, err := base64.StdEncoding.DecodeString(o.token)
	if err != nil {
		fmt.Println("decode error:", err)
		return ""
	}
	return string(decoded)
}

type AadClient struct {
	ApplicationsClient *graphrbac.ApplicationsClient
}

func NewAadClient(ctx context.Context, tenantId string) (*AadClient, error) {
	cred, err := getCred()
	//Get a token from the DefaultAzureCredential
	token, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		// okay to hardcode to PublicCloud since we should never deploy to anything else in public OSS repo
		Scopes: []string{cloud.AzurePublic.Services[cloud.ResourceManager].Endpoint + "/.default"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %v\n", err)
	}

	wrapper := &OAuthWrapper{token: token.Token}
	fmt.Println(wrapper.OAuthToken())
	authorizer := autorest.NewBearerAuthorizer(wrapper)
	//authorizer, err := auth.NewAuthorizerFromCLI()
	if err != nil {
		return nil, fmt.Errorf("getting authorizer: %w", err)
	}
	fmt.Println(tenantId)
	appsClient := graphrbac.NewApplicationsClient(tenantId)
	appsClient.Authorizer = authorizer

	return &AadClient{
		ApplicationsClient: &appsClient,
	}, nil
}

func (ac *AadClient) getObjectIdFromClientId(ctx context.Context, clientId string) (*string, error) {
	fmt.Println(ac.ApplicationsClient.BaseURI)
	res, err := ac.ApplicationsClient.GetServicePrincipalsIDByAppID(ctx, clientId)
	if err != nil {
		return nil, fmt.Errorf("getting service principal from app id: %w", err)
	}

	fmt.Printf("objectId: %s", *res.Value)
	if res.Value == nil {
		return nil, fmt.Errorf("object id is nil")
	}

	return res.Value, nil
}
