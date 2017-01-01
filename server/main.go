package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"encoding/json"
)

var copydata string

func copy(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		io.WriteString(w, fmt.Sprintf("failed to parse form data: %s", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(r.Form)
	if err != nil {
		io.WriteString(w, fmt.Sprintf("failed to marshal form data: %s", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	copydata = string(b)
}

func pastejson(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, copydata)
}


func paste(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}

	bytes := []byte(copydata)

	err := json.Unmarshal(bytes, &data)
	if err != nil {
		io.WriteString(w, fmt.Sprintf("failed to unmarshal copydata: %s", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// only print the data
	if len(data) == 1 {
		for key, val := range data {
			if fmt.Sprintf("%v", val) == "[]" {
				io.WriteString(w, key)
			} else {
				io.WriteString(w, copydata)
			}
		}
	} else {
		io.WriteString(w, copydata)
	}
}

// formatRequest generates ascii representation of a request
func formatRequest(r *http.Request) string {
	// Create return string
	var request []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		fmt.Printf("======= %s =======\n", name)
		for i, h := range headers {
			fmt.Printf("%d: %s\n", i, h)
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	}

	for i, v := range r.Form {
		fmt.Printf("--> %d: %s\n", i, v)
	}
	// Return the request as a string
	return strings.Join(request, "\n")
}

func hello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello world!")
	data := make([]byte, 0)
	i, err := r.Body.Read(data)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
	log.Printf("data[%d]: %s\n", i, data)
	log.Printf("contentlength: %d\n", r.ContentLength)
	log.Printf("   form: %s\n", r.Form.Encode())
	log.Printf("postform: %s\n", r.PostForm.Encode())
	log.Printf("Host: %s\n", r.Host)
	log.Printf("formatRequest: %s\n", formatRequest(r))

	log.Printf("\n\n")

}

func main() {
	http.HandleFunc("/", hello)
	http.HandleFunc("/copy", copy)
	http.HandleFunc("/pastejson", pastejson)
	http.HandleFunc("/paste", paste)
	http.ListenAndServe(":27182", nil)
}
