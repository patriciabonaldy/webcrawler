package httpclient

import (
	"io"
	"log"
	"net/http"
	"time"
)

type Client interface {
	Get(url string) (*Response, error)
}

// Response is the representation of a HTTP response made by a Collector
type Response struct {
	URL string
	// StatusCode is the status code of the Response
	StatusCode int
	// Body is the content of the Response
	Body []byte
	// Headers contains the Response's HTTP headers
	Headers *http.Header
}

type client struct {
	log     *log.Logger
	timeout time.Duration
}

func New() *client {
	return &client{timeout: 15 * time.Second, log: log.Default()}
}

// Get connects to the passed webpage and returns the ioReader
func (c *client) Get(url string) (*Response, error) {
	client := http.Client{
		Timeout: c.timeout,
	}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.log.Println("error creating new request: ", err)
		return &Response{}, err
	}

	res, err := client.Do(request)
	if err != nil {
		c.log.Println("error connecting to website: ", err)
		return &Response{}, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return &Response{
		URL:        url,
		StatusCode: res.StatusCode,
		Body:       body,
		Headers:    &res.Header,
	}, nil
}
