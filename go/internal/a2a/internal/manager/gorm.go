package manager

import (
	"encoding/json"
	"fmt"

	"github.com/kagent-dev/kagent/go/internal/database"
	"gorm.io/gorm"
	"trpc.group/trpc-go/trpc-a2a-go/protocol"
)

// GormStorage is a GORM-based implementation of the Storage interface
type GormStorage struct {
	db               *gorm.DB
	maxHistoryLength int
}

// NewGormStorage creates a new GORM-based storage implementation
func NewGormStorage(db *gorm.DB, options StorageOptions) (*GormStorage, error) {
	maxHistoryLength := options.MaxHistoryLength
	if maxHistoryLength <= 0 {
		maxHistoryLength = defaultMaxHistoryLength
	}

	storage := &GormStorage{
		db:               db,
		maxHistoryLength: maxHistoryLength,
	}

	return storage, nil
}

// Message operations
func (s *GormStorage) StoreMessage(message protocol.Message) error {
	// Serialize message to JSON
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	storedMessage := database.Message{
		ID:        message.MessageID,
		SessionID: message.ContextID,
		Data:      string(data),
	}

	// Begin transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Store the message
	if err := tx.Create(&storedMessage).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to store message: %w", err)
	}

	return tx.Commit().Error
}

func (s *GormStorage) GetMessage(messageID string) (protocol.Message, error) {
	var storedMessage database.Message
	err := s.db.Where("id = ?", messageID).First(&storedMessage).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return protocol.Message{}, fmt.Errorf("message not found: %s", messageID)
		}
		return protocol.Message{}, fmt.Errorf("failed to get message: %w", err)
	}

	var message protocol.Message
	if err := json.Unmarshal([]byte(storedMessage.Data), &message); err != nil {
		return protocol.Message{}, fmt.Errorf("failed to deserialize message: %w", err)
	}

	return message, nil
}

func (s *GormStorage) DeleteMessage(messageID string) error {
	return s.db.Where("id = ?", messageID).Delete(&database.Message{}).Error
}

func (s *GormStorage) GetMessages(messageIDs []string) ([]protocol.Message, error) {
	if len(messageIDs) == 0 {
		return []protocol.Message{}, nil
	}

	var storedMessages []database.Message
	err := s.db.Where("id IN ?", messageIDs).Find(&storedMessages).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	messages := make([]protocol.Message, 0, len(storedMessages))
	for _, storedMessage := range storedMessages {
		var message protocol.Message
		if err := json.Unmarshal([]byte(storedMessage.Data), &message); err != nil {
			return nil, fmt.Errorf("failed to deserialize message: %w", err)
		}
		messages = append(messages, message)
	}

	return messages, nil
}

func (s *GormStorage) ListMessagesByContextID(contextID string, limit int) ([]protocol.Message, error) {
	var messages []database.Message
	err := s.db.Where("context_id = ?", contextID).Order("created_at DESC").Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	protocolMessages := make([]protocol.Message, 0, len(messages))
	for _, message := range messages {
		var protocolMessage protocol.Message
		if err := json.Unmarshal([]byte(message.Data), &protocolMessage); err != nil {
			return nil, fmt.Errorf("failed to deserialize message: %w", err)
		}
		protocolMessages = append(protocolMessages, protocolMessage)
	}
	return protocolMessages, nil
}

// Task operations - Note: Tasks cannot be easily serialized due to context.CancelFunc
// For now, we'll store a simplified version and recreate the cancellation context
func (s *GormStorage) StoreTask(taskID string, task *MemoryCancellableTask) error {
	// Serialize the task data (without cancelFunc and ctx)
	taskData := task.Task()
	data, err := json.Marshal(taskData)
	if err != nil {
		return fmt.Errorf("failed to serialize task: %w", err)
	}

	storedTask := database.Task{
		ID:   taskID,
		Data: string(data),
	}

	return s.db.Save(&storedTask).Error
}

func (s *GormStorage) GetTask(taskID string) (*MemoryCancellableTask, error) {
	var storedTask database.Task
	err := s.db.Where("id = ?", taskID).First(&storedTask).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("task not found: %s", taskID)
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	var task protocol.Task
	if err := json.Unmarshal([]byte(storedTask.Data), &task); err != nil {
		return nil, fmt.Errorf("failed to deserialize task: %w", err)
	}

	// Recreate cancellable task
	cancellableTask := NewCancellableTask(task)
	return cancellableTask, nil
}

func (s *GormStorage) DeleteTask(taskID string) error {
	return s.db.Where("id = ?", taskID).Delete(&database.Task{}).Error
}

func (s *GormStorage) TaskExists(taskID string) bool {
	var count int64
	s.db.Model(&database.Task{}).Where("id = ?", taskID).Count(&count)
	return count > 0
}

// Push notification operations
func (s *GormStorage) StorePushNotification(taskID string, config protocol.TaskPushNotificationConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize push notification config: %w", err)
	}

	storedConfig := database.PushNotification{
		TaskID: taskID,
		Data:   string(data),
	}

	return s.db.Save(&storedConfig).Error
}

func (s *GormStorage) GetPushNotification(taskID string) (protocol.TaskPushNotificationConfig, error) {
	var storedConfig database.PushNotification
	err := s.db.Where("task_id = ?", taskID).First(&storedConfig).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return protocol.TaskPushNotificationConfig{}, fmt.Errorf("push notification config not found for task: %s", taskID)
		}
		return protocol.TaskPushNotificationConfig{}, fmt.Errorf("failed to get push notification config: %w", err)
	}

	var config protocol.TaskPushNotificationConfig
	if err := json.Unmarshal([]byte(storedConfig.Data), &config); err != nil {
		return protocol.TaskPushNotificationConfig{}, fmt.Errorf("failed to deserialize push notification config: %w", err)
	}

	return config, nil
}

func (s *GormStorage) DeletePushNotification(taskID string) error {
	return s.db.Where("task_id = ?", taskID).Delete(&database.PushNotification{}).Error
}
