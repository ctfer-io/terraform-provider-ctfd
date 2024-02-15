resource "ctfd_user" "ctfer" {
  name     = "CTFer"
  email    = "ctfer-io@protonmail.com"
  password = "password"
}

resource "ctfd_team" "cybercombattants" {
  name     = "Les cybercombattants de l'innovation"
  email    = "lucastesson@protonmail.com"
  password = "password"
  members = [
    ctfd_user.ctfer.id,
  ]
  captain = ctfd_user.ctfer.id
}
