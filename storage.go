package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func (wc *WebCrawler) SaveJob() {
	defer wc.wg.Done()

	select {
	case <-wc.ctx.Done():
		fmt.Println("контекс отменен")
		return
	case job := <-wc.Jobs:
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

		fmt.Println("job получен")
		data, err := json.MarshalIndent(arr, " ", "    ")
		if err != nil {
			log.Printf("Error with marshal: %s", err)
		}

		err = os.WriteFile("data.json", data, 0644)
		if err != nil {
			log.Printf("Error with writing in data.json: %s", err)
		}

		fmt.Println("процесс завершился")

	}

}
