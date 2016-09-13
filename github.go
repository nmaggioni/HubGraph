package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/oauth2"
)

// RateLimitSpecs is an instance of `rateLimitSpecs` that holds RL details received from the API response's headers.
var RateLimitSpecs rateLimitSpecs

type user struct {
	Login      string `json:"login"`
	ID         int    `json:"id"`
	AvatarURL  string `json:"avatar_url"`
	GravatarID string `json:"gravatar_id"`
	HTMLURL    string `json:"html_url"`
}

type repo struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Owner       user   `json:"owner"`
	HTMLURL     string `json:"html_url"`
	Description string `json:"description"`
	Fork        bool   `json:"fork"`
}

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
	Payload struct {
		Comment struct {
			HTMLURL  string `json:"html_url"`
			ID       int    `json:"id"`
			User     user   `json:"user"`
			CommitID string `json:"commit_id"`
		} `json:"comment"`
		Ref          string `json:"ref"`
		RefType      string `json:"ref_type"`
		MasterBranch string `json:"master_branch"`
		Description  string `json:"description"`
		Forkee       struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			FullName string `json:"full_name"`
			Owner    user   `json:"owner"`
		} `json:"forkee"`
		Pages []struct {
			PageName string `json:"page_name"`
			Title    string `json:"title"`
			Action   string `json:"action"`
			Sha      string `json:"sha"`
			HTMLURL  string `json:"html_url"`
		} `json:"pages"`
		Action string `json:"action"`
		Issue  struct {
			URL         string `json:"url"`
			LabelsURL   string `json:"labels_url"`
			CommentsURL string `json:"comments_url"`
			EventsURL   string `json:"events_url"`
			HTMLURL     string `json:"html_url"`
			ID          int    `json:"id"`
			Number      int    `json:"number"`
			Title       string `json:"title"`
			User        user   `json:"user"`
			Labels      []struct {
				URL   string `json:"url"`
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"labels"`
			State    string `json:"state"`
			Locked   bool   `json:"locked"`
			Comments int    `json:"comments"`
			Body     string `json:"body"`
		} `json:"issue"`
		Member     user `json:"member"`
		Repository struct {
			ID       int    `json:"id"`
			Name     string `json:"name"`
			FullName string `json:"full_name"`
			Owner    user   `json:"owner"`
		} `json:"repository"`
		Number      int `json:"number"`
		PullRequest struct {
			ID      int    `json:"id"`
			HTMLURL string `json:"html_url"`
			Number  int    `json:"number"`
			State   string `json:"state"`
			Locked  bool   `json:"locked"`
			Title   string `json:"title"`
			User    user   `json:"user"`
			Body    string `json:"body"`
			Head    struct {
				Label string `json:"label"`
				Ref   string `json:"ref"`
				Sha   string `json:"sha"`
				User  user   `json:"user"`
				Repo  repo   `json:"repo"`
			} `json:"head"`
			Base struct {
				Label string `json:"label"`
				Ref   string `json:"ref"`
				Sha   string `json:"sha"`
				User  user   `json:"user"`
				Repo  repo   `json:"repo"`
			} `json:"base"`
		} `json:"pull_request"`
		Head         string `json:"head"`
		Before       string `json:"before"`
		Size         int    `json:"size"`
		DistinctSize int    `json:"distinct_size"`
		Commits      []struct {
			Sha      string `json:"sha"`
			Distinct bool   `json:"distinct"`
			Message  string `json:"message"`
			URL      string `json:"url"`
			Author   struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"author"`
		} `json:"commits"`
	} `json:"payload"`
	Org struct {
		ID         int    `json:"id"`
		Login      string `json:"login"`
		GravatarID string `json:"gravatar_id"`
		URL        string `json:"url"`
		AvatarURL  string `json:"avatar_url"`
	} `json:"org,omitempty"`
}

type rawRateLimitSpecs struct {
	Resources struct {
		Core struct {
			Limit     int
			Remaining int
			Reset     int64
		}
	}
}

func (raw *rawRateLimitSpecs) toRateLimitSpecs() rateLimitSpecs {
	return rateLimitSpecs{
		Limit:          raw.Resources.Core.Limit,
		Remaining:      raw.Resources.Core.Remaining,
		ResetTimestamp: raw.Resources.Core.Reset,
		PollInterval:   180, // default value
	}
}

