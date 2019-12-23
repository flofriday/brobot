package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/valyala/fastjson"
)

type weather struct {
	MinTemp     float64
	MaxTemp     float64
	Weather     string
	Emoji 		string
	City        string
	Date        string
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
func requestData() ([]byte, error) {
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
	w.Weather = string(data.GetStringBytes("consolidated_weather", "0", "weather_state_name"))
	w.Emoji = weatherCodeToEmoji(string(data.GetStringBytes("consolidated_weather", "0", "weather_state_abbr")))
	w.Date = string(data.GetStringBytes("consolidated_weather", "0", "applicable_date"))
	w.City = string(data.GetStringBytes("title"))
}

// Convert the weather data into a Markdown message
func (w *weather) message() string {

	msg := fmt.Sprintf("*%v %v*\n%v %v\nMin: %s°C\nMax: %s°C",
		w.City,
		w.Date,
		w.Weather,
		w.Emoji,
		strconv.FormatFloat(w.MinTemp, 'f', 1, 32),
		strconv.FormatFloat(w.MaxTemp, 'f', 1, 32),
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
		w.Weather = string(d.GetStringBytes( "weather_state_name"))
		w.Emoji = weatherCodeToEmoji(string(d.GetStringBytes( "weather_state_abbr")))
		w.Date = string(d.GetStringBytes( "applicable_date"))
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
			dayText = "today"
		} else {
			t, _ := time.Parse("2006-01-02", day.Date)
			dayText = t.Weekday().String()[:3]
		}

		msg += fmt.Sprintf("*%v* %v %v %v°C *-* %v°C\n",
			dayText,
			day.Emoji,
			day.Weather,
			strconv.FormatFloat(day.MinTemp, 'f', 1, 32),
			strconv.FormatFloat(day.MaxTemp, 'f', 1, 32),
		)
	}

	return msg
}