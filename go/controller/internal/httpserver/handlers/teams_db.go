package handlers

import (
	"net/http"

	"github.com/kagent-dev/kagent/go/controller/internal/database"
	"github.com/kagent-dev/kagent/go/controller/internal/httpserver/errors"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// TeamRequest represents a team creation/update request
type TeamRequest struct {
	UserID    string                 `json:"user_id"`
	Component map[string]interface{} `json:"component"`
}

// HandleListTeamsDB handles GET /api/teams requests using database
func (h *TeamsHandler) HandleListTeamsDB(w ErrorResponseWriter, r *http.Request) {
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

// HandleGetTeamDB handles GET /api/teams/{teamID} requests using database
func (h *TeamsHandler) HandleGetTeamDB(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("teams-handler").WithValues("operation", "get-db")

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	teamID, err := GetIntPathParam(r, "teamID")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get team ID from path", err))
		return
	}
	log = log.WithValues("teamID", teamID)

	log.V(1).Info("Getting team from database")
	team, err := h.DatabaseService.GetTeam(uint(teamID), userID)
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

// HandleCreateTeamDB handles POST /api/teams requests using database
func (h *TeamsHandler) HandleCreateTeamDB(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("teams-handler").WithValues("operation", "create-db")

	var teamRequest TeamRequest
	if err := DecodeJSONBody(r, &teamRequest); err != nil {
		w.RespondWithError(errors.NewBadRequestError("Invalid request body", err))
		return
	}

	if teamRequest.UserID == "" {
		w.RespondWithError(errors.NewBadRequestError("user_id is required", nil))
		return
	}
	log = log.WithValues("userID", teamRequest.UserID)

	team := &database.Team{
		BaseModel: database.BaseModel{
			UserID: &teamRequest.UserID,
		},
		Component: database.JSONMap(teamRequest.Component),
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

// HandleUpdateTeamDB handles PUT /api/teams/{teamID} requests using database
func (h *TeamsHandler) HandleUpdateTeamDB(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("teams-handler").WithValues("operation", "update-db")

	teamID, err := GetIntPathParam(r, "teamID")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get team ID from path", err))
		return
	}
	log = log.WithValues("teamID", teamID)

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
	team, err := h.DatabaseService.GetTeam(uint(teamID), userID)
	if err != nil {
		w.RespondWithError(errors.NewNotFoundError("Team not found", err))
		return
	}

	// Update component
	if teamRequest.Component != nil {
		team.Component = database.JSONMap(teamRequest.Component)
	}

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

// HandleDeleteTeamDB handles DELETE /api/teams/{teamID} requests using database
func (h *TeamsHandler) HandleDeleteTeamDB(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("teams-handler").WithValues("operation", "delete-db")

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	teamID, err := GetIntPathParam(r, "teamID")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get team ID from path", err))
		return
	}
	log = log.WithValues("teamID", teamID)

	if err := h.DatabaseService.DeleteTeam(uint(teamID), userID); err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to delete team", err))
		return
	}

	log.Info("Successfully deleted team")
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "Team deleted successfully",
	})
}
