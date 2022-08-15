package azureopenshift

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/redhatopenshift/mgmt/2022-04-01/redhatopenshift"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/rh-mobb/terraform-provider-azureopenshift/azureopenshift/clients"
	"github.com/rh-mobb/terraform-provider-azureopenshift/azureopenshift/parse"
	openShiftValidate "github.com/rh-mobb/terraform-provider-azureopenshift/azureopenshift/validate"
	"github.com/rh-mobb/terraform-provider-azureopenshift/helpers/azure"
	"github.com/rh-mobb/terraform-provider-azureopenshift/helpers/suppress"
	"github.com/rh-mobb/terraform-provider-azureopenshift/helpers/tf"
	"github.com/rh-mobb/terraform-provider-azureopenshift/helpers/utils"
	"github.com/rh-mobb/terraform-provider-azureopenshift/helpers/validate"
)

var (
	randomDomainName = GenerateRandomDomainName()
)

const (
	// APIPrivate ...
	APIPrivate string = "Private"
	// Public ...
	APIPublic     string = "Public"
	StandardD8sV3 string = "Standard_D8s_v3"
	StandardD4sV3 string = "Standard_D4s_v3"
)

func resourceOpenShiftCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpenShiftClusterCreate,
		Read:   resourceOpenShiftClusterRead,
		Update: resourceOpenShiftClusterUpdate,
		Delete: resourceOpenShiftClusterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(90 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(90 * time.Minute),
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"location": commonschema.Location(),

			"resource_group_name": commonschema.ResourceGroupName(),

			"cluster_profile": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pull_secret": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"domain": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"resource_group_id": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
					},
				},
			},

			"service_principal": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: openShiftValidate.ClientID,
						},
						"client_secret": {
							Type:         schema.TypeString,
							Required:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"network_profile": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pod_cidr": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      "10.128.0.0/14",
							ValidateFunc: validate.CIDR,
						},
						"service_cidr": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      "172.30.0.0/16",
							ValidateFunc: validate.CIDR,
						},
					},
				},
			},

			"master_profile": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: azure.ValidateResourceID,
						},
						"vm_size": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							Default:          StandardD8sV3,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc:     validation.StringIsNotEmpty,
						},
						"encryption_at_host": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      redhatopenshift.EncryptionAtHostDisabled,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"kubeadmin_username": {
				Type:      schema.TypeString,
				Computed:  true,
				Optional:  true,
				Sensitive: true,
			},

			"kubeadmin_password": {
				Type:      schema.TypeString,
				Computed:  true,
				Optional:  true,
				Sensitive: true,
			},

			"worker_profile": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vm_size": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							Default:          StandardD4sV3,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc:     validation.StringIsNotEmpty,
						},
						"disk_size_gb": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							Default:      128,
							ValidateFunc: openShiftValidate.DiskSizeGB,
						},
						"node_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							Default:      3,
							ValidateFunc: validation.IntBetween(3, 20),
						},
						"subnet_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: azure.ValidateResourceID,
						},
						"encryption_at_host": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      redhatopenshift.EncryptionAtHostDisabled,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},

			"api_server_profile": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"visibility": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  APIPublic,
						},
					},
				},
			},

			"ingress_profile": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"visibility": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  redhatopenshift.VisibilityPublic,
						},
					},
				},
			},

			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"console_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": &schema.Schema{
				Type:         schema.TypeMap,
				Optional:     true,
				ValidateFunc: azure.ValidateTags,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceOpenShiftClusterCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).OpenShiftClustersClient
	ctx, cancel := context.WithTimeout(meta.(*clients.Client).StopCtx, d.Timeout(schema.TimeoutCreate))
	defer cancel()

	log.Printf("[INFO] preparing arguments for Red Hat Openshift Cluster create.")

	resourceGroupName := d.Get("resource_group_name").(string)
	subscriptionId := client.SubscriptionID

	name := d.Get("name").(string)

	existing, err := client.Get(ctx, resourceGroupName, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for presence of existing Red Hat Openshift Cluster %q (Resource Group %q): %s", name, resourceGroupName, err)
		}
	}

	if existing.ID != nil && *existing.ID != "" {
		return tf.ImportAsExistsError("azurerm_redhatopenshift_cluster", *existing.ID)
	}

	location := d.Get("location").(string)

	clusterProfileRaw := d.Get("cluster_profile").([]interface{})
	clusterProfile := expandOpenshiftClusterProfile(clusterProfileRaw, subscriptionId)

	consoleProfile := &redhatopenshift.ConsoleProfile{}

	servicePrincipalProfileRaw := d.Get("service_principal").([]interface{})
	servicePrincipalProfile := expandOpenshiftServicePrincipalProfile(servicePrincipalProfileRaw)

	networkProfileRaw := d.Get("network_profile").([]interface{})
	networkProfile := expandOpenshiftNetworkProfile(networkProfileRaw)

	masterProfileRaw := d.Get("master_profile").([]interface{})
	masterProfile := expandOpenshiftMasterProfile(masterProfileRaw)

	workerProfilesRaw := d.Get("worker_profile").([]interface{})
	workerProfiles := expandOpenshiftWorkerProfiles(workerProfilesRaw)

	apiServerProfileRaw := d.Get("api_server_profile").([]interface{})
	apiServerProfile := expandOpenshiftApiServerProfile(apiServerProfileRaw)

	ingressProfilesRaw := d.Get("ingress_profile").([]interface{})
	ingressProfiles := expandOpenshiftIngressProfiles(ingressProfilesRaw)

	t := d.Get("tags").(map[string]interface{})

	parameters := redhatopenshift.OpenShiftCluster{
		Name:     &name,
		Location: &location,
		OpenShiftClusterProperties: &redhatopenshift.OpenShiftClusterProperties{
			ClusterProfile:          clusterProfile,
			ConsoleProfile:          consoleProfile,
			ServicePrincipalProfile: servicePrincipalProfile,
			NetworkProfile:          networkProfile,
			MasterProfile:           masterProfile,
			WorkerProfiles:          workerProfiles,
			ApiserverProfile:        apiServerProfile,
			IngressProfiles:         ingressProfiles,
		},
		Tags: azure.TagsExpand(t),
	}

	future, err := client.CreateOrUpdate(ctx, resourceGroupName, name, parameters)
	if err != nil {
		return fmt.Errorf("creating Red Hat OpenShift Cluster %q (Resource Group %q): %+v", name, resourceGroupName, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for creation of Red Hat OpenShift Cluster %q (Resource Group %q): %+v", name, resourceGroupName, err)
	}

	read, err := client.Get(ctx, resourceGroupName, name)
	if err != nil {
		return fmt.Errorf("retrieving Red Hat OpenShift Cluster %q (Resource Group %q): %+v", name, resourceGroupName, err)
	}

	if read.ID == nil {
		return fmt.Errorf("cannot read ID for Red Hat OpenShift Cluster %q (Resource Group %q)", name, resourceGroupName)
	}

	d.SetId(*read.ID)

	return resourceOpenShiftClusterRead(d, meta)
}

func resourceOpenShiftClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).OpenShiftClustersClient
	ctx, cancel := context.WithTimeout(meta.(*clients.Client).StopCtx, d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	log.Printf("[INFO] preparing arguments for Red Hat OpenShift Cluster update.")

	resourceGroupName := d.Get("resource_group_name").(string)
	subscriptionId := client.SubscriptionID
	resourceGroupId := ResourceGroupID(subscriptionId, resourceGroupName)

	id, err := parse.ClusterID(d.Id())
	if err != nil {
		return err
	}

	d.Partial(true)

	existing, err := client.Get(ctx, id.ResourceGroup, id.ManagedClusterName)
	if err != nil {
		return fmt.Errorf("retrieving existing Red Hat OpenShift Cluster %q (Resource Group %q): %+v", id.ManagedClusterName, id.ResourceGroup, err)
	}
	if existing.OpenShiftClusterProperties == nil {
		return fmt.Errorf("retrieving existing Red Hat OpenShift Cluster %q (Resource Group %q): `properties` was nil", id.ManagedClusterName, id.ResourceGroup)
	}

	if d.HasChange("cluster_profile") {
		clusterProfileRaw := d.Get("cluster_profile").([]interface{})
		clusterProfile := expandOpenshiftClusterProfile(clusterProfileRaw, resourceGroupId)
		existing.OpenShiftClusterProperties.ClusterProfile = clusterProfile
	}

	if d.HasChange("master_profile") {
		masterProfileRaw := d.Get("master_profile").([]interface{})
		masterProfile := expandOpenshiftMasterProfile(masterProfileRaw)
		existing.OpenShiftClusterProperties.MasterProfile = masterProfile
	}

	if d.HasChange("worker_profile") {
		workerProfilesRaw := d.Get("worker_profile").([]interface{})
		workerProfiles := expandOpenshiftWorkerProfiles(workerProfilesRaw)
		existing.OpenShiftClusterProperties.WorkerProfiles = workerProfiles
	}

	d.Partial(false)

	return resourceOpenShiftClusterRead(d, meta)
}

func resourceOpenShiftClusterRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).OpenShiftClustersClient
	ctx, cancel := context.WithTimeout(meta.(*clients.Client).StopCtx, d.Timeout(schema.TimeoutRead))
	defer cancel()

	id, err := parse.ClusterID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.ManagedClusterName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Red Hat OpenShift Cluster %q was not found in Resource Group %q - removing from state!", id.ManagedClusterName, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Red Hat OpenShift Cluster %q (Resource Group %q): %+v", id.ManagedClusterName, id.ResourceGroup, err)
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", resp.Location)

	if props := resp.OpenShiftClusterProperties; props != nil {
		clusterProfile := flattenOpenShiftClusterProfile(props.ClusterProfile)
		if err := d.Set("cluster_profile", clusterProfile); err != nil {
			return fmt.Errorf("setting `cluster_profile`: %+v", err)
		}

		servicePrincipalProfile := flattenOpenShiftServicePrincipalProfile(props.ServicePrincipalProfile, d)
		if err := d.Set("service_principal", servicePrincipalProfile); err != nil {
			return fmt.Errorf("setting `service_principal`: %+v", err)
		}

		networkProfile := flattenOpenShiftNetworkProfile(props.NetworkProfile)
		if err := d.Set("network_profile", networkProfile); err != nil {
			return fmt.Errorf("setting `network_profile`: %+v", err)
		}

		masterProfile := flattenOpenShiftMasterProfile(props.MasterProfile)
		if err := d.Set("master_profile", masterProfile); err != nil {
			return fmt.Errorf("setting `master_profile`: %+v", err)
		}

		workerProfiles := flattenOpenShiftWorkerProfiles(props.WorkerProfiles)
		if err := d.Set("worker_profile", workerProfiles); err != nil {
			return fmt.Errorf("setting `worker_profile`: %+v", err)
		}

		apiServerProfile := flattenOpenShiftAPIServerProfile(props.ApiserverProfile)
		if err := d.Set("api_server_profile", apiServerProfile); err != nil {
			return fmt.Errorf("setting `api_server_profile`: %+v", err)
		}

		ingressProfiles := flattenOpenShiftIngressProfiles(props.IngressProfiles)
		if err := d.Set("ingress_profile", ingressProfiles); err != nil {
			return fmt.Errorf("setting `ingress_profile`: %+v", err)
		}

		d.Set("version", props.ClusterProfile.Version)
		d.Set("console_url", props.ConsoleProfile.URL)
	}

	credResponse, err := client.ListCredentials(ctx, id.ResourceGroup, id.ManagedClusterName)
	if err != nil {
		if utils.ResponseWasNotFound(credResponse.Response) {
			log.Printf("[DEBUG] Red Hat OpenShift Cluster %q:%q does not have kubeadmin username and password", id.ManagedClusterName, id.ResourceGroup)
			return nil
		}

		return fmt.Errorf("retrieving Red Hat OpenShift Cluster Credential %q (Resource Group %q): %+v", id.ManagedClusterName, id.ResourceGroup, err)
	} else {
		d.Set("kubeadmin_username", credResponse.KubeadminUsername)
		d.Set("kubeadmin_password", credResponse.KubeadminPassword)
	}

	return azure.TagsFlattenAndSet(d, resp.Tags)
}

func resourceOpenShiftClusterDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).OpenShiftClustersClient
	ctx, cancel := context.WithTimeout(meta.(*clients.Client).StopCtx, d.Timeout(schema.TimeoutDelete))
	defer cancel()

	id, err := parse.ClusterID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.ManagedClusterName)
	if err != nil {
		return fmt.Errorf("deleting Red Hat Openshift Cluster %q (Resource Group %q): %+v", id.ManagedClusterName, id.ResourceGroup, err)
	}

	if err := future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for the deletion of Red Hat Openshift Cluster %q (Resource Group %q): %+v", id.ManagedClusterName, id.ResourceGroup, err)
	}

	return nil
}

func flattenOpenShiftClusterProfile(profile *redhatopenshift.ClusterProfile) []interface{} {
	if profile == nil {
		return []interface{}{}
	}

	pullSecret := ""
	if profile.PullSecret != nil {
		pullSecret = *profile.PullSecret
	}

	clusterDomain := ""
	if profile.Domain != nil {
		clusterDomain = *profile.Domain
	}

	resourceGroupId := ""
	if profile.ResourceGroupID != nil {
		resourceGroupId = *profile.ResourceGroupID
	}

	return []interface{}{
		map[string]interface{}{
			"pull_secret":       pullSecret,
			"domain":            clusterDomain,
			"resource_group_id": resourceGroupId,
		},
	}
}

