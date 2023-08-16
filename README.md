<div align="center">
    <h1>Terraform Provider for CTFd</h1>
    <p><b>Time for CTF as Code</b><p>
    <a href="https://pkg.go.dev/github.com/pandatix/terraform-provider-ctfd"><img src="https://shields.io/badge/-reference-blue?logo=go&style=for-the-badge" alt="reference"></a>
	<a href="https://goreportcard.com/report/github.com/pandatix/terraform-provider-ctfd"><img src="https://goreportcard.com/badge/github.com/pandatix/terraform-provider-ctfd?style=for-the-badge" alt="go report"></a>
	<a href="https://coveralls.io/github/pandatix/terraform-provider-ctfd?branch=main"><img src="https://img.shields.io/coverallsCoverage/github/pandatix/terraform-provider-ctfd?style=for-the-badge" alt="Coverage Status"></a>
	<br>
	<a href=""><img src="https://img.shields.io/github/license/pandatix/terraform-provider-ctfd?style=for-the-badge" alt="License"></a>
	<a href="https://github.com/pandatix/terraform-provider-ctfd/actions?query=workflow%3Aci+"><img src="https://img.shields.io/github/actions/workflow/status/pandatix/terraform-provider-ctfd/ci.yaml?style=for-the-badge&label=CI" alt="CI"></a>
	<a href="https://github.com/pandatix/terraform-provider-ctfd/actions/workflows/codeql-analysis.yaml"><img src="https://img.shields.io/github/actions/workflow/status/pandatix/terraform-provider-ctfd/codeql-analysis.yaml?style=for-the-badge&label=CodeQL" alt="CodeQL"></a>
    <br>
    <a href="https://securityscorecards.dev/viewer/?uri=github.com/pandatix/terraform-provider-ctfd"><img src="https://img.shields.io/ossf-scorecard/github.com/pandatix/terraform-provider-ctfd?label=openssf%20scorecard&style=for-the-badge" alt="OpenSSF Scoreboard"></a>
</div>

## Why creating this ?

Terraform is used to manage resources that have lifecycles, to sum it up.

Well, that is the case of CTFd : it handles challenges that could be created, modified and deleted.
With some work to leverage the unsteady CTFd's API, Terraform is now able to manage them as cloud resources bringing you to opportunity of **CTF as Code**.

It avoids shitty scripts, `ctfcli` and other tools that does not solve the problem of reproductibility, ease of deployment and resiliency.

## How to use it ?

With the **Terraform Provider for CTFd**, you could setup your CTFd challenge using the following configuration.
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
