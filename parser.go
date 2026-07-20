package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/time/rate"
)

func (wc *WebCrawler) Start() {
	wc.wg.Add(1)
	go wc.SaveJob()

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
			return
		case u, ok := <-wc.URLs:
			if !ok {
				fmt.Println("Channel URLs is closed")
				return
			}

			parsed, err := url.Parse(u)
			if err != nil {
				continue
			}

			limiter := wc.limiter.GetLimitForHost(parsed.Host)
			if err := limiter.Wait(wc.ctx); err != nil {
				continue
			}
			delay := time.Duration(300+rand.Intn(400)) * time.Millisecond

			select {
			case <-time.After(delay):
			case <-wc.ctx.Done():
				return
			}

			wc.ProccessHTML(getHTML(u), u, id)
		}
	}
}

func (wc *WebCrawler) Stop() {
	wc.cancel()
	wc.wg.Wait()

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
		return
	}

	base, err := url.Parse(urlString)
	if err != nil {
		log.Printf("Error parse url: %s\n", err)
		return
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
							continue
						}
						fullURL := base.ResolveReference(path)
						exists := wc.Map.Visit(fullURL.String())
						if !exists {
							if wc.submitURL(fullURL.String()) {
								log.Printf("Goroutine with id: %d Wrote data in channel URLs\n", id)
							}
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

	checkJSON(rawJSON, &job, wc, id)

}

func checkJSON(rawJSON string, job *Job, wc *WebCrawler, id int) {
	if rawJSON != "" {

		if !strings.Contains(rawJSON, `"@type": "JobPosting"`) && !strings.Contains(rawJSON, `"@type":"JobPosting"`) {
			return
		}

		err := json.Unmarshal([]byte(rawJSON), job)
		if err != nil {
			log.Printf("Error marshal json: %s", err)
			return
		}
		parsedTime, err := parseDate(job.Date)
		if err != nil {
			log.Printf("Error with datePosted: %s", err)
			return
		}
		if parsedTime.After(time.Now().AddDate(0, -3, 0)) {
			job.Description = parseHTMLToText(job.Description)
			if wc.submitJob(*job) {
				log.Printf("Goroutine with id: %d wrote data in channel Jobs\n", id)
			}
		}

	}
}

func (wc *WebCrawler) submitJob(job Job) bool {
	select {
	case <-wc.ctx.Done():
		return false
	case wc.Jobs <- job:
		return true
	}
}

func (wc *WebCrawler) submitURL(url string) bool {
	select {
	case <-wc.ctx.Done():
		return false
	case wc.URLs <- url:
		return true
	}
}

func getHTML(url string) string {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Error creating request: %s", err)
		return ""
	}
	req.Header.Set("User-Agent", "job-crawler/pet-project")

	client := http.Client{Timeout: time.Second * 10}
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

func parseDate(date string) (time.Time, error) {

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05 -0700",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05.999-07:00",
		"2006-01-02",
	}

	for _, l := range layouts {
		parsedDate, err := time.Parse(l, date)
		if err == nil {
			return parsedDate, nil
		}
	}

	return time.Time{}, fmt.Errorf("Wrong format of time")

}

func (hl *hostLimiter) GetLimitForHost(host string) *rate.Limiter {
	hl.mx.Lock()
	defer hl.mx.Unlock()

	limiter, ok := hl.limiters[host]
	if ok {
		return limiter
	}
	limiter = rate.NewLimiter(rate.Every(time.Second*1), 1)
	hl.limiters[host] = limiter
	return limiter
}
