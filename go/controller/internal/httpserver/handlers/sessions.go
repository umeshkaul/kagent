package handlers

import (
	"fmt"
	"io"
	"net/http"

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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to read request body", err))
		return
	}

	ch, err := h.AutogenClient.InvokeSessionStream(sessionID, userID, string(body))
	if err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to invoke session", err))
		return
	}

	for event := range ch {
		w.Write([]byte(fmt.Sprintf("event: %s\ndata: %s\n\n", event.Event, event.Data)))
	}
}
