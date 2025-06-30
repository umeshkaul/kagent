package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/kagent-dev/kagent/go/httpserver/server/errors"
	"github.com/kagent-dev/kagent/go/internal/database"
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

// HandleListSessionsDB handles GET /api/sessions requests using database
func (h *SessionsHandler) HandleListSessionsDB(w ErrorResponseWriter, r *http.Request) {
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

// HandleCreateSessionDB handles POST /api/sessions requests using database
func (h *SessionsHandler) HandleCreateSessionDB(w ErrorResponseWriter, r *http.Request) {
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
		UserModel: database.UserModel{
			UserID: sessionRequest.UserID,
		},
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

// HandleGetSessionDB handles GET /api/sessions/{sessionName} requests using database
func (h *SessionsHandler) HandleGetSessionDB(w ErrorResponseWriter, r *http.Request) {
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

// HandleUpdateSessionDB handles PUT /api/sessions requests using database
func (h *SessionsHandler) HandleUpdateSessionDB(w ErrorResponseWriter, r *http.Request) {
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

// HandleDeleteSessionDB handles DELETE /api/sessions/{sessionName} requests using database
func (h *SessionsHandler) HandleDeleteSessionDB(w ErrorResponseWriter, r *http.Request) {
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

// HandleListSessionRunsDB handles GET /api/sessions/{sessionName}/runs requests using database
func (h *SessionsHandler) HandleListSessionRunsDB(w ErrorResponseWriter, r *http.Request) {
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

// HandleSessionInvokeDB handles POST /api/sessions/{sessionID}/invoke requests using database
func (h *SessionsHandler) HandleSessionInvokeDB(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("sessions-handler").WithValues("operation", "invoke-db")

	sessionID, err := GetIntPathParam(r, "sessionID")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get session ID from path", err))
		return
	}
	log = log.WithValues("sessionID", sessionID)

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	var runRequest RunRequest
	if err := DecodeJSONBody(r, &runRequest); err != nil {
		w.RespondWithError(errors.NewBadRequestError("Invalid request body", err))
		return
	}

	// Verify session exists and belongs to user
	session, err := h.DatabaseService.Session.Get(uint(sessionID), userID)
	if err != nil {
		w.RespondWithError(errors.NewNotFoundError("Session not found", err))
		return
	}

	// Create a new run
	run := &database.Run{
		BaseModel: database.BaseModel{
			UserID: &userID,
		},
		SessionID: session.ID,
		Status:    database.RunStatusCreated,
		Task: database.JSONMap{
			"content": runRequest.Task,
			"source":  "user",
		},
		TeamResult: database.JSONMap{},
		Messages:   database.JSONMap{},
	}

	if err := h.DatabaseService.CreateRun(run); err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to create run", err))
		return
	}

	// For now, we'll still use the autogen client for the actual execution
	// but store the results in the database
	result, err := h.AutogenClient.InvokeSession(sessionID, userID, runRequest.Task)
	if err != nil {
		// Update run status to error
		run.Status = database.RunStatusError
		errMsg := err.Error()
		run.ErrorMessage = &errMsg
		h.DatabaseService.UpdateRun(run)

		w.RespondWithError(errors.NewInternalServerError("Failed to invoke session", err))
		return
	}

	// Update run with results
	run.Status = database.RunStatusComplete
	if result != nil {
		resultBytes, _ := json.Marshal(result)
		var resultMap map[string]interface{}
		json.Unmarshal(resultBytes, &resultMap)
		run.TeamResult = database.JSONMap(resultMap)
	}

	if err := h.DatabaseService.UpdateRun(run); err != nil {
		log.Error(err, "Failed to update run with results")
	}

	log.Info("Successfully invoked session", "runID", run.ID)
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": true,
		"data":   result,
	})
}
