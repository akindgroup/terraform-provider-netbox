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
	url := fmt.Sprintf("%s/api/ipam/prefixes/%s/available-ips/", c.host, prefix_id) // Trailing slash needed

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte(fmt.Sprintf(`[{"dns_name": "%s"}]`, dns_name))))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
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

	var prefixResponse []PrefixIPAddress
	if err := json.Unmarshal(body, &prefixResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json available ip addresses response body: %w, body: %s", err, body)
	}
	if len(prefixResponse) != 1 {
		return nil, fmt.Errorf("expected just one ip address in response , got : %d, body: %s", len(prefixResponse), body)
	}
	prefix := prefixResponse[0]

	if _, _, err := net.ParseCIDR(prefix.Address); err != nil {
		return nil, fmt.Errorf("failed to parse ip returned from netbox, got %s, error: %w", prefix.Address, err)
	}
	if prefix.DNSName != dns_name {
		return nil, fmt.Errorf("dns name returned from netbox was %s, expected: %s", prefix.DNSName, dns_name)
	}
	if prefix.Status.Value != "active" {
		return nil, fmt.Errorf("expected ip to be set to active in netbox, got %s, expected: %s", prefix.Status.Value, "active")
	}
	return &prefix, nil
}

func (c *httpClient) ReadReservedIPAddress(ctx context.Context, id string) (*PrefixIPAddress, error) {
	c.Lock()
	defer c.Unlock()

	url := fmt.Sprintf("%s/api/ipam/ip-addresses/%s/", c.host, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call netbox for reserved ip address id: %s %w", id, err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read reserved ip address id: %s, response body: %w", id, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to call netbox for reserved ip address id: %s, expected status %d, got %d, body: %s", id, http.StatusOK, resp.StatusCode, body)
	}

	var prefixResponse PrefixIPAddress
	if err := json.Unmarshal(body, &prefixResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json reserved ip address response body: %w, body: %s", err, body)
	}
	if _, _, err := net.ParseCIDR(prefixResponse.Address); err != nil {
		return nil, fmt.Errorf("failed to parse ip returned from netbox, got %s, error: %w", prefixResponse.Address, err)
	}

	if strconv.Itoa(prefixResponse.ID) != id {
		return nil, fmt.Errorf("expected retrieved id to be %s, got %d", id, prefixResponse.ID)
	}
	if prefixResponse.Status.Value != "active" {
		return nil, fmt.Errorf("expected ip to be set to active in netbox, got %s, expected: %s", prefixResponse.Status.Value, "active")
	}
	return &prefixResponse, nil
}

func (c *httpClient) UpdateReservedIPAddress(ctx context.Context, id, ip_address, dns_name string) (*PrefixIPAddress, error) {
	c.Lock()
	defer c.Unlock()

	url := fmt.Sprintf("%s/api/ipam/ip-addresses/%s", c.host, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader([]byte(fmt.Sprintf(`{ "address": "%s", "dns_name": "%s" }`, ip_address, dns_name))))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call netbox to update reserved ip address id: %s %w", id, err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to update reserved ip address id: %s, response body: %w", id, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to call netbox to update reserved ip address id: %s, expected status %d, got %d, body: %s", id, http.StatusOK, resp.StatusCode, body)
	}

	var prefixResponse PrefixIPAddress
	if err := json.Unmarshal(body, &prefixResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json for reserved ip address response body: %w, body: %s", err, body)
	}
	if prefixResponse.DNSName != dns_name {
		return nil, fmt.Errorf("expected updated dns_name to be %s, got %s", dns_name, prefixResponse.DNSName)
	}

	if strconv.Itoa(prefixResponse.ID) != id {
		return nil, fmt.Errorf("expected retrieved id to be %s, got %d", id, prefixResponse.ID)
	}
	if prefixResponse.Status.Value != "active" {
		return nil, fmt.Errorf("expected ip to be set to active in netbox, got %s, expected: %s", prefixResponse.Status.Value, "active")
	}
	return &prefixResponse, nil
}

func (c *httpClient) DeleteReservedIPAddress(ctx context.Context, id string) error {
	c.Lock()
	defer c.Unlock()

	url := fmt.Sprintf("%s/api/ipam/ip-addresses/%s", c.host, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call netbox to delete reserved ip address id: %s %w", id, err)
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to call netbox to delete reserved ip address id: %s, expected status %d, got %d", id, http.StatusNoContent, resp.StatusCode)
	}

	return nil
}
