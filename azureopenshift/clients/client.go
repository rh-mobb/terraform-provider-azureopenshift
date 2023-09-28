package clients

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	redhatopenshift "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redhatopenshift/armredhatopenshift"
)

type Client struct {
	OpenShiftClustersClient *redhatopenshift.OpenShiftClustersClient
	SubscriptionID          string
	StopCtx                 context.Context
}

func NewClient(stopCtx context.Context, subscriptionId string) (*Client, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	openshiftClustersClient, err := redhatopenshift.NewOpenShiftClustersClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}

	return &Client{
		OpenShiftClustersClient: openshiftClustersClient,
		StopCtx:                 stopCtx,
		SubscriptionID:          subscriptionId,
	}, nil
}

// //Validate Subscription ID Access by getting a token
// func validateSubscriptionIDAccess(ctx context.Context, config auth.Credentials, subscriptionId string) (*string, error) {
// 	var validatedSubscriptionId string
// 	validatedSubscriptionId = subscriptionId
// 	authorizer, err := auth.NewAuthorizerFromCredentials(ctx, config, config.Environment.MicrosoftGraph)
// 	if err != nil {
// 		return nil, fmt.Errorf("unable to build authorizer for Microsoft Graph API: %+v", err)
// 	}

// 	// Acquire an access token so we can inspect the claims
// 	if _, err = authorizer.Token(ctx, &http.Request{}); err != nil {
// 		return nil, fmt.Errorf("could not validate the azure authorization: %+v", err)
// 	}
// 	if cli, ok := authorizer.(*auth.AzureCliAuthorizer); ok {
// 		// Use the subscription ID from Azure CLI when otherwise unknown
// 		if validatedSubscriptionId == "" {
// 			if cli.DefaultSubscriptionID == "" {
// 				return nil, fmt.Errorf("azure-cli could not determine subscription ID to use and no subscription was specified")
// 			}

// 			validatedSubscriptionId = cli.DefaultSubscriptionID
// 			log.Printf("[DEBUG] Using default subscription ID from Azure CLI: %q", subscriptionId)
// 		}
// 	}
// 	return &validatedSubscriptionId, nil
// }
