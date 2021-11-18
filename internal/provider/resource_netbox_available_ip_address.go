package provider

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type PrefixIPAddress struct {
	ID      int    `json:"id"`
	Address string `json:"address"`
	Status  struct {
		Value string `json:"value"`
		Label string `json:"label"`
	} `json:"status"`
	DNSName string `json:"dns_name"`
}

func resourceAvailableIPAddress() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetboxAvailableIPAddressCreate,
		// ReadContext:   resourceNetboxAvailableIPAddressRead,
		// UpdateContext: resourceNetboxAvailableIPAddressUpdate,
		// DeleteContext: resourceNetboxAvailableIPAddressDelete,
		Schema: map[string]*schema.Schema{
			"prefix_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceNetboxAvailableIPAddressCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, ok := m.(*httpClient)
	if !ok {
		return diag.Errorf("expected a http client for netbox, got: %v", c)
	}
	prefix_id, ok := d.Get("prefix_id").(string)
	if !ok {
		return diag.Errorf("required parameter prefix_id not provided")
	}
	dns_name, ok := d.Get("dns_name").(string)
	if !ok {
		return diag.Errorf("required parameter dns_name not provided")
	}

	ip, err := c.CreateAvailableIPAddress(ctx, prefix_id, dns_name)
	if err != nil {
		return diag.Errorf("failed to reserve ip address in netbox: %w", err)
	}

	d.SetId(strconv.Itoa(ip.ID))
	d.Set("ip_address", ip.Address)
	d.Set("dns_name", ip.DNSName)
	return nil
}

func resourceNetboxAvailableIPAddressRead(d *schema.ResourceData, m interface{}) diag.Diagnostics {

	d.Set("ip_address", res.GetPayload().Address)
	d.Set("status", res.GetPayload().Status.Value)
	return nil
}

func resourceNetboxAvailableIPAddressUpdate(d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// update then get
	return nil
}

func resourceNetboxAvailableIPAddressDelete(d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}
