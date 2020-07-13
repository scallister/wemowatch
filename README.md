# wemowatch
Activates/deactivates Wemo switches when certain processes are active.

```
# The Wemo device "Meeting in Progress sign" will turn on whenever a BlueJeans process is detected on the system.
wemowatch --name "Meeting in Progress sign" --processes "BlueJeans"
```
![Meeting in Progress Sign](https://raw.githubusercontent.com/scallister/wemowatch/master/signnew.jpg)

## Cron Job
To keep the script running at all times, I add it to my crontab like so. Wemowatch is designed to automatically exit if there is already a wemowatch process running so it is safe to run this once an hour.

```
0 * * * * /Users/scallister/go/bin/wemowatch --name "Meeting in Progress sign" -p "bluejeans" -v 2>> /var/log/wemowatch.err >> /var/log/wemowatch.std
```
