package internal

import (
	"context"
	"testing"

	"github.com/patriciabonaldy/webcrawler/internal/httpclient"
	"github.com/patriciabonaldy/webcrawler/internal/platform/logger"
)

var clientMock httpclient.Client = &httpClientMock{
	urls: map[string]string{
		"http://localhost:8080/":              "../test_data/index.html",
		"http://localhost:8080/engage.html":   "../test_data/engage.html",
		"http://localhost:8080/products.html": "../test_data/products.html",
	},
}

func Test_crawler_Run(t *testing.T) {
	tests := []struct {
		name   string
		client httpclient.Client
		url    string
	}{
		{
			name:   "success",
			client: clientMock,
			url:    "http://localhost:8080/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &crawler{
				client:  tt.client,
				crawled: make(map[string]bool, 0),
				log:     logger.New(),
			}
			c.Run(context.Background(), tt.url)
		})
	}
}
