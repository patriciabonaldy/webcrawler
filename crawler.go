package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/patriciabonaldy/webcrawler/httpclient"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

type crawler struct {
	client  httpclient.Client
	crawled map[string]bool
	log     *log.Logger
}

func NewCrawler() *crawler {
	return &crawler{client: httpclient.New(), crawled: make(map[string]bool), log: log.Default()}
}

func (c *crawler) Run(url string) {
	linksChannel := make(chan string)
	err := c.process(url, linksChannel)
	if err != nil {
		c.log.Println("error processing url")
	}

	var wg sync.WaitGroup
	for link := range linksChannel {
		wg.Add(1)

		go func(wg *sync.WaitGroup, lk string) {
			defer wg.Done()

			errP := c.process(lk, linksChannel)
			if errP != nil {
				c.log.Printf("error processing sub url %s, error: %v\n", lk, errP)
				return
			}
		}(&wg, link)
	}

	wg.Wait()
}

func (c *crawler) process(url string, linksChannel chan string) error {
	resp, err := c.getContent(url)
	if err != nil {
		return err
	}

	return c.readBody(resp, linksChannel)
}

func (c *crawler) getContent(url string) (*httpclient.Response, error) {
	resp, err := c.client.Get(url)
	if err != nil || resp.StatusCode >= 500 {
		return nil, err
	}

	return resp, nil
}

func (c *crawler) readBody(resp *httpclient.Response, linksChannel chan string) error {
	err := c.encodeBytes(resp)
	if err != nil {
		return err
	}

	tokenizer := html.NewTokenizer(bytes.NewReader(resp.Body))
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			c.log.Println("invalid data")
			return fmt.Errorf("invalid data")
		}

		token := tokenizer.Token()

		if tokenType == html.StartTagToken && token.DataAtom.String() == "a" {
			link := linksFromToken(token, resp.URL)
			if strings.EqualFold(resp.URL, link) {
				continue
			}

			if link != "" {
				go func(l string) {
					c.log.Printf("Processing url %s", l)
					linksChannel <- l
				}(link)
			}
		}
	}
	c.log.Printf("url %s processed", resp.URL)

	return nil
}

func (c *crawler) encodeBytes(resp *httpclient.Response) error {
	contentType := strings.ToLower(resp.Headers.Get("Content-Type"))
	if strings.Contains(contentType, "image/") ||
		strings.Contains(contentType, "video/") ||
		strings.Contains(contentType, "audio/") ||
		strings.Contains(contentType, "font/") {
		// skip these types.
		return nil
	}

	r, err := charset.NewReader(bytes.NewReader(resp.Body), contentType)
	if err != nil {
		return err
	}

	tmpBody, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	resp.Body = tmpBody

	return nil
}

func linksFromToken(token html.Token, url string) string {
	for _, attr := range token.Attr {
		if attr.Key == "href" {
			link := attr.Val
			tl := parseURL(url, link)
			if tl == "" {
				break
			}

			return tl
		}
	}
	return ""
}

func parseURL(url string, link string) string {
	base := strings.TrimSuffix(url, "/")
	if strings.Contains(link, base) {
		return link
	}

	if strings.HasPrefix(link, "/") {
		return fmt.Sprintf("%s%s", base, link)
	}

	return ""
}
