package main

func main() {
	crawler := NewCrawler()
	crawler.Run("https://www.irs.gov/")
}
