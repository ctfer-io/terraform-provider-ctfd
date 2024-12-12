resource "ctfd_challenge_dynamic" "http" {
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

resource "ctfd_flag" "http_flag" {
  challenge_id = ctfd_challenge_dynamic.http.id
  content      = "CTF{some_flag}"
}

resource "ctfd_hint" "http_hint_1" {
  challenge_id = ctfd_challenge_dynamic.http.id
  content      = "Some super-helpful hint"
  cost         = 50
}

resource "ctfd_hint" "http_hint_2" {
  challenge_id = ctfd_challenge_dynamic.http.id
  content      = "Even more helpful hint !"
  cost         = 50
  requirements = [ctfd_hint.http_hint_1.id]
}

resource "ctfd_file" "http_file" {
  challenge_id = ctfd_challenge_dynamic.http.id
  name         = "image.png"
  contentb64   = filebase64(".../image.png")
}
