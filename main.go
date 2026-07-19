package main

import "time"

func main() {
	url1 := "https://www.golangprojects.com/golang-go-job-gxb-Remote-Europe-Senior-Software-Engineer-Cast-AI-remotework.html"

	wc := NewWebCrawler(50)

	wc.Start()

	wc.URLs <- url1

	time.Sleep(time.Second * 3)
	wc.Stop()
	wc.SaveJob()

}

// func main() {
// 	url := "https://www.golangprojects.com/golang-go-job-gxb-Remote-Europe-Senior-Software-Engineer-Cast-AI-remotework.html"

// 	tr := NewURLTracker()

// 	ProccessHTML(getHTML(url), url, tr)
// }
