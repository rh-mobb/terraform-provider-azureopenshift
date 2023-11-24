package parse_test

import (
	"testing"

	redhatopenshift "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redhatopenshift/armredhatopenshift"
	"github.com/rh-mobb/terraform-provider-azureopenshift/azureopenshift/parse"
)

func TestParseInternalClusterId(t *testing.T) {
	captureClusterId(
		"test-tf",
		[]string{"test-tf-2gk5b-worker-eastus21"},
		t,
	)
	captureClusterId(
		"test-tf-001",
		[]string{"test-tf-001-2gk5b-worker-eastus22"},
		t,
	)
	captureClusterId(
		"aro-001-xxxx-cp4i-dev-use2",
		[]string{"aro-001-xxxx-cp4i-dev-2gk5b-worker-eastus21"},
		t,
	)
	captureClusterId(
		"test-tf",
		[]string{
			"test-tf-2gk5b-infra-eastus21",
			"test-tf-2gk5b-worker-eastus21",
		},
		t,
	)
	captureClusterId(
		"test-tf-001",
		[]string{
			"test-tf-001-2gk5b-infra-eastus22",
			"test-tf-001-2gk5b-worker-eastus22",
		},
		t,
	)
	captureClusterId(
		"aro-001-xxxx-cp4i-dev-use2",
		[]string{
			"aro-001-xxxx-cp4i-dev-2gk5b-infra-eastus21",
			"aro-001-xxxx-cp4i-dev-2gk5b-worker-eastus21",
		},
		t,
	)
}

func captureClusterId(clusterName string, testString []string, t *testing.T) {
	workerProfiles := make([]*redhatopenshift.WorkerProfile, 0)
	for _, profile := range testString {
		workerProfiles = append(workerProfiles, &redhatopenshift.WorkerProfile{
			Name: &profile,
		})
	}

	clusterId, err := parse.InternalClusterId(clusterName, workerProfiles)
	if err != nil {
		t.Errorf("can not get id with cluster name %s and work profile name %s: error %v", clusterName, testString, err)
	} else if *clusterId != "2gk5b" {
		t.Errorf("the cluster id shoud equal to %s, but it is %s", "2gk5b", *clusterId)
	}
}
