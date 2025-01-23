package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Weather struct {
	Location struct {
		Name      string `json:"name"`
		Region    string `json:"region"`
		Country   string `json:"country"`
		Localtime string `json:"localtime"`
	} `json:"location"`
	Current struct {
		TempC     float64 `json:"temp_c"`
		Condition struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
			Code int    `json:"code"`
		} `json:"condition"`
	} `json:"current"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <city> <api_key>")
		return
	}

	city := os.Args[1]
	apiKey := os.Args[2]

	parsedWeather := fetchWeather(city, apiKey)
	fmt.Println("City:", parsedWeather.Location.Name)
	fmt.Println("Region:", parsedWeather.Location.Region)
	fmt.Println("Country:", parsedWeather.Location.Country)
	fmt.Println("Local Time:", parsedWeather.Location.Localtime)
	fmt.Println("Temperature:", parsedWeather.Current.TempC, "°C")
	fmt.Println("Condition:", parsedWeather.Current.Condition.Text)
	fmt.Println("Icon:", parsedWeather.Current.Condition.Icon)
	fmt.Println("Code:", parsedWeather.Current.Condition.Code)
}

func fetchWeather(city string, apiKey string) Weather {
	response, err := http.Get("http://api.weatherapi.com/v1/current.json?key=" + apiKey + "&q=" + city + "&aqi=no")
	if err != nil {
		fmt.Println("Error fetching weather data:", err)
		return Weather{}
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return Weather{}
	}

	weather := parseWeather(string(body))

	return weather
}

func parseWeather(body string) Weather {
	weather := Weather{}
	json.Unmarshal([]byte(body), &weather)
	return weather
}
