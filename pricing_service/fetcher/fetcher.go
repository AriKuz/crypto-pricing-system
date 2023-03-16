package fetcher

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	zero "github.com/rs/zerolog/log"
)

const (
	apiUrl = "https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest"
)

var (
	data   map[string]interface{}
	client http.Client
)

// Fetch get the data from coinmarketcap api and store in memory
func Fetch(coins *sync.Map, apiKey string) error {
	// Set the request parameters.
	params := url.Values{}
	params.Set("start", "1")
	params.Set("limit", "3")
	params.Set("convert", "USD")

	// Create a new HTTP request with the required headers and parameters.
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		zero.Error().Msgf("Error creating request:%v", err)
		return err
	}
	req.Header.Set("X-CMC_PRO_API_KEY", apiKey)
	req.Header.Set("Accept", "application/json")
	req.URL.RawQuery = params.Encode()

	// Send the request and read the response.
	resp, err := client.Do(req)
	if err != nil {
		zero.Error().Msgf("Error sending request:%v", err)
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		zero.Error().Msgf("Error reading response:%v", err)
		return err
	}

	// Parse the JSON response and print the results.
	if err := json.Unmarshal(body, &data); err != nil {
		zero.Error().Msgf("Error parsing JSON:%v", err)
		return err
	}

	for _, coin := range data["data"].([]interface{}) {
		symbol := coin.(map[string]interface{})["symbol"]
		price := coin.(map[string]interface{})["quote"].(map[string]interface{})["USD"].(map[string]interface{})["price"]
		coins.Store(symbol, price)
	}

	return nil
}

// CheckCoinMarket check market every 60 seconds
func CheckCoinMarket(coins *sync.Map, apiKey string) {
	for {
		if err := Fetch(coins, apiKey); err != nil {
			zero.Error().Msgf("CheckCoinMarket failed:%v", err)
		}
		time.Sleep(time.Minute)
	}
}
