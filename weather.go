package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/valyala/fastjson"
)

type weather struct {
	MinTemp     float64
	MaxTemp     float64
	Weather     string
	WeatherCode string
	City        string
	Date        string
}

func (w *weather) load() {
	resp, err := http.Get("https://www.metaweather.com/api/location/551801/")
	if err != nil {
		fmt.Println("ERROR: couldn't download the weather")
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ERROR: couldn't load the response body")
		return
	}

	var parser fastjson.Parser
	data, err := parser.ParseBytes(body)

	w.MinTemp = data.GetFloat64("consolidated_weather", "0", "min_temp")
	w.MaxTemp = data.GetFloat64("consolidated_weather", "0", "max_temp")
	w.Weather = string(data.GetStringBytes("consolidated_weather", "0", "weather_state_name"))
	w.WeatherCode = string(data.GetStringBytes("consolidated_weather", "0", "weather_state_abbr"))
	w.Date = string(data.GetStringBytes("consolidated_weather", "0", "applicable_date"))
	w.City = string(data.GetStringBytes("title"))
}

func (w *weather) message() string {
	var emoji string
	switch w.WeatherCode {
	case "sn":
		emoji = "🌨"
	case "sl":
		emoji = "🌧🌨"
	case "h":
		emoji = "🌨"
	case "t":
		emoji = "⛈"
	case "hr":
		emoji = "🌧"
	case "hl":
		emoji = "🌧"
	case "s":
		emoji = "🌦"
	case "hc":
		emoji = "☁️"
	case "lc":
		emoji = "⛅️"
	case "c":
		emoji = "☀️"
	}

	msg := fmt.Sprintf("*%v %v*\nWeather: %v%v\nMin: %s°C\nMax: %s°C", w.City,
		w.Date,
		w.Weather,
		emoji,
		strconv.FormatFloat(w.MinTemp, 'f', 1, 32),
		strconv.FormatFloat(w.MaxTemp, 'f', 1, 32),
	)

	return msg
}
