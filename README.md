# Brobot
My personal telegram bot

![Screenshot](screenshot.png)

## Run the bot
```bash
go build
TELEGRAM_TOKEN=XXXX ./brobot
```
Replace the XXXX with your telegram token

## Run the bot with docker
```
docker build -t brobot-template .
docker run -e TELEGRAM_TOKEN=XXXX --rm --name brobot-container brobot-template
```
Replace the XXXX with your telegram token

## Deploy the bot
Look into `deploy.md` to see how I deploy the bot.


