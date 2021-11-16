# Install and go
* Copy amqp to /usr/local/bin
* Configure units
* Copy unit files to /etc/systemd/system/aglc.service for example
* systemctl enable aglc
* systemctl daemon-reload
* systemctl start aglc.service

# Logs
To view logs you can tail wherever syslog is writing to or use 
``` journalctl -u aglc -f ```
for example.

# Libraries 
I have modified the amqp lib to print the heartbeater events .  be aware that if you compile this yourself you won't
get those messages

# Running by hand
```amqp scheme username password hostname vhost port```

for example

```amqp amqp guest guest amqp.cuac.com /myvhost 5672```
