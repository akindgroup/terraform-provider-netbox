# Terraform Provider for Netbox

Intended to be a bare minimum terraform provider for Netbox IPAM functionality.

Get an available IP-address from a given prefix and reserve it for your server.

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.15

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the make `build_dev` command: 
```sh
$ make build_dev
```



## Using the provider

Make sure to set the given environment variables:
```sh
NETBOX_HOST=https://netbox.example.com
NETBOX_TOKEN=fawofiawjef0230-f8jqf0a8e9j
```

An optional configuration for disabling TLS verification is available:
```sh
NETBOX_TLS_VERIFY=false  # defaults to true
```


## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.
*Note:* Acceptance tests create real resources in Netbox, make sure you use a test instance.

```sh
$ make testacc
```


# Running the provider from local dev build


1. Build the provider
```sh
$ make build_dev
```
2. Set _~/.terraformrc_ to:
```hcl
provider_installation {
  dev_overrides {
    "academicwork/netbox" = "/home/YOUR-USERNAME/.terraform.d/plugins/academicwork/netbox/dev/YOUR-ARCHITECHTURE/" # architecture could be linux_amd64 for example
  }
  direct {}
}
```
