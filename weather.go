package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	url2 "net/url"
	"sync"
	"time"

	"github.com/valyala/fastjson"
)

// 2 Global variables for the cache
var weatherCache = make(map[int64]weatherCacheEntry)
var weatherCacheMutex = &sync.Mutex{}

const weatherCacheValidTime = time.Minute * 15

// A very simple cache that can save a byte array (like a http response body)
type weatherCacheEntry struct {
	data    []byte
	created time.Time
}

// Make a request to the api and return the body of the response
// This function is really fast, since it uses a cache to store the http response from the API for 15minutes
//
func requestData(location int64) ([]byte, error) {
	// Lock the cache for the lifetime of this function
	weatherCacheMutex.Lock()
	defer weatherCacheMutex.Unlock()

	cache, ok := weatherCache[location]

	// Check if the cache is fresh enough (15 min)
	if ok == true && time.Since(cache.created) < weatherCacheValidTime {
		log.Printf("Loading %d weather from cache", location)
		// Cache is still warm, so we will use that
		return cache.data, nil
	}

	// Since there is no valid cache, lets download the data from the internet
	url := fmt.Sprintf("https://www.metaweather.com/api/location/%d/", location)
	resp, err := http.Get(url)
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
	weatherCache[location] = weatherCacheEntry{
		data:    body,
		created: time.Now(),
	}

	// return the data
	return body, nil
}

// Delete all cache entries who are no longer valid
func cleanWeatherCache() {
	cleanedCache := make(map[int64]weatherCacheEntry)

	// Lock the mutex of the weatherCache
	weatherCacheMutex.Lock()
	defer weatherCacheMutex.Unlock()

	// Copy all still valid cache entries to cleanedCache
	for k, v := range weatherCache {
		if time.Since(v.created) < weatherCacheValidTime {
			cleanedCache[k] = v
		}
	}

	// Update the weatherCache
	weatherCache = cleanedCache

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

// Load the content of the current weather
func loadWeather(location int64) (weather, error) {
	var w weather

	body, err := requestData(location)
	if err != nil {
		return w, err
	}

	var parser fastjson.Parser
	data, err := parser.ParseBytes(body)
	if err != nil {
		return w, err
	}

	w.MinTemp = data.GetFloat64("consolidated_weather", "0", "min_temp")
	w.MaxTemp = data.GetFloat64("consolidated_weather", "0", "max_temp")
	w.SunRise, err = time.Parse(time.RFC3339Nano, string(data.GetStringBytes("sun_rise")))
	w.SunSet, _ = time.Parse(time.RFC3339Nano, string(data.GetStringBytes("sun_set")))
	w.Weather = string(data.GetStringBytes("consolidated_weather", "0", "weather_state_name"))
	w.Emoji = weatherCodeToEmoji(string(data.GetStringBytes("consolidated_weather", "0", "weather_state_abbr")))
	w.Date, _ = time.Parse("2006-01-02", string(data.GetStringBytes("consolidated_weather", "0", "applicable_date")))
	w.City = string(data.GetStringBytes("title"))

	return w, nil
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
func loadForecast(location int64) (forecast, error) {
	var f forecast

	body, err := requestData(location)
	if err != nil {
		fmt.Println("ERROR: couldn't download the weather")
		return f, err
	}

	var parser fastjson.Parser
	data, err := parser.ParseBytes(body)
	if err != nil {
		fmt.Println("ERROR: couldn't download the weather")
		return f, err
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

	return f, nil
}

// Convert the forecast data into a Markdown message
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

// This function gets the Woeid needed for the weather request from a city name provided
func getLocations(city string) ([]int64, []string, error) {
	// Make a request
	url := fmt.Sprintf("https://www.metaweather.com//api/location/search/?query=%s", url2.PathEscape(city))
	resp, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	// Parse the response json
	var parser fastjson.Parser
	data, err := parser.ParseBytes(body)
	if err != nil {
		return nil, nil, err
	}

	// Return all the cities found
	var woeids []int64
	var locations []string
	entries, _ := data.Array()
	for _, entry := range entries {
		woeids = append(woeids, entry.GetInt64("woeid"))
		locations = append(locations, string(entry.GetStringBytes("title")))
	}

	return woeids, locations, nil
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
