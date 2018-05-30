// Copyright 2017 alertmanager-github-receiver Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//////////////////////////////////////////////////////////////////////////////

// A client interface wrapping the Github API for creating, listing, and closing
// issues on a single repository.
package issues

import (
	"io/ioutil"
	"log"

	"github.com/google/go-github/github"
	"github.com/kr/pretty"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// A Client manages communication with the Github API.
type Client struct {
	// githubClient is an authenticated client for accessing the github API.
	GithubClient *github.Client
	// repos are the github repositories under the above owner.
	repos []string
	// owner is the github project (e.g. github.com/<owner>/<repo>).
	owner string
}

// NewClient creates an Client authenticated using the Github authToken.
// Future operations are only performed on the given github "owner/repo".
func NewClient(owner string, repos []string, authToken string) *Client {
	ctx := context.Background()
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: authToken},
	)
	client := &Client{
		GithubClient: github.NewClient(oauth2.NewClient(ctx, tokenSource)),
		repos:        repos,
		owner:        owner,
	}
	return client
}

// CreateIssue creates a new Github issue. New issues are unassigned.
func (c *Client) CreateIssue(repo, title, body string) (*github.Issue, error) {
	// alert color: #e03010
	//        name: alert:boom:
	//      search: label:"alert:boom:"

	// Construct a minimal github issue request.
	issueReq := github.IssueRequest{
		Title:  &title,
		Body:   &body,
		Labels: &([]string{"alert:boom:"}),
	}

	// Create the issue.
	// See also: https://developer.github.com/v3/issues/#create-an-issue
	// See also: https://godoc.org/github.com/google/go-github/github#IssuesService.Create
	issue, resp, err := c.GithubClient.Issues.Create(
		context.Background(), c.owner, repo, &issueReq)
	if err != nil {
		log.Printf("Error in CreateIssue: response: %v\n%s",
			err, pretty.Sprint(resp))
		return nil, err
	}
	return issue, nil
}

// ListOpenIssues returns open issues from github Github issues are either
// "open" or "closed". Closed issues have either been resolved automatically or
// by a person. So, there will be an ever increasing number of "closed" issues.
// By only listing "open" issues we limit the number of issues returned.
func (c *Client) ListOpenIssues() ([]*github.Issue, error) {
	var allIssues []*github.Issue

	//opts := &github.IssueListByRepoOptions{State: "open"}
	sopts := &github.SearchOptions{}
	for {
		// TODO: use "Search" rather than "List" -- is:issue in:title is:open <text>
		issues, resp, err := c.GithubClient.Search.Issues(
			context.Background(), `is:issue in:title is:open org:`+c.owner+` label:"alert:boom:"`, sopts)

		//issues, resp, err := c.GithubClient.Issues.ListByRepo(
		//context.Background(), c.owner, c.repos[0], opts)
		if err != nil {
			log.Printf("Failed to list open github issues: %v\n", err)
			return nil, err
		}
		b, _ := ioutil.ReadAll(resp.Body)
		pretty.Print(string(b))
		// Collect 'em all.
		for _, issue := range issues.Issues {
			log.Println("ListOpenIssues", issue.GetTitle())
			i := new(github.Issue)
			*i = issue
			allIssues = append(allIssues, i)
		}

		// Continue loading the next page until all issues are received.
		if resp.NextPage == 0 {
			break
		}
		sopts.ListOptions.Page = resp.NextPage
	}
	return allIssues, nil
}

// CloseIssue changes the issue state to "closed" unconditionally. If the issue
// is already close, then this should have no effect.
func (c *Client) CloseIssue(issue *github.Issue) (*github.Issue, error) {
	issueReq := github.IssueRequest{
		State: github.String("closed"),
	}

	// Edits the issue to have "closed" state.
	// See also: https://developer.github.com/v3/issues/#edit-an-issue
	// See also: https://godoc.org/github.com/google/go-github/github#IssuesService.Edit
	pretty.Print("ISSUE", issue)
	closedIssue, _, err := c.GithubClient.Issues.Edit(
		context.Background(),
		c.owner,
		c.repos[0],
		issue.GetNumber(),
		&issueReq)
	if err != nil {
		log.Printf("Failed to close issue: %v\n", err)
		return nil, err
	}
	return closedIssue, nil
}
