package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
)

var buildDate = "<__unknown__>"

func sendWeather(bot *tgbotapi.BotAPI) {
	// Get all users that are subscribed to the weather
	users := loadSubscribedUsers()

	// Send the weather to every client
	for _, user := range users {
		weather, err := loadWeather(user.Location)

		// Decide what to send, based on the result of loadWeather
		message := ""
		if err != nil {
			message = "Unable to load the weather for your location"
		} else {
			message = weather.message()
		}

		// Send the telegram message
		msg := tgbotapi.NewMessage(user.TelegramID, message)
		msg.ParseMode = "Markdown"
		_, _ = bot.Send(msg)
	}
}

func handleMessage(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {

	// Log the user and incoming text
	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	// Decide which function should handle the command
	cmd := strings.ToLower(update.Message.Command())
	switch cmd {
	case "weather":
		weatherCmd(bot, update)
	case "forecast":
		forecastCmd(bot, update)
	case "setlocation":
		setLocationCmd(bot, update)
	case "subscribe":
		subscribeCmd(bot, update)
	case "unsubscribe":
		unsubscribeCmd(bot, update)
	case "deleteme":
		deleteMeCmd(bot, update)
	case "help":
		helpCmd(bot, update)
	case "start":
		helpCmd(bot, update)
	case "botinfo":
		botinfoCmd(bot, update)
	case "privacy":
		privacyCmd(bot, update)
	case "screenfetch":
		screenfetchCmd(bot, update)
	default:
		message := "😅 Sorry, I didn't understand that.\nYou can type /help to see what I can understand."
		sendMessage(bot, update, message)
	}
}

func weatherCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	// Load the user
	u, err := loadUser(update.Message.Chat.ID)
	if err != nil {
		sendMessage(bot, update, "Internal Error - unable to load the user\n"+err.Error())
		return
	}

	// Load the weather
	w, err := loadWeather(u.Location)
	if err != nil {
		sendMessage(bot, update, "Unable to download the weather\n"+err.Error())
		return
	}

	// Send the message
	sendMessage(bot, update, w.message())
}

func forecastCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	// Load the user
	u, err := loadUser(update.Message.Chat.ID)
	if err != nil {
		sendMessage(bot, update, "Internal Error - unable to load the user\n"+err.Error())
		return
	}

	// Load the forecast
	f, err := loadForecast(u.Location)
	if err != nil {
		sendMessage(bot, update, "Unable to download the forecast\n"+err.Error())
		return
	}

	// Send the message
	sendMessage(bot, update, f.message())
}

func setLocationCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	// Get the query
	query := update.Message.CommandArguments()
	if strings.TrimSpace(query) == "" {
		sendMessage(bot, update, "😅 You need to tell me at which location you want to be.\n Example: `/setLocation Vienna`")
		return
	}

	// Load the user
	u, err := loadUser(update.Message.Chat.ID)
	if err != nil {
		sendMessage(bot, update, "Internal Error - unable to load the user\n"+err.Error())
		return
	}

	// Load the locations matching the search query
	locations, titles, err := getLocations(query)
	if err != nil || len(locations) <= 0 {
		sendMessage(bot, update, "😔 Sorry I couldn't find the city you are looking for.\nMaybe try something else ?")
		return
	}

	// Check if there matches more than one location
	if len(locations) > 1 {
		cities := strings.Join(titles, "\n")
		sendMessage(bot, update,
			"😔 Sorry, you need to be more specific.\nI found the following cities matching your description:\n"+cities)
		return
	}

	// Update the user in the db
	err = u.setLocation(locations[0])
	if err != nil {
		sendMessage(bot, update, "Internal Error - unable to update the user\n"+err.Error())
		return
	}

	// Send a confirmation message
	message := fmt.Sprintf("😊 *Success*\nYour location is now set to: `%s`", titles[0])
	sendMessage(bot, update, message)
}

func subscribeCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	// Load the user
	u, err := loadUser(update.Message.Chat.ID)
	if err != nil {
		sendMessage(bot, update, "Internal Error - unable to load the user\n"+err.Error())
		return
	}

	// Tell the user he is already subscribed
	if u.WeatherSubscribed {
		sendMessage(bot, update, "Hawara, you are already subscribed 😂")
		return
	}

	// Update the user
	err = u.setWeatherSubscribed(true)
	if err != nil {
		sendMessage(bot, update, "Internal Error - unable to update the user\n"+err.Error())
		return
	}

	// Send a confirmation text
	message := "🥳 *Subscribed* 🎉\nYou will now receive the weather update daily at 6am CET."
	sendMessage(bot, update, message)
}

func unsubscribeCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	// Load the user
	u, err := loadUser(update.Message.Chat.ID)
	if err != nil {
		sendMessage(bot, update, "Internal Error - unable to load the user\n"+err.Error())
		return
	}

	// Tell the user he is already subscribed
	if !u.WeatherSubscribed {
		sendMessage(bot, update, "Hawara, you aren't even subscribed 😂")
		return
	}

	// Update the user
	err = u.setWeatherSubscribed(false)
	if err != nil {
		sendMessage(bot, update, "Internal Error - unable to update the user\n"+err.Error())
		return
	}

	// Send a confirmation text
	message := "🥺 *Unsubscribed* \nYou will no longer get the daily weather update."
	sendMessage(bot, update, message)
}

func deleteMeCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	// Try to delete the user
	err := deleteUser(update.Message.Chat.ID)
	if err != nil {
		sendMessage(bot, update, "Internal Error - unable to update the user\n"+err.Error())
		return
	}

	message := "*Deleted 🗑*\nAll your user data is now deleted, however the bot will create new data, " +
		"if you use one of the following commands:\n" +
		"/weather\n/forecast\n/setlocation\n/subscribe\n/unsubscribe"
	sendMessage(bot, update, message)
}

func helpCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	commands, _ := ioutil.ReadFile("commands.txt")
	message := fmt.Sprintf("*Here is a list of things I can do 😊:*\n%s\n\nI am activly developed by [flofriday](https://github.com/flofriday) as opensource software on [GitHub](https://github.com/flofriday/brobot) and [GitLab](https://gitlab.com/flofriday/brobot).", string(commands))
	sendMessage(bot, update, message)
}

func privacyCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	message := "*Privacy*\n" +
		"Unfortunately, this bot saves some user data. However, I try to keep the data as small as possible." +
		"At the moment the bot saves exactly three datapoints with every user:\n" +
		"1.	Telegram id: This is used so that the bot knows who is talking to it. " +
		"However, neither I (the developer) nor a potential hacker can use this to contact or harass you.\n" +
		"2.	The location you selected for your weather update. " +
		"By default, this is set to Vienna, but you can change it with the /setLocation command.\n" +
		"3.	Whether or not you are subscribed. " +
		"This is used to determine if you would like a daily weather update at 6am CET.\n\n" +
		"Moreover, I promise you that I will never sell or license this data to anyone, " +
		"or try to gain a financial advantage from it."

	sendMessage(bot, update, message)
}

func botinfoCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	total, subscribed, locations := loadUserStatistics()

	message := fmt.Sprintf("*Statistics*\nUsers: %d\nSubscribed: %d\nLocations: %d\nBot compiled at: %s", total, subscribed, locations, buildDate)
	sendMessage(bot, update, message)
}

func screenfetchCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	// Execute neofetch
	bytes, err := exec.Command("neofetch", "--stdout").Output()

	// Check if the command executed successfully and create the correct text
	var output string
	if err != nil {
		output = "😔 Sorry, but neofetch is not installed"
	} else {
		output = "*Neofetch:*\n```\n" + string(bytes) + "\n```"
	}

	// Send the message
	sendMessage(bot, update, output)
}

func sendMessage(bot *tgbotapi.BotAPI, update *tgbotapi.Update, text string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true
	_, _ = bot.Send(msg)
}
