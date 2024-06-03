---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "ctfd_flag Resource - terraform-provider-ctfd"
subcategory: ""
description: |-
  A flag to solve the challenge.
---

# ctfd_flag (Resource)

A flag to solve the challenge.

## Example Usage

```terraform
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

resource "ctfd_flag" "http_flag" {
  challenge_id = ctfd_challenge.http.id
  content      = "CTF{some_flag}"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `challenge_id` (String) Challenge of the flag.
- `content` (String, Sensitive) The actual flag to match. Consider using the convention `MYCTF{value}` with `MYCTF` being the shortcode of your event's name and `value` depending on each challenge.

### Optional

- `data` (String) The flag sensitivity information, either case_sensitive or case_insensitive
- `type` (String) The type of the flag, could be either static or regex

### Read-Only

- `id` (String) Identifier of the flag, used internally to handle the CTFd corresponding object.