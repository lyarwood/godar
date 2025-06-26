package notification

import (
	"fmt"
	"sync"
)

// MockNotificationSender is a mock implementation of NotificationSender for testing
type MockNotificationSender struct {
	mu            sync.RWMutex
	notifications []NotificationCall
	shouldError   bool
	errorMessage  string
}

// NotificationCall represents a notification that was sent
type NotificationCall struct {
	Title    string
	Message  string
	IconPath string
}

// NewMockNotificationSender creates a new mock notification sender
func NewMockNotificationSender() *MockNotificationSender {
	return &MockNotificationSender{
		notifications: make([]NotificationCall, 0),
	}
}

// Notify captures the notification call without actually sending it
func (m *MockNotificationSender) Notify(title, message, iconPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldError {
		return fmt.Errorf("mock error: %s", m.errorMessage)
	}

	m.notifications = append(m.notifications, NotificationCall{
		Title:    title,
		Message:  message,
		IconPath: iconPath,
	})

	return nil
}

// GetNotifications returns all captured notifications
func (m *MockNotificationSender) GetNotifications() []NotificationCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent race conditions
	result := make([]NotificationCall, len(m.notifications))
	copy(result, m.notifications)
	return result
}

// ClearNotifications clears all captured notifications
func (m *MockNotificationSender) ClearNotifications() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifications = make([]NotificationCall, 0)
}

// SetShouldError configures the mock to return an error on the next Notify call
func (m *MockNotificationSender) SetShouldError(shouldError bool, errorMessage string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldError = shouldError
	m.errorMessage = errorMessage
}

// GetNotificationCount returns the number of notifications captured
func (m *MockNotificationSender) GetNotificationCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.notifications)
}
