resource "ctfd_challenge_standard" "example" {
  name        = "Example challenge"
  category    = "test"
  description = "Example challenge description..."
  value       = 500
}

resource "ctfd_solution" "wu" {
  challenge_id = ctfd_challenge_standard.example.id
  content      = "Here is how to solve the challenge: ..."
  state        = "visible"
}
