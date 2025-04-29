package a2a

import (
	"context"
	"fmt"
	autogen_client "github.com/kagent-dev/kagent/go/autogen/client"
	"github.com/kagent-dev/kagent/go/controller/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	reconcileLog = ctrl.Log.WithName("a2a_reconcile")
)

type A2AReconciler interface {
	ReconcileAutogenAgent(
		ctx context.Context,
		agent *v1alpha1.Agent,
		autogenTeam *autogen_client.Team,
	) error
}

type a2aReconciler struct {
	a2aTranslator AutogenA2ATranslator
	autogenClient *autogen_client.Client
	a2aHandler    A2AHandlerMux
}

func NewAutogenReconciler(
	autogenClient *autogen_client.Client,
	a2aHandler A2AHandlerMux,
	a2aBaseUrl string,
) A2AReconciler {
	return &a2aReconciler{
		a2aTranslator: NewAutogenA2ATranslator(a2aBaseUrl, autogenClient),
		autogenClient: autogenClient,
		a2aHandler:    a2aHandler,
	}
}

func (a *a2aReconciler) ReconcileAutogenAgent(
	ctx context.Context,
	agent *v1alpha1.Agent,
	autogenTeam *autogen_client.Team,
) error {
	params, err := a.a2aTranslator.TranslateHandlerForAgent(ctx, agent, autogenTeam)
	if err != nil {
		return err
	}

	return a.a2aHandler.SetAgentHandler(
		fmt.Sprintf("%s_%s", agent.Namespace, agent.Name),
		params,
	)
}
