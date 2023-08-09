terraform {
    required_providers {
        ctfd = {
            source = "registry.terraform.io/pandatix/ctfd"
        }
    }
}

provider "ctfd" {
    url = "http://localhost:8080"
}

resource "ctfd_challenge" "icmp" {
    name        = "Stealing data"
    category    = "network"
    description = <<-EOT
        L'administrateur réseau vient de nous signaler que des flux étranges étaient à destination d'un serveur. 
        Visiblement, il s'agit d'un serveur interne. Vous pouvez nous dire de quoi il s'agit ?

        (La capture a été réalisée en dehors de l'infrastructure du CTF)

        Authors:
        - NicolasFgrx
    EOT
    // TODO find a way to avoid this shitty pattern (either <value> with type="static" or <initial,decay,minimum> with type="dynamic")
    value       = 500
    initial     = 500
    decay       = 17
    minimum     = 50
    state       = "visible"

    flags = [{
        content = "24HIUT{IcmpExfiltrationIsEasy}"
    }]

    topics = [
        "Network"
    ]
    tags = [
        "network",
        "icmp"
    ]

    hints = [{
        content = "Vous ne trouvez pas qu'il ya beaucoup de requêtes ICMP ?"
        cost    = 50
    }, {
        content = "Pour l'exo, le ttl a été modifié, tente un `ip.ttl<=20`"
        cost    = 50
    }]

    files = [{
        name    = "icmp.pcap"
        content = file("${path.module}/icmp.pcap")
    }]
}

# resource "ctfd_flag" "some_flag" {
#     challenge = ctfd_challenge.icmp.id
#     content   = "24HIUT{IcmpExfiltrationIsEasy-gg}"
# }

# data "ctfd_challenges" "all" {}

# output "all_challenges" {
#     value = data.ctfd_challenges.all
# }
