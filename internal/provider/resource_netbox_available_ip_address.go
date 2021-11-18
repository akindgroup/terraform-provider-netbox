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
		CreateContext: resourceAvailableIPAddressCreate,
		ReadContext:   resourceAvailableIPAddressRead,
		UpdateContext: resourceAvailableIPAddressUpdate,
		DeleteContext: resourceAvailableIPAddressDelete,
		Schema: map[string]*schema.Schema{
			"prefix_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"dns_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAvailableIPAddressCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		return diag.Errorf("failed to reserve ip address in netbox: %v", err)
	}

	d.SetId(strconv.Itoa(ip.ID))
	d.Set("ip_address", ip.Address)
	d.Set("dns_name", ip.DNSName)
	return nil
}

func resourceAvailableIPAddressRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, ok := m.(*httpClient)
	if !ok {
		return diag.Errorf("expected a http client for netbox, got: %v", c)
	}
	id, ok := d.Get("id").(string)
	if !ok {
		return diag.Errorf("required parameter id not provided")
	}

	ip, err := c.ReadReservedIPAddress(ctx, id)
	if err != nil {
		return diag.Errorf("failed to read reserved ip address from netbox: %v", err)
	}
	d.SetId(strconv.Itoa(ip.ID))
	d.Set("ip_address", ip.Address)
	d.Set("dns_name", ip.DNSName)
	return nil
}

func resourceAvailableIPAddressUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, ok := m.(*httpClient)
	if !ok {
		return diag.Errorf("expected a http client for netbox, got: %v", c)
	}
	id, ok := d.Get("id").(string)
	if !ok {
		return diag.Errorf("required parameter id not provided")
	}
	dns_name, ok := d.Get("dns_name").(string)
	if !ok {
		return diag.Errorf("required parameter dns_name not provided")
	}
	ip_address, ok := d.Get("ip_address").(string)
	if !ok {
		return diag.Errorf("required parameter ip_address not provided")
	}

	ip, err := c.UpdateReservedIPAddress(ctx, id, ip_address, dns_name)
	if err != nil {
		return diag.Errorf("failed to update reserved ip address in netbox: %v", err)
	}
	d.SetId(strconv.Itoa(ip.ID))
	d.Set("ip_address", ip.Address)
	d.Set("dns_name", ip.DNSName)
	return nil
}

func resourceAvailableIPAddressDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, ok := m.(*httpClient)
	if !ok {
		return diag.Errorf("expected a http client for netbox, got: %v", c)
	}
	id, ok := d.Get("id").(string)
	if !ok {
		return diag.Errorf("required parameter id not provided")
	}
	err := c.DeleteReservedIPAddress(ctx, id)
	if err != nil {
		return diag.Errorf("failed to delete reserved ip address in netbox: %v", err)
	}
	return nil
}
