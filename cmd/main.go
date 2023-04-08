package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/patriciabonaldy/webcrawler/internal"
	"github.com/patriciabonaldy/webcrawler/internal/platform/logger"
)

func main() {
	log := logger.New()
	log.PrintHeader()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		sing := <-sigint

		log.Infof("service is interrupted by signal %s", sing.String())
		cancel()
	}()

	crawler := internal.NewCrawler(log)
	crawler.Run(ctx, "https://llorllale.github.io/posts/golang-generics-first-look/")
}
