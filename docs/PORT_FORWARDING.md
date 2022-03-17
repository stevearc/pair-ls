# Port Forwarding

Perhaps the simplest way to share your local webpage via your remote server is
to forward your local ports.

## Using ngrok

[ngrok](https://ngrok.com/) is a service that allows you to forward ports to
a public server. It is a paid service, but they have a free tier. The rough
steps to follow are:

- [Install ngrok](https://ngrok.com/download)
- [Create an account](https://dashboard.ngrok.com/signup)
- Verify your email address (you should receive an email)
- [Connect your account](https://dashboard.ngrok.com/get-started/setup) by running `ngrok authtoken <token>`
- `ngrok http 8080` (or whatever the local port is)
- Go to http://localhost:4040/ to inspect the tunnels and get the public link

## Using SSH

First, make sure that your server has `GatewayPorts yes` in the `sshd_config` file (and restart the service if you had to change
it).

Then from your local machine, run:

```
ssh -R 80:localhost:8080 my.server.com
```

This will forward port 80 on my.server.com to your localhost port 8080. Change
the ports and hostname as needed.

For more on port forwarding, see https://www.ssh.com/academy/ssh/tunneling/example#remote-forwarding

Note that you will still need to set up TLS certificates somewhere. Either by
keeping the certs on your local machine and enforcing TLS there, or by
terminating the TLS on the remote server using something like NGINX.
