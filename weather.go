package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/valyala/fastjson"
)

// 2 Global variables for the cache
var weatherCache cache
var cacheMutex = &sync.Mutex{}

// A very simple cache that can save a byte array (like a http response body)
type cache struct {
	data    []byte
	created time.Time
}

type weather struct {
	MinTemp float64
	MaxTemp float64
	Weather string
	Emoji   string
	City    string
	Date    time.Time
	SunSet  time.Time
	SunRise time.Time
}

// Convert the weatherCode form the API into easy human readable emojis
func weatherCodeToEmoji(weatherCode string) string {
	switch weatherCode {
	case "sn":
		return "🌨"
	case "sl":
		return "🌨"
	case "h":
		return "🌨"
	case "t":
		return "⛈"
	case "hr":
		return "🌧"
	case "lr":
		return "🌧"
	case "s":
		return "🌦"
	case "hc":
		return "☁️"
	case "lc":
		return "⛅️"
	case "c":
		return "☀️"
	}

	return ""
}

// Make a request to the api and return the body of the response
// This function is really fast, since it uses a cache to store the http response from the API for 15minutes
//
func requestData() ([]byte, error) {
	// Lock the cache for the lifetime of this function
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Check if the cache is fresh enough (15 min)
	if time.Since(weatherCache.created) < time.Minute*15 {
		// Cache is still warm, so we will use that
		return weatherCache.data, nil
	}

	// Since there is no valid cache, lets download the data from the internet
	resp, err := http.Get("https://www.metaweather.com/api/location/551801/")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// Update the cache for the next time
	weatherCache = cache{
		data:    body,
		created: time.Now(),
	}

	// return the data
	return body, nil
}

// Load the content of the current weather
func (w *weather) load() {
	body, err := requestData()
	if err != nil {
		fmt.Println("ERROR: couldn't download the weather")
		return
	}

	var parser fastjson.Parser
	data, err := parser.ParseBytes(body)
	if err != nil {
		fmt.Println("ERROR: couldn't download the weather")
		return
	}

	w.MinTemp = data.GetFloat64("consolidated_weather", "0", "min_temp")
	w.MaxTemp = data.GetFloat64("consolidated_weather", "0", "max_temp")
	w.SunRise, err = time.Parse(time.RFC3339Nano, string(data.GetStringBytes("sun_rise")))
	w.SunSet, _ = time.Parse(time.RFC3339Nano, string(data.GetStringBytes("sun_set")))
	w.Weather = string(data.GetStringBytes("consolidated_weather", "0", "weather_state_name"))
	w.Emoji = weatherCodeToEmoji(string(data.GetStringBytes("consolidated_weather", "0", "weather_state_abbr")))
	w.Date, _ = time.Parse("2006-01-02", string(data.GetStringBytes("consolidated_weather", "0", "applicable_date")))
	w.City = string(data.GetStringBytes("title"))
}

// Convert the weather data into a Markdown message
func (w *weather) message() string {

	msg := fmt.Sprintf("*%v %v*\n%v %v\nMin: %v°C\nMax: %v°C\nDaylight: %s - %s",
		w.City,
		w.Date.Format("02.01.2006"),
		w.Weather,
		w.Emoji,
		int(math.Round(w.MinTemp)),
		int(math.Round(w.MaxTemp)),
		w.SunRise.Format("15:04"),
		w.SunSet.Format("15:04"),
	)

	return msg
}

type forecast struct {
	days []weather
}

// Load the forecast data
func (f *forecast) load() {
	body, err := requestData()
	if err != nil {
		fmt.Println("ERROR: couldn't download the weather")
		return
	}

	var parser fastjson.Parser
	data, err := parser.ParseBytes(body)
	if err != nil {
		fmt.Println("ERROR: couldn't download the weather")
		return
	}

	object := data.Get("consolidated_weather")
	objects, _ := object.Array()
	for _, d := range objects {
		var w weather

		w.MinTemp = d.GetFloat64("min_temp")
		w.MaxTemp = d.GetFloat64("max_temp")
		w.Weather = string(d.GetStringBytes("weather_state_name"))
		w.Emoji = weatherCodeToEmoji(string(d.GetStringBytes("weather_state_abbr")))
		w.Date, _ = time.Parse("2006-01-02", string(d.GetStringBytes("applicable_date")))
		w.City = string(data.GetStringBytes("title"))

		f.days = append(f.days, w)
	}
}

// Convert the forcast data into a Markdown message
func (f *forecast) message() string {

	msg := fmt.Sprintf("*%v Forecast*\n", f.days[0].City)

	for i, day := range f.days {
		dayText := ""
		if i == 0 {
			dayText = "Today"
		} else {
			dayText = day.Date.Weekday().String()[:3]
		}

		msg += fmt.Sprintf("*%v* %v %v %v°C | %v°C\n",
			dayText,
			day.Emoji,
			day.Weather,
			int(math.Round(day.MinTemp)),
			int(math.Round(day.MaxTemp)),
		)
	}

	return msg
}
