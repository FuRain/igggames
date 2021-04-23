package filter

import (
	"fmt"
	"log"
	"strings"
	"regexp"
	"time"
	// "errors"
	"net/http"
	"strconv"

	"github.com/gocolly/colly/v2"
    "github.com/gocolly/colly/v2/proxy"
    "gopkg.in/olebedev/go-duktape.v3"
)

type GameLink struct {
	LinkInfo string
	Link     string
}

const UserAgent string = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.104 Safari/537.36"
const getCloudDiskUrl string = "http://bluemediafiles.com/get-url.php?url="
const downWebsite string = "MegaUp.net"
const urlMatch string = `http[s]?://[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]`
const encryptKeyMatch string = `Goroi_n_Create_Button\("(.*)"\)`

const decodeKey string = `function Goroi_n_Create_Button(d_roi){var cidkeg=["yHoix","__proto__","96407aJAasq","tJoOb","pWYll","lJfvZ","4|3|2|5|0|1","nut","block","debu","23812ntUbsU","constructor","{}.constructor(\x22return\x20this\x22)(\x20)","XKdVs","gger","while\x20(true)\x20{}","call","log","jOayE","style","nNLOk","trace","wxuso","test","apply","QwBUX","anti-adblock","^([^\x20]+(\x20+[^\x20]+)+)+[^\x20]}","chain","rfluZ","input","return\x20/\x22\x20+\x20this\x20+\x20\x22/","stateObject","EkRBD","bwVnO","length","split","abs","fvFGZ","jsZpD","action","error","BAeof","asvWx","uMAKJ","console","getElementById","866870KUWOzv","yfEGR","xFNjv","none","hONEH","pVfmY","138632ZxXhGx","zoAMP","string","toString","display","1278242qernuz","src","MbkfH","return\x20(function()\x20",".item","idModal","setAttribute","cssText","function\x20*\x5c(\x20*\x5c)","counter","506819XIXBmf","\x5c+\x5c+\x20*(?:[a-zA-Z_$][0-9a-zA-Z_$]*)","29BiYpko","exception","MjLCF","pnZPC","qZcOk","1280661XEDEAk","OgUOl","warn","Joavc","init","AqQjE","jfHoo","HMkab","position:\x20fixed;z-index:\x201;padding-top:\x20180px;left:\x200;top:\x200;width:\x20100%;height:\x20100%;overflow:\x20auto;background-color:\x20rgb(0,0,0);background-color:\x20rgba(0,0,0,0.4);"];var cidkeh=function(a,b){a=a-450;var c=cidkeg[a];return c};var cidkez=cidkeh;(function(a,b){var p=cidkeh;while(!![]){try{var c=parseInt(p(460))+parseInt(p(471))+-parseInt(p(522))+-parseInt(p(479))*parseInt(p(455))+-parseInt(p(527))+parseInt(p(453))+parseInt(p(516));if(c===b){break}else{a["push"](a["shift"]())}}catch(d){a["push"](a["shift"]())}}}(cidkeg,643335));var cidken="";for(i=d_roi[cidkez(504)]/2-5;i>=0;i=i-2){cidken+=d_roi[i]}for(i=d_roi[cidkez(504)]/2+4;i<d_roi[cidkez(504)];i=i+2){cidken+=d_roi[i]}return cidken}`
const callDecodeFunc string = `Goroi_n_Create_Button("%s");`


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

	c.UserAgent = UserAgent
	c.OnHTML("html", func(e *colly.HTMLElement) {
		linkID := e.ChildAttr("head>link[rel=shortlink]", "href")
		if "" == linkID {
			log.Println("can not found link id.")
			return
		}

		// fmt.Println(linkID)

		indexID := strings.LastIndex(linkID, "=")
		if -1 == indexID {
			log.Println("can not found game id.")
			return
		}
		indexID++
		// fmt.Println(linkID[indexID:])

		// e.ForEach(fmt.Sprintf("#post-%s>div>p>a", linkID[indexID:]), func(_ int, el *colly.HTMLElement) {
		e.ForEach(fmt.Sprintf("#post-%s>div>p", linkID[indexID:]), func(_ int, el *colly.HTMLElement) {

            downSiteName := el.ChildText("b[class=uk-heading-bullet]")
            if strings.Contains(downSiteName, downWebsite) {
                tempLink := el.ChildAttrs("p>a", "href")
                
                for k, v := range tempLink {
                    links = append(links, GameLink{
                        LinkInfo: downWebsite + ":Part" + strconv.Itoa(k+1),
                        Link:     v,
                    })
                }

                // fmt.Println("downSiteWeb", downSiteName, tempLink)
            }

			// fmt.Println(i, el.Text, el.Attr("href"))
			// tempLink := el.Attr("href")
            // fmt.Println(tempLink, "dddddd")
			// if strings.Contains(tempLink, downWebsite) {
			// 	tempIndex := strings.LastIndex(tempLink, "//")
			// 	if -1 != tempIndex {
			// 		links = append(links, GameLink{
			// 			LinkInfo: el.Text,
			// 			Link:     "https://" + tempLink[tempIndex+2:],
			// 		})
			// 	}
			// }
		})
	})

	// Before making a request print "Visiting ..."
	// c.OnRequest(func(r *colly.Request) {
	// 	fmt.Println("Visiting", r.URL.String())
	// })

	c.Visit(url)

    for k, v := range links {
        tempLink := getWebSiteLink(v.Link, proxyURL)
        links[k].Link = tempLink
        // fmt.Println(tempLink)
    }

	return links
}

