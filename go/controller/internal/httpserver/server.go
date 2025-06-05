package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	autogen_client "github.com/kagent-dev/kagent/go/autogen/client"
	"github.com/kagent-dev/kagent/go/controller/internal/a2a"
	"github.com/kagent-dev/kagent/go/controller/internal/database"
	"github.com/kagent-dev/kagent/go/controller/internal/httpserver/handlers"
	common "github.com/kagent-dev/kagent/go/controller/internal/utils"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// API Path constants
	APIPathHealth      = "/health"
	APIPathModelConfig = "/api/modelconfigs"
	APIPathRuns        = "/api/runs"
	APIPathSessions    = "/api/sessions"
	APIPathTools       = "/api/tools"
	APIPathToolServers = "/api/toolservers"
	APIPathTeams       = "/api/teams"
	APIPathAgents      = "/api/agents"
	APIPathProviders   = "/api/providers"
	APIPathModels      = "/api/models"
	APIPathMemories    = "/api/memories"
	APIPathA2A         = "/api/a2a"
	APIPathFeedback    = "/api/feedback"
)

var defaultModelConfig = types.NamespacedName{
	Name:      "default-model-config",
	Namespace: common.GetResourceNamespace(),
}

// ServerConfig holds the configuration for the HTTP server
type ServerConfig struct {
	BindAddr      string
	AutogenClient autogen_client.Client
	KubeClient    client.Client
	A2AHandler    a2a.A2AHandlerMux
	DatabasePath  string // Path to SQLite database file
}

// HTTPServer is the structure that manages the HTTP server
type HTTPServer struct {
	httpServer *http.Server
	config     ServerConfig
	router     *mux.Router
	handlers   *handlers.Handlers
	dbManager  *database.Manager
	dbService  *database.Service
}

// NewHTTPServer creates a new HTTP server instance
func NewHTTPServer(config ServerConfig) (*HTTPServer, error) {
	// Initialize database
	dbManager, err := database.NewManager(config.DatabasePath)
	if err != nil {
		return nil, err
	}

	// Initialize database tables
	if err := dbManager.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	dbService := database.NewService(dbManager)

	return &HTTPServer{
		config:    config,
		router:    mux.NewRouter(),
		handlers:  handlers.NewHandlers(config.KubeClient, config.AutogenClient, defaultModelConfig, dbService),
		dbManager: dbManager,
		dbService: dbService,
	}, nil
}

// Start initializes and starts the HTTP server
func (s *HTTPServer) Start(ctx context.Context) error {
	log := ctrllog.FromContext(ctx).WithName("http-server")
	log.Info("Starting HTTP server", "address", s.config.BindAddr)

	// Setup routes
	s.setupRoutes()

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:    s.config.BindAddr,
		Handler: s.router,
	}

	// Start the server in a separate goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(err, "HTTP server failed")
		}
	}()

	// Wait for context cancellation to shut down
	go func() {
		<-ctx.Done()
		log.Info("Shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error(err, "Failed to properly shutdown HTTP server")
		}
		// Close database connection
		if err := s.dbManager.Close(); err != nil {
			log.Error(err, "Failed to close database connection")
		}
	}()

	return nil
}

// Stop stops the HTTP server
func (s *HTTPServer) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// NeedLeaderElection implements controller-runtime's LeaderElectionRunnable interface
func (s *HTTPServer) NeedLeaderElection() bool {
	// Return false so the HTTP server runs on all instances, not just the leader
	return false
}

