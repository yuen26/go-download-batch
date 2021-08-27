package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const MAX_BATCH_SIZE int = 4

func main() {
	// Read url, file, output flag
	var url, file, outputDir string
	var batchSize int
	flag.StringVar(&url, "url", "", "URL")
	flag.StringVar(&file, "file", "", "File path")
	flag.StringVar(&outputDir, "outputDir", "", "Output directory path")
	flag.IntVar(&batchSize, "batchSize", MAX_BATCH_SIZE, "Batch size")
	flag.Parse()
	if url == "" && file == "" {
		log.Fatal("url flag or file flag not present")
	}
	if outputDir == "" {
		log.Fatal("outputDir flag not present")
	}

	// Build URLs from each case
	var urls []string
	if url != "" {
		urls = buildUrlsFromUrlTemplate(url)
	} else {
		urls = buildUrlsFromFile(file)
	}

	// Download batch
	downloadBatch(urls, outputDir, batchSize)

	fmt.Print("Press any key to exit...")
	fmt.Scanln()
}

// Build URLs from URL template
// ----------------------------------------------------

func buildUrlsFromUrlTemplate(urlTemplate string) []string {
	// Read from and to flag
	var from, to int
	flag.IntVar(&from, "from", -1, "From index")
	flag.IntVar(&to, "to", -1, "To index")
	flag.Parse()
	if from < 0 {
		log.Fatal("Invalid from flag")
	}
	if to < 0 {
		log.Fatal("Invalid to flag")
	}

	// Parse URL template
	const beginChar = "{"
	const endChar = "}"
	beginCharIndex := strings.Index(urlTemplate, beginChar)
	endCharIndex := strings.Index(urlTemplate, endChar)
	if beginCharIndex == -1 || endCharIndex == -1 {
		log.Fatal("Invalid url flag")
	}
	left := (urlTemplate)[:beginCharIndex-1]
	right := (urlTemplate)[endCharIndex+1:]
	pattern := (urlTemplate)[beginCharIndex+1 : endCharIndex-1]

	// Build urls
	var urls []string
	for i := from; i <= to; i++ {
		number := fmt.Sprintf(pattern, i)
		urls = append(urls, left+number+right)
	}
	return urls
}

// Build URLs from file
// ----------------------------------------------------

func buildUrlsFromFile(filePath string) []string {
	urls, err := readLines(filePath)
	if err != nil {
		log.Fatal("Read file failed: ", err)
	}
	return urls
}

func readLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// Download
// ----------------------------------------------------

func downloadBatch(urls []string, outputDir string, batchSize int) {
	skip := 0
	urlCount := len(urls)
	batchCount := int(math.Ceil(float64(urlCount / batchSize)))
	fmt.Printf("URLs = %d, batches = %d\n", urlCount, batchCount)

	for i := 0; i <= batchCount; i++ {
		// Extract URLs
		lowerBound := skip
		upperBound := skip + batchSize
		if upperBound > urlCount {
			upperBound = urlCount
		}
		batchUrls := urls[lowerBound:upperBound]
		skip += batchSize

		// Channels
		processingErrorChan := make(chan error)
		processingDoneChan := make(chan int)
		processingErrors := make([]error, 0)
		go func() {
			for {
				select {
				case err := <-processingErrorChan:
					processingErrors = append(processingErrors, err)
				case <-processingDoneChan:
					close(processingErrorChan)
					close(processingDoneChan)
					return
				}
			}
		}()

		// Start processing batch
		var processingGroup sync.WaitGroup
		processingGroup.Add(len(batchUrls))
		for _, url := range batchUrls {
			go func(url string) {
				defer processingGroup.Done()
				err := downloadFile(url, outputDir)
				if err != nil {
					fmt.Printf("Download file %s failed: %s\n", url, err)
				} else {
					fmt.Printf("Download file %s successfully\n", url)
				}
			}(url)
		}
		processingGroup.Wait()

		// Finish processing batch
		processingDoneChan <- 0
		fmt.Printf("Download batch %d completed\n", i)
		if len(processingErrors) > 0 {
			for _, err := range processingErrors {
				fmt.Println(err)
			}
		}
	}
}

func downloadFile(url, outputDir string) error {
	// Get the response bytes from the url
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New(response.Status)
	}

	// Create a empty file
	filePath := outputDir + "/" + filepath.Base(url)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the bytes to the fiel
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}