// rateLimitSpecs defines the fields from the API response's headers that concern Rate Limiting.
type rateLimitSpecs struct {
	Limit          int
	Remaining      int
	ResetTimestamp int64
	PollInterval   int
}

// APIError is the struct that is used to report API status errors. It is used to differ from 304 and 403 status codes.
type APIError struct {
	msg    string
	status int
}

func (e *APIError) Error() string {
	return e.msg + " (" + strconv.Itoa(e.status) + ")"
}

// parseHeader converts a header string to int and handles errors in conversion.
func parseHeader(header http.Header, fieldName string) int {
	if header.Get(fieldName) != "" {
		content, err := strconv.Atoi(header.Get(fieldName))
		if err != nil {
			log.Fatalf("Unable to parse header \"%s\"'s content: %s", fieldName, err.Error())
		}
		return content
	}
	return 0
}

// parseLongHeader converts a long header string to int64 and handles errors in conversion.
func parseLongHeader(header http.Header, fieldName string) int64 {
	if header.Get(fieldName) != "" {
		content, err := strconv.ParseInt(header.Get(fieldName), 10, 64)
		if err != nil {
			log.Fatalf("Unable to parse long header \"%s\"'s content: %s", fieldName, err.Error())
		}
		return content
	}
	return 0
}

// unauthenticatedGet performs an unauthenticated call to the GitHub's API.
// If given an `http.Client`, it will be used to perform the call instead of the standard one.
// This caveat is used to deduplicate code for the `authenticatedGet` function.
func unauthenticatedGet(url string, client *http.Client) ([]byte, error) {
	var httpClient *http.Client
	if client != nil {
		httpClient = client
	} else {
		httpClient = &http.Client{
			Timeout: time.Second * 30,
		}
	}
	r, err := httpClient.Get(url)
	if err != nil {
		log.Fatalf("Error in requesting data from API: %s", err.Error())
	}

	defer r.Body.Close()

	if r.StatusCode == http.StatusNotModified { // (304) No new content
		return nil, &APIError{"no new content", 304}
	} else if r.StatusCode == http.StatusForbidden { // (403) Rate limit reached
		return nil, &APIError{"rate limited", 403}
	}

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Fatal("Unable to read response body")
	}

	return body, nil
}

// authenticatedGet performs an authenticated call to the GitHub's API using the user-supplied token.
// A custom call to `unauthenticatedGet` is made under the hood.
func authenticatedGet(url string, token string) ([]byte, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	return unauthenticatedGet(url, tc)
}

// GetHubData returns a success/failure boolean (respectively `true`/`false`) along with the marshalled API data.
// See the `GithubEvents` struct.
func GetHubData(pages int, page int, token string) (GithubEvents, error) {
	var responseBody []byte
	var err error

	if token == "" {
		responseBody, err =
			unauthenticatedGet("https://api.github.com/events?page="+strconv.Itoa(page), nil)
	} else {
		responseBody, err =
			authenticatedGet("https://api.github.com/events?page="+strconv.Itoa(page), token)
	}

	if err != nil {
		return nil, err
	}

	var events GithubEvents
	json.Unmarshal(responseBody, &events)

	return events, nil
}

// GetRateLimits fetches rate limits from GitHub API server for the current client. It returns a `rateLimitSpecs`
// data structures containg reset timestamp, used requests, max requests and a default polling interval
func GetRateLimits(token string) rateLimitSpecs {
	// TODO: Check if rateLimitSpecs is already initialized

	var response []byte
	var err error

	const rateLimitURL = "https://api.github.com/rate_limit"
	if token == "" {
		response, err = unauthenticatedGet(rateLimitURL, nil)
	} else {
		response, err = authenticatedGet(rateLimitURL, token)
	}

	if err != nil {
		panic("Error while fetching rate limits for GitHub API")
	}

	var limits rawRateLimitSpecs
	err = json.Unmarshal(response, &limits)

	if err != nil {
		panic("Error while unmarshalling GitHub rate limit response")
	}

	RateLimitSpecs = limits.toRateLimitSpecs()
	return RateLimitSpecs
}

// GetSpecsFromEventType returns the integer to use as group ID in the frontend graph.
// Each group is used to specifically colour a different type of event.
func GetSpecsFromEventType(eventType string) (int, string) {
	switch eventType { // https://developer.github.com/v3/activity/events/types/
	case "CommitCommentEvent":
		return 1, "Comment to commit" // TODO: should be appended to commit node, not to repo
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
