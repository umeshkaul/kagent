package handlers

import (
	"fmt"
	"net/http"

	autogen_client "github.com/kagent-dev/kagent/go/autogen/client"
	"github.com/kagent-dev/kagent/go/controller/internal/httpserver/errors"
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

func (h *SessionsHandler) HandleSessionInvokeStream(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("sessions-handler").WithValues("operation", "invoke-stream")

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

	var invokeRequest *autogen_client.InvokeRequest
	if err := DecodeJSONBody(r, &invokeRequest); err != nil {
		w.RespondWithError(errors.NewBadRequestError("Invalid request body", err))
		return
	}

	if invokeRequest.Task == "" {
		w.RespondWithError(errors.NewBadRequestError("task is required", nil))
		return
	}

	if invokeRequest.TeamConfig == nil {
		w.RespondWithError(errors.NewBadRequestError("team_config is required", nil))
		return
	}

	ch, err := h.AutogenClient.InvokeSessionStream(sessionID, userID, invokeRequest)
	if err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to invoke session", err))
		return
	}

	for event := range ch {
		w.Write([]byte(fmt.Sprintf("event: %s\ndata: %s\n\n", event.Event, event.Data)))
	}
}
