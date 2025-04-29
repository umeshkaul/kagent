package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	autogen_client "github.com/kagent-dev/kagent/go/autogen/client"
	"github.com/kagent-dev/kagent/go/controller/api/v1alpha1"
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
	autogenClient *autogen_client.Client
}

var _ AutogenA2ATranslator = &autogenA2ATranslator{}

func NewAutogenA2ATranslator(
	a2aBaseUrl string,
	autogenClient *autogen_client.Client,
) AutogenA2ATranslator {
	return &autogenA2ATranslator{
		a2aBaseUrl:    a2aBaseUrl,
		autogenClient: autogenClient,
	}
}

func (a *autogenA2ATranslator) TranslateHandlerForAgent(ctx context.Context, agent *v1alpha1.Agent, autogenTeam *autogen_client.Team) (*A2AHandlerParams, error) {
	card, err := a.translateCardForAgent(ctx, agent)
	if err != nil {
		return nil, err
	}

	handler, err := a.makeHandlerForTeam(ctx, autogenTeam)
	if err != nil {
		return nil, err
	}

	return &A2AHandlerParams{
		AgentCard:  *card,
		HandleTask: handler,
	}, nil
}

func (a *autogenA2ATranslator) translateCardForAgent(
	ctx context.Context,
	agent *v1alpha1.Agent,
) (*server.AgentCard, error) {

	return &server.AgentCard{
		Name:        agent.Name,
		Description: common.MakePtr(agent.Spec.Description),
		URL:         fmt.Sprintf("%s/%s", a.a2aBaseUrl, agent.Name),
		//Provider:           nil,
		Version: fmt.Sprintf("%v", agent.Generation),
		//DocumentationURL:   nil,
		//Capabilities:       server.AgentCapabilities{},
		//Authentication:     nil,
		DefaultInputModes:  []string{"text"},
		DefaultOutputModes: []string{"text"},
		//Skills:             nil,
	}, nil
}

func (a *autogenA2ATranslator) makeHandlerForTeam(
	ctx context.Context,
	autogenTeam *autogen_client.Team,
) (TaskHandler, error) {
	return func(ctx context.Context, task string) (string, error) {
		resp, err := a.autogenClient.InvokeTask(&autogen_client.InvokeTaskRequest{
			Task:       task,
			TeamConfig: autogenTeam.Component,
		})
		if err != nil {
			return "", nil
		}

		if !resp.Status {
			return "", fmt.Errorf("failed to invoke task: %s", resp.Message)
		}

		b, err := json.Marshal(resp.Data)

		return string(b), err
	}, nil
}
