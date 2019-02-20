# wemowatch
Activates/deactivates Wemo switches when certain processes are active.

```
# The Wemo device "Meeting in Progress sign" will turn on whenever a BlueJeans process is detected on the system.
wemowatch --name "Meeting in Progress sign" --processes "BlueJeans"
```
![Meeting in Progress Sign](https://raw.githubusercontent.com/scallister/wemowatch/master/newsign.jpg)
