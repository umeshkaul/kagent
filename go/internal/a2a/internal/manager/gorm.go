package manager

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"trpc.group/trpc-go/trpc-a2a-go/protocol"
)

// GORM models
type StoredMessage struct {
	ID        uint   `gorm:"primaryKey"`
	MessageID string `gorm:"unique;not null;index"`
	ContextID *string
	Data      string `gorm:"type:text;not null"` // JSON serialized protocol.Message
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (StoredMessage) TableName() string {
	return "messages"
}

type StoredConversation struct {
	ID             uint   `gorm:"primaryKey"`
	ContextID      string `gorm:"unique;not null;index"`
	MessageIDs     string `gorm:"type:text"` // JSON array of message IDs
	LastAccessTime time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (StoredConversation) TableName() string {
	return "conversations"
}

type StoredTask struct {
	ID        uint   `gorm:"primaryKey"`
	TaskID    string `gorm:"unique;not null;index"`
	Data      string `gorm:"type:text;not null"` // JSON serialized task data
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (StoredTask) TableName() string {
	return "tasks"
}

type StoredPushNotification struct {
	ID        uint   `gorm:"primaryKey"`
	TaskID    string `gorm:"unique;not null;index"`
	Data      string `gorm:"type:text;not null"` // JSON serialized push notification config
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (StoredPushNotification) TableName() string {
	return "push_notifications"
}

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

	// Auto migrate tables
	err := db.AutoMigrate(
		&StoredMessage{},
		&StoredConversation{},
		&StoredTask{},
		&StoredPushNotification{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate tables: %w", err)
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

	storedMessage := StoredMessage{
		MessageID: message.MessageID,
		ContextID: message.ContextID,
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

	// If the message has a contextID, handle conversation history
	if message.ContextID != nil {
		contextID := *message.ContextID

		var conversation StoredConversation
		err := tx.Where("context_id = ?", contextID).First(&conversation).Error

		if err == gorm.ErrRecordNotFound {
			// Create new conversation
			messageIDs := []string{message.MessageID}
			messageIDsJSON, _ := json.Marshal(messageIDs)

			conversation = StoredConversation{
				ContextID:      contextID,
				MessageIDs:     string(messageIDsJSON),
				LastAccessTime: time.Now(),
			}

			if err := tx.Create(&conversation).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create conversation: %w", err)
			}
		} else if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to query conversation: %w", err)
		} else {
			// Update existing conversation
			var messageIDs []string
			if err := json.Unmarshal([]byte(conversation.MessageIDs), &messageIDs); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to unmarshal message IDs: %w", err)
			}

			messageIDs = append(messageIDs, message.MessageID)

			// Limit history length
			if len(messageIDs) > s.maxHistoryLength {
				// Remove oldest messages
				removedMsgIDs := messageIDs[:len(messageIDs)-s.maxHistoryLength]
				messageIDs = messageIDs[len(messageIDs)-s.maxHistoryLength:]

				// Delete old messages from database
				if err := tx.Where("message_id IN ?", removedMsgIDs).Delete(&StoredMessage{}).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to delete old messages: %w", err)
				}
			}

			messageIDsJSON, _ := json.Marshal(messageIDs)
			conversation.MessageIDs = string(messageIDsJSON)
			conversation.LastAccessTime = time.Now()

			if err := tx.Save(&conversation).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update conversation: %w", err)
			}
		}
	}

	return tx.Commit().Error
}

func (s *GormStorage) GetMessage(messageID string) (protocol.Message, error) {
	var storedMessage StoredMessage
	err := s.db.Where("message_id = ?", messageID).First(&storedMessage).Error
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
	return s.db.Where("message_id = ?", messageID).Delete(&StoredMessage{}).Error
}

func (s *GormStorage) GetMessages(messageIDs []string) ([]protocol.Message, error) {
	if len(messageIDs) == 0 {
		return []protocol.Message{}, nil
	}

	var storedMessages []StoredMessage
	err := s.db.Where("message_id IN ?", messageIDs).Find(&storedMessages).Error
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

// Conversation operations
func (s *GormStorage) StoreConversation(contextID string, history *ConversationHistory) error {
	messageIDsJSON, err := json.Marshal(history.MessageIDs)
	if err != nil {
		return fmt.Errorf("failed to serialize message IDs: %w", err)
	}

	conversation := StoredConversation{
		ContextID:      contextID,
		MessageIDs:     string(messageIDsJSON),
		LastAccessTime: history.LastAccessTime,
	}

	return s.db.Save(&conversation).Error
}

func (s *GormStorage) GetConversation(contextID string) (*ConversationHistory, error) {
	var conversation StoredConversation
	err := s.db.Where("context_id = ?", contextID).First(&conversation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("conversation not found: %s", contextID)
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	var messageIDs []string
	if err := json.Unmarshal([]byte(conversation.MessageIDs), &messageIDs); err != nil {
		return nil, fmt.Errorf("failed to deserialize message IDs: %w", err)
	}

	return &ConversationHistory{
		MessageIDs:     messageIDs,
		LastAccessTime: conversation.LastAccessTime,
	}, nil
}

func (s *GormStorage) UpdateConversationAccess(contextID string, timestamp time.Time) error {
	return s.db.Model(&StoredConversation{}).
		Where("context_id = ?", contextID).
		Update("last_access_time", timestamp).Error
}

func (s *GormStorage) DeleteConversation(contextID string) error {
	return s.db.Where("context_id = ?", contextID).Delete(&StoredConversation{}).Error
}

func (s *GormStorage) GetExpiredConversations(maxAge time.Duration) ([]string, error) {
	cutoff := time.Now().Add(-maxAge)

	var conversations []StoredConversation
	err := s.db.Where("last_access_time < ?", cutoff).Find(&conversations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get expired conversations: %w", err)
	}

	contextIDs := make([]string, len(conversations))
	for i, conv := range conversations {
		contextIDs[i] = conv.ContextID
	}

	return contextIDs, nil
}

func (s *GormStorage) GetConversationStats() (map[string]interface{}, error) {
	var totalConversations int64
	var totalMessages int64

	if err := s.db.Model(&StoredConversation{}).Count(&totalConversations).Error; err != nil {
		return nil, fmt.Errorf("failed to count conversations: %w", err)
	}

	if err := s.db.Model(&StoredMessage{}).Count(&totalMessages).Error; err != nil {
		return nil, fmt.Errorf("failed to count messages: %w", err)
	}

	stats := map[string]interface{}{
		"total_conversations": totalConversations,
		"total_messages":      totalMessages,
	}

	if totalConversations > 0 {
		var oldestAccess, newestAccess time.Time

		s.db.Model(&StoredConversation{}).Select("MIN(last_access_time)").Scan(&oldestAccess)
		s.db.Model(&StoredConversation{}).Select("MAX(last_access_time)").Scan(&newestAccess)

		stats["oldest_access"] = oldestAccess
		stats["newest_access"] = newestAccess
	}

	return stats, nil
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

	storedTask := StoredTask{
		TaskID: taskID,
		Data:   string(data),
	}

	return s.db.Save(&storedTask).Error
}

func (s *GormStorage) GetTask(taskID string) (*MemoryCancellableTask, error) {
	var storedTask StoredTask
	err := s.db.Where("task_id = ?", taskID).First(&storedTask).Error
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
	return s.db.Where("task_id = ?", taskID).Delete(&StoredTask{}).Error
}

func (s *GormStorage) TaskExists(taskID string) bool {
	var count int64
	s.db.Model(&StoredTask{}).Where("task_id = ?", taskID).Count(&count)
	return count > 0
}

// Push notification operations
func (s *GormStorage) StorePushNotification(taskID string, config protocol.TaskPushNotificationConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize push notification config: %w", err)
	}

	storedConfig := StoredPushNotification{
		TaskID: taskID,
		Data:   string(data),
	}

	return s.db.Save(&storedConfig).Error
}

func (s *GormStorage) GetPushNotification(taskID string) (protocol.TaskPushNotificationConfig, error) {
	var storedConfig StoredPushNotification
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
	return s.db.Where("task_id = ?", taskID).Delete(&StoredPushNotification{}).Error
}

// Cleanup operations
func (s *GormStorage) CleanupExpiredConversations(maxAge time.Duration) (int, error) {
	cutoff := time.Now().Add(-maxAge)

	// Begin transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get expired conversations
	var expiredConversations []StoredConversation
	err := tx.Where("last_access_time < ?", cutoff).Find(&expiredConversations).Error
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to find expired conversations: %w", err)
	}

	if len(expiredConversations) == 0 {
		tx.Commit()
		return 0, nil
	}

	// Collect all message IDs from expired conversations
	var allMessageIDs []string
	var contextIDs []string

	for _, conv := range expiredConversations {
		contextIDs = append(contextIDs, conv.ContextID)

		var messageIDs []string
		if err := json.Unmarshal([]byte(conv.MessageIDs), &messageIDs); err == nil {
			allMessageIDs = append(allMessageIDs, messageIDs...)
		}
	}

	// Delete messages from expired conversations
	if len(allMessageIDs) > 0 {
		if err := tx.Where("message_id IN ?", allMessageIDs).Delete(&StoredMessage{}).Error; err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to delete expired messages: %w", err)
		}
	}

	// Delete expired conversations
	if err := tx.Where("context_id IN ?", contextIDs).Delete(&StoredConversation{}).Error; err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("failed to delete expired conversations: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("failed to commit cleanup transaction: %w", err)
	}

	return len(expiredConversations), nil
}
