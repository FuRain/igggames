package main

import (
	"flag"
	"log"
	"os"
	"fmt"
    "igggames/filter"
    "igggames/downloader"
)

func main() {
	var gamePage string
	var downProxy string
	// var threadCount int

	flag.StringVar(&gamePage, "url", "", "start download game page.")
	flag.StringVar(&downProxy, "proxy", "", "support http/https/socks5 download proxy.")
	// flag.IntVar(&threadCount, "thread", 2, "simultaneous download threads count.")
    verbose := flag.Bool("v", false, "print game download link and exit.")

	flag.Parse()

	if "" == gamePage {
		log.Println("game start download page cannot is null.")
		flag.PrintDefaults()
		os.Exit(1)
	}
    // log.Println(*verbose)

    gameLinks := filter.ProcessGamePage(gamePage, downProxy)
    if 0 == len(gameLinks) {
        log.Println("cannot get game download link on the game page.")
        return
    }

    if *verbose {
        for i, v := range gameLinks {
            fmt.Println(i+1, v.LinkInfo, v.Link)
        }
        return
    }

    for _, v := range gameLinks {
        realLink := filter.GetDownloadLink(v.Link, downProxy)
        downloader.DownloadShowBar(realLink.Link, "/home/bruce/rain/temp/my_download/" + realLink.LinkInfo)
    }
}
