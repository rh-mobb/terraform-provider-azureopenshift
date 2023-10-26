package clients

import (
	"context"

	armpolicy "github.com/Azure/azure-sdk-for-go/sdk/azcore/arm/policy"
	redhatopenshift "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redhatopenshift/armredhatopenshift"

	"github.com/rh-mobb/terraform-provider-azureopenshift/azureopenshift/auth"
)

type Client struct {
	OpenShiftClustersClient *redhatopenshift.OpenShiftClustersClient
	SubscriptionID          string
	StopCtx                 context.Context
}

func NewClient(stopCtx context.Context, config auth.Config) (*Client, error) {
	cred, err := auth.NewDefaultAroCredential(config)
	if err != nil {
		return nil, err
	}

	options := &armpolicy.ClientOptions{}
	options.ClientOptions = auth.GetOptions(config)

	openshiftClustersClient, err := redhatopenshift.NewOpenShiftClustersClient(config.SubscriptionId, cred, options)
	if err != nil {
		return nil, err
	}

	return &Client{
		OpenShiftClustersClient: openshiftClustersClient,
		StopCtx:                 stopCtx,
		SubscriptionID:          config.SubscriptionId,
	}, nil
}
