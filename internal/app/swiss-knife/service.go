package swissknife

import (
	"context"
	jira "github.com/andygrunwald/go-jira"
	githubeventcache "github.com/siriusfreak/swiss-knife/internal/app/github-event-cache"
	savedjql "github.com/siriusfreak/swiss-knife/internal/app/saved-jql"
	pb "github.com/siriusfreak/swiss-knife/internal/pkg/generated/api/swiss-knife"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"log"
	"net"
)

type Issue = jira.Issue
type SavedJQL = savedjql.SavedJQL
type JiraClient interface {
	FetchAllIssues(jql string) ([]jira.Issue, error)
}

type SavedJQLService interface {
	SaveJQL(ctx context.Context, jql *SavedJQL) error
	GetAllJQL(ctx context.Context) ([]*SavedJQL, error)
	DeleteJQL(ctx context.Context, id string) error
}

type Event = githubeventcache.Event

type GithubEventsCacheService interface {
	FetchAndSaveEvents(owner, repoName string, limit int64) error
	GetGithubEvents(owner, repoName string, startTimestamp, endTimestamp int64) ([]Event, error)
}

type Server struct {
	pb.UnimplementedSwissKnifeServer

	jiraClient      JiraClient
	savedJQLService SavedJQLService
	githubEvents    GithubEventsCacheService
}

func (s *Server) CacheGithubEvents(ctx context.Context, in *pb.CacheGithubEventsRequest) (*pb.CacheGithubEventsResponse, error) {
	err := s.githubEvents.FetchAndSaveEvents(in.Owner, in.Repo, in.Limit)
	if err != nil {
		return nil, err
	}

	return &pb.CacheGithubEventsResponse{}, nil
}

func (s *Server) GetGithubEvents(ctx context.Context, in *pb.GetGithubEventsRequest) (*pb.GetGithubEventsResponse, error) {
	events, err := s.githubEvents.GetGithubEvents(in.Owner, in.Repo, in.StartTimestamp, in.EndTimestamp)
	if err != nil {
		return nil, err
	}

	responseEvents := make([]*pb.GithubEvent, len(events))
	for i, event := range events {
		responseEvents[i] = &pb.GithubEvent{
			Id:        event.Id,
			Type:      event.Type,
			Repo:      event.Repo,
			Owner:     event.Owner,
			Payload:   event.Payload,
			Timestamp: event.Timestamp,
			User:      event.User,
		}
	}

	return &pb.GetGithubEventsResponse{
		Events: responseEvents,
	}, nil
}

func (s *Server) GetSavedJQL(ctx context.Context, in *pb.GetSavedJQLRequest) (*pb.GetSavedJQLResponse, error) {
	log.Println("Received saved JQL request")
	jqls, err := s.savedJQLService.GetAllJQL(ctx)
	if err != nil {
		log.Println("Error fetching saved JQLs", err)
		return nil, err
	}

	res := make([]*pb.SavedJQL, 0, len(jqls))
	for _, jql := range jqls {
		res = append(res, &pb.SavedJQL{
			Id:   jql.ID,
			Name: jql.Name,
			Jql:  jql.JQL,
		})
	}

	return &pb.GetSavedJQLResponse{
		SavedJQL: res,
	}, nil
}

func (s *Server) SaveJQL(ctx context.Context, in *pb.SaveJQLRequest) (*pb.SaveJQLResponse, error) {
	log.Printf("Received save JQL request: %v", in.Jql)
	if in.Jql == "" {
		log.Println("JQL query is empty")
		return nil, status.Errorf(codes.InvalidArgument, "JQL query is empty")
	}
	err := s.savedJQLService.SaveJQL(ctx, &SavedJQL{
		Name: in.Name,
		JQL:  in.Jql,
	})
	if err != nil {
		log.Println("Error saving JQL", err)
		return nil, err
	}

	return &pb.SaveJQLResponse{}, nil
}

func (s *Server) DeleteJQL(ctx context.Context, in *pb.DeleteSavedJQLRequest) (*pb.DeleteSavedJQLResponse, error) {
	log.Printf("Received delete JQL request: %v", in.Id)
	if in.Id == "" {
		log.Println("JQL id is empty")
		return nil, status.Errorf(codes.InvalidArgument, "JQL id is empty")
	}
	err := s.savedJQLService.DeleteJQL(ctx, in.Id)
	if err != nil {
		log.Println("Error deleting JQL", err)
		return nil, err
	}

	return &pb.DeleteSavedJQLResponse{}, nil
}

func (s *Server) GetJIRATasks(ctx context.Context, in *pb.GetJIRATasksRequest) (*pb.GetJIRATasksResponse, error) {
	log.Printf("Received JIRA tasks request: %v", in.Jql)
	if in.Jql == "" {
		log.Println("JQL query is empty")
		return nil, status.Errorf(codes.InvalidArgument, "JQL query is empty")
	}
	issues, err := s.jiraClient.FetchAllIssues(in.GetJql())

	if err != nil {
		log.Println("Error fetching issues from JIRA", err)
		return nil, err
	}

	res := make([]*pb.JIRATask, 0, len(issues))
	for _, issue := range issues {
		if issue.Fields == nil {
			continue
		}

		links := make([]*pb.JIRAFieldIssueLink, len(issue.Fields.IssueLinks))
		for i, link := range issue.Fields.IssueLinks {
			var inwardIssue, outwardIssue *pb.JIRAFieldInOutwardIssue
			if link.InwardIssue != nil {
				inwardIssue = &pb.JIRAFieldInOutwardIssue{
					Key: link.InwardIssue.Key,
				}
			}
			if link.OutwardIssue != nil {
				outwardIssue = &pb.JIRAFieldInOutwardIssue{
					Key: link.OutwardIssue.Key,
				}
			}

			links[i] = &pb.JIRAFieldIssueLink{
				InwardIssue:  inwardIssue,
				OutwardIssue: outwardIssue,
				Type: &pb.JIRAFieldIssueType{
					Name: link.Type.Name,
				},
			}
		}

		fields := &pb.JIRATaskFields{
			Summary:    issue.Fields.Summary,
			IssueLinks: links,
			IssueType:  &pb.JIRAFieldIssueType{Name: issue.Fields.Type.Name},
		}

		if issue.Fields.Epic != nil {
			fields.EpicKey = issue.Fields.Epic.Key
		}
		if issue.Fields.Status != nil {
			fields.Status = &pb.JIRAFieldStatus{Name: issue.Fields.Status.Name}
		}
		if issue.Fields.Parent != nil {
			fields.Parent = &pb.JIRAFieldParent{Key: issue.Fields.Parent.Key}
		}

		task := &pb.JIRATask{
			Key:    issue.Key,
			Fields: fields,
		}
		res = append(res, task)
	}

	return &pb.GetJIRATasksResponse{
		Tasks: res,
	}, nil
}

func StartGRPCServer(address string, jiraClient JiraClient, savedJQLService SavedJQLService, githubEvents GithubEventsCacheService) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterSwissKnifeServer(s, &Server{jiraClient: jiraClient, savedJQLService: savedJQLService, githubEvents: githubEvents})

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
