package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type PrefixAPIResponse struct {
	Count    int         `json:"count"`
	Next     interface{} `json:"next"`
	Previous interface{} `json:"previous"`
	Results  []struct {
		ID      int    `json:"id"`
		URL     string `json:"url"`
		Display string `json:"display"`
		Family  struct {
			Value int    `json:"value"`
			Label string `json:"label"`
		} `json:"family"`
		Prefix string `json:"prefix"`
		Site   struct {
			ID      int    `json:"id"`
			URL     string `json:"url"`
			Display string `json:"display"`
			Name    string `json:"name"`
			Slug    string `json:"slug"`
		} `json:"site"`
		Vrf    interface{} `json:"vrf"`
		Tenant interface{} `json:"tenant"`
		Vlan   interface{} `json:"vlan"`
		Status struct {
			Value string `json:"value"`
			Label string `json:"label"`
		} `json:"status"`
		Role         interface{}   `json:"role"`
		IsPool       bool          `json:"is_pool"`
		MarkUtilized bool          `json:"mark_utilized"`
		Description  string        `json:"description"`
		Tags         []interface{} `json:"tags"`
		CustomFields struct {
		} `json:"custom_fields"`
		Created     string    `json:"created"`
		LastUpdated time.Time `json:"last_updated"`
		Children    int       `json:"children"`
		Depth       int       `json:"_depth"`
	} `json:"results"`
}

func dataSourceIPAMPrefix() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIpamPrefixRead,
		Schema: map[string]*schema.Schema{
			"prefix_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"status": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"label": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"value": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceIpamPrefixRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	c := m.(*netboxHTTPClient)

	region, site := d.Get("region").(string), d.Get("site").(string)

	// https://netbox.academicwork.net/api/ipam/prefixes/?region=helsinki&site=dev-hz1

	url := fmt.Sprintf("%s/api/ipam/prefixes/?region=%s&site=%s&format=json", host, region, site)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return diag.Errorf("failed to create request: %v", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return diag.Errorf("failed to call netbox for prefix: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("failed to call netbox for prefix, expected status 200, got %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return diag.Errorf("failed to read prefix response body: %v", err)
	}
	defer resp.Body.Close()

	var prefixResponse PrefixAPIResponse
	if err := json.Unmarshal(body, &prefix); err != nil {
		return diag.Errorf("failed to unmarshal json prefix response body: %v", err)
	}

	if prefixResponse.Count < 1 {
		return diag.Errorf("prefix not found in region: %s and site: %s", region, site)
	}

	d.Set("prefix_id")

	// d.SetId(strconv.FormatInt(resp.Payload.ID, 10))
	// d.Set("family", flattenIpamPrefixFamily(resp.Payload.Family))
	// d.Set("prefix", resp.Payload.Prefix)
	// d.Set("site", flattenIpamPrefixSite(resp.Payload.Site))
	// d.Set("vrf", flattenIpamPrefixVRF(resp.Payload.Vrf))
	// d.Set("tenant", flattenIpamPrefixTenant(resp.Payload.Tenant))
	// d.Set("vlan", flattenIpamPrefixVLAN(resp.Payload.Vlan))
	// d.Set("status", flattenIpamPrefixStatus(resp.Payload.Status))
	// d.Set("role", flattenIpamPrefixRole(resp.Payload.Role))
	// d.Set("is_pool", resp.Payload.IsPool)
	// d.Set("description", resp.Payload.Description)
	// d.Set("tags", flattenTags(resp.Payload.Tags))
	// d.Set("custom_fields", resp.Payload.CustomFields)

	return diags
}
