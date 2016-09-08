package main

import (
	"io"
	"log"
	"net/http"

	"github.com/GeertJohan/go.rice"
)

// replyGraphData serves the in-memory `D3GraphData` JSON string to the frontend graph.
func replyGraphData(res http.ResponseWriter, req *http.Request) {
	io.WriteString(res, D3GraphData)
}

// replyDashboardData serves the in-memory `DashboardData` JSON string to the frontend graph.
func replyDashboardData(res http.ResponseWriter, req *http.Request) {
	io.WriteString(res, DashboardData)
}

// Listen configures and starts a web server, enclosing it in an asynchronous goroutine.
func Listen(port string) {
	go func() {
		http.Handle("/", http.FileServer(rice.MustFindBox("public").HTTPBox()))
		http.HandleFunc("/graphdata.json", replyGraphData)
		http.HandleFunc("/dashboarddata.json", replyDashboardData)
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			log.Fatalf("Unable to start web server: %s", err.Error())
		}
	}()
}
