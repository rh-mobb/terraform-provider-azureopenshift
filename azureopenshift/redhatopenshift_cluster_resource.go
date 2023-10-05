package azureopenshift

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	redhatopenshift "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redhatopenshift/armredhatopenshift"
	"github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/rh-mobb/terraform-provider-azureopenshift/azureopenshift/clients"
	"github.com/rh-mobb/terraform-provider-azureopenshift/azureopenshift/parse"
	openShiftValidate "github.com/rh-mobb/terraform-provider-azureopenshift/azureopenshift/validate"
	"github.com/rh-mobb/terraform-provider-azureopenshift/helpers/aro"
	"github.com/rh-mobb/terraform-provider-azureopenshift/helpers/azure"
	"github.com/rh-mobb/terraform-provider-azureopenshift/helpers/suppress"
	"github.com/rh-mobb/terraform-provider-azureopenshift/helpers/tf"
	"github.com/rh-mobb/terraform-provider-azureopenshift/helpers/utils"
	"github.com/rh-mobb/terraform-provider-azureopenshift/helpers/validate"
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

			"cluster_resource_group": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

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
							Computed:     true,
						},
						"version": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							Computed:     true,
						},
						"resource_group_id": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"fips_validated_modules": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      redhatopenshift.FipsValidatedModulesDisabled,
							ValidateFunc: validation.StringIsNotEmpty,
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
						"outbound_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      redhatopenshift.OutboundTypeLoadbalancer,
							ValidateFunc: validate.ValidateOutBoundType,
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
						"disk_encryption_set": {
							Type:     schema.TypeString,
							Optional: true,
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

			"internal_cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
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
						"disk_encryption_set": {
							Type:     schema.TypeString,
							Optional: true,
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
						"url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
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
						"ip": {
							Type:     schema.TypeString,
							Computed: true,
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
	subscriptionId := meta.(*clients.Client).SubscriptionID

	name := d.Get("name").(string)

	existing, err := client.Get(ctx, resourceGroupName, name, nil)
	if err != nil {
		responseError, ok := err.(*azcore.ResponseError)
		if !ok || responseError.StatusCode != 404 {
			return fmt.Errorf("checking for presence of existing Red Hat Openshift Cluster %q (Resource Group %q): %s", name, resourceGroupName, err)
		}
	}

	if existing.ID != nil && *existing.ID != "" {
		return tf.ImportAsExistsError("azurerm_redhatopenshift_cluster", *existing.ID)
	}

	location := d.Get("location").(string)

	clusterProfileRaw := d.Get("cluster_profile").([]interface{})
	clusterResourceGroup := d.Get("cluster_resource_group").(string)
	clusterProfile := aro.NewClusterProfileHelper(subscriptionId, clusterResourceGroup).Expand(clusterProfileRaw)

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
		Properties: &redhatopenshift.OpenShiftClusterProperties{
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

	future, err := client.BeginCreateOrUpdate(ctx, resourceGroupName, name, parameters, nil)
	if err != nil {
		return fmt.Errorf("creating Red Hat OpenShift Cluster %q (Resource Group %q): %+v", name, resourceGroupName, err)
	}

	if _, err = future.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("waiting for creation of Red Hat OpenShift Cluster %q (Resource Group %q): %+v", name, resourceGroupName, err)
	}

	read, err := client.Get(ctx, resourceGroupName, name, nil)
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

	subscriptionId := meta.(*clients.Client).SubscriptionID

	id, err := parse.ClusterID(d.Id())
	if err != nil {
		return err
	}

	d.Partial(true)

	existing, err := client.Get(ctx, id.ResourceGroup, id.ManagedClusterName, nil)
	if err != nil {
		return fmt.Errorf("retrieving existing Red Hat OpenShift Cluster %q (Resource Group %q): %+v", id.ManagedClusterName, id.ResourceGroup, err)
	}
	if existing.Properties == nil {
		return fmt.Errorf("retrieving existing Red Hat OpenShift Cluster %q (Resource Group %q): `properties` was nil", id.ManagedClusterName, id.ResourceGroup)
	}

	if d.HasChange("cluster_profile") {
		clusterProfileRaw := d.Get("cluster_profile").([]interface{})
		clusterResourceGroup := d.Get("cluster_resource_group").(string)
		clusterProfile := aro.NewClusterProfileHelper(subscriptionId, clusterResourceGroup).Expand(clusterProfileRaw)
		existing.Properties.ClusterProfile = clusterProfile
	}

	if d.HasChange("master_profile") {
		masterProfileRaw := d.Get("master_profile").([]interface{})
		masterProfile := expandOpenshiftMasterProfile(masterProfileRaw)
		existing.Properties.MasterProfile = masterProfile
	}

	if d.HasChange("worker_profile") {
		workerProfilesRaw := d.Get("worker_profile").([]interface{})
		workerProfiles := expandOpenshiftWorkerProfiles(workerProfilesRaw)
		existing.Properties.WorkerProfiles = workerProfiles
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

	resp, err := client.Get(ctx, id.ResourceGroup, id.ManagedClusterName, nil)
	if err != nil {
		responseError, ok := err.(*azcore.ResponseError)
		if !ok || responseError.StatusCode != 404 {
			return fmt.Errorf("checking for presence of existing Red Hat Openshift Cluster %q (Resource Group %q): %s", id.ManagedClusterName, id.ResourceGroup, err)
		}
	}

	d.Set("name", resp.Name)
	d.Set("resource_group_name", id.ResourceGroup)
	d.Set("location", resp.Location)

	if props := resp.Properties; props != nil {
		clusterProfile := flattenOpenShiftClusterProfile(props.ClusterProfile, d)
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

		internalClusterId, err := parse.InternalClusterId(*resp.Name, props.WorkerProfiles)
		if err != nil {
			return err
		}

		d.Set("internal_cluster_id", internalClusterId)

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

	credResponse, err := client.ListCredentials(ctx, id.ResourceGroup, id.ManagedClusterName, nil)
	if err != nil {
		responseError, ok := err.(*azcore.ResponseError)
		if !ok || responseError.StatusCode != 404 {
			return fmt.Errorf("checking for presence of existing Red Hat Openshift Cluster %q (Resource Group %q): %s", id.ManagedClusterName, id.ResourceGroup, err)
		}
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

	future, err := client.BeginDelete(ctx, id.ResourceGroup, id.ManagedClusterName, nil)
	if err != nil {
		return fmt.Errorf("deleting Red Hat Openshift Cluster %q (Resource Group %q): %+v", id.ManagedClusterName, id.ResourceGroup, err)
	}

	if _, err := future.PollUntilDone(ctx, nil); err != nil {
		return fmt.Errorf("waiting for the deletion of Red Hat Openshift Cluster %q (Resource Group %q): %+v", id.ManagedClusterName, id.ResourceGroup, err)
	}

	return nil
}

func flattenOpenShiftClusterProfile(profile *redhatopenshift.ClusterProfile, d *schema.ResourceData) []interface{} {
	if profile == nil {
		return []interface{}{}
	}

	// pull secret isn't returned by the API so pass the existing value along
	var pullSecret interface{}

	var version interface{}
	clusterProfileRaw := d.Get("cluster_profile").([]interface{})
	if len(clusterProfileRaw) != 0 {
		pullSecretRaw := d.Get("cluster_profile").([]interface{})[0].(map[string]interface{})["pull_secret"]
		if pullSecretRaw != nil {
			pullSecret = pullSecretRaw.(string)
		}

		versionRaw := d.Get("cluster_profile").([]interface{})[0].(map[string]interface{})["version"]
		if versionRaw != nil {
			version = versionRaw.(string)
		}
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
			"pull_secret":            pullSecret,
			"domain":                 clusterDomain,
			"resource_group_id":      resourceGroupId,
			"fips_validated_modules": profile.FipsValidatedModules,
			"version":                version,
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

	var outboundType *redhatopenshift.OutboundType
	if profile.OutboundType != nil {
		outboundType = profile.OutboundType
	}

	return []interface{}{
		map[string]interface{}{
			"pod_cidr":      podCidr,
			"service_cidr":  serviceCidr,
			"outbound_type": outboundType,
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
			"vm_size":             profile.VMSize,
			"subnet_id":           subnetId,
			"disk_encryption_set": profile.DiskEncryptionSetID,
			"encryption_at_host":  profile.EncryptionAtHost,
		},
	}
}

func flattenOpenShiftWorkerProfiles(profiles []*redhatopenshift.WorkerProfile) []interface{} {
	if profiles == nil {
		return []interface{}{}
	}

	results := make([]interface{}, 0)

	result := make(map[string]interface{})
	result["node_count"] = int32(len(profiles))

	for _, profile := range profiles {
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
		if result["disk_encryption_set"] == nil && profile.DiskEncryptionSetID != nil {
			result["disk_encryption_set"] = profile.DiskEncryptionSetID
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
			"visibility": string(*profile.Visibility),
			"url":        string(*profile.URL),
			"ip":         string(*profile.IP),
		},
	}
}

func flattenOpenShiftIngressProfiles(profiles []*redhatopenshift.IngressProfile) []interface{} {
	if profiles == nil {
		return []interface{}{}
	}

	results := make([]interface{}, 0)

	for _, profile := range profiles {
		result := make(map[string]interface{})
		result["visibility"] = string(*profile.Visibility)
		result["ip"] = string(*profile.IP)

		results = append(results, result)
	}

	return results
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
			PodCidr:      utils.String("10.128.0.0/14"),
			ServiceCidr:  utils.String("172.30.0.0/16"),
			OutboundType: to.Ptr(redhatopenshift.OutboundTypeLoadbalancer),
		}
	}

	config := input[0].(map[string]interface{})

	podCidr := config["pod_cidr"].(string)
	serviceCidr := config["service_cidr"].(string)
	outboundType := config["outbound_type"].(string)

	return &redhatopenshift.NetworkProfile{
		PodCidr:      utils.String(podCidr),
		ServiceCidr:  utils.String(serviceCidr),
		OutboundType: to.Ptr(redhatopenshift.OutboundType(outboundType)),
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

	var diskEncryptionSetId *string
	if config["disk_encryption_set"] != nil {
		diskEncryptionSetId = utils.String(config["disk_encryption_set"].(string))
	}

	return &redhatopenshift.MasterProfile{
		VMSize:              utils.String(vmSize),
		SubnetID:            utils.String(subnetId),
		EncryptionAtHost:    to.Ptr(redhatopenshift.EncryptionAtHost(encryptionAtHost)),
		DiskEncryptionSetID: diskEncryptionSetId,
	}
}

func expandOpenshiftWorkerProfiles(inputs []interface{}) []*redhatopenshift.WorkerProfile {
	if len(inputs) == 0 {
		return nil
	}

	profiles := make([]*redhatopenshift.WorkerProfile, 0)
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
	var diskEncryptionSetId *string
	if config["disk_encryption_set"] != nil {
		diskEncryptionSetId = utils.String(config["disk_encryption_set"].(string))
	}

	profile := &redhatopenshift.WorkerProfile{
		Name:                utils.String(workerName),
		VMSize:              utils.String(vmSize),
		DiskSizeGB:          utils.Int32(diskSizeGb),
		SubnetID:            utils.String(subnetId),
		Count:               utils.Int32(nodeCount),
		EncryptionAtHost:    to.Ptr(redhatopenshift.EncryptionAtHost(encryptionAtHost)),
		DiskEncryptionSetID: diskEncryptionSetId,
	}

	profiles = append(profiles, profile)

	return profiles
}

func expandOpenshiftApiServerProfile(input []interface{}) *redhatopenshift.APIServerProfile {
	if len(input) == 0 {
		return &redhatopenshift.APIServerProfile{
			Visibility: to.Ptr(redhatopenshift.Visibility(APIPublic)),
		}
	}
	config := input[0].(map[string]interface{})

	visibility := config["visibility"].(string)

	return &redhatopenshift.APIServerProfile{
		Visibility: to.Ptr(redhatopenshift.Visibility(visibility)),
	}
}

func expandOpenshiftIngressProfiles(inputs []interface{}) []*redhatopenshift.IngressProfile {
	profiles := make([]*redhatopenshift.IngressProfile, 0)

	name := utils.String("default")
	visibility := string(redhatopenshift.VisibilityPublic)

	if len(inputs) > 0 {
		input := inputs[0].(map[string]interface{})
		visibility = input["visibility"].(string)
	}

	profile := &redhatopenshift.IngressProfile{
		Name:       name,
		Visibility: to.Ptr(redhatopenshift.Visibility(visibility)),
	}

	profiles = append(profiles, profile)

	return profiles
}
