package autogen

import (
	"context"
	"fmt"

	autogen_client "github.com/kagent-dev/kagent/go/autogen/client"
	"github.com/kagent-dev/kagent/go/controller/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AutogenReconciler interface {
	ReconcileAutogenAgent(ctx context.Context, req ctrl.Request) error
	ReconcileAutogenModelConfig(ctx context.Context, req ctrl.Request) error
	ReconcileAutogenTeam(ctx context.Context, req ctrl.Request) error
	ReconcileAutogenApiKeySecret(ctx context.Context, req ctrl.Request) error
}

type autogenReconciler struct {
	translator ApiTranslator

	kube          client.Client
	autogenClient *autogen_client.Client

	defaultModelConfig types.NamespacedName
}

func NewAutogenReconciler(
	translator ApiTranslator,
	kube client.Client,
	autogenClient *autogen_client.Client,
	defaultModelConfig types.NamespacedName,
) AutogenReconciler {
	return &autogenReconciler{
		translator:         translator,
		kube:               kube,
		autogenClient:      autogenClient,
		defaultModelConfig: defaultModelConfig,
	}
}

func (a *autogenReconciler) ReconcileAutogenAgent(ctx context.Context, req ctrl.Request) error {
	// reconcile the agent team itself
	agent := &v1alpha1.Agent{}
	if err := a.kube.Get(ctx, req.NamespacedName, agent); err != nil {
		return fmt.Errorf("failed to get agent %s: %v", req.Name, err)
	}
	if err := a.reconcileAgents(ctx, agent); err != nil {
		return fmt.Errorf("failed to reconcile agent %s: %v", req.Name, err)
	}

	// find and reconcile all teams which use this agent
	teams, err := a.findTeamsUsingAgent(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to find teams for agent %s: %v", req.Name, err)
	}

	return a.reconcileTeams(ctx, teams...)
}

func (a *autogenReconciler) ReconcileAutogenModelConfig(ctx context.Context, req ctrl.Request) error {
	agents, err := a.findAgentsUsingModel(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to find agents for model %s: %v", req.Name, err)
	}

	if err := a.reconcileAgents(ctx, agents...); err != nil {
		return fmt.Errorf("failed to reconcile agents for model %s: %v", req.Name, err)
	}

	teams, err := a.findTeamsUsingModel(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to find teams for model %s: %v", req.Name, err)
	}

	return a.reconcileTeams(ctx, teams...)
}

func (a *autogenReconciler) ReconcileAutogenTeam(ctx context.Context, req ctrl.Request) error {
	team := &v1alpha1.Team{}
	if err := a.kube.Get(ctx, req.NamespacedName, team); err != nil {
		return fmt.Errorf("failed to get team %s: %v", req.Name, err)
	}

	return a.reconcileTeams(ctx, team)
}

func (a *autogenReconciler) ReconcileAutogenApiKeySecret(ctx context.Context, req ctrl.Request) error {
	agents, err := a.findAgentsUsingApiKeySecret(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to find agents for secret %s: %v", req.Name, err)
	}

	if err := a.reconcileAgents(ctx, agents...); err != nil {
		return fmt.Errorf("failed to reconcile agents for secret %s: %v", req.Name, err)
	}

	teams, err := a.findTeamsUsingApiKeySecret(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to find teams for api key secret %s: %v", req.Name, err)
	}

	return a.reconcileTeams(ctx, teams...)
}

