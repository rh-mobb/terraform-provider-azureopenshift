package aro_test

import (
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/redhatopenshift/mgmt/redhatopenshift"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rh-mobb/terraform-provider-azureopenshift/helpers/aro"
)

var _ = Describe("Cluster Test", func() {

	var input []interface{}
	var cph *aro.ClusterProfileHelper
	var subscription_id string
	var expectedPullSecret, domain string
	var cluster_resource_group string

	BeforeEach(func() {
		expectedPullSecret = "this is my pull secret"
		domain = "example.com"
		cluster_resource_group = ""
		input = []interface{}{
			map[string]interface{}{
				"pull_secret":            expectedPullSecret,
				"domain":                 domain,
				"fips_validated_modules": "Enabled",
			},
		}
		subscription_id = "123456"
		cluster_resource_group = ""
		cph = aro.NewClusterProfileHelper(subscription_id, cluster_resource_group)
	})

	Context("When both pull secret and domain is provided", func() {
		It("Should return cluster profile with provided domain and pull secret", func() {
			cp := cph.Expand(input)
			Ω(*cp.Domain).Should(Equal(domain))
			Ω(*cp.PullSecret).Should(Equal(expectedPullSecret))
		})
	})

	Context("When domain is not provided", func() {
		BeforeEach(func() {
			domain = ""
		})
		It("Should return cluster profile with a random domain", func() {
			cp := cph.Expand(input)
			Ω(*cp.Domain).Should(Not(Equal("")))
		})
	})

	Context("When cluster resource group is not provided", func() {
		It("Should return resource group id with aro prefix random name", func() {
			cp := cph.Expand(input)
			Ω(strings.Contains(*cp.ResourceGroupID, "aro-")).Should(Equal(true))
		})
	})

	Context("When cluster resource group is provided", func() {
		BeforeEach(func() {
			cluster_resource_group = "custom_resource_group"
			cph = aro.NewClusterProfileHelper(subscription_id, cluster_resource_group)
		})
		It("Should return resource group id with resource group name", func() {
			cp := cph.Expand(input)
			Ω(*cp.ResourceGroupID).Should(Equal("/subscriptions/123456/resourceGroups/custom_resource_group"))
		})
	})

	Context("When fips validated module is enabled", func() {
		It("Should return enabled", func() {
			cp := cph.Expand(input)
			Ω(cp.FipsValidatedModules).Should(Equal(redhatopenshift.FipsValidatedModulesEnabled))
		})
	})
})
