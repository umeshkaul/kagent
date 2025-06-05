package handlers

// TeamsHandler handles team-related requests
type TeamsHandler struct {
	*Base
}

// NewTeamsHandler creates a new TeamsHandler
func NewTeamsHandler(base *Base) *TeamsHandler {
	return &TeamsHandler{Base: base}
}
