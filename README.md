# pair-ls

Pair-ls is a lightweight, editor-agnostic tool for remote pair-programming. It
allows you to easily share the files you are working on in a read-only manner.
Pair-ls is _not_ a collaborative editor. If you're wondering why you would use
pair-ls, read the [comparison](#comparison) section.

:warning: pair-ls is currently in **Alpha** status. Expect breaking API changes.
Backwards compatibility will be respected _after_ the first release.

- [Installation](#installation)
- [Setup](#setup)
- [Sharing](#sharing)
- [Configuration](#configuration)
- [Comparison](#comparison)
- [Alternatives](#alternatives)
- [Technical Overview](#technical-overview)

## Installation

You can download a binary from [the releases
page](https://github.com/stevearc/pair-ls/releases)

If you want to build from source, `git clone` this repo and run `yarn build`
(you will need to [install
yarn](https://classic.yarnpkg.com/lang/en/docs/install/) and [go build tools](https://go.dev/doc/install))

## Setup

If the option is available, it is recommended to use a plugin for your editor.
If your editor is not yet supported, you can still use pair-ls with minimal
configuration (also file an issue to ask for support). Using pair-ls without an
editor plugin will mean some degradation in the cursor tracking.

**Editor plugins:**

See the link for installation instructions

- Neovim: [pair-ls.nvim](https://github.com/stevearc/pair-ls.nvim)

### Generic Setup - any editor

pair-ls supports any editor with a [LSP
client](https://microsoft.github.io/language-server-protocol/).

Configure your LSP client to run this command as a server: `pair-ls lsp -port 8080`.
Now you can visit http://localhost:8080 to see a mirror of your code. For more
info on how to share this page, see [Sharing](#sharing).

## Sharing

Running `pair-ls lsp -port 8080` is an easy way to get started, but how can you
share this across the internet?

The quickest way with no setup required is to use [WebRTC
connections](docs/WEBRTC.md).

If you have access to a server with a public IP address that is reachable by
both parties, you have options that will be a bit more convenient and more
reliable. The most straightforward is [ssh port
forwarding](docs/PORT_FORWARDING.md), but you can also set up [a relay
server](docs/RELAY.md).

For completeness, you can also set up a [signal server](docs/SIGNAL.md), but
that has all the drawbacks of both WebRTC relay server, so it's not recommended.

### Password protection

If your pair-ls webpage is visible on the public internet, you should make sure
to enable password protection to prevent anyone on the internet from seeing your
code. Simply provide a password, either via the [config file](#configuration) or
with the environment variable `PAIR_WEB_PASS`. The webserver will now require
this password before allowing access. You should also make sure your site it
hosted over https so the password can't be trivially sniffed (see
[encryption](docs/RELAY.md#encryption)).

## Configuration

The configuration file can be found at `$XDG_CONFIG_HOME/pair-ls.toml`. Most
values can be specified on the command line instead, if you prefer (run
`pair-ls` for detailed help).

```toml
# Default log file is $XDG_CACHE_HOME/pair-ls.log
logFile = "/path/to/file.log"

# For the relay server. When false (the default) all files are cleared from the
# server when the last editor connection is closed.
relayPersist = false

# The static site hosting the WebRTC connection code
staticRTCSite = "https://code.stevearc.com/"

# The one-time WebRTC connection token generated from the static WebRTC site
# Editor plugins provide a better way to pass this in, so only use the option if
# your editor doesn't have a plugin.
callToken = ""

[server]
# If provided, will require password auth from web client
webPassword = "passw0rd"
# If provided, will require connecting pair-ls LSP to provide this password in
# the [client] section (only used for relay & signal servers)
lspPassword = "secur3"
# If provided, will secure all connections with TLS
certFile = "/path/to/cert.pem"
# If the private key is not in the certFile PEM, you can pass it in separately here
keyFile = "/path/to/cert.key.pem"
# If true, will require pair-ls LSP to provide a matching client cert.
# This is the certFile under the [client] section.
requireClientCert = false
# PEM file with one or more certs that pair-ls LSP can match
# (when requireClientCert = true; only used for relay & signal servers)
clientCAs = "/path/to/pool.pem"

[client]
# Provide this certificate to the relay/signal server when connecting
certFile = "/path/to/cert.pem"
# If the private key is not in the certFile PEM, you can pass it in separately here
keyFile = "/path/to/cert.key.pem"
# If the relay/signal server requires a password, supply it here
password = "secur3"
```

## Comparison

Pairing tools fall into roughly 3 categories: **screen sharing**, **web
editors**, and **editor plugins**.

**Screen sharing**: Easiest to use, with the worst functionality

- **Pros**:
  - Very easy, they're built into the tools you're already using to call your partners
  - You see exactly what the sharer is doing, across all windows and applications
- **Cons**:
  - Video artifacts can make text hard to read
  - Text is often too small unless the sharer increases the size dramatically
  - Viewer has no control over what they're looking at
  - Read-only sharing

**Web editor**: Easy to share, but only if you buy into their ecosystem

- **Pros**:
  - No installation required
  - Often have collaborative editing functionality
- **Cons**:
  - You have to use their editor
  - You have to use their entire editing ecosystem, since it's not simply working with files on your own computer

**Editor plugin**: Best experience as long as everyone's preferred editor is supported

- **Pros**:
  - You can use your own editor
  - Often have collaborative editing functionality
- **Cons**:
  - Requires installation
  - Your editor has to have a plugin available
  - Everyone has to be using the same editor
  - Only shares editor state, nothing in other windows

**Pair-ls**: Sacrifices features of editor plugins for broader support

- **Pros**:
  - You can use your own editor
- **Cons**:
  - Requires installation
  - Read-only sharing
  - Only shares files. You can't see open terminals or anything else the sharer is doing in the editor

## Alternatives

Most of these are paid apps/services, though many of those have a free tier.

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
LSP (e.g. tracking cursor movement).
