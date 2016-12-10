package graphite

import (
	"encoding/json"
	"fmt"
	"log"

	"sync"
)

type Graphite struct {
	URL string
}

func NewGraphite(url string) *Graphite {
	return &Graphite{
		URL: url,
	}
}

type APIMetric struct {
	Leaf uint8 `json:"leaf"`
	//Context []string? TODO
	Text          string `json:"text"`
	Expandable    uint8  `json:"expandable"`
	Id            string `json:"id"`
	AllowChildren uint8  `json:"allowchildren"`
}

func (g *Graphite) APIFindMetrics(query string) ([]*APIMetric, error) {

	parms := make(map[string]string)
	parms["query"] = query

	findData, err := CURL(g.URL+"/metrics/find", "GET", parms, nil)
	if err != nil {
		log.Printf("network err: %s\n", err)
		return nil, err
	}

	var results []*APIMetric
	err = json.Unmarshal([]byte(findData), &results)
	if err != nil {
		log.Printf("marshal err: %s\n", err)
		return nil, err
	}

	return results, nil
}

func (g *Graphite) APIExpand(query string) ([]string, error) {

	type apiExpandResult struct {
		Results []string `json:"results"`
	}

	parms := make(map[string]string)
	parms["query"] = query

	expandData, err := CURL(g.URL+"/metrics/expand", "GET", parms, nil)
	if err != nil {
		return nil, err
	}

	results := apiExpandResult{}
	err = json.Unmarshal([]byte(expandData), &results)
	if err != nil {
		return nil, err
	}

	return results.Results, nil
}

//TODO rename
func (g *Graphite) WalkMetrics(search string, continueWalk func(m *APIMetric) bool, metrics chan<- *APIMetric, done chan<- struct{}) {

	workerCount := 100
	work := make(chan string)
	wg := &sync.WaitGroup{}

	//start with search
	wg.Add(1)
	go func() {
		work <- search
	}()
	for i := 0; i < workerCount; i++ {
		go func() {
			for nextSearch := range work {

				//TODO first use expand instead...findmetrics has a hard time expanding '*'s
				//log.Printf("%s\n", metric.Id)
				//expandData, err:= g.APIExpand(fmt.Sprintf("%s.*",metric.Id))
				//if err !=nil{
				//	log.Printf("expand failed: %s\n", err)
				//	expandData = nil
				//}

				ms, err := g.APIFindMetrics(nextSearch)
				if err != nil {
					log.Printf("%s\n", err)
				}
				for _, metric := range ms {

					if continueWalk(metric) {
						wg.Add(1)
						go func(m *APIMetric) {
							metrics <- m
							work <- fmt.Sprintf("%s.*", m.Id)
						}(metric)
					}
				}
				wg.Done()
			}
		}()
	}
	wg.Wait()

	done <- struct{}{}
}
