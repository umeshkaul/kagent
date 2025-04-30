package a2a

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kagent-dev/kagent/go/autogen/api"
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
	teamComponent, err := removeUserProxyParticipant(autogenTeam.Component)
	if err != nil {
		return nil, fmt.Errorf("failed to get team component: %w", err)
	}

	return func(ctx context.Context, task string) (string, error) {
		resp, err := a.autogenClient.InvokeTask(&autogen_client.InvokeTaskRequest{
			Task:       task,
			TeamConfig: teamComponent,
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

// TODO(ilackarms): remove this once we stop translating the user proxy agent
// this is a hack to remove the user proxy agent participant from the team component
func removeUserProxyParticipant(teamComponent *api.Component) (*api.Component, error) {
	teamConfig := &api.RoundRobinGroupChatConfig{}
	err := teamConfig.FromConfig(teamComponent.Config)
	if err != nil {
		return nil, err
	}

	for i, participant := range teamConfig.Participants {
		if participant.Provider == "autogen_agentchat.agents.UserProxyAgent" {
			teamConfig.Participants = append(teamConfig.Participants[:i], teamConfig.Participants[i+1:]...)
			break
		}
	}

	teamComponentConfig, err := teamConfig.ToConfig()
	if err != nil {
		return nil, err
	}

	return &api.Component{
		Provider:         teamComponent.Provider,
		ComponentType:    teamComponent.ComponentType,
		Version:          teamComponent.Version,
		ComponentVersion: teamComponent.ComponentVersion,
		Description:      teamComponent.Description,
		Label:            teamComponent.Label,
		Config:           teamComponentConfig,
	}, nil
}
