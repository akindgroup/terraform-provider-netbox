package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDataSourceIpamPrefixRead(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceIpamPrefixRead,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.netbox_ipam_prefix.foo", "cidr", regexp.MustCompile("^ba")),
				),
			},
		},
	})
}

const testDataSourceIpamPrefixRead = `
data "netbox_ipam_prefix" "foo" {
  site = "foo"
  region = "bar"
}
`
