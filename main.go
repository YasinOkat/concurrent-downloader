package main

import (
	"flag"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cheggaaa/pb/v3"
)

func main() {
	outputDir := flag.String("o", ".", "Output directory")
	maxGoroutines := flag.Int("n", 5, "Maximum number of concurrent downloads")
	flag.Parse()
	urls := flag.Args()

	if len(urls) == 0 {
		fmt.Println("Please provide at least one URL to download.")
		return
	}

	if err := os.MkdirAll(*outputDir, os.ModePerm); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	semaphore := make(chan struct{}, *maxGoroutines)
	var wg sync.WaitGroup

	bars := make([]*pb.ProgressBar, len(urls))

	for i := range urls {
		bar := pb.New(0)
		bar.Set("prefix", "Downloading... ")
		bars[i] = bar
	}

	pool, err := pb.StartPool(bars...)
	if err != nil {
		fmt.Printf("Error starting progress bar pool: %v\n", err)
		return
	}
	defer pool.Stop()

	for i, url := range urls {
		semaphore <- struct{}{}
		wg.Add(1)
		go func(u string, bar *pb.ProgressBar) {
			defer func() {
				<-semaphore
				wg.Done()
			}()
			err := downloadFile(u, *outputDir, bar)
			if err != nil {
				fmt.Printf("Error downloading %s: %v\n", u, err)
				bar.Finish()
			}
		}(url, bars[i])
	}

	wg.Wait()
	fmt.Println("All downloads completed.")
}

func downloadFile(url string, outputDir string, bar *pb.ProgressBar) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Downloader/1.0)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: %s", url, resp.Status)
	}

	size := resp.ContentLength
	if size <= 0 {
		size = 0
	}

	filename := path.Base(resp.Request.URL.Path)
	if filename == "" || filename == "/" || filename == "." {
		cd := resp.Header.Get("Content-Disposition")
		if cd != "" {
			if _, params, err := mime.ParseMediaType(cd); err == nil {
				if params["filename"] != "" {
					filename = params["filename"]
				}
			}
		}

		if filename == "" || filename == "/" || filename == "." {
			filename = "downloaded_file"
		}
	} else if strings.Contains(filename, "?") {
		filename = strings.Split(filename, "?")[0]
	}

	bar.Set("prefix", filename+" ")

	bar.SetTotal(size)

	filePath := filepath.Join(outputDir, filename)
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	reader := bar.NewProxyReader(resp.Body)

	_, err = io.Copy(out, reader)
	if err != nil {
		return err
	}

	bar.Finish()
	return nil
}
