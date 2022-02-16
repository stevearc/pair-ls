# WebRTC

WebRTC is a standard for peer-to-peer communication over the internet. This
allows us to host a static site and connect that webpage directly to the
`pair-ls` LSP server without the need for an intermediate relay server.

The pair-ls static site relies on free
[STUN](https://en.wikipedia.org/wiki/STUN) and
[TURN](https://en.wikipedia.org/wiki/Traversal_Using_Relays_around_NAT) servers
powered by [Metered Video](https://www.metered.ca/tools/openrelay/) and backup
STUN servers by [Google](https://google.com). Thanks!

## Editor initiates

Your editor plugin will provide some command (consult the docs) to initiate a
WebRTC call. It should generate a url that looks like
`https://code.stevearc.com?t=V2h5IHdvdWxkIHlvdSBkZWNvZGUgdGhpcz8gV2hhdCB3ZXJlIHlvdSBsb29raW5nIGZvcj8`.
When your partner visits that URL they will receive a long base64 token in
return that they send to you. Again, your editor will have a way to input it,
and that should start the connection.

This process can be performed multiple times to share with multiple people. It
creates a new P2P connection each time.

Note that this method is only available via the editor plugins. There is no way
to initiate a WebRTC call using only the bare-bones LSP server. You can,
however, use the pattern below.

## Viewer initiates

The viewer goes to https://code.stevearc.com/ and is presented with a button.
Upon clicking the button, they should receive a long base64 token that they can
share with the editor. The editor will again have a mechanism to input it, and
they will receive a response token in return. The viewer then inputs the
response token on the static site and it should complete the connection.

If you do not have an editor plugin, there are two ways to pass the viewer token
in. One is via command line args `pair-ls lsp -call-token <token>`, and the
other is using the `callToken` key in the configuration file. The response token
will be sent back via the LSP notification `window/showMessage`.
