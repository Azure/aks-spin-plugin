package azure

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azcertificates"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/azure/spin-aks-plugin/pkg/logger"
)

type Akv struct {
	Uri            string
	Id             string
	TenantId       string
	SubscriptionId string
	ResourceGroup  string
	Name           string
}

// CertOpt specifies what kind of certificate to create
type CertOpt func(cert *azcertificates.CreateCertificateParameters) error

type Cert struct {
	name string
}

func LoadAkv(id arm.ResourceID) *Akv {
	return &Akv{
		Id:             id.String(),
		Name:           id.Name,
		ResourceGroup:  id.ResourceGroupName,
		SubscriptionId: id.SubscriptionID,
	}
}

func GetKeyVault(ctx context.Context, subscriptionId, resourceGroup, name string) (*Akv, error) {
	lgr := logger.FromContext(ctx).With("name", name, "resourceGroup", resourceGroup, "subscriptionId", subscriptionId)
	ctx = logger.WithContext(ctx, lgr)
	lgr.Info("starting to get keyvault")
	defer lgr.Info("finished getting keyvault")

	cred, err := getCred()
	if err != nil {
		return nil, fmt.Errorf("getting az credentials: %w", err)
	}

	vaultsClient, err := armkeyvault.NewVaultsClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}
	
	resp, err := vaultsClient.Get(ctx, resourceGroup, name, nil)
	if err != nil {
		return nil, fmt.Errorf("getting keyvault: %w", err)
	}
	kv := resp.Vault

	id := resp.Vault.ID
	kvId, err := arm.ParseResourceID(*id)
	if err != nil {
		return nil, fmt.Errorf("parsing resource id: %w", err)
	}
	newAkv := LoadAkv(*kvId)
	newAkv.Uri = *kv.Properties.VaultURI

	return newAkv, nil
}

func ListKeyVaults(ctx context.Context, subscriptionId, resourceGroup string) ([]Akv, error) {
	lgr := logger.FromContext(ctx).With("resourceGroup", resourceGroup, "subscriptionId", subscriptionId)
	ctx = logger.WithContext(ctx, lgr)
	lgr.Info("starting to list keyvaults")
	defer lgr.Info("finished listing keyvaults")

	cred, err := getCred()
	if err != nil {
		return nil, fmt.Errorf("getting az credentials: %w", err)
	}

	client, err := armkeyvault.NewVaultsClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}

	pager := client.NewListByResourceGroupPager(resourceGroup, nil)
	var akvs []Akv
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting next page: %w", err)
		}

		for _, kv := range page.Value {

			if kv == nil {
				return nil, fmt.Errorf("nil keyvault")
			}
			id, err := arm.ParseResourceID(*kv.ID)
			if err != nil {
				return nil, fmt.Errorf("parsing resource id: %w", err)
			}
			newAkv := LoadAkv(*id)
			newAkv.Uri = *kv.Properties.VaultURI

			akvs = append(akvs, *newAkv)
			lgr.Info("keyvault", "name", kv.Name)
		}
	}
	return akvs, nil
}


func NewAkv(ctx context.Context, tenantId, subscriptionId, resourceGroup, name, location string) (*Akv, error) {
	name = truncate(name, 24)

	lgr := logger.FromContext(ctx).With("name", name, "resourceGroup", resourceGroup, "location", location, "subscriptionId", subscriptionId)
	ctx = logger.WithContext(ctx, lgr)
	lgr.Info("starting to create akv")
	defer lgr.Info("finished creating akv")

	cred, err := getCred()
	if err != nil {
		return nil, fmt.Errorf("getting az credentials: %w", err)
	}

	factory, err := armkeyvault.NewClientFactory(subscriptionId, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating client factory: %w", err)
	}

	clientObjectId, err := getObjectId(ctx, cred)
	if err != nil {
		return nil, fmt.Errorf("getting client object id: %w", err)
	}

	v := &armkeyvault.VaultCreateOrUpdateParameters{
		Location: to.Ptr(location),
		Properties: &armkeyvault.VaultProperties{
			AccessPolicies: []*armkeyvault.AccessPolicyEntry{
				{
					ObjectID: to.Ptr(clientObjectId),
					Permissions: &armkeyvault.Permissions{
						Certificates: []*armkeyvault.CertificatePermissions{
							to.Ptr(armkeyvault.CertificatePermissionsCreate),
						},
					},
					TenantID:      to.Ptr(tenantId),
					ApplicationID: nil,
				},
			},
			TenantID: to.Ptr(tenantId),
			SKU: &armkeyvault.SKU{
				Name: to.Ptr(armkeyvault.SKUNameStandard),
			},
		},
	}
	poller, err := factory.NewVaultsClient().BeginCreateOrUpdate(ctx, resourceGroup, name, *v, nil)
	if err != nil {
		return nil, fmt.Errorf("starting to create vault: %w", err)
	}

	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("waiting for vault creation to complete: %w", err)
	}

	return &Akv{
		Uri:            *result.Properties.VaultURI,
		Id:             *result.ID,
		ResourceGroup:  resourceGroup,
		Name:           *result.Name,
		SubscriptionId: subscriptionId,
		TenantId:       tenantId,
	}, nil
}

