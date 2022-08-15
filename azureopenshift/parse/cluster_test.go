package parse_test

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/redhatopenshift/mgmt/redhatopenshift"
	"github.com/rh-mobb/terraform-provider-azureopenshift/azureopenshift/parse"
)

func TestParseInternalClusterId(t *testing.T) {
	clusterName := "test-tf"
	testString := "test-tf-2gk5b-worker-eastus2"
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