func (a *autogenReconciler) reconcileTeams(ctx context.Context, teams ...*v1alpha1.Team) error {
	errs := map[types.NamespacedName]error{}
	for _, team := range teams {
		autogenTeam, err := a.translator.TranslateGroupChatForTeam(ctx, team)
		if err != nil {
			errs[types.NamespacedName{Name: team.Name, Namespace: team.Namespace}] = fmt.Errorf("failed to translate team %s: %v", team.Name, err)
			continue
		}
		if err := a.upsertTeam(autogenTeam); err != nil {
			errs[types.NamespacedName{Name: team.Name, Namespace: team.Namespace}] = fmt.Errorf("failed to upsert team %s: %v", team.Name, err)
			continue
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to reconcile teams: %v", errs)
	}

	return nil
}

func (a *autogenReconciler) reconcileAgents(ctx context.Context, agents ...*v1alpha1.Agent) error {
	errs := map[types.NamespacedName]error{}
	for _, agent := range agents {
		autogenTeam, err := a.translator.TranslateGroupChatForAgent(ctx, agent)
		if err != nil {
			errs[types.NamespacedName{Name: agent.Name, Namespace: agent.Namespace}] = fmt.Errorf("failed to translate agent %s: %v", agent.Name, err)
			continue
		}
		if err := a.upsertTeam(autogenTeam); err != nil {
			errs[types.NamespacedName{Name: agent.Name, Namespace: agent.Namespace}] = fmt.Errorf("failed to upsert agent %s: %v", agent.Name, err)
			continue
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to reconcile agents: %v", errs)
	}

	return nil
}

func (a *autogenReconciler) upsertTeam(team *autogen_client.Team) error {
	// delete if team exists
	existingTeam, err := a.autogenClient.GetTeam(*team.Component.Label, GlobalUserID)
	if err != nil {
		return fmt.Errorf("failed to get existing team %s: %v", *team.Component.Label, err)
	}
	if existingTeam != nil {
		err = a.autogenClient.DeleteTeam(existingTeam.Id, GlobalUserID)
		if err != nil {
			return fmt.Errorf("failed to delete existing team %s: %v", *team.Component.Label, err)
		}
	}

	// validate the team
	req := autogen_client.ValidationRequest{
		Component: team.Component,
	}
	resp, err := a.autogenClient.Validate(&req)
	if err != nil {
		return fmt.Errorf("failed to validate team %s: %v", *team.Component.Label, err)
	}
	if !resp.IsValid {
		return fmt.Errorf("team %s is invalid: %v", *team.Component.Label, resp.Errors)
	}

	return a.autogenClient.CreateTeam(team)
}

func (a *autogenReconciler) findAgentsUsingModel(ctx context.Context, req ctrl.Request) ([]*v1alpha1.Agent, error) {
	var agentsList v1alpha1.AgentList
	if err := a.kube.List(
		ctx,
		&agentsList,
		client.InNamespace(req.Namespace),
	); err != nil {
		return nil, fmt.Errorf("failed to list agents: %v", err)
	}

	var agents []*v1alpha1.Agent
	appendAgentIfUsesModel := func(agent *v1alpha1.Agent) {
		// TODO currently all agents use the default model config
		// eventually we will want to support per-agent overrides
		// then we will want to update this to check the agent's spec
		if a.defaultModelConfig.Name == req.Name && a.defaultModelConfig.Namespace == req.Namespace {
			agents = append(agents, agent)
		}
	}
	for _, agent := range agentsList.Items {
		agent := agent
		appendAgentIfUsesModel(&agent)
	}

	return agents, nil
}

func (a *autogenReconciler) findAgentsUsingApiKeySecret(ctx context.Context, req ctrl.Request) ([]*v1alpha1.Agent, error) {
	var modelsList v1alpha1.ModelConfigList
	if err := a.kube.List(
		ctx,
		&modelsList,
		client.InNamespace(req.Namespace),
	); err != nil {
		return nil, fmt.Errorf("failed to list secrets: %v", err)
	}

	var models []string
	appendModelIfUsesApiKeySecret := func(model v1alpha1.ModelConfig) {
		if model.Spec.APIKeySecretName == req.Name {
			models = append(models, model.Name)
		}
	}
	for _, model := range modelsList.Items {
		appendModelIfUsesApiKeySecret(model)
	}

	var agents []*v1alpha1.Agent
	appendUniqueAgent := func(agent *v1alpha1.Agent) {
		for _, t := range agents {
			if t.Name == agent.Name {
				return
			}
		}
		agents = append(agents, agent)
	}

	for _, model := range models {
		agentsUsingModel, err := a.findAgentsUsingModel(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: req.Namespace,
				Name:      model,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to find agents for model %s: %v", model, err)
		}
		for _, agent := range agentsUsingModel {
			appendUniqueAgent(agent)
		}
	}

	return agents, nil

}

func (a *autogenReconciler) findTeamsUsingAgent(ctx context.Context, req ctrl.Request) ([]*v1alpha1.Team, error) {
	var teamsList v1alpha1.TeamList
	if err := a.kube.List(
		ctx,
		&teamsList,
		client.InNamespace(req.Namespace),
	); err != nil {
		return nil, fmt.Errorf("failed to list teams: %v", err)
	}

	var teams []*v1alpha1.Team
	appendTeamIfUsesAgent := func(team *v1alpha1.Team) {
		for _, participant := range team.Spec.Participants {
			if participant == req.Name {
				teams = append(teams, team)
				break
			}
		}
	}
	for _, team := range teamsList.Items {
		team := team
		appendTeamIfUsesAgent(&team)
	}

	return teams, nil
}

func (a *autogenReconciler) findTeamsUsingModel(ctx context.Context, req ctrl.Request) ([]*v1alpha1.Team, error) {
	var teamsList v1alpha1.TeamList
	if err := a.kube.List(
		ctx,
		&teamsList,
		client.InNamespace(req.Namespace),
	); err != nil {
		return nil, fmt.Errorf("failed to list teams: %v", err)
	}

	var teams []*v1alpha1.Team
	appendTeamIfUsesModel := func(team *v1alpha1.Team) {
		if team.Spec.ModelConfig == req.Name {
			teams = append(teams, team)
		}
	}
	for _, team := range teamsList.Items {
		team := team
		appendTeamIfUsesModel(&team)
	}

	return teams, nil
}

func (a *autogenReconciler) findTeamsUsingApiKeySecret(ctx context.Context, req ctrl.Request) ([]*v1alpha1.Team, error) {
	var modelsList v1alpha1.ModelConfigList
	if err := a.kube.List(
		ctx,
		&modelsList,
		client.InNamespace(req.Namespace),
	); err != nil {
		return nil, fmt.Errorf("failed to list secrets: %v", err)
	}

	var models []string
	appendModelIfUsesApiKeySecret := func(model v1alpha1.ModelConfig) {
		if model.Spec.APIKeySecretName == req.Name {
			models = append(models, model.Name)
		}
	}
	for _, model := range modelsList.Items {
		appendModelIfUsesApiKeySecret(model)
	}

	var teams []*v1alpha1.Team
	appendUniqueTeam := func(team *v1alpha1.Team) {
		for _, t := range teams {
			if t.Name == team.Name {
				return
			}
		}
		teams = append(teams, team)
	}

	for _, model := range models {
		teamsUsingModel, err := a.findTeamsUsingModel(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: req.Namespace,
				Name:      model,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to find teams for model %s: %v", model, err)
		}
		for _, team := range teamsUsingModel {
			appendUniqueTeam(team)
		}
	}

	return teams, nil

}
