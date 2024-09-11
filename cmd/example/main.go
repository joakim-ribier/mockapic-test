package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
	"github.com/joakim-ribier/go-utils/pkg/stringsutil"
	"github.com/joakim-ribier/mockapic-example-go/internal"
)

func main() {
	URL := "https://api.exchangeratesapi.io/v1/latest"
	API_KEY := ""
	from := "EUR"
	to := "USD"
	amount := 1

	// get application parameters
	args := slicesutil.ToMap(os.Args[1:])
	if arg, ok := args["--API_KEY"]; ok {
		API_KEY = arg
	} else {
		fmt.Printf("%v", errors.New("parameter --API_KEY is required"))
		return
	}
	if arg, ok := args["--currency"]; ok {
		to = arg
	}
	if arg, ok := args["--amount"]; ok {
		amount = stringsutil.Int(arg, 1)
	}

	// call the convert service with the URL as parameter
	value, time, err := internal.NewCurrencyConverter(URL, API_KEY).Convert(from, to, amount)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf(
			"~ A light application which converts an EUR amount to a specific currency.\n(using %s)\n\n%d %s equals\n%v %s\n%s ~\n",
			URL,
			amount, from,
			value, to, time)
	}
}
