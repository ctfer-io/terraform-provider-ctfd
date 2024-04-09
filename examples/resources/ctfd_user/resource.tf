resource "ctfd_user" "ctfer" {
  username = "CTFer"
  email    = "ctfer-io@protonmail.com"
  password = "password"

  type     = "admin"
  verified = true
  hidden   = true
}
