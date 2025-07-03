package client

// ClientSetInterface represents the main interface for the KAgent client
type ClientSetInterface interface {
	Health() HealthInterface
	Version() VersionInterface
	ModelConfigs() ModelConfigInterface
	Sessions() SessionInterface
	Teams() TeamInterface
	Tools() ToolInterface
	ToolServers() ToolServerInterface
	Memories() MemoryInterface
	Providers() ProviderInterface
	Models() ModelInterface
	Namespaces() NamespaceInterface
	Feedback() FeedbackInterface
}

// ClientSet contains all the sub-clients for different resource types
type ClientSet struct {
	baseClient *BaseClient

	health      HealthInterface
	version     VersionInterface
	modelConfig ModelConfigInterface
	session     SessionInterface
	team        TeamInterface
	tool        ToolInterface
	toolServer  ToolServerInterface
	memory      MemoryInterface
	provider    ProviderInterface
	model       ModelInterface
	namespace   NamespaceInterface
	feedback    FeedbackInterface
}

// NewClientSet creates a new KAgent client set
func NewClientSet(baseURL string, options ...ClientOption) ClientSetInterface {
	// Create a temporary client to extract configuration using existing option system
	tempClient := New(baseURL, options...)

	baseClient := NewBaseClient(baseURL, tempClient.HTTPClient, tempClient.UserID)

	return &ClientSet{
		baseClient:  baseClient,
		health:      NewHealthClient(baseClient),
		version:     NewVersionClient(baseClient),
		modelConfig: NewModelConfigClient(baseClient),
		session:     NewSessionClient(baseClient),
		team:        NewTeamClient(baseClient),
		tool:        NewToolClient(baseClient),
		toolServer:  NewToolServerClient(baseClient),
		memory:      NewMemoryClient(baseClient),
		provider:    NewProviderClient(baseClient),
		model:       NewModelClient(baseClient),
		namespace:   NewNamespaceClient(baseClient),
		feedback:    NewFeedbackClient(baseClient),
	}
}

// Health returns the health client
func (c *ClientSet) Health() HealthInterface {
	return c.health
}

// Version returns the version client
func (c *ClientSet) Version() VersionInterface {
	return c.version
}

// ModelConfigs returns the model config client
func (c *ClientSet) ModelConfigs() ModelConfigInterface {
	return c.modelConfig
}

// Sessions returns the session client
func (c *ClientSet) Sessions() SessionInterface {
	return c.session
}

// Teams returns the team client
func (c *ClientSet) Teams() TeamInterface {
	return c.team
}

// Tools returns the tool client
func (c *ClientSet) Tools() ToolInterface {
	return c.tool
}

// ToolServers returns the tool server client
func (c *ClientSet) ToolServers() ToolServerInterface {
	return c.toolServer
}

// Memories returns the memory client
func (c *ClientSet) Memories() MemoryInterface {
	return c.memory
}

// Providers returns the provider client
func (c *ClientSet) Providers() ProviderInterface {
	return c.provider
}

// Models returns the model client
func (c *ClientSet) Models() ModelInterface {
	return c.model
}

// Namespaces returns the namespace client
func (c *ClientSet) Namespaces() NamespaceInterface {
	return c.namespace
}

// Feedback returns the feedback client
func (c *ClientSet) Feedback() FeedbackInterface {
	return c.feedback
}

// NewClient creates a new KAgent client set (alias for NewClientSet)
func NewClient(baseURL string, options ...ClientOption) ClientSetInterface {
	return NewClientSet(baseURL, options...)
}
