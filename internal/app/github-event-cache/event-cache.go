package githubeventcache

import (
	"encoding/json"
	githubclient "github.com/siriusfreak/swiss-knife/internal/app/github-client"
	"github.com/siriusfreak/swiss-knife/internal/app/github-event-cache/repository"
	"strconv"
)

type Service struct {
	client GitHubClient
	repo   Repository
}

type GitHubClient interface {
	ListRepoEvents(owner, repo string, page int) ([]*GithubEvent, int, error)
}

type Repository interface {
	CreateTables() error
	SaveEvent(eventID int64, eventType, repo, owner, payload, user string, timestamp int64) error
	GetEvents(owner, repo string, startTimestamp, endTimestamp int64) ([]Event, error)
}

type GithubEvent = githubclient.Event
type Event = repository.Event

func NewService(client GitHubClient, repo Repository) (*Service, error) {

	if err := repo.CreateTables(); err != nil {
		return nil, err
	}

	return &Service{
		client: client,
		repo:   repo,
	}, nil
}

func (s *Service) FetchAndSaveEvents(owner, repoName string, limit int64) error {
	var fetchedCount int64 = 0
	var page int = 0
	for {
		events, lastPage, err := s.client.ListRepoEvents(owner, repoName, page)
		if err != nil {
			return err
		}

		for _, event := range events {
			if fetchedCount >= limit {
				return nil
			}

			payload, _ := json.Marshal(event) // converting payload to string

			// I'm assuming the `ID` field of the GitHub event is a string, so we parse it to int64.
			eventID, _ := strconv.ParseInt(*event.ID, 10, 64)

			err = s.repo.SaveEvent(eventID, *event.Type, repoName, owner, string(payload), *event.Actor.Login, event.GetCreatedAt().Unix())
			if err != nil {
				return err
			}

			fetchedCount++
		}

		if fetchedCount >= limit || lastPage == 0 {
			break
		}
		page++
	}

	return nil
}

func (s *Service) GetGithubEvents(owner, repoName string, startTimestamp, endTimestamp int64) ([]Event, error) {
	return s.repo.GetEvents(owner, repoName, startTimestamp, endTimestamp)
}
