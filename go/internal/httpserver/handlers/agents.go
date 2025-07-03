package handlers

import (
	"fmt"
	"net/http"

	"github.com/kagent-dev/kagent/go/client"
	"github.com/kagent-dev/kagent/go/internal/database"
	"github.com/kagent-dev/kagent/go/internal/httpserver/errors"
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
	teams, err := h.DatabaseService.ListTeams(userID)
	if err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to list teams", err))
		return
	}

	log.Info("Successfully listed teams", "count", len(teams))
	data := client.NewResponse(teams, "Successfully listed teams", false)
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

	log.V(1).Info("Getting team from database")
	team, err := h.DatabaseService.GetTeam(fmt.Sprintf("%s/%s", agentNamespace, agentName))
	if err != nil {
		w.RespondWithError(errors.NewNotFoundError("Team not found", err))
		return
	}

	log.Info("Successfully retrieved team")
	data := client.NewResponse(team, "Successfully retrieved team", false)
	RespondWithJSON(w, http.StatusOK, data)
}

// HandleCreateAgent handles POST /api/agents requests using database
func (h *AgentsHandler) HandleCreateAgent(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("agents-handler").WithValues("operation", "create-db")

	var teamRequest client.TeamRequest
	if err := DecodeJSONBody(r, &teamRequest); err != nil {
		w.RespondWithError(errors.NewBadRequestError("Invalid request body", err))
		return
	}
	log = log.WithValues("agentRef", teamRequest.AgentRef)

	team := &database.Team{
		Component: teamRequest.Component,
	}

	log.V(1).Info("Creating team in database")
	if err := h.DatabaseService.CreateTeam(team); err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to create team", err))
		return
	}

	log.Info("Successfully created team", "teamID", team.ID)
	data := client.NewResponse(team, "Successfully created team", false)
	RespondWithJSON(w, http.StatusCreated, data)
}

// HandleUpdateAgent handles PUT /api/agents/{namespace}/{name} requests using database
func (h *AgentsHandler) HandleUpdateAgent(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("agents-handler").WithValues("operation", "update-db")

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

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	var teamRequest client.TeamRequest
	if err := DecodeJSONBody(r, &teamRequest); err != nil {
		w.RespondWithError(errors.NewBadRequestError("Invalid request body", err))
		return
	}

	// Get existing team
	team, err := h.DatabaseService.GetTeam(fmt.Sprintf("%s/%s", agentNamespace, agentName))
	if err != nil {
		w.RespondWithError(errors.NewNotFoundError("Team not found", err))
		return
	}

	team.Component = teamRequest.Component

	if err := h.DatabaseService.UpdateTeam(team); err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to update team", err))
		return
	}

	log.Info("Successfully updated team")
	data := client.NewResponse(team, "Successfully updated team", false)
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

	if err := h.DatabaseService.DeleteTeam(fmt.Sprintf("%s/%s", agentNamespace, agentName)); err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to delete team", err))
		return
	}

	log.Info("Successfully deleted team")
	data := client.NewResponse(struct{}{}, "Successfully deleted team", false)
	RespondWithJSON(w, http.StatusOK, data)
}
