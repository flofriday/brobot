// This package should get the hours I worked on this week, to show me how much I need to work. The hours for my
// hobby-startup are stored in a google sheet. So this bot has to download that google sheet, find my name and add all
// the hours I worked this week
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

var (
	spreadSheetID = "1J-Ny04bnMOfilT6SLwjvIGZ-eVl1dXW57xS_mL2H7Yg"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file googletoken.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "googletoken.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func timeMessage(name string) (message string) {
	b, err := ioutil.ReadFile("googlecredentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved googletoken.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// Prints the names and majors of students in a sample spreadsheet:
	// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	readRange := "Antworten!A2:C"
	resp, err := srv.Spreadsheets.Values.Get(spreadSheetID, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		return "No data found."
	}

	// Parse the data
	totalHours := 0.0
	weekHours := 0.0
	for _, row := range resp.Values {
		name := fmt.Sprintf("%s", row[1])
		timeRaw := fmt.Sprintf("%s", row[0])
		durationRaw := fmt.Sprintf("%s", row[2])

		// Filter only my edits
		if (name != "Florian Freitag") {
			continue
		}

		// Get the duration
		tmp := strings.Split(durationRaw, ":")
		duration, err := time.ParseDuration(tmp[0] + "h" + tmp[1] + "m")
		if err != nil {
			continue
		}

		totalHours += duration.Hours()

		// Check if it was from this week
		workTime, err := time.Parse("02.01.2006 15:04:05", timeRaw)
		if err != nil {

		}

		nowTime := time.Now()
		y1, w1 := workTime.ISOWeek()
		y2, w2 := nowTime.ISOWeek()
		if y1 == y2 && w1 == w2 {
			totalHours += duration.Hours()
		}
	}

	// Craft the message
	return fmt.Sprintf("Total hours: %.1f\nWeek hours %.1f", totalHours, weekHours)
}
