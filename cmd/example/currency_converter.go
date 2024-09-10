package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/joakim-ribier/go-utils/pkg/httpsutil"
	"github.com/joakim-ribier/go-utils/pkg/jsonsutil"
)

type CurrencyRate struct {
	Timestamp int
	Rates     map[string]float32
}

type CurrencyConverter struct {
	url    string
	apiKey string
}

func NewCurrencyConverter(url, apiKey string) CurrencyConverter {
	return CurrencyConverter{
		url:    url,
		apiKey: apiKey,
	}
}

// Convert calls the external API service and converts the {amount} in a the new currency {to}
func (hiw CurrencyConverter) Convert(from, to string, amount int) (float32, string, error) {
	// fetch exchange rates from the external API service
	URL := fmt.Sprintf("%s?access_key=%s&base=%s", hiw.url, hiw.apiKey, from)
	httpRequest, err := httpsutil.NewHttpRequest(URL, "")
	if err != nil {
		return -1, "N/A", err
	}

	response, err := httpRequest.AsJson().Call()
	if err != nil {
		return -1, "N/A", err
	}

	if response.StatusCode != 200 {
		return -1, "N/A", errors.New(string(response.Body))
	}

	currencyRate, err := jsonsutil.Unmarshal[CurrencyRate](response.Body)
	if err != nil {
		return -1, "N/A", err
	}

	// convert the value
	if value, ok := currencyRate.Rates[to]; ok {
		t := time.Unix(int64(currencyRate.Timestamp), 0)
		return value * float32(amount), t.Format("2006-01-02 15:04:05"), nil
	} else {
		return -1, "N/A", fmt.Errorf("'%s' rate not found", to)
	}
}
