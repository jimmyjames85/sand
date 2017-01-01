package main

import (
	"io"
	"log"
	"net/http"
	"fmt"
	"strings"
)
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
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
	r.ParseForm()
	request = append(request, "\n")
	request = append(request, r.Form.Encode())
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
r.Header
	log.Printf("\n\n")

}

func main() {
	http.HandleFunc("/", hello)
	http.ListenAndServe(":1234", nil)
}
