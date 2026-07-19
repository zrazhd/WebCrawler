package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func (wc *WebCrawler) SaveJob() {
	// defer wc.wg.Done()

	for {
		select {
		case <-wc.ctx.Done():
			fmt.Println("контекс отменен")
			return
		case job, ok := <-wc.Jobs:
			if !ok {
				fmt.Println("Channel Jobs is closed")
				return
			}
			var arr []Job

			fileData, err := os.ReadFile("data.json")
			if err != nil {
				log.Printf("Error reading file: %s", err)
			}

			if len(fileData) > 0 {
				if err = json.Unmarshal(fileData, &arr); err != nil {
					log.Printf("Error unmarshal file: %s", err)
				}
			}

			arr = append(arr, job)

			data, err := json.MarshalIndent(arr, " ", "    ")
			if err != nil {
				log.Printf("Error with marshal: %s", err)
			}

			err = os.WriteFile("data.json", data, 0644)
			if err != nil {
				log.Printf("Error with writing in data.json: %s", err)
			}

			fmt.Println("Данные записались в файл")

		}
	}

}
