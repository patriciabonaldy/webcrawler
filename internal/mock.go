package internal

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/patriciabonaldy/webcrawler/internal/httpclient"
)

type httpClientMock struct {
	urls map[string]string
}

func (h httpClientMock) Get(_ context.Context, url string) (*httpclient.Response, error) {
	path, ok := h.urls[url]
	if !ok {
		return nil, errors.New("unknown error")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	body, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var response = createMockResponse(url, body)

	return &response, nil
}

func createMockResponse(url string, body []byte) httpclient.Response {
	return httpclient.Response{
		URL: url,
		Headers: &http.Header{
			"Accept":           {"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
			"Accept-Language":  {"en-us,en;q=0.5"},
			"Accept-Encoding":  {"gzip,deflate"},
			"Accept-Charset":   {"ISO-8859-1,utf-8;q=0.7,*;q=0.7"},
			"Keep-Alive":       {"300"},
			"Proxy-Connection": {"keep-alive"},
			"Content-Length":   {"7"},
			"User-Agent":       {"Fake"},
			"Content-Type":     {"text/html; charset=UTF-8"},
		},
		Body:       body,
		StatusCode: 200}
}
