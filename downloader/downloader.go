package downloader

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
)

type downloadInfo struct {
	Url      string
	FilePath string
	FileName string
}

type downloadBar struct {
	syncWG   sync.WaitGroup
	mpbObj   *mpb.Progress
	jobFlag  chan bool
	chanData chan downloadInfo
}

func New() *downloadBar {
	db := downloadBar{}
	db.jobFlag = make(chan bool)
	db.chanData = make(chan downloadInfo, 10)
	db.mpbObj = mpb.New(
		mpb.WithWidth(70),
		mpb.WithRefreshRate(180*time.Millisecond),
		mpb.WithWaitGroup(&db.syncWG),
	)

	return &db
}

func (DB *downloadBar) AddJob(url string, filePath string, fileName string) {
	DB.chanData <- downloadInfo{
		Url:      url,
		FilePath: filePath,
		FileName: fileName,
	}
}

func (DB *downloadBar) jobFinish() {
	close(DB.chanData)
}

const TEMP_SUFFIX string = ".temp"

func (DB *downloadBar) DownloadFile(threadCount int) {
	// DB.chanData = make(chan downloadInfo, 10)

	for i := 0; i < threadCount; i++ {
		DB.syncWG.Add(1)
		go func() {
			defer DB.syncWG.Done()
			for val := range DB.chanData {
				if !strings.HasSuffix(val.FilePath, "/") {
					val.FilePath += "/"
				}

				tempFilePath := val.FilePath + val.FileName + TEMP_SUFFIX
				realFilePath := val.FilePath + val.FileName

				if _, err := os.Stat(realFilePath); err == nil {
					// log.Printf("%s already download finish!\n", val.FileName)
                    bar := DB.mpbObj.AddBar(1,
                        mpb.PrependDecorators(
                            decor.Name(val.FileName),
                            decor.Percentage(decor.WCSyncSpace),
                        ),
                        mpb.AppendDecorators(
                            decor.EwmaETA(decor.ET_STYLE_GO, 90),
                            decor.EwmaSpeed(decor.UnitKiB, " % .2f", 60),
                        ),
                    )
                    bar.Increment()

					continue
				}

				req, err := http.NewRequest("GET", val.Url, nil)
				if err != nil {
					log.Println(err)
					continue
				}
				req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.104 Safari/537.36")
				// req.Header.Add("Accept", "*/*")

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					log.Println(err)
					continue
				}

				defer resp.Body.Close()

				f, err := os.OpenFile(tempFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil {
					log.Println(err)
					continue
				}
				defer f.Close()
				// log.Println(resp.ContentLength, "len")

				bar := DB.mpbObj.AddBar(resp.ContentLength,
					mpb.PrependDecorators(
						decor.Name(val.FileName),
						decor.Percentage(decor.WCSyncSpace),
					),
					mpb.AppendDecorators(
						decor.EwmaETA(decor.ET_STYLE_GO, 90),
						decor.EwmaSpeed(decor.UnitKiB, " % .2f", 60),
					),
				)

				proxyReader := bar.ProxyReader(resp.Body)
				defer proxyReader.Close()

				// copy from proxyReader, ignoring errors
				byteLen, err := io.Copy(f, proxyReader)
				if err != nil {
					// TODO: http redownload process.
					log.Printf("%s download occur error, cause: [%s], already download length: [%d]\n", val.FileName, err.Error(), byteLen)
					os.Remove(tempFilePath)
					continue
				}
				os.Rename(tempFilePath, realFilePath)
			}

		}()
	}
	DB.mpbObj.Wait()
	DB.jobFlag <- true
}

func (DB *downloadBar) WaitExit() {
	DB.jobFinish()
	<-DB.jobFlag
}
