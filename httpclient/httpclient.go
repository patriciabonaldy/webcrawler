package httpclient

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
)

type Client interface {
	Get(ctx context.Context, url string) (*Response, error)
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
func (c *client) Get(ctx context.Context, url string) (*Response, error) {
	client := http.Client{
		Timeout: c.timeout,
	}

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.log.Println("error creating new request: ", err)
		return &Response{}, err
	}

	res, err := client.Do(request)
	if err != nil {
		c.log.Println("error connecting to website: ", err)
		return &Response{}, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.log.Println("error closing the file ", err)
		}
	}(res.Body)

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

// Save writes response body to disk
func (r *Response) Save() error {
	folderName := strings.ReplaceAll(r.URL, "https://", "")

	if strings.HasSuffix(folderName, "/") {
		folderName = folderName[:len(folderName)-1]
	}

	if err := os.MkdirAll(folderName, 0750); err != nil {
		return err
	}

	fileName := getFileName(strings.TrimPrefix(r.URL, "/"))

	return os.WriteFile(fmt.Sprintf("%s/%s", folderName, fileName), r.Body, 0644)
}

func getFileName(fileName string) string {
	ext := filepath.Ext(fileName)

	cleanExt := sanitize.BaseName(ext)
	if cleanExt == "" {
		cleanExt = ".html"
	}

	return strings.Replace(fmt.Sprintf(
		"%s.%s",
		sanitize.BaseName(fileName[:len(fileName)-len(ext)]),
		cleanExt[1:],
	), "-", "_", -1)
}
