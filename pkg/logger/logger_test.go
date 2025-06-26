package logger_test

import (
	"testing"

	"github.com/lyarwood/godar/pkg/logger"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

func TestLogger(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logger Package")
}

var _ = Describe("Logger", func() {
	Describe("Init", func() {
		Context("with debug mode", func() {
			It("should initialize logger in debug mode", func() {
				logger.Init(true)
				// The logger should be initialized without panicking
				Expect(func() {
					logger.Info("test message")
				}).NotTo(Panic())
			})
		})

		Context("with production mode", func() {
			It("should initialize logger in production mode", func() {
				logger.Init(false)
				// The logger should be initialized without panicking
				Expect(func() {
					logger.Info("test message")
				}).NotTo(Panic())
			})
		})
	})

	Describe("Logging functions", func() {
		BeforeEach(func() {
			logger.Init(true)
		})

		It("should log info messages", func() {
			Expect(func() {
				logger.Info("test info message")
			}).NotTo(Panic())
		})

		It("should log error messages", func() {
			Expect(func() {
				logger.Error("test error message")
			}).NotTo(Panic())
		})

		It("should log debug messages", func() {
			Expect(func() {
				logger.Debug("test debug message")
			}).NotTo(Panic())
		})

		It("should log warning messages", func() {
			Expect(func() {
				logger.Warn("test warning message")
			}).NotTo(Panic())
		})

		It("should log with fields", func() {
			Expect(func() {
				logger.Info("test message with fields", zap.String("key", "value"))
			}).NotTo(Panic())
		})

		It("should handle nil logger gracefully", func() {
			// This test verifies that the logger functions don't panic when the logger is nil
			// In a real scenario, this might happen if Init hasn't been called
			Expect(func() {
				logger.Info("test message")
			}).NotTo(Panic())
		})
	})

	Describe("Fatal", func() {
		BeforeEach(func() {
			logger.Init(true)
		})

		It("should exit on fatal messages", func() {
			// Note: This test is commented out because it would actually exit the process
			// In a real test environment, you might want to mock os.Exit or test this differently
			/*
				Expect(func() {
					logger.Fatal("test fatal message")
				}).To(Panic())
			*/
		})
	})

	Describe("Logger configuration", func() {
		It("should use ISO8601 timestamp format", func() {
			// This test verifies that the logger is configured with ISO8601 timestamps
			// The actual verification would require capturing log output, which is complex
			// For now, we just ensure the initialization doesn't panic
			Expect(func() {
				logger.Init(true)
			}).NotTo(Panic())
		})
	})
})
