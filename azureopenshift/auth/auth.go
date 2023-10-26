package auth

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

const (
	AzurePublicString       = "public"
	AzureUSGovernmentString = "usgovernment"

	// TODO: remove China support for now until ARO supports it.
	// AzureChinaString        = "china"
)

type Config struct {
	SubscriptionId string
	TenantId       string
	ClientId       string
	ClientSecret   string
	Environment    string
}

type DefaultAroCredential struct {
	chain   *azidentity.ChainedTokenCredential
	options *azidentity.ClientSecretCredentialOptions
}

func NewDefaultAroCredential(config Config) (*DefaultAroCredential, error) {
	// create the credential with the options pointed to the appropriate selected cloud
	cred := &DefaultAroCredential{
		options: &azidentity.ClientSecretCredentialOptions{
			ClientOptions: policy.ClientOptions{
				Cloud: getCloud(config),
			},
		},
	}

	clientSecretCred, err := azidentity.NewClientSecretCredential(config.TenantId, config.ClientId, config.ClientSecret, cred.options)
	if err != nil {
		return cred, fmt.Errorf("AroClientSecretCredential: %w", err)
	}

	cliCred, err := azidentity.NewAzureCLICredential(nil)
	if err != nil {
		return cred, fmt.Errorf("AroCLICredential: %w", err)
	}

	chain, err := azidentity.NewChainedTokenCredential([]azcore.TokenCredential{clientSecretCred, cliCred}, nil)
	if err != nil {
		return nil, err
	}

	cred.chain = chain

	return cred, nil
}

// GetToken requests an access token from Azure Active Directory. This method is called automatically by Azure SDK clients.
func (c *DefaultAroCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return c.chain.GetToken(ctx, opts)
}

// GetClientOptions returns the options as set on the credential.  It is used to pass in consistent options to other providers
// e.g. ARO when creating the individual service requests.
func (c *DefaultAroCredential) GetClientOptions() *policy.ClientOptions {
	if c.options == nil {
		return nil
	}

	return &c.options.ClientOptions
}

func getCloud(config Config) cloud.Configuration {
	switch config.Environment {
	// TODO: remove China support for now until ARO supports it.
	// case AzureChinaString:
	// 	return cloud.AzureChina
	case AzureUSGovernmentString:
		return cloud.AzureGovernment
	default:
		return cloud.AzurePublic
	}
}
