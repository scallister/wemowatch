# wemowatch
Activates/deactivates Wemo switches when certain processes are active.

```
# The Wemo device "Meeting in Progress sign" will turn on whenever a BlueJeans process is detected on the system.
wemowatch --name "Meeting in Progress sign" --processes "BlueJeans"
```
![Meeting in Progress Sign](https://raw.githubusercontent.com/scallister/wemowatch/master/signnew.jpg)

## Install and use
Make sure you have Go installed and then run this to download/install:
```bash
go install github.com/scallister/wemowatch
```

## Buy a sign
This is the meeting in progress sign I use:
https://www.amazon.com/Meeting-Progress-Pollution-Prohibit-Distract/dp/B01HEDN0CA

## Buy a wemo mini plug
This is the wemo plug I use (but any wemo branded switch should work):
https://smile.amazon.com/Smart-Enabled-Google-Assistant-HomeKit/dp/B01NBI0A6R/

## Cron Job
To keep the script running at all times, I add it to my crontab like so. Wemowatch is designed to automatically exit if there is already a wemowatch process running so it is safe to run this once an hour.

```
0 * * * * /Users/scallister/go/bin/wemowatch --name "Meeting in Progress sign" -p "bluejeans,zoom.us" 2>> /var/log/wemowatch.err >> /var/log/wemowatch.std
```
