package jsonsearcher

import (
	"encoding/json"
	"log"
	"fmt"
	"strings"
	"reflect"
)


// https://www.tutorialspoint.com/json/json_data_types.htm
//
// Number 	double- precision floating-point format in JavaScript
// String 	double-quoted Unicode with backslash escaping
// Boolean 	true or false
// Array 	an ordered sequence of values
// Value 	it can be a string, a number, true or false, null etc
// Object 	an unordered collection of key:value pairs
// Whitespace 	can be used between any pair of tokens
// null 	empty

const (
	Null = 0
	Number = 1
	String = 2
	Boolean = 3
	Array = 4
	Value = 5
	Object = 6
	Whitespace = 7 //todo remove if not used
)

type node struct {
	name      string
	value     struct{}
	valueJson string
	nodetype  int
}

func Parse(b []byte) map[string]interface{} {
	var objmap map[string]interface{}
	err := json.Unmarshal(b, &objmap)
	if err != nil {
		log.Fatal("%v", err)
	}
	return objmap
}

func printMap(m map[string]interface{}, depth int) {
	prefix := strings.Repeat("|   ", depth)
	for k, v := range m {

		fmt.Printf("%s%s: ", prefix, k)
		if reflect.ValueOf(v).Kind() == reflect.Map{
			fmt.Printf("\n")
			printMap(v.(map[string]interface{}), depth + 1)
		} else {
			fmt.Printf("%v\n", v)
		}
	}
}

func PrintMap(m map[string]interface{}) {
	printMap(m, 0)
}

