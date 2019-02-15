# wemowatch
Watches for processes and turns the specified Wemo device on whenever they are detected, and off when they are not detected.

```
# The Wemo device "Meeting in Progress sign" will turn on whenever a BlueJeans process is detected on the system.
wemowatch --name "Meeting in Progress sign" --processes "BlueJeans"
```
