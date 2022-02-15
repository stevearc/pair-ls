# pair-ls

pair-ls is a lightweight, editor-agnostic tool for remote pair-programming.

:warning: pair-ls is currently in **Alpha** status. Expect breaking API changes.
Backwards compatibility will be respected _after_ the first release.

- [Installation](#installation)
- [Setup](#setup)
- [Sharing](#sharing)
- [Configuration](#configuration)
- [Alternatives](#alternatives)
- [Technical Overview](#technical-overview)

## Installation

You can download a binary from [the releases
page](https://github.com/stevearc/pair-ls/releases)

If you want to build from source, `git clone` this repo and run `yarn build`
(you will need to [install
yarn](https://classic.yarnpkg.com/lang/en/docs/install/) and [go build tools](https://go.dev/doc/install))

## Setup

| Feature            | No plugin | Neovim |
| ------------------ | --------- | ------ |
| Read-only web view | ✓         | ✓      |
| Cursor tracking    | ✓\*       | ✓      |

\*Cursor tracking is limited

**Editor plugins:**

- Neovim: [pair-ls.nvim](https://github.com/stevearc/pair-ls.nvim)

### Generic Setup - any editor

pair-ls supports any editor with a [LSP
client](https://microsoft.github.io/language-server-protocol/).

Configure your LSP client to run this command as a server: `pair-ls lsp -port 8080`. Now you can visit http://localhost:8080 to see a mirror of your code. For
more info on how to share this page, see [Sharing](#sharing).

## Sharing

Running `pair-ls lsp -port 8080` is an easy way to get started, but how can you
share this with other people? There are two basic ways, and both of them require
you to have a server that is accessible by both users.

### Port forwarding

The easiest option is to forward a port from your remote server. First, make
sure that your server has `GatewayPorts yes` in the `sshd_config` file (and
restart the service if you had to change it).

Then from your local machine, run:

```
ssh -R 80:localhost:8080 my.server.com
```

This will forward port 80 on my.server.com to your localhost port 8080. Change
the ports and hostname as needed.

For more on port forwarding, see https://www.ssh.com/academy/ssh/tunneling/example#remote-forwarding

### Relay server

If you run `pair-ls lsp` with the `-forward <host>` option, it will forward all
of its data to a relay server. The relay server (run with `pair-ls relay`) will
listen for connections from a forwarding server and can also host the webserver
(with `-web-port`).

Note that the relay server _requires_ an encrypted channel, so both your local
machine and the relay server must have matching certificates (see the section on
[encryption](#encryption) below).

### Password protection

If any part of this is accessible from the public internet, you should probably
enable password-protection. Simply provide a password, either via the [config
file](#configuration) or with the environment variable `PAIR_WEB_PASS`. The
webserver will now require this password before allowing access. Whenever you do
this, you should also enable [Encryption](#encryption) so your password can't be
sniffed.

### Encryption

You can provide a x509 certificate and private key file to enable TLS (https)
support for the webserver. These can be passed in with the `-web-cert` and
`-web-key` arguments, or they can be put in the [config file](#configuration).
An easy and free way to get certificates is using [Let's
Encrypt](https://letsencrypt.org/).

If you are using a relay server, both the forwarding client and the relay server
need to have the same certificate. It should be passed in with `-relay-cert`, or
be put in the config file. There is an easy helper command `pair-ls cert` that
will generate a self-signed certificate for you that will work for this purpose.

You can also use the `pair-ls cert` certs for your webserver, but since it's
self-signed your browser will give you a big "Your connection is not private"
warning message that you will have to click through (and if you don't manually
verify the certificate your browser shows you, you could be vulnerable to a [MITM
attack](https://en.wikipedia.org/wiki/Man-in-the-middle_attack)). You will also
have to split the generated pem file into separate "certificate" and "key"
files.

A full configuration for a local forwarding server and a remote relay server
that is also running the webserver looks like this:

Local server (`pair-ls lsp`):

```json
{
  "forwardHost": "remote.server:8888",
  "relayCertFile": "/path/to/relay.pem"
}
```

Relay server (`pair-ls relay`):

```json
{
  "relayPort": 8888,
  "relayCertFile": "/path/to/relay.pem",
  "webPort": 80,
  "webKeyFile": "/path/to/server.key.pem",
  "webCertFile": "/path/to/server.pem",
  "webPassword": "asdfasdf"
}
```

## Configuration

The configuration file can be found at `$XDG_CONFIG_HOME/pair-ls.json`. Most
values can be specified on the command line instead, if you prefer (run
`pair-ls` for detailed help). The possible values are:

| Key             | Description                                                                               |
| --------------- | ----------------------------------------------------------------------------------------- |
| `logFile`       | Logs will be written here (default `$XDG_CACHE_HOME/pair-ls.log`)                         |
| `webKeyFile`    | Private key file for webserver TLS                                                        |
| `webCertFile`   | Certificate file for webserver TLS                                                        |
| `webHostname`   | Webserver binds to this hostname                                                          |
| `webPassword`   | Password to restrict access to webpage                                                    |
| `forwardHost`   | Address of the relay server                                                               |
| `relayHostname` | Relay server binds to this hostname                                                       |
| `relayPort`     | Relay server listens on this port                                                         |
| `relayPersist`  | Relay server retains file data even after forwarding clients disconnect (default `false`) |
| `relayCertFile` | Certificate file to encrypt connection to relay server                                    |

## Alternatives

Other options fall into roughly 3 categories: **screen sharing** tools, **web
editors** that require you to use their in-browser editor, and **editor
plugins** that, best-case, support a few popular editors. Most of these are paid
apps/services, though many of those have a free tier.

- Screen sharing
  - [Tuple](https://tuple.app/) (paid)
  - [Pop](https://pop.com/) (paid)
  - [Coscreen](https://www.coscreen.co/) (paid)
  - [Drovio](https://www.drovio.com/) (paid)
  - Plus all of the generic services like Zoom, Google Meet, Facebook Messenger,
    etc.
- Web editor
  - [Replit](https://replit.com/) (paid)
  - [Codeanywhere](https://codeanywhere.com/) (paid)
  - [CodeSandbox](https://codesandbox.io/) (paid)
  - [Cloud9](https://aws.amazon.com/cloud9/) (only AWS usage fees)
  - [Red Hat
    CodeReady](https://developers.redhat.com/products/codeready-workspaces/overview)
- Editor plugin
  - [Microsoft Live
    Share](https://visualstudio.microsoft.com/services/live-share/)
    (VSCode, Visual Studio)
  - [Duckly](https://duckly.com/) (VSCode, IntelliJ) (paid)
  - [CodeTogether](https://www.codetogether.com/) (VSCode, IntelliJ, Eclipse)
    (paid)
  - [Teletype](https://teletype.atom.io/) (Atom)
  - [instant.nvim](https://github.com/jbyuki/instant.nvim) (Neovim)
  - [crdt.el](https://code.librehq.com/qhong/crdt.el) (Emacs)

## Technical Overview

pair-ls is implemented as a [Language
Server](https://microsoft.github.io/language-server-protocol/), so it receives
file open and edit information from any LSP client. It is a simple matter to
then expose that information to a web client, or to replicate it to a relay server.

Editor plugins add extensions on top of LSP to allow for enhanced features (see
differences in [Setup](#setup)) that aren't possible with the current state of
LSP.
