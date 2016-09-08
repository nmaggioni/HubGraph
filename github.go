package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/oauth2"
)

// RateLimitSpecs is an instance of `rateLimitSpecs` that holds RL details received from the API response's headers.
var RateLimitSpecs rateLimitSpecs

// GithubEvents is the model to marshal the received JSON API data with.
type GithubEvents []struct { // https://mholt.github.io/json-to-go/
	ID    string `json:"id"`
	Type  string `json:"type"`
	Actor struct {
		ID           int    `json:"id"`
		DisplayLogin string `json:"display_login"`
		AvatarURL    string `json:"avatar_url"`
	} `json:"actor"`
	Repo struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"repo"`
	CreatedAt time.Time `json:"created_at"`
	Org       struct {
		ID         int    `json:"id"`
		Login      string `json:"login"`
		GravatarID string `json:"gravatar_id"`
		URL        string `json:"url"`
		AvatarURL  string `json:"avatar_url"`
	} `json:"org,omitempty"`
}

// rateLimitSpecs defines the fields from the API response's headers that concern Rate Limiting.
type rateLimitSpecs struct {
	Limit          int
	Remaining      int
	ResetTimestamp int64
	PollInterval   int
}

// unauthenticatedGet performs an unauthenticated call to the GitHub's API.
// If given an `http.Client`, it will be used to perform the call instead of the standard one.
// This caveat is used to deduplicate code for the `authenticatedGet` function.
func unauthenticatedGet(pages int, page int, client *http.Client) (GithubEvents, bool) {
	var httpClient *http.Client
	if client != nil {
		httpClient = client
	} else {
		httpClient = &http.Client{
			Timeout: time.Second * 30,
		}
	}
	r, err := httpClient.Get("https://api.github.com/events?page=" + strconv.Itoa(page))
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	RateLimitSpecs.Limit, _ = strconv.Atoi(r.Header.Get("x-ratelimit-limit"))
	RateLimitSpecs.Remaining, _ = strconv.Atoi(r.Header.Get("x-ratelimit-remaining"))
	RateLimitSpecs.ResetTimestamp, _ = strconv.ParseInt(r.Header.Get("x-ratelimit-reset"), 10, 64)
	RateLimitSpecs.PollInterval, _ = strconv.Atoi(r.Header.Get("x-poll-interval"))
	RateLimitSpecs.PollInterval = RateLimitSpecs.PollInterval * pages

	if r.StatusCode == http.StatusNotModified { // (304) No new content
		return nil, false
	} else if r.StatusCode == http.StatusForbidden { // (403) Rate limit reached
		return nil, true
	}

	body, _ := ioutil.ReadAll(r.Body)
	var events GithubEvents
	json.Unmarshal(body, &events)

	return events, false
}

// authenticatedGet performs an authenticated call to the GitHub's API using the user-supplied token.
// A custom call to `unauthenticatedGet` is made under the hood.
func authenticatedGet(pages int, page int, token string) (GithubEvents, bool) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	return unauthenticatedGet(pages, page, tc)
}

// GetHubData returns a success/failure boolean (respectively `true`/`false`) along with the marshalled API data.
// See the `GithubEvents` struct.
func GetHubData(pages int, page int, token string) (GithubEvents, bool) {
	if token != "" {
		return authenticatedGet(pages, page, token)
	}
	return unauthenticatedGet(pages, page, nil)
}

// GetSpecsFromEventType returns the integer to use as group ID in the frontend graph.
// Each group is used to specifically colour a different type of event.
func GetSpecsFromEventType(eventType string) (int, string) {
	switch eventType { // https://developer.github.com/v3/activity/events/types/
	case "CommitCommentEvent":
		return 1, "Comment to commit"  // TODO: should be appended to commit node, not to repo
	case "CreateEvent":
		return 2, "New repo created"
	case "DeleteEvent":
		return 3, "Something has been deleted"
	case "ForkEvent": // Fired on the parent repo!
		return 4, "Repo has been forked"
	case "GollumEvent":
		return 5, "Wiki page edited"
	case "IssueCommentEvent":
		return 6, "Issue has been commented"
	case "IssuesEvent":
		return 7, "An issue has changed"
	case "MemberEvent":
		return 8, "New collaborator added"
	case "PublicEvent":
		return 9, "Repo made public!"
	case "PullRequestEvent":
		return 10, "New pull request"
	case "PullRequestReviewCommentEvent":
		return 11, "PR's code has been commented"
	case "PushEvent":
		return 12, "New commit pushed"
	case "ReleaseEvent":
		return 13, "New release created"
	case "WatchEvent":
		return 14, "Repo has been starred"
	default:
		return 99, "Unknown event"
	}
}
