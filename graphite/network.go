package graphite

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"io"
)

//var tokens = make(chan struct{}, 40)

func CURL(url , method string, parms map[string]string, body io.Reader) (string , error){

	client := &http.Client{}

	req, _ := http.NewRequest(method, url, body)
	q := req.URL.Query()
	for k,v := range parms {
		q.Add(k,v)
	}
	req.URL.RawQuery = q.Encode()
	//log.Println(req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(resp_body), nil
}

func LoadMetric(graphiteURL, metric, format, from , until string) string {

	client := &http.Client{}

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/render", graphiteURL), nil)
	q := req.URL.Query()
	q.Add("target", metric)
	if from != "" {
		q.Add("from", from)
	}
	if until != "" {
		q.Add("until", until)
	}
	q.Add("format", format)
	req.URL.RawQuery = q.Encode()
	log.Println(req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	resp_body, _ := ioutil.ReadAll(resp.Body)

	//fmt.Println(resp.Status)
	return string(resp_body)
}
