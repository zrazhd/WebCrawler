package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func (wc *WebCrawler) SaveJob() {
	defer wc.wg.Done()
	var arr []Job

	fileData, err := os.ReadFile("data.json")
	if err != nil {
		log.Printf("Error reading file: %s", err)
	}

	if len(fileData) > 0 {
		if err = json.Unmarshal(fileData, &arr); err != nil {
			log.Printf("Error unmarshal file: %s", err)
			return
		}
	}

	cnt := 0
	for {
		select {
		case <-wc.ctx.Done():
			fmt.Println("Context canceled in SaveJob")
			fmt.Printf("Count of Jobs: %d\n", cnt)
			save(arr)
			return
		case job, ok := <-wc.Jobs:
			if !ok {
				fmt.Println("Channel Jobs is closed")
				fmt.Printf("Count of Jobs: %d\n", cnt)
				save(arr)
				return
			}
			arr = append(arr, job)
			cnt++
		}
	}
}

func save(arr []Job) {
	data, err := json.MarshalIndent(arr, " ", "    ")
	if err != nil {
		log.Printf("Error with marshal: %s", err)
		return
	}

	err = os.WriteFile("data.json", data, 0644)
	if err != nil {
		log.Printf("Error with writing in data.json: %s", err)
		return
	}

	log.Println("Data is written in JSON")
}
