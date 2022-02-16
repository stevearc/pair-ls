# Signal Server

A signal server hosts the sharing webpage, but connects clients using WebRTC.
The main difference between this and the [WebRTC solution](WEBRTC.md) is that
this requires the LSP server to connect to it like a [relay server](RELAY.md).
There is currently no advantage to using this over a relay server.

The sharing flow for using the signal server is:

- Editor connects and gets a URL with a unique path
- Anyone that visits that unique URL will automatically start a WebRTC connection

Compared to the WebRTC flow, this is much simplified as it doesn't require
copy/pasting multiple long token strings. The downside is that it requires a
dedicated server. If you have a dedicated server, you might as well use a relay
server instead since the connection there is more reliable than WebRTC.
