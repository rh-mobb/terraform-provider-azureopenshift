package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	redhatopenshift "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redhatopenshift/armredhatopenshift"
	"github.com/rh-mobb/terraform-provider-azureopenshift/helpers/azure"
)

type ClusterId struct {
	SubscriptionId     string
	ResourceGroup      string
	ManagedClusterName string
}

func NewClusterID(subscriptionId, resourceGroup, managedClusterName string) ClusterId {
	return ClusterId{
		SubscriptionId:     subscriptionId,
		ResourceGroup:      resourceGroup,
		ManagedClusterName: managedClusterName,
	}
}

func (id ClusterId) String() string {
	segments := []string{
		fmt.Sprintf("Managed Cluster Name %q", id.ManagedClusterName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Cluster", segmentsStr)
}

// ClusterID parses a Cluster ID into an ClusterId struct
func ClusterID(input string) (*ClusterId, error) {
	id, err := azure.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := ClusterId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.ManagedClusterName, err = id.PopSegment("openShiftClusters"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}

//heack: we want to get load balancer name. But API does not expose that yet.

func InternalClusterId(clusterName string, workerProfiles []*redhatopenshift.WorkerProfile) (*string, error) {
	if len(workerProfiles) < 1 {
		return nil, errors.New("need at least 1 worker profile to calculate internal cluster id")
	}
	profile := (workerProfiles)[0]
	es := `(.+)-(.+?)-worker-.+`
	rgx, err := regexp.Compile(es)
	if err != nil {
		return nil, err
	}
	matches := rgx.FindStringSubmatch(*profile.Name)
	if len(matches) != 3 {
		return nil, fmt.Errorf("can not capture the internal cluster id with cluster name %s, profile worker name %s, matches: %v", clusterName, *profile.Name, matches)
	}
	return &matches[2], nil
}
