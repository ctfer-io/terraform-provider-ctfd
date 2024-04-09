resource "ctfd_user" "ctfer" {
  name     = "CTFer"
  email    = "ctfer-io@protonmail.com"
  password = "password"

  # Make the user administrator of the CTFd instance
  type     = "admin"
  verified = true
  hidden   = true
}
