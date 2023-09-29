package azureopenshift

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/rh-mobb/terraform-provider-azureopenshift/azureopenshift/auth"
	"github.com/rh-mobb/terraform-provider-azureopenshift/azureopenshift/clients"
)

// Provider -
func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"subscription_id": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_SUBSCRIPTION_ID", ""),
				Description: "The Subscription ID which should be used.",
			},

			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_ID", ""),
				Description: "The Client ID which should be used.",
			},

			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_SECRET", ""),
				Description: "The Client Secret which should be used. For use When authenticating as a Service Principal using a Client Secret.",
			},

			"tenant_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_TENANT_ID", ""),
				Description: "The Tenant ID which should be used.",
			},

			"environment": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_ENVIRONMENT", "public"),
				Description: "The Cloud Environment which should be used. Possible values are public, usgovernment, and china. Defaults to public.",
			},

			"metadata_host": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_METADATA_HOSTNAME", ""),
				Description: "The Hostname which should be used for the Azure Metadata Service.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"azureopenshift_redhatopenshift_cluster": resourceOpenShiftCluster(),
		},
		DataSourcesMap: map[string]*schema.Resource{},
	}
	p.ConfigureContextFunc = providerConfigure(p)
	return p
}

func providerConfigure(p *schema.Provider) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		stopCtx, ok := schema.StopContext(ctx)
		if !ok {
			stopCtx = ctx
		}
		config := auth.Config{
			SubscriptionId: d.Get("subscription_id").(string),
			TenantId:       d.Get("tenant_id").(string),
			ClientSecret:   d.Get("client_secret").(string),
			ClientId:       d.Get("client_id").(string),
		}
		client, err := clients.NewClient(stopCtx, config)
		if err != nil {
			return nil, diag.Errorf("building AzureRM Client: %s", err)
		}
		return client, nil
	}
}
