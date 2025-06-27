package database

import (
	"fmt"

	"gorm.io/gorm"
)

func NewClient(manager *Manager) *Client {
	return &Client{
		Team:         NewService[Team](manager),
		Session:      NewService[Session](manager),
		Run:          NewService[Run](manager),
		Message:      NewService[Message](manager),
		Feedback:     NewService[Feedback](manager),
		Tool:         NewService[Tool](manager),
		ToolServer:   NewService[ToolServer](manager),
		EvalTask:     NewService[EvalTask](manager),
		EvalCriteria: NewService[EvalCriteria](manager),
		EvalRun:      NewService[EvalRun](manager),
	}
}

type Model interface {
	TableName() string
}

type Client struct {
	Team         *Service[Team]
	Session      *Service[Session]
	Run          *Service[Run]
	Message      *Service[Message]
	Feedback     *Service[Feedback]
	Tool         *Service[Tool]
	ToolServer   *Service[ToolServer]
	EvalTask     *Service[EvalTask]
	EvalCriteria *Service[EvalCriteria]
	EvalRun      *Service[EvalRun]
}

// Service provides high-level database operations
type Service[T Model] struct {
	db *gorm.DB
}

// NewService creates a new database service
func NewService[T Model](manager *Manager) *Service[T] {
	return &Service[T]{db: manager.db}
}

func (s *Service[T]) List(userID string) ([]T, error) {
	var models []T
	err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	return models, nil
}

func (s *Service[T]) Get(id uint, userID string) (*T, error) {
	var model T
	err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&model).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get model: %w", err)
	}
	return &model, nil
}

func (s *Service[T]) Create(model *T) error {
	err := s.db.Create(model).Error
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}
	return nil
}

func (s *Service[T]) Update(model *T) error {
	err := s.db.Save(model).Error
	if err != nil {
		return fmt.Errorf("failed to update model: %w", err)
	}
	return nil
}

func (s *Service[T]) Delete(id uint, userID string) error {
	t := new(T)
	result := s.db.Where("id = ? AND user_id = ?", id, userID).Delete(t)
	if result.Error != nil {
		return fmt.Errorf("failed to delete model: %w", result.Error)
	}
	return nil
}
