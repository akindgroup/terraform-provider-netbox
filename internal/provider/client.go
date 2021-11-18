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
	url := fmt.Sprintf("%s/api/ipam/prefixes?region=%s&site=%s", c.host, region, site)
	type prefixAPIResponse struct {
		Count   int `json:"count"`
		Results []struct {
			ID     int    `json:"id"`
			Prefix string `json:"prefix"`
		} `json:"results"`
	}
	var prefixResponse prefixAPIResponse
	if err := c.call(ctx, url, http.MethodGet, nil, http.StatusOK, &prefixResponse); err != nil {
		return nil, fmt.Errorf("failed to get available ip address from prefix: %w", err)
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
	url := fmt.Sprintf("%s/api/ipam/prefixes/%s/available-ips/", c.host, prefix_id)
	reqbody := []byte(fmt.Sprintf(`[{"dns_name": "%s"}]`, dns_name))
	var prefixResponse []PrefixIPAddress
	if err := c.call(ctx, url, http.MethodPost, reqbody, http.StatusCreated, &prefixResponse); err != nil {
		return nil, fmt.Errorf("failed to create reserved ip address: %w", err)
	}
	prefix := prefixResponse[0]
	if _, _, err := net.ParseCIDR(prefix.Address); err != nil {
		return nil, fmt.Errorf("failed to parse ip returned from netbox: %s, error: %w", prefix.Address, err)
	}
	if prefix.DNSName != dns_name {
		return nil, fmt.Errorf("expected returned dns_name: %s, got: %s", dns_name, prefix.DNSName)
	}
	if prefix.Status.Value != "active" {
		return nil, fmt.Errorf("expected returned ip address status: active, got %s", prefix.Status.Value)
	}
	return &prefix, nil
}

func (c *httpClient) ReadReservedIPAddress(ctx context.Context, id string) (*PrefixIPAddress, error) {
	url := fmt.Sprintf("%s/api/ipam/ip-addresses/%s/", c.host, id)
	var prefixResponse PrefixIPAddress
	if err := c.call(ctx, url, http.MethodGet, nil, http.StatusOK, &prefixResponse); err != nil {
		return nil, fmt.Errorf("failed to read reserved ip address: %w", err)
	}
	if _, _, err := net.ParseCIDR(prefixResponse.Address); err != nil {
		return nil, fmt.Errorf("failed to parse ip returned from netbox, got %s, error: %w", prefixResponse.Address, err)
	}
	if strconv.Itoa(prefixResponse.ID) != id {
		return nil, fmt.Errorf("expected returned id to be %s, got %d", id, prefixResponse.ID)
	}
	if prefixResponse.Status.Value != "active" {
		return nil, fmt.Errorf("expected returned ip address status: active, got %s", prefixResponse.Status.Value)
	}
	return &prefixResponse, nil
}

func (c *httpClient) UpdateReservedIPAddress(ctx context.Context, id, ip_address, dns_name string) (*PrefixIPAddress, error) {
	url := fmt.Sprintf("%s/api/ipam/ip-addresses/%s/", c.host, id)
	reqbody := []byte(fmt.Sprintf(`{ "address": "%s", "dns_name": "%s" }`, ip_address, dns_name))
	var prefixResponse PrefixIPAddress
	if err := c.call(ctx, url, http.MethodPut, reqbody, http.StatusOK, &prefixResponse); err != nil {
		return nil, fmt.Errorf("failed to updated reserved ip address: %w", err)
	}
	if prefixResponse.DNSName != dns_name {
		return nil, fmt.Errorf("expected updated dns_name to be %s, got %s", dns_name, prefixResponse.DNSName)
	}
	if strconv.Itoa(prefixResponse.ID) != id {
		return nil, fmt.Errorf("expected returned prefix_id to be %s, got %d", id, prefixResponse.ID)
	}
	if prefixResponse.Status.Value != "active" {
		return nil, fmt.Errorf("expected returned ip address status: active, got %s", prefixResponse.Status.Value)
	}
	return &prefixResponse, nil
}

func (c *httpClient) DeleteReservedIPAddress(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/api/ipam/ip-addresses/%s/", c.host, id)
	if err := c.call(ctx, url, http.MethodDelete, nil, http.StatusNoContent, nil); err != nil {
		return fmt.Errorf("failed to delete reserved ip address: %w", err)
	}
	return nil
}

func (c *httpClient) call(ctx context.Context, url, method string, reqbody []byte, expectedStatusCode int, responseStruct interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqbody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", c.token))
	if reqbody != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call netbox: %w", err)
	}
	if expectedStatusCode == http.StatusNoContent &&
		resp.StatusCode == http.StatusNoContent {
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != expectedStatusCode {
		return fmt.Errorf("incorrect status code returned, expected status: %d, got: %d, body: %s", expectedStatusCode, resp.StatusCode, body)
	}
	if responseStruct != nil {
		if err := json.Unmarshal(body, &responseStruct); err != nil {
			return fmt.Errorf("failed to unmarshal json from response, error: %w, body: %s", err, body)
		}
	}
	return nil
}
