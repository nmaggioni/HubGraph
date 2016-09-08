package main

import (
	"encoding/json"
	"io/ioutil"
	"strings"
)

// D3JSON is the JSON string that the frontend requests for sourcing graph data.
// It is served directly from memory via a custom http route.
var D3JSON string

// MarshalToMemory converts a JSON object to a string and saves it in-memory, in the `D3JSON` variable.
func MarshalToMemory(d3Data D3) {
	d3JSON, err := json.MarshalIndent(d3Data, "", "  ")
	if err != nil {
		panic(err)
	}
	D3JSON = strings.Replace(string(d3JSON), "null", "{}", -1)
}

// MarshalToFile converts a JSON object to a string and saves it in a file, specified by the `filePath` string.
func MarshalToFile(d3Data D3, filePath string) {
	d3JSON, err := json.MarshalIndent(d3Data, "", "  ")
	if err != nil {
		panic(err)
	}
	d3JSON = []byte(strings.Replace(string(d3JSON), "null", "{}", -1))
	err = ioutil.WriteFile(filePath, d3JSON, 0644)
	if err != nil {
		panic(err)
	}
}