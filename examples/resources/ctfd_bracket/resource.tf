resource "ctfd_bracket" "juniors" {
  name        = "Juniors"
  description = "Bracket for 14-25 years old players."
  type        = "users"
}

resource "ctfd_user" "player1" {
  name     = "player1"
  email    = "player1@ctfer.io"
  password = "password"

  bracket_id = ctfd_bracket.juniors.id
}
