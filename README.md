# mockapic-example-go

# How it works

Before reading this, please make sure to understand correctly the `Mockapic` [README.md](https://github.com/joakim-ribier/mockapic).

### How to test correctly your code when it calls external API services?

I created a light version of an application of currency converter to understand how to use `Mockapic`.

The application takes an amount in EUR and converts it to a specific currency by using an external API service `api.exchangeratesapi.io` which returns the current exchange rates.

```go
// example of the application, it is converting 10 EUR to USD = 11.08777 USD

& ./example --API_KEY {API_KEY} --currency USD --amount 10
~ A light application which converts an EUR amount to a specific currency.
(using https://api.exchangeratesapi.io/v1/latest)

10 EUR equals
11.08777 USD
2024-09-06 22:23:16 ~
```

So, I want to test the converter service but I don't want to call the external exchange rates API every time I run the test.

1. `CurrencyConverter` takes the URL as parameter to configure different behaviors between production and testing
2. Create several mocked requests
3. Assert your expected data in each test depend on the mocked request URL

### How I do

### Production mode

```go
// Main.go
amount, _, _ := NewCurrencyConverter(
	"https://api.exchangeratesapi.io/v1/latest", API_KEY).Convert("EUR", "USD", 10)

println(amount) // 11.08777
```

### Testing mode

Start the `Mockapic` server.

For the first one, I'm testing the nominal use case where the external API returns the exchange rates.

```bash
$ curl -X POST 'http://localhost:3333/v1/new' \
--header 'Content-Type: application/json' \
--data '{
	"status": 200,
	"contentType": "application/json",
	"charset": "UTF-8",
	"body": "{\"timestamp\":1725967696,\"rates\":{\"EUR\":1,\"GBP\":0.842772,\"KZT\":527.025041,\"USD\":1.103546}}",
	"path": "/exchangeratesapi"
}' | jq
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
								 Dload  Upload   Total   Spent    Left  Speed
100   408  100    48  100   360  13892   101k --:--:-- --:--:-- --:--:--  132k
{
  "id": "c1403100-3aa0-484f-8e0f-f2c1db80f371"
}
```

```go
expectedValue, _, _ := NewCurrencyConverter(
	"http://localhost:3333/v1/c1403100-3aa0-484f-8e0f-f2c1db80f371", ""). Convert("EUR", "USD", 10) // or with the {path}: http://localhost:3333/v1/exchangeratesapi

expectedValue == 11.03546 // the expected value is always equals to 11.03546, the rate does not change because the data is mocked
```

For the second one, I'm testing the case where the external API returns an internal server error.

```bash
$ curl -X POST 'http://localhost:3333/v1/new' \
--header 'Content-Type: application/json' \
--data '{
	"status": 500,
	"contentType": "application/json",
	"charset": "UTF-8"
}' | jq
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
								 Dload  Upload   Total   Spent    Left  Speed
100   408  100    48  100   360  13892   101k --:--:-- --:--:-- --:--:--  132k
{
  "id": "79090265-a1af-47ec-a177-88668582ce28"
}
```

```go
expectedValue, _, err := NewCurrencyConverter(
	"http://localhost:3333/v1/79090265-a1af-47ec-a177-88668582ce28", "").
	Convert("EUR", "USD", 10)

expectedValue == -1
err.Errors() == "Error: Internal Server Error from the API"
```

Full example in Go in the `howitworks` folder.

* [main](cmd/example/main.go)
* [service](internal/currency_converter.go)
* [test](internal/currency_converter_test.go)