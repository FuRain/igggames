package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

func DownloadShowBar(url string, filePath string) {
	var bar *progressbar.ProgressBar
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.104 Safari/537.36")

	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	f, _ := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	startT := time.Now().Unix()
	bar = progressbar.NewOptions(int(resp.ContentLength),
		// progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(70),
		progressbar.OptionSetDescription(filePath),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		// progressbar.OptionFullWidth(),
		// progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			endT := time.Now().Unix()
			fmt.Fprint(os.Stderr, fmt.Sprintf(" download spend time: %s\n", (time.Duration(endT-startT)*time.Second).String()))
		}),
	)
	io.Copy(io.MultiWriter(f, bar), resp.Body)
}
