package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type PrefixAPIResponse struct {
	Count   int `json:"count"`
	Results []struct {
		ID     string `json:"id"`
		Prefix string `json:"prefix"`
	} `json:"results"`
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
		},
	}
}

func dataSourceIpamPrefixRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*netboxHTTPClient)

	region, site := d.Get("region").(string), d.Get("site").(string)
	url := fmt.Sprintf("%s/api/ipam/prefixes/?region=%s&site=%s&format=json", c.host, region, site)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return diag.Errorf("failed to create request: %v", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))

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
	if err := json.Unmarshal(body, &prefixResponse); err != nil {
		return diag.Errorf("failed to unmarshal json prefix response body: %v", err)
	}
	if prefixResponse.Count < 1 {
		return diag.Errorf("prefix not found in region: %s and site: %s", region, site)
	}
	if prefixResponse.Count > 1 {
		return diag.Errorf("multiple prefixes found, but expected just one in region: %s and site: %s", region, site)
	}

	d.SetId(prefixResponse.Results[0].ID)
	d.Set("cidr", prefixResponse.Results[0].Prefix)

	return diags
}