func GetDownloadLink(gameLink string, proxyURL string) GameLink {
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
        r.Headers.Set("Host", downWebsite)
        // r.Headers.Set("accept-encoding", "gzip, deflate, br")
        r.Headers.Set("accept", "*/*")
        // r.Headers.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

        if "" != downCookie {
            r.Headers.Set("Cookie", downCookie)
        }
		// fmt.Println("Downloader Visiting", r.URL.String(), r.Headers)
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

        time.Sleep(time.Duration(6)*time.Second)
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

	c.Visit(gameLink)
    // fmt.Println(downFileName, realGameDownLink)
    return GameLink {
        LinkInfo : downFileName,
        Link : realGameDownLink,
    }
}

// because igggames encrypts game url, need decrypts it. 
func getWebSiteLink(igggamesLink string, proxyURL string) string {
	var downLinks string
	c := colly.NewCollector()
    redirectGet := colly.NewCollector()

    if proxyURL != "" {
        // c.SetProxy(proxyURL)
        rp, err := proxy.RoundRobinProxySwitcher(proxyURL)
        if err != nil {
            log.Fatal(err)
        }
        c.SetProxyFunc(rp)
        redirectGet.SetProxyFunc(rp)
    }

	c.UserAgent = UserAgent
    redirectGet.UserAgent = UserAgent
	c.OnHTML("body>script:first-of-type", func(e *colly.HTMLElement) {
        // fmt.Println(e.Text)
        r, err := regexp.Compile(encryptKeyMatch)
        if err != nil {
            log.Println(err)
            return
        }
        r.Longest()

        realDownLink := r.FindStringSubmatch(e.Text)
        if len(realDownLink) < 2 {
            log.Println("not match encrypt key!")
            return
        }

        ctx := duktape.New()
        ctx.PevalString(decodeKey + fmt.Sprintf(callDecodeFunc, realDownLink[1]))
        result := ctx.GetString(-1)
        ctx.Pop()
        // fmt.Println("result is:", result)
        // To prevent memory leaks, don't forget to clean up after
        // yourself when you're done using a context.

        if result != "" {
            redirectGet.Visit(getCloudDiskUrl + result)
        }

        ctx.DestroyHeap()
    })

    redirectGet.SetRedirectHandler(func(req *http.Request, via []*http.Request) error {
        return http.ErrUseLastResponse
    })

    redirectGet.OnError(func(r *colly.Response, err error) {
        // get redirect response.
        downLinks = r.Headers.Get("location")
    })

	c.Visit(igggamesLink)

    return downLinks
}
