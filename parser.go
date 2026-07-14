package main

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

type Job struct {
	Title        string
	Company      string
	URL          string
	Date         string
	Salary       string
	Requirements []string
}

func IsTargetURL(url string) bool {
	return false
}

func GetURL(htmlString string) {

	doc, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					fmt.Println("------------------------")
					fmt.Printf("URL: %s\n", a.Val)
				}
			}
		}
	}
}
