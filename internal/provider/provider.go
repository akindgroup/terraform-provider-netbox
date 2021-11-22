package provider

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"host": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("NETBOX_HOST", nil),
				},
				"token": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("NETBOX_TOKEN", nil),
				},
				"tls_verify": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
					DefaultFunc: schema.EnvDefaultFunc("NETBOX_TLS_VERIFY", nil),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"netbox_ipam_prefix": dataSourceIPAMPrefix(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"netbox_available_ip": resourceAvailableIPAddress(),
			},
		}
		p.ConfigureContextFunc = configure
		return p
	}
}

func configure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	transport := http.DefaultTransport
	if tlsv := d.Get("tls_verify").(bool); !tlsv {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	return &httpClient{
		host:  d.Get("host").(string),
		token: d.Get("token").(string),
		Client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
	}, nil
}
