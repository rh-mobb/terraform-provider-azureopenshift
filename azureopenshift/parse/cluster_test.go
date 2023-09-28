package parse_test

import (
	"testing"

	redhatopenshift "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redhatopenshift/armredhatopenshift"
	"github.com/rh-mobb/terraform-provider-azureopenshift/azureopenshift/parse"
)

func TestParseInternalClusterId(t *testing.T) {
	captureClusterId("test-tf", "test-tf-2gk5b-worker-eastus21", t)
	captureClusterId("test-tf-001", "test-tf-001-2gk5b-worker-eastus22", t)
	captureClusterId("aro-001-xxxx-cp4i-dev-use2", "aro-001-xxxx-cp4i-dev-2gk5b-worker-eastus21", t)
}

func captureClusterId(clusterName, testString string, t *testing.T) {
	workerProfiles := &[]redhatopenshift.WorkerProfile{
		{
			Name: &testString,
		},
	}
	clusterId, err := parse.InternalClusterId(clusterName, workerProfiles)
	if err != nil {
		t.Errorf("can not get id with cluster name %s and work profile name %s: error %v", clusterName, testString, err)
	} else if *clusterId != "2gk5b" {
		t.Errorf("the cluster id shoud equal to %s, but it is %s", "2gk5b", *clusterId)
	}
}
