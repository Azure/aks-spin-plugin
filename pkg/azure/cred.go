package azure

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/azure/spin-aks-plugin/pkg/usererror"
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
