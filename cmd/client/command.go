package client

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/patriciabonaldy/webcrawler/internal"
	"github.com/patriciabonaldy/webcrawler/internal/platform/logger"
	"github.com/spf13/cobra"
)

// CobraFn function definion of run cobra command
type CobraFn func(cmd *cobra.Command, args []string)

const idFlag = "url"

var log logger.Logger

func Execute() {
	log = logger.New()

	root := initClientCmd()

	if err := root.Execute(); err != nil {
		log.ErrorFatal(err.Error())
	}
}

func initClientCmd() *cobra.Command {
	crawlerCmd := &cobra.Command{
		Use:   "url",
		Short: "Scrape the website",
		Run:   runCrawlerFn(),
	}

	crawlerCmd.Flags().StringP(idFlag, "u", "", "url of site to scrape")

	return crawlerCmd
}

func runCrawlerFn() CobraFn {
	return func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetString(idFlag)

		if url == "" {
			log.ErrorFatal("url param can not be empty")
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			sigint := make(chan os.Signal, 1)
			signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
			sing := <-sigint

			log.Infof("service is interrupted by signal %s\n", sing.String())
			cancel()
		}()

		crawler := internal.NewCrawler(log)
		crawler.Run(ctx, url)
	}
}
