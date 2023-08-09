# Terraform Provider for CTFd

> Why creating this ?

For fun.
No, basically CTFd handles ressources that has life-cycles (mainly challenges), and providing them as cloud ressources enables your infrastructure to become modular and avoid shitty scripts i.e. `ctfcli`.

With the Terraform Provider for CTFd, you could setup your CTFd challenge using the following configuration.
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

By combining it with existing Terraform providers (e.g. Kubernetes), you can make use of real Infrastructure as Code, providing reproductibility, security and efficiency to the operations !

This provider also leverages the CTFd API that straigly exposes its data model as an API rather than providing the resources and handling the business layer internals.
