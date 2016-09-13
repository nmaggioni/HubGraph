package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// D3GraphData is the JSON string that the frontend requests for sourcing graph data.
// It is served directly from memory via a custom http route.
var D3GraphData string

// DashboardData is the JSON string that the frontend requests for sourcing dashboard data.
// It is served directly from memory via a custom http route.
var DashboardData string

// MarshalD3ToMemory converts a JSON object to a string and saves it in-memory, in the `D3GraphData` variable.
func MarshalD3ToMemory(d3Data D3) {
	d3JSON, err := json.MarshalIndent(d3Data, "", "  ")
	if err != nil {
		log.Fatalf("Unable to marshal graph data to in-memory JSON: %s", err.Error())
	}

	D3GraphData = string(d3JSON)
}

// GetLastUpdateTime fetches the last time graph data (see: `D3` structure) was updated into a RFC1123Z formatted string
func GetLastUpdateTime() string {
	var d3Data D3

	err := json.Unmarshal([]byte(D3GraphData), &d3Data)

	if err != nil {
		fmt.Printf("Unable to unmarshal d3Data in-memory JSON to data structure:%s - ", err.Error())
		log.Fatal(D3GraphData)
	}

	return d3Data.LastUpdate
}

// MarshalDashboardToMemory converts a JSON object to a string and saves it in-memory, in the `DashboardData` variable.
func MarshalDashboardToMemory(dashboardData Dashboard) {
	dashboardJSON, err := json.MarshalIndent(dashboardData, "", "  ")
	if err != nil {
		log.Fatalf("Unable to marshal dashboard data to in-memory JSON: %s", err.Error())
	}
	DashboardData = string(dashboardJSON)
}
