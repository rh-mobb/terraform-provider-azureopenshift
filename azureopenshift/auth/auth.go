package auth

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

const (
	credNameAzureCLI = "AroCLICredential"

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
	chain *azidentity.ChainedTokenCredential
}

func NewDefaultAroCredential(config Config) (*DefaultAroCredential, error) {
	var errorMessages []string

	options := &azidentity.ClientSecretCredentialOptions{
		ClientOptions: GetOptions(config),
	}

	clientSecretCred, err := azidentity.NewClientSecretCredential(config.TenantId, config.ClientId, config.ClientSecret, options)
	if err != nil {
		errorMessages = append(errorMessages, "AroClientSecretCredential: "+err.Error())
	}

	cliCred, err := azidentity.NewAzureCLICredential(nil)
	if err != nil {
		errorMessages = append(errorMessages, "AroCLICredential: "+err.Error())
	}

	creds := []azcore.TokenCredential{clientSecretCred, cliCred}

	err = defaultAroCredentialConstructorErrorHandler(len(creds), errorMessages)
	if err != nil {
		return nil, err
	}

	chain, err := azidentity.NewChainedTokenCredential(creds, nil)
	if err != nil {
		return nil, err
	}
	return &DefaultAroCredential{chain: chain}, nil
}

// GetToken requests an access token from Azure Active Directory. This method is called automatically by Azure SDK clients.
func (c *DefaultAroCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return c.chain.GetToken(ctx, opts)
}

func defaultAroCredentialConstructorErrorHandler(numberOfSuccessfulCredentials int, errorMessages []string) (err error) {
	errorMessage := strings.Join(errorMessages, "\n\t")

	if numberOfSuccessfulCredentials == 0 {
		return errors.New(errorMessage)
	}

	if len(errorMessages) != 0 {
		log.Printf("NewDefaultAroCredential failed to initialize some credentials:\n\t%s", errorMessage)
	}

	return nil
}

func GetOptions(config Config) policy.ClientOptions {
	switch config.Environment {
	// TODO: remove China support for now until ARO supports it.
	// case AzureChinaString:
	// 	return cloud.AzureChina
	case AzureUSGovernmentString:
		return policy.ClientOptions{Cloud: cloud.AzureGovernment}
	default:
		return policy.ClientOptions{Cloud: cloud.AzurePublic}
	}
}
