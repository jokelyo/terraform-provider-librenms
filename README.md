# Terraform Provider LibreNMS

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.11
- [Go](https://golang.org/doc/install) >= 1.23

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

To use this provider, first, add the `required_providers` block to your Terraform configuration and then define the provider.

```terraform
terraform {
  required_providers {
    librenms = {
      source  = "registry.terraform.io/jokelyo/librenms"
      version = ">=0.1.0" # Replace with the desired version
    }
  }
}

provider "librenms" {
  host = "https://your-librenms-instance/"  # or use LIBRENMS_HOST environment variable
  token = "your_api_token"                  # or use LIBRENMS_TOKEN environment variable
}

# Example resource
# resource "librenms_device" "example" {
#   # ... resource configuration ...
# }
```

See [main.tf](examples/example-plan/main.tf) for a more complete example.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```

For local development, create a file named `~/.terraformrc` (or `terraform.rc` on Windows) with the following content, 
adjusting the path to where your go install command places the binary (typically `$GOPATH/bin` or `$GOBIN`):
```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/jokelyo/librenms" = "/path/to/your/gopath/bin" # Or C:\Users\YourUser\go\bin on Windows
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

## Development Environment

For information about the development Docker stack and available make commands for local development and testing, 
please see the [development/README.md](development/README.md) file.
