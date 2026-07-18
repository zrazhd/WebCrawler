package main

import "time"

func main() {
	url := "https://web3.career/graduate-junior-software-engineer-backend-elwoodtechnologies/151199"

	wc := NewWebCrawler(50)

	wc.Start()

	wc.URLs <- url

	time.Sleep(time.Second * 5)

	wc.SaveJob()
	wc.Stop()

}
