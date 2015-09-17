Self-monitoring tool (sketch)
=============================

Usage
-----
Create own config file:
```shell
cp example.config.json config.json
```
Edit configuration:
```shell
vim config.json
```
Run:
```shell
go run *.go
```

Configuration
-------------
Example:
```javascript
{
    "interval": 15, // Update interval in seconds
    "notifyAtStart": false, // Send notification with offline services on start
	"processList": // Process list for monitoring
    [
        "acrypt",
		"capella",
		"docker"
    ],
    "logger": true, // Print status info in log
    "telegram": {
        "enable": true, // Enable Telegram notification
        "token": "", // Telegram Bot API Token
        "contactID": 0, // Contact ID (user, chat, etc.)
        "debug": false // Print debug info
    }
}
```
