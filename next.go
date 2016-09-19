package main

import(
	"github.com/SurgeNews/SurgeServer/scrapper"
	"time"
) 

func main() {
	scrap := scrapper.NewClient()
	scrap.Request(0)
	time.Sleep(5)
}