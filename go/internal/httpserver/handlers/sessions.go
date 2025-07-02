package handlers

import (
	"net/http"

	"github.com/kagent-dev/kagent/go/internal/database"
	"github.com/kagent-dev/kagent/go/internal/httpserver/errors"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// SessionsHandler handles session-related requests
type SessionsHandler struct {
	*Base
}

// NewSessionsHandler creates a new SessionsHandler
func NewSessionsHandler(base *Base) *SessionsHandler {
	return &SessionsHandler{Base: base}
}

// SessionRequest represents a session creation/update request
type SessionRequest struct {
	TeamID *uint  `json:"team_id,omitempty"`
	Name   string `json:"name"`
	UserID string `json:"user_id"`
}

// RunRequest represents a run creation request
type RunRequest struct {
	Task string `json:"task"`
}

// HandleListSessions handles GET /api/sessions requests using database
func (h *SessionsHandler) HandleListSessions(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("sessions-handler").WithValues("operation", "list-db")

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	log.V(1).Info("Listing sessions from database")
	sessions, err := h.DatabaseService.ListSessions(userID)
	if err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to list sessions", err))
		return
	}

	log.Info("Successfully listed sessions", "count", len(sessions))
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": true,
		"data":   sessions,
	})
}

// HandleCreateSession handles POST /api/sessions requests using database
func (h *SessionsHandler) HandleCreateSession(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("sessions-handler").WithValues("operation", "create-db")

	var sessionRequest SessionRequest
	if err := DecodeJSONBody(r, &sessionRequest); err != nil {
		w.RespondWithError(errors.NewBadRequestError("Invalid request body", err))
		return
	}

	if sessionRequest.UserID == "" {
		w.RespondWithError(errors.NewBadRequestError("user_id is required", nil))
		return
	}
	log = log.WithValues("userID", sessionRequest.UserID)

	session := &database.Session{
		UserID: sessionRequest.UserID,
		TeamID: sessionRequest.TeamID,
		Name:   sessionRequest.Name,
	}

	log.V(1).Info("Creating session in database",
		"teamID", sessionRequest.TeamID,
		"name", sessionRequest.Name)

	if err := h.DatabaseService.CreateSession(session); err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to create session", err))
		return
	}

	log.Info("Successfully created session", "sessionID", session.ID)
	RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"status":  true,
		"data":    session,
		"message": "Session created successfully",
	})
}

// HandleGetSession handles GET /api/sessions/{sessionName} requests using database
func (h *SessionsHandler) HandleGetSession(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("sessions-handler").WithValues("operation", "get-db")

	sessionName, err := GetPathParam(r, "session_name")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get session name from path", err))
		return
	}
	log = log.WithValues("session_name", sessionName)

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	log.V(1).Info("Getting session from database")
	session, err := h.DatabaseService.GetSession(sessionName, userID)
	if err != nil {
		w.RespondWithError(errors.NewNotFoundError("Session not found", err))
		return
	}

	log.Info("Successfully retrieved session")
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": true,
		"data":   session,
	})
}

// HandleUpdateSession handles PUT /api/sessions requests using database
func (h *SessionsHandler) HandleUpdateSession(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("sessions-handler").WithValues("operation", "update-db")

	var sessionRequest SessionRequest
	if err := DecodeJSONBody(r, &sessionRequest); err != nil {
		w.RespondWithError(errors.NewBadRequestError("Invalid request body", err))
		return
	}

	// Get existing session
	session, err := h.DatabaseService.GetSession(sessionRequest.Name, sessionRequest.UserID)
	if err != nil {
		w.RespondWithError(errors.NewNotFoundError("Session not found", err))
		return
	}

	// Update fields
	session.TeamID = sessionRequest.TeamID

	if err := h.DatabaseService.UpdateSession(session); err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to update session", err))
		return
	}

	log.Info("Successfully updated session")
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"data":    session,
		"message": "Session updated successfully",
	})
}

// HandleDeleteSession handles DELETE /api/sessions/{sessionName} requests using database
func (h *SessionsHandler) HandleDeleteSession(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("sessions-handler").WithValues("operation", "delete-db")

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	sessionName, err := GetPathParam(r, "session_name")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get session ID from path", err))
		return
	}
	log = log.WithValues("session_name", sessionName)

	if err := h.DatabaseService.DeleteSession(sessionName, userID); err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to delete session", err))
		return
	}

	log.Info("Successfully deleted session")
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "Session deleted successfully",
	})
}

// HandleListSessionRuns handles GET /api/sessions/{sessionName}/runs requests using database
func (h *SessionsHandler) HandleListSessionRuns(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("sessions-handler").WithValues("operation", "list-runs-db")

	sessionName, err := GetPathParam(r, "session_name")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get session ID from path", err))
		return
	}
	log = log.WithValues("session_name", sessionName)

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	log.V(1).Info("Getting session runs from database")
	runs, err := h.DatabaseService.ListSessionRuns(sessionName, userID)
	if err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to get session runs", err))
		return
	}

	// Build response with messages per run
	runData := make([]map[string]interface{}, 0, len(runs))
	for _, run := range runs {
		// Get messages for this run
		messages, err := h.DatabaseService.ListMessagesForRun(run.ID)
		if err != nil {
			log.Error(err, "Failed to get messages for run", "runID", run.ID)
			messages = []database.Message{} // Continue with empty messages
		}

		runData = append(runData, map[string]interface{}{
			"id":          run.ID,
			"created_at":  run.CreatedAt,
			"status":      run.Status,
			"task":        run.Task,
			"team_result": run.TeamResult,
			"messages":    messages,
		})
	}

	log.Info("Successfully retrieved session runs", "count", len(runs))
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": true,
		"data":   map[string]interface{}{"runs": runData},
	})
}
