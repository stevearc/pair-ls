# Relay Server

A relay server hosts the sharing webpage and listens for connections from a
`pair-ls lsp` server. You can run the relay server with `pair-ls relay`, and
connect to it with `pair-ls lsp -forward wss://my.relay.host.com`.

## Encryption

You can provide a x509 certificate and private key file to enable TLS (https)
support for the webserver. These can be passed in with the `-cert` and
`-key` arguments, or they can be put in the config file. An easy and free way to
get certificates is using [Let's Encrypt](https://letsencrypt.org/).

You can authenticate the LSP client from the relay server either by requiring
a password (`lspPassword` in the config file) or a client certificate
(`requireClientCert = true`). You can use the helper `pair-ls cert` command to
generate a PEM file with a self-signed certificate and private key inside it for this purpose. While you _can_ also use that as the main certificate for the webserver, it is not recommended as it will show a big error page in the browser and also be vulnerable to man-in-the-middle attacks.

A full configuration for a local forwarding server and a remote relay server
using password auth:

Relay server (`pair-ls relay -port 443`):

```toml
[server]
webPassword = "passw0rd"
lspPassword = "secur3"
# Your certificates from Let's Encrypt or similar
certFile = "/path/to/cert.pem"
keyFile = "/path/to/cert.key.pem"
```

Local server (`pair-ls lsp -forward wss://remote.server.com`):

```toml
[client]
password = "secur3"
```

A full configuration for a local forwarding server and a remote relay server
using client certs for auth:

Relay server (`pair-ls relay -port 443`):

```toml
[server]
webPassword = "passw0rd"
# Your certificates from Let's Encrypt or similar
certFile = "/path/to/cert.pem"
keyFile = "/path/to/cert.key.pem"
# Generated from pair-ls cert
clientCAs = "/path/to/shared.pem"
requireClientCert = true
```

Local server (`pair-ls lsp -forward wss://remote.server.com`):

```toml
[client]
# The same file, generated from pair-ls cert
certFile = "/path/to/shared.pem"
```
