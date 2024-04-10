terraform {
  required_providers {
    ctfd = {
      source = "registry.terraform.io/ctfer-io/ctfd"
    }
  }
}

provider "ctfd" {
  url = "http://localhost:8080"
}

resource "ctfd_challenge" "http" {
  name        = "HTTP Authentication"
  category    = "network"
  description = <<-EOT
        Oh non ! Je n'avais pas vu que ma connexion n'était pas chiffrée ! 
        J'espère que personne ne m'espionnait...

        Authors:
        - NicolasFgrx
    EOT
  value       = 500
  initial     = 500
  decay       = 17
  minimum     = 50
  state       = "visible"
  function    = "logarithmic"

  flags = [{
    content = "24HIUT{Http_1s_n0t_s3cuR3}"
  }]

  topics = [
    "Network"
  ]
  tags = [
    "network",
    "http"
  ]

  hints = [{
    content = "Les flux http ne sont pas chiffrés"
    cost    = 50
    }, {
    content = "Les informations sont POSTées en HTTP :)"
    cost    = 50
  }]

  files = [{
    name       = "capture.pcapng"
    contentb64 = filebase64("${path.module}/capture.pcapng")
  }]
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
  value       = 500
  decay       = 17
  minimum     = 50
  state       = "visible"
  requirements = {
    behavior      = "anonymized"
    prerequisites = [ctfd_challenge.http.id]
  }

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
    name       = "icmp.pcap"
    contentb64 = filebase64("${path.module}/icmp.pcap")
  }]
}
