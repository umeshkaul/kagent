package handlers

import (
	"fmt"
	"net/http"

	"github.com/kagent-dev/kagent/go/httpserver/server/errors"
	"github.com/kagent-dev/kagent/go/internal/autogen/api"
	"github.com/kagent-dev/kagent/go/internal/database"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// TeamsHandler handles team-related requests
type TeamsHandler struct {
	*Base
}

// NewTeamsHandler creates a new TeamsHandler
func NewTeamsHandler(base *Base) *TeamsHandler {
	return &TeamsHandler{Base: base}
}

// TeamRequest represents a team creation/update request
type TeamRequest struct {
	AgentRef  string        `json:"agent_ref"`
	Component api.Component `json:"component"`
}

// HandleListTeams handles GET /api/teams requests using database
func (h *TeamsHandler) HandleListTeams(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("teams-handler").WithValues("operation", "list-db")

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	log.V(1).Info("Listing teams from database")
	teams, err := h.DatabaseService.ListTeams(userID)
	if err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to list teams", err))
		return
	}

	log.Info("Successfully listed teams", "count", len(teams))
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": true,
		"data":   teams,
	})
}

// HandleGetTeam handles GET /api/teams/{agent_name}/{agent_namespace} requests using database
func (h *TeamsHandler) HandleGetTeam(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("teams-handler").WithValues("operation", "get-db")

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	agentName, err := GetPathParam(r, "agent_name")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get team ID from path", err))
		return
	}
	log = log.WithValues("agentName", agentName)

	agentNamespace, err := GetPathParam(r, "agent_namespace")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get agent namespace from path", err))
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
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": true,
		"data":   team,
	})
}

// HandleCreateTeam handles POST /api/teams requests using database
func (h *TeamsHandler) HandleCreateTeam(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("teams-handler").WithValues("operation", "create-db")

	var teamRequest TeamRequest
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
	RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"status":  true,
		"data":    team,
		"message": "Team created successfully",
	})
}

// HandleUpdateTeam handles PUT /api/teams/{agent_name}/{agent_namespace} requests using database
func (h *TeamsHandler) HandleUpdateTeam(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("teams-handler").WithValues("operation", "update-db")

	agentName, err := GetPathParam(r, "agent_name")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get team ID from path", err))
		return
	}
	log = log.WithValues("agentName", agentName)

	agentNamespace, err := GetPathParam(r, "agent_namespace")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get agent namespace from path", err))
		return
	}
	log = log.WithValues("agentNamespace", agentNamespace)

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	var teamRequest TeamRequest
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
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"data":    team,
		"message": "Team updated successfully",
	})
}

// HandleDeleteTeam handles DELETE /api/teams/{agent_name}/{agent_namespace} requests using database
func (h *TeamsHandler) HandleDeleteTeam(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("teams-handler").WithValues("operation", "delete-db")

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	agentName, err := GetPathParam(r, "agent_name")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get team ID from path", err))
		return
	}
	log = log.WithValues("agentName", agentName)

	agentNamespace, err := GetPathParam(r, "agent_namespace")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get agent namespace from path", err))
		return
	}
	log = log.WithValues("agentNamespace", agentNamespace)

	if err := h.DatabaseService.DeleteTeam(fmt.Sprintf("%s/%s", agentNamespace, agentName)); err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to delete team", err))
		return
	}

	log.Info("Successfully deleted team")
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "Team deleted successfully",
	})
}
