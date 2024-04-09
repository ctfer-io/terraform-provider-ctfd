resource "ctfd_challenge" "http" {
  name        = "My Challenge"
  category    = "misc"
  description = "..."
  value       = 500
  decay       = 100
  minimum     = 50
  state       = "visible"
  function    = "logarithmic"

  flags = [{
    content = "CTF{some_flag}"
  }]

  topics = [
    "Misc"
  ]
  tags = [
    "misc",
    "basic"
  ]

  hints = [{
    content = "Some super-helpful hint"
    cost    = 50
    }, {
    content = "Even more helpful hint !"
    cost    = 50
  }]

  files = [{
    name       = "image.png"
    contentb64 = filebase64(".../image.png")
  }]
}