func flattenOpenShiftServicePrincipalProfile(profile *redhatopenshift.ServicePrincipalProfile, d *schema.ResourceData) []interface{} {
	if profile == nil {
		return []interface{}{}
	}

	clientID := ""
	if profile.ClientID != nil {
		clientID = *profile.ClientID
	}

	// client secret isn't returned by the API so pass the existing value along
	clientSecret := ""
	if sp, ok := d.GetOk("service_principal"); ok {
		var val []interface{}

		// prior to 1.34 this was a *pluginsdk.Set, now it's a List - try both
		if v, ok := sp.([]interface{}); ok {
			val = v
		} else if v, ok := sp.(*schema.Set); ok {
			val = v.List()
		}

		if len(val) > 0 && val[0] != nil {
			raw := val[0].(map[string]interface{})
			clientSecret = raw["client_secret"].(string)
		}
	}

	return []interface{}{
		map[string]interface{}{
			"client_id":     clientID,
			"client_secret": clientSecret,
		},
	}
}

func flattenOpenShiftNetworkProfile(profile *redhatopenshift.NetworkProfile) []interface{} {
	if profile == nil {
		return []interface{}{}
	}

	podCidr := ""
	if profile.PodCidr != nil {
		podCidr = *profile.PodCidr
	}

	serviceCidr := ""
	if profile.ServiceCidr != nil {
		serviceCidr = *profile.ServiceCidr
	}

	return []interface{}{
		map[string]interface{}{
			"pod_cidr":     podCidr,
			"service_cidr": serviceCidr,
		},
	}
}

func flattenOpenShiftMasterProfile(profile *redhatopenshift.MasterProfile) []interface{} {
	if profile == nil {
		return []interface{}{}
	}

	subnetId := ""
	if profile.SubnetID != nil {
		subnetId = *profile.SubnetID
	}

	return []interface{}{
		map[string]interface{}{
			"vm_size":            profile.VMSize,
			"subnet_id":          subnetId,
			"encryption_at_host": profile.EncryptionAtHost,
		},
	}
}

func flattenOpenShiftWorkerProfiles(profiles *[]redhatopenshift.WorkerProfile) []interface{} {
	if profiles == nil {
		return []interface{}{}
	}

	results := make([]interface{}, 0)

	result := make(map[string]interface{})
	result["node_count"] = int32(len(*profiles))

	for _, profile := range *profiles {
		if result["disk_size_gb"] == nil && profile.DiskSizeGB != nil {
			result["disk_size_gb"] = profile.DiskSizeGB
		}

		vmSize := profile.VMSize

		if result["vm_size"] == nil && *vmSize != "" {
			result["vm_size"] = vmSize
		}

		if result["subnet_id"] == nil && profile.SubnetID != nil {
			result["subnet_id"] = profile.SubnetID
		}
		result["encryption_at_host"] = profile.EncryptionAtHost
	}

	results = append(results, result)

	return results
}

func flattenOpenShiftAPIServerProfile(profile *redhatopenshift.APIServerProfile) []interface{} {
	if profile == nil {
		return []interface{}{}
	}

	return []interface{}{
		map[string]interface{}{
			"visibility": string(profile.Visibility),
		},
	}
}

func flattenOpenShiftIngressProfiles(profiles *[]redhatopenshift.IngressProfile) []interface{} {
	if profiles == nil {
		return []interface{}{}
	}

	results := make([]interface{}, 0)

	for _, profile := range *profiles {
		result := make(map[string]interface{})
		result["visibility"] = string(profile.Visibility)

		results = append(results, result)
	}

	return results
}

func expandOpenshiftClusterProfile(input []interface{}, subscriptionId string) *redhatopenshift.ClusterProfile {
	resourceGroupName := fmt.Sprintf("aro-%s", randomDomainName)
	resourceGroupId := ResourceGroupID(subscriptionId, resourceGroupName)

	if len(input) == 0 {
		return &redhatopenshift.ClusterProfile{
			ResourceGroupID:      utils.String(resourceGroupId),
			Domain:               utils.String(randomDomainName),
			FipsValidatedModules: redhatopenshift.FipsValidatedModulesDisabled,
		}
	}

	config := input[0].(map[string]interface{})

	pullSecret := config["pull_secret"].(string)

	domain := config["domain"].(string)
	if domain == "" {
		domain = randomDomainName
	}

	return &redhatopenshift.ClusterProfile{
		ResourceGroupID:      utils.String(resourceGroupId),
		Domain:               utils.String(domain),
		PullSecret:           utils.String(pullSecret),
		FipsValidatedModules: redhatopenshift.FipsValidatedModulesDisabled,
	}
}

