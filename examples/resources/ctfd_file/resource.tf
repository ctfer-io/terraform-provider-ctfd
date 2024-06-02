resource "ctfd_challenge" "http" {
  name        = "My Challenge"
  category    = "misc"
  description = "..."
  value       = 500
  decay       = 100
  minimum     = 50
  state       = "visible"
  function    = "logarithmic"

  topics = [
    "Misc"
  ]
  tags = [
    "misc",
    "basic"
  ]
}

resource "ctfd_file" "http_file" {
  challenge_id = ctfd_challenge.http.id
  name         = "image.png"
  contentb64   = filebase64(".../image.png")
}
