package handlers

// FeedbackHandler handles user feedback submissions
type FeedbackHandler struct {
	*Base
}

// NewFeedbackHandler creates a new feedback handler
func NewFeedbackHandler(base *Base) *FeedbackHandler {
	return &FeedbackHandler{Base: base}
}
