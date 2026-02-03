terraform {
  required_providers {
    ctfd = {
      source = "ctfer-io/ctfd"
    }
  }
}

provider "ctfd" {
  url      = "http://localhost:8000"
  username = "ctfer"
  password = "ctfer"
}

resource "ctfd_challenge_standard" "http" {
  name        = "My Challenge"
  category    = "misc"
  description = "..."
  value       = 500

  topics = [
    "Misc"
  ]
  tags = [
    "misc",
    "basic"
  ]
}

resource "ctfd_flag" "http_flag" {
  challenge_id = ctfd_challenge_standard.http.id
  content      = "CTF{some_flag}"
}

resource "ctfd_hint" "http_hint_1" {
  challenge_id = ctfd_challenge_standard.http.id
  content      = "Some super-helpful hint"
  cost         = 50
}

resource "ctfd_hint" "http_hint_2" {
  challenge_id = ctfd_challenge_standard.http.id
  content      = "Even more helpful hint !"
  cost         = 50
  requirements = [ctfd_hint.http_hint_1.id]
}

resource "ctfd_file" "http_file" {
  challenge_id = ctfd_challenge_standard.http.id
  name         = "dump.txt"
  contentb64   = "aHR0cHM6Ly95b3V0dS5iZS9kUXc0dzlXZ1hjUQo="
}
