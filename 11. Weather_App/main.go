package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
)

const currentWeatherAPI = "openweatherapp"
const forecastAPI = "openweatherappforecast"

func main() {
	apiKey := "YOUR_API_KEY"
	cityList := []string{"서울"}

	units := "metric"

	if apiKey == "" {
		fmt.Println("API 키가 필요합니다. 코드에서 apiKey 변수를 설정해주세요.")
		os.Exit(1)
	}

	if len(cityList) == 0 {
		fmt.Println("도시 목록이 비어 있습니다. 코드에서 cityList 변수를 설정해주세요.")
		os.Exit(1)
	}

	for _, city := range cityList {
		city = strings.TrimSpace(city)
		if city == "" {
			continue
		}
		fmt.Printf("\n%s의 현재 날씨와 5일 예보:\n", city)
		printCurrentWeather(city, units, apiKey)
		printForecast(city, units, apiKey)
	}
}

func printCurrentWeather(city, units, apiKey string) {
	url := fmt.Sprintf("%s?q=%s&appid=%s&units=%s&lang=kr", currentWeatherAPI, city, apiKey, units)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("데이터를 가져오는 데 실패했습니다: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("API 요청 실패: %s\n", resp.Status)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("응답을 읽는 데 실패했습니다: %v\n", err)
		return
	}

	var data CurrentWeatherData
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("JSON 파싱에 실패했습니다: %v\n", err)
		return
	}

	unitSymbol := "°C"
	if units == "imperial" {
		unitSymbol = "°F"
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	fmt.Printf("%s: %s\n", cyan("날씨"), data.Weather[0].Description)
	fmt.Printf("%s: %s%s\n", yellow("온도"), fmt.Sprintf("%.2f", data.Main.Temp), unitSymbol)
	fmt.Printf("%s: %s%%\n", green("습도"), fmt.Sprintf("%d", data.Main.Humidity))
}

func printForecast(city, units, apiKey string) {
	url := fmt.Sprintf("%s?q=%s&appid=%s&units=%s&lang=kr", forecastAPI, city, apiKey, units)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("데이터를 가져오는 데 실패했습니다: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("API 요청 실패: %s\n", resp.Status)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("응답을 읽는 데 실패했습니다: %v\n", err)
		return
	}

	var data ForecastData
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("JSON 파싱에 실패했습니다: %v\n", err)
		return
	}

	unitSymbol := "°C"
	if units == "imperial" {
		unitSymbol = "°F"
	}

	fmt.Println("\n5일 예보:")
	for _, item := range data.List {
		fmt.Printf("%s - %s: %.2f%s, %s\n", item.DtTxt, item.Weather[0].Description, item.Main.Temp, unitSymbol, fmt.Sprintf("습도: %d%%", item.Main.Humidity))
	}
}

type CurrentWeatherData struct {
	Cod     int    `json:"cod"`
	Message string `json:"message"`
	Name    string `json:"name"`
	Main    struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
}

type ForecastData struct {
	Cod     string `json:"cod"`
	Message int    `json:"message"`
	Cnt     int    `json:"cnt"`
	List    []struct {
		Dt   int64 `json:"dt"`
		Main struct {
			Temp     float64 `json:"temp"`
			Humidity int     `json:"humidity"`
		} `json:"main"`
		Weather []struct {
			Description string `json:"description"`
		} `json:"weather"`
		DtTxt string `json:"dt_txt"`
	} `json:"list"`
}
