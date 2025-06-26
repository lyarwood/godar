package notification_test

import (
	"testing"
	"time"

	"github.com/lyarwood/godar/pkg/notification"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

func TestNotification(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Notification Package")
}

var _ = Describe("Notification", func() {
	var (
		logger     *zap.Logger
		mockSender *notification.MockNotificationSender
	)

	BeforeEach(func() {
		var err error
		logger, err = zap.NewDevelopment()
		Expect(err).NotTo(HaveOccurred())
		mockSender = notification.NewMockNotificationSender()
	})

	AfterEach(func() {
		if logger != nil {
			_ = logger.Sync()
		}
		mockSender.ClearNotifications()
	})

	Describe("NewNotifier", func() {
		It("should create a new notifier with enabled state", func() {
			notifier := notification.NewNotifier(true, 30*time.Second, logger)
			Expect(notifier).NotTo(BeNil())
		})

		It("should create a new notifier with disabled state", func() {
			notifier := notification.NewNotifier(false, 30*time.Second, logger)
			Expect(notifier).NotTo(BeNil())
		})

		It("should create a new notifier with nil logger", func() {
			notifier := notification.NewNotifier(true, 30*time.Second, nil)
			Expect(notifier).NotTo(BeNil())
		})
	})

	Describe("NewNotifierWithSender", func() {
		It("should create a new notifier with custom sender", func() {
			notifier := notification.NewNotifierWithSender(true, 30*time.Second, logger, mockSender)
			Expect(notifier).NotTo(BeNil())
		})
	})

	Describe("SendAircraftNotification", func() {
		Context("when notifications are enabled", func() {
			It("should send notification with correct content", func() {
				notifier := notification.NewNotifierWithSender(true, 30*time.Second, logger, mockSender)
				err := notifier.Send("TEST123", "A320", 35000, 450, 25.5, "NE")
				Expect(err).To(BeNil())

				notifications := mockSender.GetNotifications()
				Expect(notifications).To(HaveLen(1))
				Expect(notifications[0].Title).To(Equal("Aircraft Detected: TEST123"))
				Expect(notifications[0].Message).To(ContainSubstring("Type: A320"))
				Expect(notifications[0].Message).To(ContainSubstring("Altitude: 35000 ft"))
				Expect(notifications[0].Message).To(ContainSubstring("Speed: 450.0 knots"))
				Expect(notifications[0].Message).To(ContainSubstring("Distance: 25.50 km"))
				Expect(notifications[0].Message).To(ContainSubstring("Direction: NE"))
			})

			It("should handle empty callsign", func() {
				notifier := notification.NewNotifierWithSender(true, 30*time.Second, logger, mockSender)
				err := notifier.Send("", "A320", 35000, 450, 25.5, "N")
				Expect(err).To(BeNil())

				notifications := mockSender.GetNotifications()
				Expect(notifications).To(HaveLen(1))
				Expect(notifications[0].Title).To(Equal("Aircraft Detected: "))
			})

			It("should handle empty aircraft type", func() {
				notifier := notification.NewNotifierWithSender(true, 30*time.Second, logger, mockSender)
				err := notifier.Send("TEST123", "", 35000, 450, 25.5, "S")
				Expect(err).To(BeNil())

				notifications := mockSender.GetNotifications()
				Expect(notifications).To(HaveLen(1))
				Expect(notifications[0].Message).To(ContainSubstring("Type: "))
			})

			It("should handle zero values", func() {
				notifier := notification.NewNotifierWithSender(true, 30*time.Second, logger, mockSender)
				err := notifier.Send("TEST123", "A320", 0, 0, 0.0, "W")
				Expect(err).To(BeNil())

				notifications := mockSender.GetNotifications()
				Expect(notifications).To(HaveLen(1))
				Expect(notifications[0].Message).To(ContainSubstring("Altitude: 0 ft"))
				Expect(notifications[0].Message).To(ContainSubstring("Speed: 0.0 knots"))
				Expect(notifications[0].Message).To(ContainSubstring("Distance: 0.00 km"))
			})

			It("should include previous distance when aircraft is getting closer", func() {
				notifier := notification.NewNotifierWithSender(true, 30*time.Second, logger, mockSender)
				err := notifier.Send("TEST123", "A320", 35000, 450, 20.0, "NE", 25.5)
				Expect(err).To(BeNil())

				notifications := mockSender.GetNotifications()
				Expect(notifications).To(HaveLen(1))
				Expect(notifications[0].Message).To(ContainSubstring("Previous: 25.50 km (closer by 5.50 km)"))
			})

			It("should include previous distance when aircraft is moving away", func() {
				notifier := notification.NewNotifierWithSender(true, 30*time.Second, logger, mockSender)
				err := notifier.Send("TEST123", "A320", 35000, 450, 30.0, "SW", 25.5)
				Expect(err).To(BeNil())

				notifications := mockSender.GetNotifications()
				Expect(notifications).To(HaveLen(1))
				Expect(notifications[0].Message).To(ContainSubstring("Previous: 25.50 km (farther by 4.50 km)"))
			})

			It("should handle zero previous distance", func() {
				notifier := notification.NewNotifierWithSender(true, 30*time.Second, logger, mockSender)
				err := notifier.Send("TEST123", "A320", 35000, 450, 25.5, "E", 0.0)
				Expect(err).To(BeNil())

				notifications := mockSender.GetNotifications()
				Expect(notifications).To(HaveLen(1))
				// Should not include previous distance line when previous distance is 0
				Expect(notifications[0].Message).NotTo(ContainSubstring("Previous:"))
			})

			It("should handle notification sender error", func() {
				mockSender.SetShouldError(true, "test error")
				notifier := notification.NewNotifierWithSender(true, 30*time.Second, logger, mockSender)
				err := notifier.Send("TEST123", "A320", 35000, 450, 25.5, "N")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("mock error: test error"))
			})
		})

		Context("when notifications are disabled", func() {
			It("should not send notification and return nil", func() {
				notifier := notification.NewNotifierWithSender(false, 30*time.Second, logger, mockSender)
				err := notifier.Send("TEST123", "A320", 35000, 450, 25.5, "SE")
				Expect(err).To(BeNil())
				Expect(mockSender.GetNotificationCount()).To(Equal(0))
			})

			It("should not send notification with previous distance and return nil", func() {
				notifier := notification.NewNotifierWithSender(false, 30*time.Second, logger, mockSender)
				err := notifier.Send("TEST123", "A320", 35000, 450, 25.5, "NW", 30.0)
				Expect(err).To(BeNil())
				Expect(mockSender.GetNotificationCount()).To(Equal(0))
			})
		})
	})

	Describe("Send method", func() {
		It("should be an alias for SendAircraftNotification", func() {
			notifier := notification.NewNotifierWithSender(true, 30*time.Second, logger, mockSender)
			err := notifier.Send("TEST123", "A320", 35000, 450, 25.5, "NE")
			Expect(err).To(BeNil())
			Expect(mockSender.GetNotificationCount()).To(Equal(1))
		})
	})
})
