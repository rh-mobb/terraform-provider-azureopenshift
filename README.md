# Terraform Provider Azure Redhat Openshift

This is a provider to create [Azure Redhat Openshift](https://docs.microsoft.com/en-us/azure/openshift/)


## Prerequistes


### [Create Azure network with two empty subnets](https://docs.microsoft.com/en-us/azure/openshift/tutorial-create-cluster#create-a-virtual-network-containing-two-empty-subnets)
* Azure Resource Group
* Azure network
* Master subnet & Worker subnet

### [Create a service principal](https://docs.microsoft.com/en-us/azure/openshift/howto-create-service-principal?pivots=aro-azurecli)

### Give Red Hat Openshift Resource Provider network contributor role of Azure network

```bash
OPENSHIFT_RP_OBJECT_ID=$(az ad sp list --filter "displayname eq 'Azure Red Hat OpenShift RP'" --query "[?appDisplayName=='Azure Red Hat OpenShift RP'].objectId" --only-show-errors --output tsv)
az role assignment create --role "Contributor" --assignee-object-id ${OPENSHIFT_RP_OBJECT_ID} --scope [NETWORK_ID]
```

## Test sample configuration

```bash
cd examples
terraform init
terraform apply
```