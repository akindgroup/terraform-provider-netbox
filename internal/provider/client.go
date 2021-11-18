package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"sync"
)

type httpClient struct {
	host  string
	token string
	*http.Client
	sync.Mutex
}

func (c *httpClient) GetIPAMPrefix(ctx context.Context, site, region string) (*IPAMPrefix, error) {
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

func (c *httpClient) CreateAvailableIPAddress(ctx context.Context, prefix_id, dns_name string) (*PrefixIPAddress, error) {
	c.Lock()
	defer c.Unlock()
	url := fmt.Sprintf("%s/api/ipam/prefixes/%s/available-ips", c.host, prefix_id)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte(fmt.Sprintf(`[{"dns_name": "%s"}]`))))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call netbox for available ip addresses for prefix id: %s %w", prefix_id, err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read available ip addresses for prefix id: %s, response body: %w", prefix_id, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to call netbox for available ip addresses for prefix id: %s, expected status %d, got %d, body: %s", prefix_id, http.StatusCreated, resp.StatusCode, body)
	}

	var prefixResponse PrefixIPAddress
	if err := json.Unmarshal(body, &prefixResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json available ip addresses response body: %w, body: %s", err, body)
	}
	if _, _, err := net.ParseCIDR(prefixResponse.Address); err != nil {
		return nil, fmt.Errorf("failed to parse ip returned from netbox, got %s, error: %w", prefixResponse.Address, err)
	}
	if prefixResponse.DNSName != dns_name {
		return nil, fmt.Errorf("dns name returned from netbox was %s, expected: %s", prefixResponse.DNSName, dns_name)
	}
	if prefixResponse.Status.Value != "active" {
		return nil, fmt.Errorf("expected ip to be set to active in netbox, got %s, expected: %s", prefixResponse.Status.Value, "active")
	}
	return &prefixResponse, nil
}

type mockclient struct{}

func (m *mockclient) GetIPAMPrefix(ctx context.Context, site, region string) (*IPAMPrefix, error) {
	return &IPAMPrefix{
		ID:   "foo",
		CIDR: "bar",
	}, nil
}
