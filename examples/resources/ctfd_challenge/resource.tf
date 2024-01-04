resource "ctfd_challenge" "http" {
  name        = "HTTP Authentication"
  category    = "network"
  description = <<-EOT
        Oh no ! I did not see my connection was no encrypted !
        I hope no one spied me...

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
    content = "HTTP exchanges are not ciphered."
    cost    = 50
    }, {
    content = "Content is POSTed in HTTP :)"
    cost    = 50
  }]

  files = [{
    name       = "capture.pcapng"
    contentb64 = filebase64("${path.module}/capture.pcapng")
  }]
}
