package config

const exampleConfig = `# wisirc configuration
# Each server goes in its own [servers.<name>] section.

[servers.libera]
  nickname = "my-nick"
  username = "my-nick"
  server = "irc.libera.chat"
  port = 6697
  channels = ["#wisirc"]
  # dangerously_accept_invalid_certs = true   # only for self-signed certificates

  # [servers.libera.sasl.plain]
  # username = "my-nick"
  # password = "change-me"

# [notifications]
#   enabled = true

# [theme]
#   name = "dark"   # or "light"
`

const starterConfig = `# wisirc configuration
# Your servers are saved in servers.toml (≡ → Servers). You can also
# write them here, each in its own [servers.<name>] section:
#
# [servers.libera]
#   nickname = "my-nick"
#   server = "irc.libera.chat"
#   port = 6697
#   channels = ["#wisirc"]

# [notifications]
#   enabled = true

# [theme]
#   name = "dark"
`
