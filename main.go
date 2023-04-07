package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := log.Default()
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
		<-sigint

		log.Println("shutdown service")
		cancel()
	}()

	crawler := NewCrawler(log)
	crawler.Run(ctx, "https://learnenglish.britishcouncil.org/")
}
