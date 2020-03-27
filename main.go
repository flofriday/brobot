package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
	bolt "go.etcd.io/bbolt"
	"log"
	"os"
	"path"
)

var (
	DB     *bolt.DB
	DbPath = path.Join("data", "data.db")
)

func main() {
	// Load the token
	token := os.Getenv("TELEGRAM_TOKEN")

	// Check if there is a token
	if token == "" {
		log.Println("ERROR: Telegram Token missing")
		log.Println("You need to specify a token for telegram.")
		log.Println("example: TELEGRAM_TOKEN=t0k3n ./brobot")
		os.Exit(1)
	}

	// Register the bot at telegram
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Setup the bot
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Setup the database
	err = initDB(DbPath)
	if err != nil {
		log.Fatalf("Unable to initialize the Database: %s", err.Error())
	}
	defer DB.Close()

	// Start the background task
	go background(bot)

	// Handle the updates as the come in
	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		// Only handle message updates
		if update.Message == nil {
			continue
		}

		// Call the handle message concurrently so the execution of one command cannot block another one
		up := update
		go handleMessage(bot, &up)
	}
}

// This function must be called as a go routine because it is blocking
func background(bot *tgbotapi.BotAPI) {
	gocron.Every(1).Day().At("6:00").Do(func() { sendWeather(bot) })
	gocron.Every(1).Day().At("00:00").Do(cleanWeatherCache)
	<-gocron.Start()
}

func initDB(path string) error {
	// Open the Database
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return err
	}
	DB = db

	// Create all the needed buckets
	err = DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Users"))
		if err != nil {
			return fmt.Errorf("unable to create bucket: %s", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
