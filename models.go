package main

import (
	"context"
	"encoding/json"
	"sync"

	"golang.org/x/time/rate"
)

type Job struct {
	Title       string       `json:"title"`
	Company     Organization `json:"hiringOrganization"`
	Date        string       `json:"datePosted"`
	Description string       `json:"description"`
	Salary      BaseSalary   `json:"baseSalary"`
	URL         string       `json:"URL"`
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

func NewURLTracker() *URLTracker {
	return &URLTracker{visited: make(map[string]bool)}
}

type WebCrawler struct {
	WorkersCount int
	Jobs         chan Job
	URLs         chan string
	Map          URLTracker

	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	limiter *hostLimiter
}

func NewWebCrawler(workers int) *WebCrawler {
	ctx, cancel := context.WithCancel(context.Background())
	return &WebCrawler{
		WorkersCount: workers,
		Jobs:         make(chan Job, workers),
		URLs:         make(chan string, workers*5),
		Map:          URLTracker{make(map[string]bool), sync.Mutex{}},
		ctx:          ctx,
		cancel:       cancel,
		limiter:      &hostLimiter{limiters: make(map[string]*rate.Limiter)},
	}
}

type hostLimiter struct {
	mx       sync.Mutex
	limiters map[string]*rate.Limiter
}
