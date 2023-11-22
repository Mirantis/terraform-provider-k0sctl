# Mirantis K0S Ctl provider

A terraform provider which integrates the k0s project tooling to natively
convert a set of hosts to a kubernetes cluster using terraform resources.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.4
- [Go](https://golang.org/doc/install) >= 1.20
- [GoReleaser](https://goreleaser.com/) : If you want to use it locally

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the `make local` command (uses goreleaser)

```shell
 $/> make local
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using
Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

The provider, once installed properly can be used in any terraform root/chart.

### Using the released provider

Go to the terraform registry page and follow the instructions for declaring
the provider version in your chart/module

@see https://registry.terraform.io/providers/Mirantis/k0sctl/latest

### Using the local source code provider

The `make local` target will use goreleaser to build the provider, and
then provide instructions on how to configure `terraform` to use the
provider locally,

@see https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers

## Developing the Provider

### Using the local provider

You can develop the provider locally, and test the development version by building
the plugin locally, and then configuring terraform to use the local version as a
`dev_override` for the production version.

To build the local plugin:
```
make local
```

It is recommended that you use the `dev_override` by using a special TF config file
and running terraform with an environment variable telling it to use the special file.
This avoids using the development version globally, preventing simple mistakes.

First create a file like my_tf_config_file:

```
provider_installation {
# This disables the version and checksum verifications for this provider
# and forces Terraform to look for the k0sctl provider plugin in the
# given directory.
dev_overrides {
	"mirantis/k0sctl" = "path/to/this/repo/dist/terraform-provider-k0sctl_linux_amd64_v1"
}
# For all other providers, install them directly from their origin provider
# registries as normal. If you omit this, Terraform will _only_ use
# the dev_overrides block, and so no other providers will be available.
direct {}
}
```

@NOTE that you mus replace `linux` and `amd64` if you are on a Mac/Windows machine
  or not on a 64bit intel/amd processor.  See `go env GOOS` and `go env GOARCH` for
  the correct values.

then run terraform with a config file override pointing to the new file:
```
 $/> TF_CLI_CONFIG_FILE=my_tf_config_file terraform plan
```
(or use an environment variable export)

@see: https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers"

### Contributing

To generate or update documentation, run `go generate`.

In order to run the testing mode unit test suite:

```
make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests require that you have an environment set up for
		testing that k0sctl can use.

```shell
make testacc
```

<!-- BEGIN_TF_DOCS -->
## Requirements

No requirements.

## Providers

No providers.

## Modules

No modules.

## Resources

No resources.

## Inputs

No inputs.

## Outputs

No outputs.
<!-- END_TF_DOCS -->