func (a *Akv) PutSecret(ctx context.Context, name, value string) error {
	lgr := logger.FromContext(ctx).With("name", name, "resourceGroup", a.ResourceGroup, "subscriptionId", a.SubscriptionId)
	ctx = logger.WithContext(ctx, lgr)
	lgr.Info("starting to put secret")
	defer lgr.Info("finished putting secret")

	cred, err := getCred()
	if err != nil {
		return fmt.Errorf("getting az credentials: %w", err)
	}

	secretClient, err := azsecrets.NewClient(a.Uri, cred, nil)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	//TODO add some validation for access check so we can validate access when selecting the keyvault
	resp,err := secretClient.GetSecret(ctx,name,"",&azsecrets.GetSecretOptions{})
	if err != nil {
		var respErr *azcore.ResponseError
		ok := errors.As(err, &respErr)
		if !ok {
			return fmt.Errorf("extracting ResponseError while getting secret '%s': %w",name, err)
		}
		if respErr.StatusCode == http.StatusNotFound{
			lgr.Info(fmt.Sprintf("existing secret not found for key '%s'", name))
		}
		if respErr.StatusCode != http.StatusNotFound{
			return fmt.Errorf("getting secret '%s': %w",name, err)
		}
	}
	if err == nil {
		lgr.Info(fmt.Sprintf("existing secret found for key '%s', checking if secret value changed", name))
		if value == *resp.Value{
			lgr.Info(fmt.Sprintf("existing secret value matches for key '%s'", name))
			return nil
		}
	}

	_,err = secretClient.SetSecret(ctx,name,azsecrets.SetSecretParameters{Value:to.Ptr(value)},nil)
	if err != nil {
		return fmt.Errorf("getting key: %w", err)
	}

	return nil
}


func (a *Akv) GetId() string {
	return a.Id
}

func (a *Akv) AddAccessPolicy(ctx context.Context, objectId string, permissions armkeyvault.Permissions) error {
	lgr := logger.FromContext(ctx).With("objectId", objectId, "name", a.Name, "resourceGroup", a.ResourceGroup, "subscriptionId", a.SubscriptionId)
	ctx = logger.WithContext(ctx, lgr)
	lgr.Info("starting to add access policy")
	defer lgr.Info("finished adding access policy")

	cred, err := getCred()
	if err != nil {
		return fmt.Errorf("getting az credentials: %w", err)
	}

	client, err := armkeyvault.NewVaultsClient(a.SubscriptionId, cred, nil)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	addition := armkeyvault.VaultAccessPolicyParameters{
		Properties: &armkeyvault.VaultAccessPolicyProperties{
			AccessPolicies: []*armkeyvault.AccessPolicyEntry{
				{
					TenantID:    to.Ptr(a.TenantId),
					ObjectID:    to.Ptr(objectId),
					Permissions: &permissions,
				},
			},
		},
	}
	if _, err := client.UpdateAccessPolicy(ctx, a.ResourceGroup, a.Name, armkeyvault.AccessPolicyUpdateKindAdd, addition, nil); err != nil {
		return fmt.Errorf("adding access policy: %w", err)
	}

	return nil
}

func LoadCert(name string) *Cert {
	return &Cert{
		name: name,
	}
}

func (c *Cert) GetName() string {
	return c.name
}
