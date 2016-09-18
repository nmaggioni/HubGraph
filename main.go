/*
HubGraph grabs the latest events from the GitHub's API and builds an entertaining graph upon them.

A frontend web page is exposed with a D3-powered (https://d3js.org/) force graph in it.
Both unauthenticated and authenticated (OAUTH2 token) requests are supported, enabling the use of the 60 req/hr and 5000 req/hr rate limits.

Consult the help by running `./hubgraph -h` to learn more about the configuration options.
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"reflect"
	"time"
)

var (
	port  string
	pages int
	token string
	delay int
)

type node struct {
	ID    string `json:"id"`
	Group int    `json:"group"`
	Title string `json:"title"`
}

type link struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Value  int    `json:"value"`
}

// D3 is the structure used to construct the data for the frontend D3 graph.
type D3 struct {
	Nodes []node `json:"nodes"`
	Links []link `json:"links"`
}

// Dashboard is the structure containing the necessary details for the dashboard.
type Dashboard struct {
	RequestsUsed    int    `json:"requestsUsed"`
	MaxRequests     int    `json:"maxRequests"`
	RefreshInterval int64  `json:"refreshInterval"`
	LastUpdate      string `json:"lastUpdate"`
}

// stringInSlice determines whenever a string is already present in a slice.
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// extractReposAsNodes parses the obtained data and creates the main nodes for every present repo.
func extractReposAsNodes(events GithubEvents, d3Data *D3) {
	var repos []string
	for _, evt := range events {
		if !stringInSlice(evt.Repo.Name, repos) {
			repos = append(repos, evt.Repo.Name)
		}
	}
	for _, repoName := range repos {
		d3Data.Nodes = append(d3Data.Nodes, node{repoName, 0, ""})
	}
}

// extractEventsAsLinks parses the obtained data and creates the links between the nodes.
func extractEventsAsLinks(events GithubEvents, d3Data *D3) {
	for _, evt := range events {
		// Generic handler
		group, title := GetSpecsFromEventType(evt.Type)
		d3Data.Nodes = append(d3Data.Nodes, node{evt.ID, group, title})
		d3Data.Links = append(d3Data.Links, link{evt.Repo.Name, evt.ID, 1})
		// Fork event
		if evt.Type == "ForkEvent" {
			d3Data.Nodes = append(d3Data.Nodes, node{evt.Payload.Forkee.FullName, 0, ""})
			d3Data.Links = append(d3Data.Links, link{evt.Repo.Name, evt.Payload.Forkee.FullName, 10})
		}
	}
}

// buildGraph wraps around the other graph-building functions to generate new graph data.
// It iterates on as many API event pages as specified via the CLI flag or the default value.
func buildGraph() {
	// Prepare graph struct
	var d3Data D3

	// Loop around API pages
	for page := 1; page < pages+1; page++ {
		// Get latest events from GitHub
		events, err := GetHubData(pages, page, token)

		if _err, ok := err.(*APIError); ok {
			if _err.status == 403 {
				secondsToWait := RateLimitSpecs.ResetTimestamp - time.Now().UTC().Unix() + 3
				for {
					if secondsToWait <= 0 {
						clearLine()
						break
					}
					fmt.Printf("Rate limit reached. Will reset in %d seconds.    \r", secondsToWait)
					time.Sleep(time.Second * 1)
					secondsToWait--
				}
				buildGraph()
				return
			} else if _err.status == 304 {
				clearLine()
				fmt.Println("No new data available.")
				time.Sleep(time.Second * 1)
				return
			}
		} else if events == nil {
			fmt.Println("No new data available!")
			return
		}

		// Create graph nodes
		extractReposAsNodes(events, &d3Data)
		// Create graph links
		extractEventsAsLinks(events, &d3Data)
		// Output progress to console
		clearLine()
		fmt.Printf("Page %d analyzed...\r", page)
	}

	// Output graph data to memory
	MarshalD3ToMemory(d3Data)
}

func buildDashboard(refreshInterval int64) {
	var dashboardData Dashboard

	dashboardData.RequestsUsed = RateLimitSpecs.Limit - RateLimitSpecs.Remaining
	dashboardData.MaxRequests = RateLimitSpecs.Limit
	dashboardData.RefreshInterval = refreshInterval
	dashboardData.LastUpdate = time.Now().Format(time.RFC1123Z)

	MarshalDashboardToMemory(dashboardData)
}

// clearLine makes sure the terminal line is (theoretically...) empty before writing on it.
func clearLine() {
	fmt.Printf("                                                                                                          \r")
}

func main() {
	flag.StringVar(&port, "port", "3000", "The port to listen on")
	flag.IntVar(&pages, "pages", 3, "How many pages to read (will impact rate limiting dramatically!)")
	flag.IntVar(&delay, "delay", (60 * pages), "Delay in seconds between data refreshes. Defaults to (60 * pages), a safe timing for unauthenticated requests")
	flag.StringVar(&token, "token", "", "The token to authenticate requests with (will bring rate limiting to 5000/hr instead of 60/hr - https://github.com/settings/tokens/new)")
	flag.Parse()

	Listen(port)
	fmt.Println("Listening on port " + port + " - http://localhost:" + port + "/\n")

	GetRateLimits(token)

	var duration time.Duration
	if delay != (60 * pages) {
		duration = time.Second * time.Duration(delay)
	} else {
		if reflect.TypeOf(RateLimitSpecs.PollInterval).String() != "int" || RateLimitSpecs.PollInterval <= 0 {
			RateLimitSpecs.PollInterval = 60 * pages
		}
		duration = time.Second * time.Duration(RateLimitSpecs.PollInterval)
	}

	refreshInterval := int64(duration.Seconds())

	buildGraph()
	GetRateLimits(token)
	buildDashboard(refreshInterval)

	secondsToWait := refreshInterval
	lastUpdated := GetLastUpdateTime()

	for {
		for {
			if secondsToWait <= 0 {
				clearLine()
				break
			}

			nextRefresh, err := time.Parse(time.RFC1123Z, lastUpdated)

			if err != nil {
				log.Printf("Error while trying to parse last update time: \"%s\"\n", nextRefresh)
			} else {

				nextRefresh = nextRefresh.Add(time.Duration(refreshInterval) * time.Second)

				secondsToWait = int64(nextRefresh.Sub(time.Now()).Seconds())

				fmt.Printf("Content updated at %s - Next refresh in: %d (RL: %d/%d req/hr used)\r",
					lastUpdated, secondsToWait, (RateLimitSpecs.Limit - RateLimitSpecs.Remaining), RateLimitSpecs.Limit)
			}

			time.Sleep(time.Second * 1)
		}

		buildGraph()
		GetRateLimits(token)
		buildDashboard(refreshInterval)

		lastUpdated = GetLastUpdateTime()
		secondsToWait = refreshInterval
	}
}
