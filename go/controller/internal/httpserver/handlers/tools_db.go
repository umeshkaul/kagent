package handlers

import (
	"net/http"

	"github.com/kagent-dev/kagent/go/controller/internal/database"
	"github.com/kagent-dev/kagent/go/controller/internal/httpserver/errors"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// ToolRequest represents a tool creation/update request
type ToolRequest struct {
	UserID    string                 `json:"user_id"`
	Component map[string]interface{} `json:"component"`
	ServerID  *uint                  `json:"server_id,omitempty"`
}

// HandleListToolsDB handles GET /api/tools requests using database
func (h *ToolsHandler) HandleListToolsDB(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("tools-handler").WithValues("operation", "list-db")

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	log.V(1).Info("Listing tools from database")
	tools, err := h.DatabaseService.Tool.List(userID)
	if err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to list tools", err))
		return
	}

	log.Info("Successfully listed tools", "count", len(tools))
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": true,
		"data":   tools,
	})
}

// HandleCreateToolDB handles POST /api/tools requests using database
func (h *ToolsHandler) HandleCreateToolDB(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("tools-handler").WithValues("operation", "create-db")

	var toolRequest ToolRequest
	if err := DecodeJSONBody(r, &toolRequest); err != nil {
		w.RespondWithError(errors.NewBadRequestError("Invalid request body", err))
		return
	}

	if toolRequest.UserID == "" {
		w.RespondWithError(errors.NewBadRequestError("user_id is required", nil))
		return
	}
	log = log.WithValues("userID", toolRequest.UserID)

	tool := &database.Tool{
		BaseModel: database.BaseModel{
			UserID: &toolRequest.UserID,
		},
		Component: database.JSONMap(toolRequest.Component),
		ServerID:  toolRequest.ServerID,
	}

	log.V(1).Info("Creating tool in database")
	if err := h.DatabaseService.CreateTool(tool); err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to create tool", err))
		return
	}

	log.Info("Successfully created tool", "toolID", tool.ID)
	RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"status":  true,
		"data":    tool,
		"message": "Tool created successfully",
	})
}

// HandleUpdateToolDB handles PUT /api/tools/{toolID} requests using database
func (h *ToolsHandler) HandleUpdateToolDB(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("tools-handler").WithValues("operation", "update-db")

	toolID, err := GetIntPathParam(r, "toolID")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get tool ID from path", err))
		return
	}
	log = log.WithValues("toolID", toolID)

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	var toolRequest ToolRequest
	if err := DecodeJSONBody(r, &toolRequest); err != nil {
		w.RespondWithError(errors.NewBadRequestError("Invalid request body", err))
		return
	}

	// Get existing tool
	tool, err := h.DatabaseService.GetTool(uint(toolID), userID)
	if err != nil {
		w.RespondWithError(errors.NewNotFoundError("Tool not found", err))
		return
	}

	// Update component
	if toolRequest.Component != nil {
		tool.Component = database.JSONMap(toolRequest.Component)
	}
	if toolRequest.ServerID != nil {
		tool.ServerID = toolRequest.ServerID
	}

	if err := h.DatabaseService.UpdateTool(tool); err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to update tool", err))
		return
	}

	log.Info("Successfully updated tool")
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"data":    tool,
		"message": "Tool updated successfully",
	})
}

// HandleDeleteToolDB handles DELETE /api/tools/{toolID} requests using database
func (h *ToolsHandler) HandleDeleteToolDB(w ErrorResponseWriter, r *http.Request) {
	log := ctrllog.FromContext(r.Context()).WithName("tools-handler").WithValues("operation", "delete-db")

	userID, err := GetUserID(r)
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get user ID", err))
		return
	}
	log = log.WithValues("userID", userID)

	toolID, err := GetIntPathParam(r, "toolID")
	if err != nil {
		w.RespondWithError(errors.NewBadRequestError("Failed to get tool ID from path", err))
		return
	}
	log = log.WithValues("toolID", toolID)

	if err := h.DatabaseService.DeleteTool(uint(toolID), userID); err != nil {
		w.RespondWithError(errors.NewInternalServerError("Failed to delete tool", err))
		return
	}

	log.Info("Successfully deleted tool")
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "Tool deleted successfully",
	})
}
