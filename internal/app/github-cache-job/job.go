package githubcachejob

import (
	"fmt"
	githubclient "github.com/siriusfreak/swiss-knife/internal/app/github-client"
	"time"
)

type RepositoryInfo struct {
	Owner    string
	RepoName string
}

type CacheService interface {
	FetchAndSaveEvents(owner, repoName string, limit int64) error
}

type Repository = githubclient.Repository
type GithubClient interface {
	GetUserSubscriptions(username string) ([]*Repository, error)
}

type Job struct {
	cache  CacheService
	client GithubClient
}

func New(cache CacheService, client GithubClient) *Job {

	j := &Job{
		cache:  cache,
		client: client,
	}

	j.StartInternalJob()

	return j
}

func (j *Job) StartInternalJob() {
	ticker := time.NewTicker(10 * time.Minute) // This will run the job every hour. Adjust the duration as needed.

	go func() {
		j.FetchAllRepos()
		for {
			select {
			case <-ticker.C:
				j.FetchAllRepos()
			}
		}
	}()
}

func (j *Job) FetchAllRepos() {
	repositories, err := j.client.GetUserSubscriptions("siriusfreak")
	if err != nil {
		fmt.Println("Error fetching user subscriptions:", err)
		return
	}
	for _, repoInfo := range repositories {
		owner := ""
		if repoInfo.Owner.Login != nil {
			owner = *repoInfo.Owner.Login
		} else if repoInfo.Owner.Name != nil {
			owner = *repoInfo.Owner.Name
		}
		err := j.cache.FetchAndSaveEvents(owner, *repoInfo.Name, 1000)
		if err != nil {
			// Handle the error, e.g., log it
			fmt.Println("Error fetching and saving events:", err)
		}
	}
}
