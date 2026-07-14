package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	res, err := http.Get("https://github.com/zrazhd/WebCrawler")
	if err != nil {
		log.Printf("Error with gettin HTTP: %s\n", err)
	}
	defer res.Body.Close()

	htmlString, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error with reading res body: %s\n", err)
	}
	// fmt.Println(string(htmlString))

	GetURL(string(htmlString))
}
