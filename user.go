package main

import (
	"encoding/json"
	bolt "go.etcd.io/bbolt"
	"log"
	"strconv"
)

const (
	defaultLocation = 551801 // Vienna as this is my location
)

type user struct {
	TelegramID        int64
	WeatherSubscribed bool
	Location          int64
}

func (u *user) update() error {
	jsonBytes, err := json.Marshal(u)
	if err != nil {
		return err
	}

	err = DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Users"))
		err := b.Put([]byte(strconv.FormatInt(u.TelegramID, 10)), jsonBytes)
		return err
	})

	return err

}

func (u *user) setWeatherSubscribed(subscribed bool) error {
	u.WeatherSubscribed = subscribed
	return u.update()
}

func (u *user) setLocation(location int64) error {
	u.Location = location
	return u.update()
}

func loadUser(telegramID int64) (user, error) {
	var bytes []byte

	_ = DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Users"))
		v := b.Get([]byte(strconv.FormatInt(telegramID, 10)))
		bytes = make([]byte, len(v))
		copy(bytes, v)
		return nil
	})

	var u user
	err := json.Unmarshal(bytes, &u)
	if err != nil {
		return createUser(telegramID)
	}
	return u, nil
}

func createUser(telegramID int64) (user, error) {
	log.Printf("Creating user: [%d]", telegramID)

	u := user{
		TelegramID:        telegramID,
		WeatherSubscribed: false,
		Location:          defaultLocation,
	}

	err := u.update()
	return u, err
}

func loadSubscribedUsers() []user {
	var users []user

	DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Users"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			// Convert the json to a user
			var u user
			err := json.Unmarshal(v, &u)
			if err != nil {
				continue
			}

			// Check if the user is subscribed and if so add it to the list of subbscribed users
			if u.WeatherSubscribed == true {
				users = append(users, u)
			}
		}

		return nil
	})

	return users
}

// This function returns the number of users and subscribed users
func loadUserStatistics() (int, int) {
	total := 0
	subscribed := 0

	DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Users"))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			// Convert the json to a user
			var u user
			err := json.Unmarshal(v, &u)
			if err != nil {
				log.Println(err.Error())
				continue
			}

			total++
			// Check if the user is subscribed and if so add it to the list of subbscribed users
			if u.WeatherSubscribed == true {
				subscribed++
			}
		}

		return nil
	})

	return total, subscribed
}
