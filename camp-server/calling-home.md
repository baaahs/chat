# Calling Home

The MQTT servers (primarily the camp server) maintains a connection to a "home" server as best it can. It does this by ssh'ing to a well known location `tompop.tomseago.com` and using ssh to tunnel port 10015 back to itself on port 22. By doing this an appropriately credentials external party can then ssh to that well known location and get all the way into the `camp-msg` machine regardless of where it might be in the wide world of networks.

This is configured by creating a user on the tompop machine. In this case the user is `tun-camp-msg` which has an ssh credential configured that matches the public key installed on the camp-msg machine.

The camp-msg machine is then configured via systemd to keep retrying this ssh connection.

On tompop the sshd_config file has to have the `GatewayPorts yes` option set so that the incoming ssh connection can bind to all addresses.

Then the following is set in /etc/systemd/system/ssh_tunnel.service

```
[Unit]
Description=SSH Tunnel to tompop.tomseago.com
After=network-online.target
Wants=network-online.target

[Service]
User=baaahs
Restart=always
RestartSec=5
ExecStart=/usr/bin/ssh -NT -o ServerAliveInterval=60 -o ExitOnForwardFailure=yes -i /home/baaahs/.ssh/id_tomseago.com_tunnel -R *:10015:127.0.0.1:22 tun-camp-msg@tompop.tomseago.com

[Install]
WantedBy=multi-user.target
```

Then use systemctl to load that and start it

    systemctl daemon-reload
    systemctl enable ssh_tunnel.service
    systemctl sstart ssh_tunnel.service

So now one can

     ssh -p 10015 baaahs@tompop.tomseago.com


