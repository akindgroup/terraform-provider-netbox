package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceAvailableIP(t *testing.T) {
	site, region, dns_name := "mock-hz1", "helsinki", "bastion-0.mock-hz1.aw-platform.com"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAvailableIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAvailableIPConfigBasic(site, region, dns_name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAvailableIPExists("netbox_available_ip.bar"),
				),
			},
		},
	})
}

func testAccResourceAvailableIPConfigBasic(site, region, dns_name string) string {
	return fmt.Sprintf(`
data "netbox_ipam_prefix" "foo" {
  site = "%s"
  region = "%s"
}
resource "netbox_available_ip" "bar" {
  prefix_id     = data.netbox_ipam_prefix.foo.id
  dns_name 		= "%s"
}
`, site, region, dns_name)
}

func testAccCheckAvailableIPDestroy(s *terraform.State) error {
	c := testAccProvider.Meta().(*httpClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "netbox_available_ip" {
			continue
		}
		prefix, err := c.ReadReservedIPAddress(context.Background(), rs.Primary.ID)
		if err != nil {
			switch err {
			case ErrReservedIPNotFound:
				continue
			}
		}
		return fmt.Errorf("prefix_id still exists: %d", prefix.ID)
	}
	return nil
}

func testAccCheckAvailableIPExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no id set")
		}
		c := testAccProvider.Meta().(*httpClient)
		if _, err := c.ReadReservedIPAddress(context.Background(), rs.Primary.ID); err != nil {
			return err
		}
		return nil
	}
}
