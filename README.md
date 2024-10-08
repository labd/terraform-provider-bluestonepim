# Bluestone PIM Terraform Provider

The Terraform provider allows you to configure your
[Bluestone PIM](https://www.bluestonepim.com/) project with infrastructure-as-code
principles.

# Commercial support

Need support implementing this terraform provider in your organization? Or are
you missing features that need to be added, then we are able to offer support.
Please contact us at opensource@labdigital.nl

# Quick start

[Read our documentation](https://registry.terraform.io/providers/labd/bluestonepim/latest/docs)

## Usage

The provider is distributed via the Terraform registry. To use it you need to configure
the [`required_provider`](https://www.terraform.io/language/providers/requirements#requiring-providers) block. For example:

```hcl
terraform {
  required_providers {
    bluestonepim = {
      source = "labd/bluestonepim"

      # It's recommended to pin the version, e.g.:
      # version = "~> 0.0.1"
    }
  }
}

provider "bluestonepim" {
  client_secret = "your mapi client secret (api key)"
}


```

# Contributing

## Requirements

- [Go](https://golang.org/doc/install) >= 1.20
- [Task](https://taskfile.dev/installation/)
- [Changie](https://github.com/miniscruff/changie)

## Building the provider

Clone repository to

Enter the provider directory and build the provider

```sh
$ task build-local
```

A build is created `terraform-provider-bluestonepim.0.0` in the root directory
and added to plugin folder available locally:


Use version `99.0.0` in the provider to test your changes locally

```hcl
terraform {
  required_providers {
    bluestonepim = {
      source  = "labd/bluestonepim"
      version = "99.0.0"
    }
  }
}
```

## Debugging / Troubleshooting

There are two environment settings for troubleshooting:

- `TF_LOG=INFO` enables debug output for Terraform.
- `BSP_DEBUG=1` enables debug output to see request/responses to Bluestone PIM

Note this generates a lot of output!

## Releasing

When creating a PR with changes, please include a changie file in the
`changelogs/unreleased` folder. This file can be interactively generated by
running `changie new` in the root of the project. Pick a suitable category for
the change. We recommend `Fixed` or `Added` for most cases. See the
[changie configuration](./.changie.yaml) for the full list of categories.

Once a new version is released all the unreleased changelog files will be merged
and added to the general CHANGELOG.md file.

## Testing

### Running the unit tests

```sh
$ task test
```

### Running the unit tests with coverage

```sh
$ task coverage
```



## Authors

This project is developed by [Lab Digital](https://www.labdigital.nl). We
welcome additional contributors. Please see our
[GitHub repository](https://github.com/labd/terraform-provider-bluestonepim)
for more information.
