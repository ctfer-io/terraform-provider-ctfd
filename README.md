<div align="center">
    <h1>Terraform Provider for CTFd</h1>
    <p><b>Time for CTF(d) as Code</b><p>
    <a href="https://pkg.go.dev/github.com/ctfer-io/terraform-provider-ctfd"><img src="https://shields.io/badge/-reference-blue?logo=go&style=for-the-badge" alt="reference"></a>
	<a href="https://goreportcard.com/report/github.com/ctfer-io/terraform-provider-ctfd"><img src="https://goreportcard.com/badge/github.com/ctfer-io/terraform-provider-ctfd?style=for-the-badge" alt="go report"></a>
	<a href="https://coveralls.io/github/ctfer-io/terraform-provider-ctfd?branch=main"><img src="https://img.shields.io/coverallsCoverage/github/ctfer-io/terraform-provider-ctfd?style=for-the-badge" alt="Coverage Status"></a>
	<br>
	<a href=""><img src="https://img.shields.io/github/license/ctfer-io/terraform-provider-ctfd?style=for-the-badge" alt="License"></a>
	<a href="https://github.com/ctfer-io/terraform-provider-ctfd/actions?query=workflow%3Aci+"><img src="https://img.shields.io/github/actions/workflow/status/ctfer-io/terraform-provider-ctfd/ci.yaml?style=for-the-badge&label=CI" alt="CI"></a>
	<a href="https://github.com/ctfer-io/terraform-provider-ctfd/actions/workflows/codeql-analysis.yaml"><img src="https://img.shields.io/github/actions/workflow/status/ctfer-io/terraform-provider-ctfd/codeql-analysis.yaml?style=for-the-badge&label=CodeQL" alt="CodeQL"></a>
    <br>
    <a href="https://securityscorecards.dev/viewer/?uri=github.com/ctfer-io/terraform-provider-ctfd"><img src="https://img.shields.io/ossf-scorecard/github.com/ctfer-io/terraform-provider-ctfd?label=openssf%20scorecard&style=for-the-badge" alt="OpenSSF Scoreboard"></a>
</div>

## Why creating this ?

Terraform is used to manage resources that have lifecycles, configurations, to sum it up.

That is the case of CTFd: it handles challenges that could be created, modified and deleted.
With some work to leverage the unsteady CTFd's API, Terraform is now able to manage them as cloud resources bringing you to opportunity of **CTF as Code**.

With a paradigm-shifting vision of setting up CTFs, the Terraform Provider for CTFd avoid shitty scripts, `ctfcli` and other tools that does not solve the problem of reproductibility, ease of deployment and resiliency.

## How to use it ?

Install the **Terraform Provider for CTFd** by setting the following in your `main.tf file`.
```hcl
terraform {
    required_providers {
        ctfd = {
            source = "registry.terraform.io/ctfer-io/ctfd"
        }
    }
}

provider "ctfd" {
    url = "https://my-ctfd.lan"
}
```

We recommend setting the environment variable `CTFD_API_KEY` to enable the provider to communicate with your CTFd instance.

Then, you could use a `ctfd_challenge` resource to setup your CTFd challenge, with for instance the following configuration.
```hcl
resource "ctfd_challenge" "my_challenge" {
    name        = "My Challenge"
    category    = "Some category"
    description = <<-EOT
        My superb description !

        And it's multiline :o
    EOT
    state       = "visible"
    value       = 500
}
```
