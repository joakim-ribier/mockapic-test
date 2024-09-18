package internal

import (
	"fmt"
	"log"
	"net"
	"os"
	"testing"

	"github.com/joakim-ribier/go-utils/pkg/httpsutil"
	"github.com/joakim-ribier/go-utils/pkg/jsonsutil"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
)

var exposeHost string = "0.0.0.0:3333"

func TestMain(m *testing.M) {
	if os.Getenv("ENV_MODE") == "CI" {
		// on Github action use directly the services container
		// without dockertest (to have more control on the steps)
		os.Exit(m.Run())
	} else {
		// uses a sensible default on windows (tcp/http) and linux/osx (socket)
		pool, err := dockertest.NewPool("")
		if err != nil {
			log.Fatalf("Could not construct pool: %s", err)
		}

		// uses pool to try to connect to Docker
		err = pool.Client.Ping()
		if err != nil {
			log.Fatalf("Could not connect to Docker: %s", err)
		}

		resource, err := pool.RunWithOptions(&dockertest.RunOptions{
			Repository:   "joakimribier/mockapic",
			Tag:          "latest",
			Env:          []string{"MOCKAPIC_PORT=3333"},
			ExposedPorts: []string{"3333"},
		}, func(config *docker.HostConfig) {
			// set AutoRemove to true so that stopped container goes away by itself
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		})
		if err != nil {
			log.Fatalf("Could not start resource: %s", err)
		}

		resource.Expire(30) // hard kill the container in 3 minutes (180 Seconds)
		exposeHost = net.JoinHostPort("0.0.0.0", resource.GetPort("3333/tcp"))

		if err := pool.Retry(func() error {
			req, err := httpsutil.NewHttpRequest(fmt.Sprintf("http://%s/", exposeHost), "")
			if err != nil {
				return err
			}
			_, err = req.Timeout("150ms").Call()
			return err
		}); err != nil {
			log.Fatalf("Could not connect to mockapic server: %s", err)
		}

		code := m.Run()

		// cannot defer this because os.Exit doesn't care for defer
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}

		os.Exit(code)
	}
}

// TestConvertWithRightData calls CurrencyConverter.Convert(string, string, int),
// checking for a valid return value.
func TestConvertWithRightData(t *testing.T) {
	// simulate the external API with a right data
	uuid := createANewMockedRequest(t, exposeHost, 200, `{"timestamp":1725647345,"base":"EUR","rates":{"USD":1.108469}}`)

	// build the mocked URL with the UUID result and
	// mock the call of https://api.exchangeratesapi.io/v1/latest
	mockedURL := fmt.Sprintf("http://%s/v1/%s", exposeHost, uuid)

	// call the service with the mocked URL instead of the external API service
	r, _, err := NewCurrencyConverter(mockedURL, "API_KEY").Convert("EUR", "USD", 10)

	// assert the same result each time to run the test because the data is mocked
	if r != 11.08469 || err != nil {
		t.Errorf(`result: {%v} but expected: {%s}`, r, "1.108469")
	}
}

// TestConvertWith500ErrorFromAPI calls CurrencyConverter.Convert(string, string, int),
// checking for a valid return value.
func TestConvertWith500ErrorFromAPI(t *testing.T) {
	// simulate the external API which returns a 504 - Intenal Server Error
	uuid := createANewMockedRequest(t, exposeHost, 500, `{"error":"the service is down...."}`)

	// build the mocked URL with the UUID result and
	// mock the call of https://api.exchangeratesapi.io/v1/latest
	mockedURL := fmt.Sprintf("http://%s/v1/%s", exposeHost, uuid)

	// call the service with the mocked URL instead of the external API service
	r, _, err := NewCurrencyConverter(mockedURL, "API_KEY").Convert("EUR", "USD", 10)

	// assert the same result each time to run the test because the data is mocked
	if err.Error() != `{"error":"the service is down...."}` {
		t.Errorf(`result: {%v} but expected: {%s}`, r, "error")
	}
}

// TestConvertWithBadAPIResponse calls CurrencyConverter.Convert(string, string, int),
// checking for a valid return value.
func TestConvertWithBadAPIResponse(t *testing.T) {
	// simulate the external API which returns data but USD rate is missing
	uuid := createANewMockedRequest(t, exposeHost, 200, `{"timestamp":1725647345,"base":"EUR","rates":{"GBP":1.108469}}`)

	// build the mocked URL with the UUID result and
	// mock the call of https://api.exchangeratesapi.io/v1/latest
	mockedURL := fmt.Sprintf("http://%s/v1/%s", exposeHost, uuid)

	// call the service with the mocked URL instead of the external API service
	r, _, err := NewCurrencyConverter(mockedURL, "API_KEY").Convert("EUR", "USD", 10)

	// assert the same result each time to run the test because the data is mocked
	if err.Error() != `'USD' rate not found` {
		t.Errorf(`result: {%v} but expected: {%s}`, r, "error")
	}
}

// createANewMockedRequest creates a new mocked request on the 'Mockapic' server
// and returns the UUID of the new request
func createANewMockedRequest(t *testing.T, hostAndPort string, status int, body string) string {
	httpRequest, err := httpsutil.NewHttpRequest(fmt.Sprintf(
		"http://%s/v1/new?status=%d&contentType=application/json&charset=UTF-8", hostAndPort, status), body)
	if err != nil {
		log.Fatalf("Could not create a http server: %s", err)
	}

	httpResponse, err := httpRequest.AsJson().Call()
	if err != nil {
		t.Fatalf("Could not build a http request struct: %s", err)
	}

	values, _ := jsonsutil.Unmarshal[map[string]interface{}](httpResponse.Body)
	return values["uuid"].(string)
}

type MockedRequest struct {
	Status      int
	ContentType string
	Charset     string
	Body        string
}
