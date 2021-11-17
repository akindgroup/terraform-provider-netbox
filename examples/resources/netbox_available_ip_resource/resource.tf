resource "netbox_available_ip" "foo" {
  prefix_id = "123"
  dns_name  = "host.example.com"
}
