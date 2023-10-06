<div align="center">
    <h1>OpenTofu Provider for CTFd</h1>
    <p><b>Time for CTF(d) as Code</b><p>
    <a href="https://pkg.go.dev/github.com/ctfer-io/tofu-provider-ctfd"><img src="https://shields.io/badge/-reference-blue?logo=go&style=for-the-badge" alt="reference"></a>
	<a href="https://goreportcard.com/report/github.com/ctfer-io/tofu-provider-ctfd"><img src="https://goreportcard.com/badge/github.com/ctfer-io/tofu-provider-ctfd?style=for-the-badge" alt="go report"></a>
	<a href="https://coveralls.io/github/ctfer-io/tofu-provider-ctfd?branch=main"><img src="https://img.shields.io/coverallsCoverage/github/ctfer-io/tofu-provider-ctfd?style=for-the-badge" alt="Coverage Status"></a>
	<br>
	<a href=""><img src="https://img.shields.io/github/license/ctfer-io/tofu-provider-ctfd?style=for-the-badge" alt="License"></a>
	<a href="https://github.com/ctfer-io/tofu-provider-ctfd/actions?query=workflow%3Aci+"><img src="https://img.shields.io/github/actions/workflow/status/ctfer-io/tofu-provider-ctfd/ci.yaml?style=for-the-badge&label=CI" alt="CI"></a>
	<a href="https://github.com/ctfer-io/tofu-provider-ctfd/actions/workflows/codeql-analysis.yaml"><img src="https://img.shields.io/github/actions/workflow/status/ctfer-io/tofu-provider-ctfd/codeql-analysis.yaml?style=for-the-badge&label=CodeQL" alt="CodeQL"></a>
    <br>
    <a href="https://securityscorecards.dev/viewer/?uri=github.com/ctfer-io/tofu-provider-ctfd"><img src="https://img.shields.io/ossf-scorecard/github.com/ctfer-io/tofu-provider-ctfd?label=openssf%20scorecard&style=for-the-badge" alt="OpenSSF Scoreboard"></a>
</div>

## Why creating this ?

OpenTofu is used to manage resources that have lifecycles, configurations, to sum it up.

That is the case of CTFd: it handles challenges that could be created, modified and deleted.
With some work to leverage the unsteady CTFd's API, OpenTofu is now able to manage them as cloud resources bringing you to opportunity of **CTF as Code**.

It avoids shitty scripts, `ctfcli` and other tools that does not solve the problem of reproductibility, ease of deployment and resiliency.

## How to use it ?

With the **OpenTofu Provider for CTFd**, you could setup your CTFd challenge using the following configuration.
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
