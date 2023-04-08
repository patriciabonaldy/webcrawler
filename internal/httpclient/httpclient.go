package httpclient

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/patriciabonaldy/webcrawler/internal/platform/logger"
)

type Client interface {
	Get(ctx context.Context, url string) (*Response, error)
}

type client struct {
	log     logger.Logger
	timeout time.Duration
}

func New() *client {
	return &client{timeout: 15 * time.Second, log: logger.New()}
}

// Get connects to the passed webpage and returns the ioReader
func (c *client) Get(ctx context.Context, url string) (*Response, error) {
	client := http.Client{
		Timeout: c.timeout,
	}

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.log.Info("error creating new request: ", err)
		return &Response{}, err
	}

	res, err := client.Do(request)
	if err != nil {
		c.log.Info("error connecting to website: ", err)
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
