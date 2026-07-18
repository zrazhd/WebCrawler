package main

import (
	"context"
	"encoding/json"
	"sync"
)

type Job struct {
	Title       string       `json:"title"`
	Company     Organization `json:"hiringOrganization"`
	Date        string       `json:"datePosted"`
	Description string       `json:"description"`
	Salary      BaseSalary   `json:"baseSalary"`
	URL         string       `json:"-"`
}
type Organization struct {
	Type string `json:"@type"`
	Name string `json:"name"`
}

type SalaryValue struct {
	Type     string      `json:"@type"`
	MinValue json.Number `json:"minValue"`
	MaxValue json.Number `json:"maxValue"`
	UnitText string      `json:"unitText"`
}

type BaseSalary struct {
	Type     string      `json:"@type"`
	Currency string      `json:"currency"`
	Value    SalaryValue `json:"value"`
}

type URLTracker struct {
	visited map[string]bool
	mx      sync.Mutex
}

type WebCrawler struct {
	WorkersCount int
	Jobs         chan Job
	URLs         chan string
	Map          URLTracker

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewWebCrawler(workers int) *WebCrawler {
	ctx, cancel := context.WithCancel(context.Background())
	return &WebCrawler{
		WorkersCount: workers,
		Jobs:         make(chan Job, workers),
		URLs:         make(chan string, workers),
		Map:          URLTracker{make(map[string]bool), sync.Mutex{}},
		ctx:          ctx,
		cancel:       cancel,
	}
}
