package provider

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
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
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("NETBOX_TOKEN", nil),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"netbox_ipam_prefix": dataSourceIPAMPrefix(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"netbox_ipam_available_ip_address": resourceIPAMAvailableIPAddress(),
			},
		}
		p.ConfigureContextFunc = configure
		return p
	}
}

type netboxHTTPClient struct {
	host  string
	token string
	*http.Client
}

func configure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return &netboxHTTPClient{
		host:  d.Get("host").(string),
		token: d.Get("token").(string),
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}
