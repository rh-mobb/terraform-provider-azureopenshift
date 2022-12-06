package aro

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/redhatopenshift/mgmt/redhatopenshift"
	"github.com/rh-mobb/terraform-provider-azureopenshift/helpers/utils"
)

type ClusterProfileHelper struct {
	subscriptionId       string
	clusterResourceGroup string
}

func NewClusterProfileHelper(subscriptionId, clusterResourceGroup string) *ClusterProfileHelper {
	return &ClusterProfileHelper{
		subscriptionId:       subscriptionId,
		clusterResourceGroup: clusterResourceGroup,
	}
}

func (cpe *ClusterProfileHelper) Expand(input []interface{}) *redhatopenshift.ClusterProfile {
	randomDomainName := generateRandomDomainName()
	resourceGroupName := fmt.Sprintf("aro-%s", randomDomainName)
	var resourceGroupId string = ""
	if cpe.clusterResourceGroup == "" {
		resourceGroupId = resourceGroupID(cpe.subscriptionId, resourceGroupName)
	} else {
		resourceGroupId = resourceGroupID(cpe.subscriptionId, cpe.clusterResourceGroup)
	}

	if len(input) == 0 {
		return &redhatopenshift.ClusterProfile{
			ResourceGroupID:      utils.String(resourceGroupId),
			Domain:               utils.String(randomDomainName),
			FipsValidatedModules: redhatopenshift.FipsValidatedModulesDisabled,
		}
	}

	config := input[0].(map[string]interface{})

	pullSecret := config["pull_secret"].(string)

	fipsValidatedModules := config["fips_validated_modules"].(string)

	domain := config["domain"].(string)
	if domain == "" {
		domain = randomDomainName
	}

	return &redhatopenshift.ClusterProfile{
		ResourceGroupID:      utils.String(resourceGroupId),
		Domain:               utils.String(domain),
		PullSecret:           utils.String(pullSecret),
		FipsValidatedModules: redhatopenshift.FipsValidatedModules(fipsValidatedModules),
	}
}

func generateRandomDomainName() string {
	randomPrefix := randomString("abcdefghijklmnopqrstuvwxyz", 1)
	randomName := randomString("abcdefghijklmnopqrstuvwxyz1234567890", 7)

	return fmt.Sprintf("%s%s", randomPrefix, randomName)
}

func resourceGroupID(subscriptionId string, resourceGroupName string) string {
	fmtString := "/subscriptions/%s/resourceGroups/%s"
	return fmt.Sprintf(fmtString, subscriptionId, resourceGroupName)
}

func randomString(acceptedChars string, size int) string {
	charSet := []rune(acceptedChars)
	randomChars := make([]rune, size)

	rand.Seed(time.Now().UnixNano())

	for i := range randomChars {
		randomChars[i] = charSet[rand.Intn(len(charSet))]
	}

	return string(randomChars)
}
