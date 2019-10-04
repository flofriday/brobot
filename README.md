# Brobot
My personal telegram bot

![screenshot.png]

## Setup 
* Create a 

## To start on a raspberry pi 1 B+
On the host:
```bash
env GOOS=linux GOARCH=arm GOARM=6 go build
scp ./brobot pi@192.168.8.148:/home/pi/brobot/brobot
```

### On the pi
```bash
sudo apt install screenfetch
mkdir $HOME/brobot
sudo nano /etc/systemd/system/brobot.service #Paste the content below
sudo systemctl daemon-reload
sudo systemctl start brobot
```

Content of brobot.service:
```
[Unit]
Description=brobot Service
After=network.target

[Service]
WorkingDirectory=/home/pi/brobot
Type=simple
User=pi
ExecStart=/home/pi/brobot/brobot
Restart=always

[Install]
WantedBy=multi-user.target
```
