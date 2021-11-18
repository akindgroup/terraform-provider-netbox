package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type IPAMPrefix struct {
	ID   string
	CIDR string
}

func dataSourceIPAMPrefix() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIpamPrefixRead,
		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"site": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cidr": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceIpamPrefixRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c, ok := m.(*httpClient)
	if !ok {
		return diag.Errorf("expected a http client for netbox, got: %v", c)
	}
	region, ok := d.Get("region").(string)
	if !ok {
		return diag.Errorf("required parameter region not provided")
	}
	site, ok := d.Get("site").(string)
	if !ok {
		return diag.Errorf("required parameter site not provided")
	}
	prefix, err := c.GetIPAMPrefix(ctx, site, region)
	if err != nil {
		return diag.Errorf("failed to get prefix from netbox: %v", err)
	}
	d.SetId(prefix.ID)
	if err := d.Set("cidr", prefix.CIDR); err != nil {
		return diag.Errorf("failed to set cidr value of ipam prefix: %v", err)
	}
	return diags
}
