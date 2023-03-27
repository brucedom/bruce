# Home Assistant Examples

The home assistant example is rather rudimentary, but it does provide a more meaningful way on how home-assistant is run, 
which effectively is docker, but not by arbitrary run command instead it sets up a systemd service which will be used to 
start / stop and manage the service more appropriately over time.  In order to update home assistant you effectively now just
have to do a service restart of home-assistant.

## How do I update my home-assistant with this setup:

As mentioned previously it is quite simple the systemd file is set up to pull the latest on restart so just issue the command:
```bash
# Note: you may have to use sudo / privilege escalation with this command
/usr/bin/systemctl restart home-assistant
```

## Where is my config data

All of your config data will be stored under /data/home-assistant which is very convenient from a backup perspective
you can create a job that will back up that directory every so often and ship it off to a proper backup location to ensure you never 
lose your important config data.