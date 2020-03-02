package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
)

var (
	clients    []int64
	clientFile = "./data/clients.json"
)

func loadClients() error {
	jsonFile, err := os.Open(clientFile)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	_ = json.Unmarshal(byteValue, &clients)
	return nil
}

func saveClients() error {
	bytes, err := json.Marshal(clients)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(clientFile, bytes, 0777)
	if err != nil {
		return err
	}

	return nil
}

func addClient(id int64) error {
	for _, client := range clients {
		if client == id {
			return errors.New("Hawara, you are already subscribed.")
		}
	}

	clients = append(clients, id)
	err := saveClients()
	if err != nil {
		return err
	}

	return nil
}

func removeClient(id int64) error {
	for i, client := range clients {
		if client == id {
			clients = append(clients[:i], clients[i+1:]...)
			err := saveClients()
			if err != nil {
				log.Printf("Error: %s", err.Error())
				return err
			}

			return nil
		}
	}

	return errors.New("You cannot unsubscibe if you aren't subscibed")
}

func sendWeather(bot *tgbotapi.BotAPI) {
	// Load the weather
	var w weather
	w.load()
	text := w.message()

	// Send the weather to every client
	for _, client := range clients {
		msg := tgbotapi.NewMessage(client, text)
		msg.ParseMode = "Markdown"
		_, _ = bot.Send(msg)
	}
}

func background(bot *tgbotapi.BotAPI) {
	gocron.Every(1).Day().At("6:00").Do(func() { sendWeather(bot) })
	<-gocron.Start()
}

func createAnswer(input *tgbotapi.Message) string {
	var cmd = input.Command()

	if cmd == "weather" {
		var w weather
		w.load()
		return w.message()

	} else if cmd == "forecast" {
		var f forecast
		f.load()
		return f.message()

	} else if cmd == "subscribe" {
		err := addClient(input.Chat.ID)
		if err != nil {
			return err.Error()
		}
		return "You are now subscribed to the daily weather feed at 07:00."

	} else if cmd == "unsubscribe" {
		err := removeClient(input.Chat.ID)
		if err != nil {
			return err.Error()
		}
		return "You are now unsubscribed from the daily feed."

	} else if cmd == "help" || cmd == "start" {
		bytes, _ := ioutil.ReadFile("commands.txt")
		return string(bytes)

	} else if cmd == "botinfo" {
		msg := fmt.Sprintf("Subscribed users: %v", len(clients))
		return msg

	} else if cmd == "screenfetch" || cmd == "neofetch" {
		bytes, err := exec.Command("neofetch", "--stdout").Output()
		var output string
		if err != nil {
			return "😔 Sorry, but neofetch is not installed"
		}
		output += "*Neofetch:*\n```\n" + string(bytes) + "\n```"
		return output

	} else {
		return "Sry I don't understand that command. /help for more information."
	}
}

func main() {
	// Load the token
	token := os.Getenv("TELEGRAM_TOKEN")

	// Check if there is a token
	if token == "" {
		log.Println("ERROR: Telegram Token missing")
		log.Println("You need to specify a token for telegram.")
		log.Println("example: TELEGRAM_TOKEN=t0k3n ./brobot")
		return
	}

	// Register the bot at telegram
	bot, err := tgbotapi.NewBotAPI(string(token))
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	_ = loadClients()

	go background(bot)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		// TODO: this API is a mess, fix it
		// TODO: Also make this run in parallel, there is no need for this to be blocking
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, createAnswer(update.Message))
		msg.ParseMode = "Markdown"
		_, _ = bot.Send(msg)
	}
}
