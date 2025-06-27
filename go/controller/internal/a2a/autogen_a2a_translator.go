package a2a

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kagent-dev/kagent/go/controller/api/v1alpha1"
	autogen_client "github.com/kagent-dev/kagent/go/controller/internal/autogen/client"
	"github.com/kagent-dev/kagent/go/controller/internal/database"
	common "github.com/kagent-dev/kagent/go/controller/internal/utils"
	"trpc.group/trpc-go/trpc-a2a-go/server"
)

// translates A2A Handlers from autogen agents/teams
type AutogenA2ATranslator interface {
	TranslateHandlerForAgent(
		ctx context.Context,
		agent *v1alpha1.Agent,
		autogenTeam *autogen_client.Team,
	) (*A2AHandlerParams, error)
}

type autogenA2ATranslator struct {
	a2aBaseUrl    string
	autogenClient autogen_client.Client
	dbService     database.Client
}

var _ AutogenA2ATranslator = &autogenA2ATranslator{}

func NewAutogenA2ATranslator(
	a2aBaseUrl string,
	autogenClient autogen_client.Client,
	dbService database.Client,
) AutogenA2ATranslator {
	return &autogenA2ATranslator{
		a2aBaseUrl:    a2aBaseUrl,
		autogenClient: autogenClient,
		dbService:     dbService,
	}
}

func (a *autogenA2ATranslator) TranslateHandlerForAgent(
	ctx context.Context,
	agent *v1alpha1.Agent,
	autogenTeam *autogen_client.Team,
) (*A2AHandlerParams, error) {
	card, err := a.translateCardForAgent(agent)
	if err != nil {
		return nil, err
	}
	if card == nil {
		return nil, nil
	}

	handler, err := a.makeHandlerForTeam(autogenTeam)
	if err != nil {
		return nil, err
	}

	return &A2AHandlerParams{
		AgentCard:  *card,
		HandleTask: handler,
	}, nil
}

func (a *autogenA2ATranslator) translateCardForAgent(
	agent *v1alpha1.Agent,
) (*server.AgentCard, error) {
	a2AConfig := agent.Spec.A2AConfig
	if a2AConfig == nil {
		return nil, nil
	}

	agentRef := common.GetObjectRef(agent)

	skills := a2AConfig.Skills
	if len(skills) == 0 {
		return nil, fmt.Errorf("no skills found for agent %s", agentRef)
	}

	var convertedSkills []server.AgentSkill
	for _, skill := range skills {
		convertedSkills = append(convertedSkills, server.AgentSkill(skill))
	}

	return &server.AgentCard{
		Name:        agentRef,
		Description: common.MakePtr(agent.Spec.Description),
		URL:         fmt.Sprintf("%s/%s", a.a2aBaseUrl, agentRef),
		//Provider:           nil,
		Version: fmt.Sprintf("%v", agent.Generation),
		//DocumentationURL:   nil,
		//Capabilities:       server.AgentCapabilities{},
		//Authentication:     nil,
		DefaultInputModes:  []string{"text"},
		DefaultOutputModes: []string{"text"},
		Skills:             convertedSkills,
	}, nil
}

func (a *autogenA2ATranslator) makeHandlerForTeam(
	autogenTeam *autogen_client.Team,
) (TaskHandler, error) {
	return func(ctx context.Context, task string, sessionID *string) (string, error) {
		var taskResult *autogen_client.TaskResult
		if sessionID != nil && *sessionID != "" {
			session, err := a.dbService.Session.Get(database.Clause{
				Key:   "user_id",
				Value: common.GetGlobalUserID(),
			}, database.Clause{
				Key:   "name",
				Value: *sessionID,
			})
			if err != nil {
				return "", fmt.Errorf("failed to get session: %w", err)
			}
			if err != nil {
				if errors.Is(err, autogen_client.NotFoundError) {
					session = &database.Session{
						Name: *sessionID,
					}
					err := a.dbService.Session.Create(session)
					if err != nil {
						return "", fmt.Errorf("failed to create session: %w", err)
					}
				} else {
					return "", fmt.Errorf("failed to get session: %w", err)
				}
			}
			resp, err := a.autogenClient.InvokeTask(session.ID, common.GetGlobalUserID(), &autogen_client.InvokeRequest{
				Task:       task,
				TeamConfig: autogenTeam.Component,
			})
			if err != nil {
				return "", fmt.Errorf("failed to invoke task: %w", err)
			}
			taskResult = &resp.TaskResult
		} else {

			resp, err := a.autogenClient.InvokeTask(&autogen_client.InvokeTaskRequest{
				Task:       task,
				TeamConfig: autogenTeam.Component,
			})
			if err != nil {
				return "", fmt.Errorf("failed to invoke task: %w", err)
			}
			taskResult = &resp.TaskResult
		}

		var lastMessageContent string
		for _, msg := range taskResult.Messages {
			switch msg["content"].(type) {
			case string:
				lastMessageContent = msg["content"].(string)
			default:
				b, err := json.Marshal(msg["content"])
				if err != nil {
					return "", fmt.Errorf("failed to marshal message content: %w", err)
				}
				lastMessageContent = string(b)
			}
		}

		return lastMessageContent, nil
	}, nil
}
