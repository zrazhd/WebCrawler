package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func (wc *WebCrawler) Start() {
	for i := 1; i <= wc.WorkersCount; i++ {
		wc.wg.Add(1)
		go wc.Work(i)
	}
}

func (tracker *URLTracker) Visit(url string) bool {
	tracker.mx.Lock()
	defer tracker.mx.Unlock()
	if tracker.visited[url] {
		return true
	}
	tracker.visited[url] = true
	return false
}

func (wc *WebCrawler) Work(id int) {
	defer wc.wg.Done()
	for {
		select {
		case <-wc.ctx.Done():
			log.Println("Work done due to context")
			return
		case url, ok := <-wc.URLs:
			if !ok {
				fmt.Println("Канал URLS закрыт")
				return
			}

			wc.ProccessHTML(getHTML(url), url, id)
			// wc.SaveJob()
		}
	}
}

func (wc *WebCrawler) Stop() {
	close(wc.URLs)
	wc.wg.Wait()
	close(wc.Jobs)
}

func (wc *WebCrawler) Finish() {
	wc.cancel()
}

func isTargetURL(url string) bool {
	if strings.Contains(url, "golang") {
		return true
	}
	if strings.Contains(url, "intern") {
		return true
	}
	if strings.Contains(url, "job") {
		return true
	}
	if strings.Contains(url, "junior") {
		return true
	}
	if strings.Contains(url, "career") {
		return true
	}
	return false
}

func parseHTMLToText(htmlString string) string {
	doc, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		fmt.Printf("Error parse html: %s", err)
	}

	res := ""
	for n := range doc.Descendants() {
		if n.Type == html.TextNode {
			res += n.Data
		}
	}

	res = strings.ReplaceAll(res, "<br>", " ")
	return res
}

func (wc *WebCrawler) ProccessHTML(htmlString string, urlString string, id int) {

	doc, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		fmt.Printf("Error parse html: %s", err)
	}

	base, err := url.Parse(urlString)
	if err != nil {
		log.Printf("Error parse url: %s\n", err)
	}
	var job Job
	job.URL = urlString
	rawJSON := ""

	for n := range doc.Descendants() {
		if n.Type == html.ElementNode {

			if n.DataAtom == atom.A {
				for _, a := range n.Attr {
					if a.Key == "href" && isTargetURL(a.Val) {
						path, err := url.Parse(a.Val)
						if err != nil {
							log.Printf("Error href: %s\n", err)
						}
						fullURL := base.ResolveReference(path)
						exists := wc.Map.Visit(fullURL.String())
						if !exists {
							wc.URLs <- fullURL.String()
							fmt.Printf("Goroutine with id: %d записала данные в канал URLs\n", id)
						}

					}
				}
			}

			if n.DataAtom == atom.Script && rawJSON == "" {
				isJSONLd := false
				for _, a := range n.Attr {
					if a.Key == "type" && a.Val == "application/ld+json" {
						isJSONLd = true
						continue
					}
				}

				if isJSONLd && n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
					rawJSON = n.FirstChild.Data
					continue
				}
			}
		}
	}

	if rawJSON != "" {

		if !strings.Contains(rawJSON, `"@type": "JobPosting"`) {
			return
		}

		err = json.Unmarshal([]byte(rawJSON), &job)
		if err != nil {
			log.Printf("Error marshal json: %s", err)
		}
		job.Description = parseHTMLToText(job.Description)
		// wc.submitJob(job)
		wc.Jobs <- job
		fmt.Printf("Goroutine with id: %d записала данные в канал Jobs\n", id)
	}

}

// func (wc *WebCrawler) submitJob(job Job) {
// 	select {
// 	case <-wc.ctx.Done():

// 	case wc.Jobs <- job:

// 	}
// }

func getHTML(url string) string {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Error creating request: %s", err)
		return ""
	}
	req.Header.Set("User-Agent", "job-crawler/pet-projec")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("Error client doesn't do request: %s", err)
		return ""
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		log.Printf("Unexpected status code: %d", res.StatusCode)
		return ""
	}

	htmlString, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading body: %s", err)
		return ""
	}
	return string(htmlString)
}

func ProccessHTML(htmlString string, urlString string, tracker *URLTracker) {

	doc, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		fmt.Printf("Error parse html: %s", err)
	}

	base, err := url.Parse(urlString)
	if err != nil {
		log.Printf("Error parse url: %s\n", err)
	}
	var job Job
	job.URL = urlString
	rawJSON := ""

	for n := range doc.Descendants() {
		if n.Type == html.ElementNode {

			if n.DataAtom == atom.A {
				for _, a := range n.Attr {
					if a.Key == "href" && isTargetURL(a.Val) {
						path, err := url.Parse(a.Val)
						if err != nil {
							log.Printf("Error href: %s\n", err)
						}
						fullURL := base.ResolveReference(path)
						fmt.Println(fullURL)
						exists := tracker.Visit(fullURL.String())
						if !exists {
							// wc.URLs <- fullURL.String()
							// fmt.Printf("Goroutine with id: %d записала данные в канал URLs\n", id)
						}

					}
				}
			}

			if n.DataAtom == atom.Script && rawJSON == "" {
				isJSONLd := false
				for _, a := range n.Attr {
					if a.Key == "type" && a.Val == "application/ld+json" {
						isJSONLd = true
						continue
					}
				}

				if isJSONLd && n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
					rawJSON = n.FirstChild.Data
					continue
				}
			}
		}
	}

	if rawJSON != "" {

		if !strings.Contains(rawJSON, `"@type": "JobPosting"`) {
			return
		}

		err = json.Unmarshal([]byte(rawJSON), &job)
		if err != nil {
			log.Printf("Error marshal json: %s", err)
		}
		job.Description = parseHTMLToText(job.Description)

	}

}
