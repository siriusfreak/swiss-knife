package main

import (
	"context"
	"database/sql"
	"github.com/siriusfreak/swiss-knife/config"
	githubcachejob "github.com/siriusfreak/swiss-knife/internal/app/github-cache-job"
	githubclient "github.com/siriusfreak/swiss-knife/internal/app/github-client"
	githubeventcache "github.com/siriusfreak/swiss-knife/internal/app/github-event-cache"
	eventCacheRepository "github.com/siriusfreak/swiss-knife/internal/app/github-event-cache/repository"
	savedjql "github.com/siriusfreak/swiss-knife/internal/app/saved-jql"
	"github.com/siriusfreak/swiss-knife/internal/app/saved-jql/repository"
	"log"

	_ "github.com/mattn/go-sqlite3"

	jiraclient "github.com/siriusfreak/swiss-knife/internal/app/jira-client"
	swissknifeservice "github.com/siriusfreak/swiss-knife/internal/app/swiss-knife"
)

func InitializeDB(filepath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	ctx := context.Background()

	cfg, err := config.Load(ctx, "./config.yaml")

	db, err := InitializeDB("./data/saved_jql.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	repo := repository.New(db)
	service, err := savedjql.New(ctx, repo)
	if err != nil {
		panic(err)
	}

	jiraClient, err := jiraclient.New(
		cfg.Jira.URL,
		cfg.Jira.Username,
		cfg.Jira.Token)

	githubCleint, err := githubclient.New(cfg.GitHub.Token)
	if err != nil {
		log.Fatalf("failed to create github client: %s", err)
	}

	eventRepo := eventCacheRepository.New(db)
	eventCache, err := githubeventcache.NewService(githubCleint, eventRepo)
	if err != nil {
		log.Fatalf("failed to create event cache service: %s", err)
	}

	githubcachejob.New(eventCache, githubCleint)

	go func() {
		log.Println("Starting gRPC server...")
		swissknifeservice.StartGRPCServer(":50051", jiraClient, service, eventCache)
	}()

	log.Println("Starting gRPC-gateway...")
	err = swissknifeservice.StartGateway(ctx, "localhost:50051", "0.0.0.0:8080")
	if err != nil {
		log.Fatalf("failed to start gateway: %s", err)
	}

}
