package clients

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/redhatopenshift/mgmt/2022-04-01/redhatopenshift"
	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/go-azure-helpers/sender"
)

type Client struct {
	OpenShiftClustersClient *redhatopenshift.OpenShiftClustersClient
	StopCtx                 context.Context
}

func NewClient(stopCtx context.Context, auth autorest.Authorizer, resourceManagerEndpoint, subscriptionId string) *Client {
	openshiftClustersClient := redhatopenshift.NewOpenShiftClustersClientWithBaseURI(resourceManagerEndpoint, subscriptionId)

	openshiftClustersClient.Authorizer = auth
	openshiftClustersClient.Sender = sender.BuildSender("AzureRM")
	openshiftClustersClient.UserAgent = "Terraform Azure Openshift - By Mobb"

	return &Client{
		OpenShiftClustersClient: &openshiftClustersClient,
		StopCtx:                 stopCtx,
	}
}
