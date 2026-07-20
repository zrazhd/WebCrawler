package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	url1 := "https://web3.career/graduate-junior-software-engineer-backend-elwoodtechnologies/151199"

	wc := NewWebCrawler(50)

	wc.Start()

	wc.URLs <- url1

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	timer := time.After(time.Second * 60)

	select {
	case <-timer:
		fmt.Println("Таймер завершился")
	case <-sigChan:
		fmt.Println("Перехвачен сигнал")

	}

	wc.Stop()
}