// setupRoutes configures all the routes for the server
func (s *HTTPServer) setupRoutes() {
	// Health check endpoint
	s.router.HandleFunc(APIPathHealth, adaptHealthHandler(s.handlers.Health.HandleHealth)).Methods(http.MethodGet)

	// Model configs
	s.router.HandleFunc(APIPathModelConfig, adaptHandler(s.handlers.ModelConfig.HandleListModelConfigs)).Methods(http.MethodGet)
	s.router.HandleFunc(APIPathModelConfig+"/{configName}", adaptHandler(s.handlers.ModelConfig.HandleGetModelConfig)).Methods(http.MethodGet)
	s.router.HandleFunc(APIPathModelConfig, adaptHandler(s.handlers.ModelConfig.HandleCreateModelConfig)).Methods(http.MethodPost)
	s.router.HandleFunc(APIPathModelConfig+"/{configName}", adaptHandler(s.handlers.ModelConfig.HandleDeleteModelConfig)).Methods(http.MethodDelete)
	s.router.HandleFunc(APIPathModelConfig+"/{configName}", adaptHandler(s.handlers.ModelConfig.HandleUpdateModelConfig)).Methods(http.MethodPut)

	// Sessions - using database handlers
	s.router.HandleFunc(APIPathSessions, adaptHandler(s.handlers.Sessions.HandleListSessionsDB)).Methods(http.MethodGet)
	s.router.HandleFunc(APIPathSessions, adaptHandler(s.handlers.Sessions.HandleCreateSessionDB)).Methods(http.MethodPost)
	s.router.HandleFunc(APIPathSessions+"/{sessionID}", adaptHandler(s.handlers.Sessions.HandleGetSessionDB)).Methods(http.MethodGet)
	s.router.HandleFunc(APIPathSessions+"/{sessionID}/invoke", adaptHandler(s.handlers.Sessions.HandleSessionInvokeDB)).Methods(http.MethodPost)
	s.router.HandleFunc(APIPathSessions+"/{sessionID}/invoke/stream", adaptHandler(s.handlers.Sessions.HandleSessionInvokeStream)).Methods(http.MethodPost)
	s.router.HandleFunc(APIPathSessions+"/{sessionID}/runs", adaptHandler(s.handlers.Sessions.HandleListSessionRunsDB)).Methods(http.MethodGet)
	s.router.HandleFunc(APIPathSessions+"/{sessionID}", adaptHandler(s.handlers.Sessions.HandleDeleteSessionDB)).Methods(http.MethodDelete)
	s.router.HandleFunc(APIPathSessions+"/{sessionID}", adaptHandler(s.handlers.Sessions.HandleUpdateSessionDB)).Methods(http.MethodPut)

	// Tools - using database handlers
	s.router.HandleFunc(APIPathTools, adaptHandler(s.handlers.Tools.HandleListToolsDB)).Methods(http.MethodGet)
	s.router.HandleFunc(APIPathTools, adaptHandler(s.handlers.Tools.HandleCreateToolDB)).Methods(http.MethodPost)
	s.router.HandleFunc(APIPathTools+"/{toolID}", adaptHandler(s.handlers.Tools.HandleUpdateToolDB)).Methods(http.MethodPut)
	s.router.HandleFunc(APIPathTools+"/{toolID}", adaptHandler(s.handlers.Tools.HandleDeleteToolDB)).Methods(http.MethodDelete)

	// Tool Servers
	s.router.HandleFunc(APIPathToolServers, adaptHandler(s.handlers.ToolServers.HandleListToolServers)).Methods(http.MethodGet)
	s.router.HandleFunc(APIPathToolServers, adaptHandler(s.handlers.ToolServers.HandleCreateToolServer)).Methods(http.MethodPost)
	s.router.HandleFunc(APIPathToolServers+"/{toolServerName}", adaptHandler(s.handlers.ToolServers.HandleDeleteToolServer)).Methods(http.MethodDelete)

	// Teams - using database handlers
	s.router.HandleFunc(APIPathTeams, adaptHandler(s.handlers.Teams.HandleListTeamsDB)).Methods(http.MethodGet)
	s.router.HandleFunc(APIPathTeams, adaptHandler(s.handlers.Teams.HandleCreateTeamDB)).Methods(http.MethodPost)
	s.router.HandleFunc(APIPathTeams+"/{teamID}", adaptHandler(s.handlers.Teams.HandleUpdateTeamDB)).Methods(http.MethodPut)
	s.router.HandleFunc(APIPathTeams+"/{teamID}", adaptHandler(s.handlers.Teams.HandleGetTeamDB)).Methods(http.MethodGet)
	s.router.HandleFunc(APIPathTeams+"/{teamID}", adaptHandler(s.handlers.Teams.HandleDeleteTeamDB)).Methods(http.MethodDelete)

	// Agents
	s.router.HandleFunc(APIPathAgents+"/{agentId}/invoke", adaptHandler(s.handlers.Invoke.HandleInvokeAgent)).Methods(http.MethodPost)
	s.router.HandleFunc(APIPathAgents+"/{agentId}/invoke/stream", adaptHandler(s.handlers.Invoke.HandleInvokeAgentStream)).Methods(http.MethodPost)

	// Providers
	s.router.HandleFunc(APIPathProviders+"/models", adaptHandler(s.handlers.Provider.HandleListSupportedModelProviders)).Methods(http.MethodGet)
	s.router.HandleFunc(APIPathProviders+"/memories", adaptHandler(s.handlers.Provider.HandleListSupportedMemoryProviders)).Methods(http.MethodGet)

	// Models
	s.router.HandleFunc(APIPathModels, adaptHandler(s.handlers.Model.HandleListSupportedModels)).Methods(http.MethodGet)

	// Memories
	s.router.HandleFunc(APIPathMemories, adaptHandler(s.handlers.Memory.HandleListMemories)).Methods(http.MethodGet)
	s.router.HandleFunc(APIPathMemories, adaptHandler(s.handlers.Memory.HandleCreateMemory)).Methods(http.MethodPost)
	s.router.HandleFunc(APIPathMemories+"/{memoryName}", adaptHandler(s.handlers.Memory.HandleDeleteMemory)).Methods(http.MethodDelete)
	s.router.HandleFunc(APIPathMemories+"/{memoryName}", adaptHandler(s.handlers.Memory.HandleGetMemory)).Methods(http.MethodGet)
	s.router.HandleFunc(APIPathMemories+"/{memoryName}", adaptHandler(s.handlers.Memory.HandleUpdateMemory)).Methods(http.MethodPut)

	// Feedback - using database handlers
	s.router.HandleFunc(APIPathFeedback, adaptHandler(s.handlers.Feedback.HandleCreateFeedbackDB)).Methods(http.MethodPost)
	s.router.HandleFunc(APIPathFeedback, adaptHandler(s.handlers.Feedback.HandleListFeedbackDB)).Methods(http.MethodGet)

	// A2A
	s.router.PathPrefix(APIPathA2A).Handler(s.config.A2AHandler)

	// Use middleware for common functionality
	s.router.Use(contentTypeMiddleware)
	s.router.Use(loggingMiddleware)
	s.router.Use(errorHandlerMiddleware)
}

func adaptHandler(h func(handlers.ErrorResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(w.(handlers.ErrorResponseWriter), r)
	}
}

func adaptHealthHandler(h func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return h
}
