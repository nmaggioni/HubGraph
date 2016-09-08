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
		group, title := GetSpecsFromEventType(evt.Type)
		d3Data.Nodes = append(d3Data.Nodes, node{evt.ID, group, title})
		d3Data.Links = append(d3Data.Links, link{evt.Repo.Name, evt.ID, 1})
	}
}

// buildGraph wraps around the other graph-building functions to generate new graph data.
// It iterates on as many API event pages as specified via the CLI flag or the default value.
func buildGraph() bool {
	// Prepare graph
	var d3Data D3
	for page := 1; page < pages+1; page++ {
		// Get latest events from GitHub
		events, rateLimited := GetHubData(pages, page, token)
		if rateLimited {
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
			return buildGraph()
		} else if events == nil {
			fmt.Println("No new data available!")
			return false
		}
		// Create graph nodes
		extractReposAsNodes(events, &d3Data)
		// Create graph links
		extractEventsAsLinks(events, &d3Data)
		clearLine()
		fmt.Printf("Page %d analyzed...\r", page)
	}
	// Output to memory
	MarshalToMemory(d3Data)
	return true
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

	go Listen(port)
	fmt.Println("Listening on port " + port + " - http://localhost:" + port + "/\n")
	buildGraph()
	var duration time.Duration
	if delay != (60 * pages) {
		duration = time.Second * time.Duration(delay)
	} else {
		if reflect.TypeOf(RateLimitSpecs.PollInterval).String() != "int" {
			RateLimitSpecs.PollInterval = 60 * pages
		}
		duration = time.Second * time.Duration(RateLimitSpecs.PollInterval)
	}
	for {
		secondsToWait := duration.Seconds()
		for {
			if secondsToWait <= 0 {
				clearLine()
				break
			}
			fmt.Printf("Content updated at %s - Next refresh in: %fs (RL: %d/%d req/hr used)\r",
				time.Now().Format(time.RFC822Z), secondsToWait, (RateLimitSpecs.Limit - RateLimitSpecs.Remaining), RateLimitSpecs.Limit)
			time.Sleep(time.Second * 1)
			secondsToWait--
		}
		buildGraph()
	}
}