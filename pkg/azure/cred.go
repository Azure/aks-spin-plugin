package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/azure/spin-aks-plugin/pkg/usererror"
	"github.com/golang-jwt/jwt/v4"
)

var (
	cred *azidentity.DefaultAzureCredential
)

func getCred() (*azidentity.DefaultAzureCredential, error) {
	if cred != nil {
		return cred, nil
	}

	var err error
	cred, err = azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, usererror.New(fmt.Errorf("authenticating to Azure: %w", err), "Unable to authenticate to Azure. Try running \"az login\".")
	}

	return cred, nil
}

// adapted from https://stackoverflow.com/a/75658185
func getObjectId(ctx context.Context, cred azcore.TokenCredential) (string, error) {
	token, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		// okay to hardcode to PublicCloud since this is a hackathon project
		// TODO: make this configurable
		Scopes: []string{cloud.AzurePublic.Services[cloud.ResourceManager].Endpoint + "/.default"},
	})
	if err != nil {
		return "", fmt.Errorf("getting token: %w", err)
	}

	type t struct {
		ObjectId string `json:"oid"`
		jwt.RegisteredClaims
	}

	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	claim := &t{}
	if _, _, err := parser.ParseUnverified(token.Token, claim); err != nil {
		return "", fmt.Errorf("parsing token: %w", err)
	}

	objectId := claim.ObjectId
	if objectId == "" {
		return "", fmt.Errorf("object id is empty")
	}

	return objectId, nil
}
