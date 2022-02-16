# Port Forwarding

Perhaps the simplest way to share your local webpage via your remote server is
to forward your local ports. First, make sure that your server has `GatewayPorts yes` in the `sshd_config` file (and restart the service if you had to change
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
