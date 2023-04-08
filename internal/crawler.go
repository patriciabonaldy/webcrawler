package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"

	"github.com/patriciabonaldy/webcrawler/internal/httpclient"
	"github.com/patriciabonaldy/webcrawler/internal/platform/logger"
)

type crawler struct {
	mux      sync.Mutex
	client   httpclient.Client
	crawled  map[string]bool
	log      logger.Logger
	saveSite bool
}

var (
	errInvalidURL  = fmt.Errorf("invalid url")
	errInvalidData = fmt.Errorf("invalid data")
	errGettingURL  = fmt.Errorf("error getting url")
)

func NewCrawler(log logger.Logger) *crawler {
	return &crawler{client: httpclient.New(), crawled: make(map[string]bool), log: log, saveSite: true}
}

func (c *crawler) Run(ctx context.Context, url string) {
	c.log.PrintHeader()
	linksChannel := make(chan string)

	err := c.processURL(ctx, url, linksChannel)
	if err != nil {
		c.log.Infof("error processing url: %s\n", url)
	}

	// we registered the url parent
	c.registerURL(url)

	var wg sync.WaitGroup
	go c.process(ctx, linksChannel, &wg)
	wg.Wait()
}

func (c *crawler) process(ctx context.Context, lkChn chan string, wg *sync.WaitGroup) {
	defer func() {
		if r := recover(); r != nil {
			c.log.Info("finish proccess")
		}
	}()

	for { //nolint:wsl
		wg.Add(1)
		select {
		case link := <-lkChn:
			func(wg *sync.WaitGroup, lk string) { //nolint:wsl
				defer wg.Done()

				errP := c.processURL(ctx, lk, lkChn)
				if errP != nil {
					c.log.Infof("error processing sub url %s, error: %v\n", lk, errP)
					return
				}
			}(wg, link)
		case <-ctx.Done():
			return
		}
	}
}

func (c *crawler) processURL(ctx context.Context, url string, linksChannel chan string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	resp, err := c.getContent(ctx, url)
	if err != nil {
		return err
	}

	err = c.readBody(resp, linksChannel)
	if err != nil {
		return err
	}

	return c.downloadPage(resp)
}

func (c *crawler) downloadPage(resp *httpclient.Response) error {
	c.log.Infof("downloading site: %s\n", resp.URL)
	if !c.saveSite {
		return nil
	}

	return resp.Save()
}

func (c *crawler) getContent(ctx context.Context, url string) (*httpclient.Response, error) {
	resp, err := c.client.Get(ctx, url)
	if err != nil || resp.StatusCode >= 400 {
		return nil, errGettingURL
	}

	return resp, nil
}

func (c *crawler) registerURL(url string) {
	defer c.mux.Unlock()

	c.mux.Lock()
	c.crawled[url] = true
}

func (c *crawler) existsURL(url string) bool {
	defer c.mux.Unlock()

	c.mux.Lock()
	_, exists := c.crawled[url]

	return exists
}

func (c *crawler) readBody(resp *httpclient.Response, linksChannel chan string) error {
	err := c.encodeBytes(resp)
	if err != nil {
		return err
	}

	tokenizer := html.NewTokenizer(bytes.NewReader(resp.Body))
	for { //nolint:wsl
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			break
		}

		token := tokenizer.Token()

		if tokenType == html.StartTagToken && token.DataAtom.String() == "a" {
			link := c.linksFromToken(token, resp.URL)
			if strings.EqualFold(resp.URL, link) {
				continue
			}

			if link != "" && !c.existsURL(link) {
				c.registerURL(link)
				c.log.Infof("processing link %s\n", link)

				go func(l string) {
					linksChannel <- l
				}(link)
			}
		}
	}

	return nil
}

func (c *crawler) encodeBytes(resp *httpclient.Response) error {
	contentType := strings.ToLower(resp.Headers.Get("Content-Type"))
	if strings.Contains(contentType, "image/") ||
		strings.Contains(contentType, "video/") ||
		strings.Contains(contentType, "audio/") ||
		strings.Contains(contentType, "font/") {
		// skip these types.
		return errInvalidURL
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

func (c *crawler) linksFromToken(token html.Token, url string) string {
	for _, attr := range token.Attr {
		if attr.Key == "href" {
			link := attr.Val
			if !validateLink(link) {
				c.log.Infof("skipping link %s\n", link)
				continue
			}

			tl := c.parseURL(url, link)
			if tl == "" {
				break
			}

			c.log.Infof("link found %s", link)

			return tl
		}
	}

	return "" //nolint:wsl
}

func validateLink(link string) bool {
	if strings.Contains(link, ".pdf") || link == "#search" || link == "#signin" || link == "/" {
		return false
	}

	return true
}

func (c *crawler) parseURL(url string, link string) string {
	base := strings.TrimSuffix(url, "/")
	if strings.Contains(link, base) {
		return link
	}

	if strings.HasPrefix(link, "/") {
		return fmt.Sprintf("%s%s", base, link)
	}

	c.log.Infof("skipping link %s\n", link)

	return ""
}
