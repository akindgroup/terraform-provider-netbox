package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type httpclient struct {
	host  string
	token string
	*http.Client
}

func (c *httpclient) GetIPAMPrefix(ctx context.Context, site, region string) (*IPAMPrefix, error) {
	url := fmt.Sprintf("%s/api/ipam/prefixes?region=%s&site=%s&format=json", c.host, region, site)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call netbox for prefix: %w", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read prefix response body: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to call netbox for prefix, expected status 200, got %d, body: %s", resp.StatusCode, body)
	}
	type prefixAPIResponse struct {
		Count   int `json:"count"`
		Results []struct {
			ID     int    `json:"id"`
			Prefix string `json:"prefix"`
		} `json:"results"`
	}
	var prefixResponse prefixAPIResponse
	if err := json.Unmarshal(body, &prefixResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json prefix response body: %w, body: %s", err, body)
	}
	if prefixResponse.Count < 1 {
		return nil, fmt.Errorf("prefix not found in region: %s and site: %s", region, site)
	}
	if prefixResponse.Count > 1 {
		return nil, fmt.Errorf("multiple prefixes found, but expected just one in region: %s and site: %s", region, site)
	}
	return &IPAMPrefix{
		ID:   strconv.Itoa(prefixResponse.Results[0].ID),
		CIDR: prefixResponse.Results[0].Prefix,
	}, nil
}

type mockclient struct{}

func (m *mockclient) GetIPAMPrefix(ctx context.Context, site, region string) (*IPAMPrefix, error) {
	return &IPAMPrefix{
		ID:   "foo",
		CIDR: "bar",
	}, nil
}
