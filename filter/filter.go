package filter

import (
	"fmt"
	"log"
	"strings"

	"github.com/gocolly/colly/v2"
)

type GameLink struct {
	LinkInfo string
	Link     string
}

const downWebsite string = "megaup.net"

func ProcessGamePage(url string, proxy string) []GameLink {
	var links []GameLink
	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.104 Safari/537.36"
	c.OnHTML("html", func(e *colly.HTMLElement) {
		linkID := e.ChildAttr("head>link[rel=shortlink]", "href")
		if "" == linkID {
			log.Println("can not found link id.")
			return
		}

		indexID := strings.LastIndex(linkID, "=")
		if -1 == indexID {
			log.Println("can not found game id.")
			return
		}
		indexID++
		// fmt.Println(linkID[indexID:])

		e.ForEach(fmt.Sprintf("#post-%s>div>p>a", linkID[indexID:]), func(_ int, el *colly.HTMLElement) {
			// fmt.Println(i, el.Text, el.Attr("href"))
			tempLink := el.Attr("href")
			if strings.Contains(tempLink, downWebsite) {
				tempIndex := strings.LastIndex(tempLink, "//")
				if -1 != tempIndex {
					links = append(links, GameLink{
						LinkInfo: el.Text,
						Link:     "https://" + tempLink[tempIndex+2:],
					})
				}
			}
		})
	})

	// Before making a request print "Visiting ..."
	// c.OnRequest(func(r *colly.Request) {
	// 	fmt.Println("Visiting", r.URL.String())
	// })

	c.Visit(url)
	return links
}

func StartDownload(gameList []GameLink) {
	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.104 Safari/537.36"
	c.OnHTML("body>section.section-padding>div>div", func(e *colly.HTMLElement) {
        // fmt.Println(e.Attr("href"), "dddd")
        // fmt.Println(e.Text, "dddd")
        fileName := e.ChildText("div[class=heading-1]")
        fmt.Println(fileName)

        linkScript := e.ChildText("div>script:first-of-type")
        fmt.Println(linkScript)
    })

	c.OnRequest(func(r *colly.Request) {
        r.Headers.Set("accept-encoding", "gzip, deflate, br")
        r.Headers.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		fmt.Println("Visiting", r.URL.String())
	})

    for i, val := range gameList {
        log.Printf("start download, id: %d, link: %s\n", i+1, val.Link)
	    c.Visit(val.Link)
        log.Println("over.")
        return
    }
}
