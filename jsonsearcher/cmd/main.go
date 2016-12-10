package main

import (
	"log"
	"fmt"
	"io/ioutil"
	"os"
	"github.com/jimmyjames85/sand/jsonsearcher"
)

func main() {
	var fileIn string
	var err error
	var in *os.File
	var search string

	if len(os.Args) > 2 {
		fileIn = os.Args[1]
		search = os.Args[2]
	} else if len(os.Args) > 1 {
		search = os.Args[1]
	}
	if fileIn == "" {
		in = os.Stdin
	} else {
		in, err = os.Open(fileIn)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}
	b, err := ioutil.ReadAll(in)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Skipping search: %s\n", search)
	obj := jsonsearcher.Parse(b)
	jsonsearcher.PrintMap(obj)



}
