package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceIpamPrefixRead(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceIpamPrefixRead,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.netbox_ipam_prefix.foo", "id", "59"),
					resource.TestCheckResourceAttr("data.netbox_ipam_prefix.foo", "cidr", "10.34.0.0/21"),
				),
			},
		},
	})
}

const testDataSourceIpamPrefixRead = `
data "netbox_ipam_prefix" "foo" {
  site = "dev-hz1"
  region = "helsinki"
}
`