func expandOpenshiftServicePrincipalProfile(input []interface{}) *redhatopenshift.ServicePrincipalProfile {
	if len(input) == 0 {
		return nil
	}

	config := input[0].(map[string]interface{})

	clientId := config["client_id"].(string)
	clientSecret := config["client_secret"].(string)

	return &redhatopenshift.ServicePrincipalProfile{
		ClientID:     utils.String(clientId),
		ClientSecret: utils.String(clientSecret),
	}
}

func expandOpenshiftNetworkProfile(input []interface{}) *redhatopenshift.NetworkProfile {
	if len(input) == 0 {
		return &redhatopenshift.NetworkProfile{
			PodCidr:     utils.String("10.128.0.0/14"),
			ServiceCidr: utils.String("172.30.0.0/16"),
		}
	}

	config := input[0].(map[string]interface{})

	podCidr := config["pod_cidr"].(string)
	serviceCidr := config["service_cidr"].(string)

	return &redhatopenshift.NetworkProfile{
		PodCidr:     utils.String(podCidr),
		ServiceCidr: utils.String(serviceCidr),
	}
}

func expandOpenshiftMasterProfile(input []interface{}) *redhatopenshift.MasterProfile {
	if len(input) == 0 {
		return nil
	}

	config := input[0].(map[string]interface{})

	vmSize := config["vm_size"].(string)
	subnetId := config["subnet_id"].(string)
	encryptionAtHost := config["encryption_at_host"].(string)

	return &redhatopenshift.MasterProfile{
		VMSize:           utils.String(vmSize),
		SubnetID:         utils.String(subnetId),
		EncryptionAtHost: redhatopenshift.EncryptionAtHost(encryptionAtHost),
	}
}

func expandOpenshiftWorkerProfiles(inputs []interface{}) *[]redhatopenshift.WorkerProfile {
	if len(inputs) == 0 {
		return nil
	}

	profiles := make([]redhatopenshift.WorkerProfile, 0)
	config := inputs[0].(map[string]interface{})

	// Hardcoded name required by ARO interface
	workerName := "worker"

	vmSize := config["vm_size"].(string)
	if vmSize == "" {
		vmSize = "Standard_D4s_v3"
	}

	diskSizeGb := int32(config["disk_size_gb"].(int))
	if diskSizeGb == 0 {
		diskSizeGb = 128
	}

	nodeCount := int32(config["node_count"].(int))
	if nodeCount == 0 {
		nodeCount = 3
	}

	subnetId := config["subnet_id"].(string)
	encryptionAtHost := config["encryption_at_host"].(string)

	profile := redhatopenshift.WorkerProfile{
		Name:             utils.String(workerName),
		VMSize:           utils.String(vmSize),
		DiskSizeGB:       utils.Int32(diskSizeGb),
		SubnetID:         utils.String(subnetId),
		Count:            utils.Int32(nodeCount),
		EncryptionAtHost: redhatopenshift.EncryptionAtHost(encryptionAtHost),
	}

	profiles = append(profiles, profile)

	return &profiles
}

func expandOpenshiftApiServerProfile(input []interface{}) *redhatopenshift.APIServerProfile {
	if len(input) == 0 {
		return &redhatopenshift.APIServerProfile{
			Visibility: redhatopenshift.Visibility(APIPublic),
		}
	}
	config := input[0].(map[string]interface{})

	visibility := config["visibility"].(string)

	return &redhatopenshift.APIServerProfile{
		Visibility: redhatopenshift.Visibility(visibility),
	}
}

func expandOpenshiftIngressProfiles(inputs []interface{}) *[]redhatopenshift.IngressProfile {
	profiles := make([]redhatopenshift.IngressProfile, 0)

	name := utils.String("default")
	visibility := string(redhatopenshift.VisibilityPublic)

	if len(inputs) > 0 {
		input := inputs[0].(map[string]interface{})
		visibility = input["visibility"].(string)
	}

	profile := redhatopenshift.IngressProfile{
		Name:       name,
		Visibility: redhatopenshift.Visibility(visibility),
	}

	profiles = append(profiles, profile)

	return &profiles
}
