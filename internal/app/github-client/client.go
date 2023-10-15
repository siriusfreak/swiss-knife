package githubclient

import (
	"context"
	"fmt"
	"github.com/google/go-github/v54/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	user   *github.User
}

type Event = github.Event
type PullRequest = github.PullRequest
type Repository = github.Repository

func New(token string) (*Client, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)

	client := github.NewClient(tc)
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
		user:   user,
	}, nil
}

// ListRepoEvents fetches recent events for the specified repository
func (c *Client) ListRepoEvents(owner, repo string, page int) ([]*Event, int, error) {
	opts := &github.ListOptions{PerPage: 100, Page: page}
	events, resp, err := c.client.Activity.ListRepositoryEvents(context.Background(), owner, repo, opts)
	if err != nil {
		return nil, 0, err
	}
	return events, resp.LastPage, nil
}

func (c *Client) GetUserSubscriptions(username string) ([]*Repository, error) {
	// The GitHub API paginates its results. The following loop ensures all pages are fetched.
	opt := &github.ListOptions{PerPage: 50} // GitHub defaults to 30 results per page; you can go up to 100.
	var allRepos []*github.Repository

	for {
		repos, resp, err := c.client.Activity.ListWatched(context.Background(), username, opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func (c *Client) FetchAuthoredPRs(repoName string) ([]*PullRequest, error) {
	prs, _, err := c.client.PullRequests.List(context.Background(), c.user.GetLogin(), repoName, &github.PullRequestListOptions{
		State: "all",
	})
	if err != nil {
		return nil, err
	}
	return prs, nil
}

func (c *Client) FetchAssignedPRs(repoName string) ([]*PullRequest, error) {
	opts := &github.SearchOptions{}
	query := fmt.Sprintf("type:pr repo:%s assignee:%s", repoName, c.user.GetLogin())
	prs, _, err := c.client.Search.Issues(context.Background(), query, opts)
	if err != nil {
		return nil, err
	}
	var pullRequests []*github.PullRequest
	for _, issue := range prs.Issues {
		if issue.PullRequestLinks != nil {
			pr, _, err := c.client.PullRequests.Get(context.Background(), c.user.GetLogin(), repoName, issue.GetNumber())
			if err != nil {
				return nil, err
			}
			pullRequests = append(pullRequests, pr)
		}
	}
	return pullRequests, nil
}

func (c *Client) FetchCommentedPRs(repoName string) ([]*PullRequest, error) {
	opts := &github.SearchOptions{}
	query := fmt.Sprintf("type:pr repo:%s commenter:%s", repoName, c.user.GetLogin())
	prs, _, err := c.client.Search.Issues(context.Background(), query, opts)
	if err != nil {
		return nil, err
	}
	var pullRequests []*github.PullRequest
	for _, issue := range prs.Issues {
		if issue.PullRequestLinks != nil {
			pr, _, err := c.client.PullRequests.Get(context.Background(), c.user.GetLogin(), repoName, issue.GetNumber())
			if err != nil {
				return nil, err
			}
			pullRequests = append(pullRequests, pr)
		}
	}
	return pullRequests, nil
}

func (c *Client) CheckPendingReviews(repoName string, prNumber int) (bool, error) {
	reviews, _, err := c.client.PullRequests.ListReviews(context.Background(), c.user.GetLogin(), repoName, prNumber, nil)
	if err != nil {
		return false, err
	}
	for _, review := range reviews {
		if review.GetUser().GetLogin() == c.user.GetLogin() && review.GetState() == "PENDING" {
			return true, nil
		}
	}
	return false, nil
}
