resource "ctfd_user" "ctfer" {
  username = "CTFer"
  email    = "ctfer-io@protonmail.com"
  password = "password"

  # Define as an administration account
  type     = "admin"
  verified = true
  hidden   = true
}
