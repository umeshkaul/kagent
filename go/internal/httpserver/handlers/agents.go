package handlers

import (
	"fmt"
	"net/http"

	"github.com/kagent-dev/kagent/go/internal/database"
	"github.com/kagent-dev/kagent/go/internal/httpserver/errors"
	"github.com/kagent-dev/kagent/go/internal/utils"
	"github.com/kagent-dev/kagent/go/pkg/client/api"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// AgentsHandler handles agent-related requests
type AgentsHandler struct {
	*Base
}

// NewAgentsHandler creates a new AgentsHandler
func NewAgentsHandler(base *Base) *AgentsHandler {
	return &AgentsHandler{Base: base}
}

// HandleListAgents handles GET /api/agents requests using database
func (h *AgentsHandler) HandleListAgents(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("agents-handler").WithValues("operation", "list-db")

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	log.V(1).Info("Listing agents from database")
	agents, err := h.DatabaseService.ListAgents(userID)
	if err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to list agents", err))
		return
	}

	log.Info("Successfully listed agents", "count", len(agents))
	data := api.NewResponse(agents, "Successfully listed agents", false)
	RespondWithJSON(w, http.StatusOK, data)
}

// HandleGetAgent handles GET /api/agents/{namespace}/{name} requests using database
func (h *AgentsHandler) HandleGetAgent(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("agents-handler").WithValues("operation", "get-db")

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	agentName, err := GetPathParam(r, "name")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get name from path", err))
		return
	}
	log = log.WithValues("agentName", agentName)

	agentNamespace, err := GetPathParam(r, "namespace")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get namespace from path", err))
		return
	}
	log = log.WithValues("agentNamespace", agentNamespace)

	log.V(1).Info("Getting agent from database")
	agent, err := h.DatabaseService.GetAgent(fmt.Sprintf("%s/%s", agentNamespace, agentName))
	if err != nil {
		w.RespondWithError(errors.NewNotFoundError("Agent not found", err))
		return
	}

	log.Info("Successfully retrieved agent")
	data := api.NewResponse(agent, "Successfully retrieved agent", false)
	RespondWithJSON(w, http.StatusOK, data)
}

// HandleCreateAgent handles POST /api/agents requests using database
func (h *AgentsHandler) HandleCreateAgent(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("agents-handler").WithValues("operation", "create-db")

	var teamRequest api.AgentRequest
	if err := DecodeJSONBody(r, &teamRequest); err != nil {
		w.RespondWithError(errors.NewBadRequestError("Invalid request body", err))
		return
	}
	log = log.WithValues("agentRef", teamRequest.AgentRef)

	agent := &database.Agent{
		Component: teamRequest.Component,
	}

	log.V(1).Info("Creating agent in database")
	if err := h.DatabaseService.CreateAgent(agent); err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to create agent", err))
		return
	}

	log.Info("Successfully created agent", "agentID", agent.ID)
	data := api.NewResponse(agent, "Successfully created agent", false)
	RespondWithJSON(w, http.StatusCreated, data)
}

// HandleUpdateAgent handles PUT /api/agents/{namespace}/{name} requests using database
func (h *AgentsHandler) HandleUpdateAgent(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("agents-handler").WithValues("operation", "update-db")

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	var teamRequest api.AgentRequest
	if err := DecodeJSONBody(r, &teamRequest); err != nil {
		w.RespondWithError(errors.NewBadRequestError("Invalid request body", err))
		return
	}

	nns, err := utils.ParseRefString(teamRequest.AgentRef, utils.GetResourceNamespace())
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Invalid agent ref", err))
		return
	}

	// Get existing agent
	agent, err := h.DatabaseService.GetAgent(nns.String())
	if err != nil {
		w.RespondWithError(errors.NewNotFoundError("Agent not found", err))
		return
	}

	agent.Component = teamRequest.Component

	if err := h.DatabaseService.UpdateAgent(agent); err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to update agent", err))
		return
	}

	log.Info("Successfully updated agent")
	data := api.NewResponse(agent, "Successfully updated agent", false)
	RespondWithJSON(w, http.StatusOK, data)
}

// HandleDeleteAgent handles DELETE /api/agents/{namespace}/{name} requests using database
func (h *AgentsHandler) HandleDeleteAgent(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("agents-handler").WithValues("operation", "delete-db")

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	agentName, err := GetPathParam(r, "name")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get name from path", err))
		return
	}
	log = log.WithValues("agentName", agentName)

	agentNamespace, err := GetPathParam(r, "namespace")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get namespace from path", err))
		return
	}
	log = log.WithValues("agentNamespace", agentNamespace)

	if err := h.DatabaseService.DeleteAgent(fmt.Sprintf("%s/%s", agentNamespace, agentName)); err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to delete agent", err))
		return
	}

	log.Info("Successfully deleted agent")
	data := api.NewResponse(struct{}{}, "Successfully deleted agent", false)
	RespondWithJSON(w, http.StatusOK, data)
}
