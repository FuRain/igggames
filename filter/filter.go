package filter

import (
	"fmt"
	"log"
	"strings"
	"regexp"
	"time"
	// "errors"
	"net/http"

	"github.com/gocolly/colly/v2"
    "github.com/gocolly/colly/v2/proxy"
)

type GameLink struct {
	LinkInfo string
	Link     string
}

const downWebsite string = "megaup.net"
const urlMatch string = `http[s]?://[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]`

func ProcessGamePage(url string, proxyURL string) []GameLink {
	var links []GameLink
	c := colly.NewCollector()

    if proxyURL != "" {
        // c.SetProxy(proxyURL)
        rp, err := proxy.RoundRobinProxySwitcher(proxyURL)
        if err != nil {
            log.Fatal(err)
        }
        c.SetProxyFunc(rp)
    }

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

func StartDownload(gameList []GameLink, proxyURL string) {
    // create downloader.
    var downCookie string = ""
    var downFileName string = ""
    var realGameDownLink string = ""

    downloader := colly.NewCollector(
        colly.AllowURLRevisit(),
        // colly.MaxBodySize(0),
    )

    if proxyURL != "" {
        rp, err := proxy.RoundRobinProxySwitcher(proxyURL)
        if err != nil {
            log.Fatal(err)
        }
        downloader.SetProxyFunc(rp)
    }

    downloader.SetRedirectHandler(func(req *http.Request, via []*http.Request) error {
        return http.ErrUseLastResponse
    })

    downloader.OnError(func(r *colly.Response, err error) {
        // get redirect response.
        realGameDownLink = r.Headers.Get("location")
    })

	downloader.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.104 Safari/537.36"
	downloader.OnRequest(func(r *colly.Request) {
        r.Headers.Set("accept-encoding", "gzip, deflate, br")
        r.Headers.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

        if "" != downCookie {
            r.Headers.Set("Set-Cookie", downCookie)
        }
		// fmt.Println("Downloader Visiting", r.URL.String())
	})

    downloader.OnResponse(func(r *colly.Response) {
        if "" == downCookie {
            downCookie = r.Headers.Get("Set-Cookie")
        }
        // fmt.Println(r.StatusCode, len(r.Body), r.Headers, downCookie)
    })

    // get real download link.
	c := colly.NewCollector()
    if proxyURL != "" {
        rp, err := proxy.RoundRobinProxySwitcher(proxyURL)
        if err != nil {
            log.Fatal(err)
        }
        c.SetProxyFunc(rp)
    }
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.104 Safari/537.36"
	c.OnHTML("body>section.section-padding>div>div", func(e *colly.HTMLElement) {

        fileName := e.ChildText("div[class=heading-1]")
        linkScript := e.ChildText("div>script:first-of-type")
        // fmt.Println(linkScript)

        r, err := regexp.Compile(urlMatch)
        if err != nil {
            log.Println(err)
            return
        }
        r.Longest()

        realDownLink := r.FindString(linkScript)
        if "" == realDownLink {
            log.Println("can not match real download link in the html.")
            return
        }

        // fmt.Println(fileName, realDownLink)
        downFileName = fileName

        // need sleep 5 second.
        err = downloader.Visit(realDownLink)
        if err != nil {
            log.Println(err)
            return
        }

        time.Sleep(time.Duration(5)*time.Second)
        if "" != downCookie {
            err = downloader.Visit(realDownLink)
            if err != nil && err.Error() != "Found" {
                log.Println(err)
                return
            }
        }
    })

	c.OnRequest(func(r *colly.Request) {
        r.Headers.Set("accept-encoding", "gzip, deflate, br")
        r.Headers.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		// fmt.Println("Visiting", r.URL.String())
	})

    for i, val := range gameList {
        log.Printf("start download, id: %d, link: %s\n", i+1, val.Link)

        // downCookie = ""
        downFileName = ""
        realGameDownLink = ""

	    c.Visit(val.Link)

        fmt.Println(downFileName, realGameDownLink)

        log.Println("over.")
        // return
    }
}
