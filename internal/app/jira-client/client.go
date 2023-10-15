package jiraclient

import (
	"fmt"

	jira "github.com/andygrunwald/go-jira"
)

type Client struct {
	client *jira.Client
}

func New(host, email, apiToken string) (*Client, error) {
	tp := jira.BasicAuthTransport{
		Username: email,
		Password: apiToken,
	}

	client, err := jira.NewClient(tp.Client(), host)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}

func (je *Client) FetchAllIssues(jql string) ([]jira.Issue, error) {
	var allIssues []jira.Issue
	var startAt int = 0
	const maxResults int = 50 // JIRA's API might have a limit per page. Adjust as needed.

	for {
		opt := &jira.SearchOptions{
			StartAt:    startAt,
			MaxResults: maxResults,
		}

		chunk, _, err := je.client.Issue.Search(jql, opt)
		if err != nil {
			return nil, fmt.Errorf("error executing JQL request: %v", err)
		}

		allIssues = append(allIssues, chunk...)

		if len(chunk) < maxResults {
			break
		}

		startAt += len(chunk)
	}

	return allIssues, nil
}
