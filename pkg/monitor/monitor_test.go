package monitor

import (
	"time"

	"github.com/lyarwood/godar/pkg/aircraft"
	"github.com/lyarwood/godar/pkg/config"
	"github.com/lyarwood/godar/pkg/notification"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
)

type mockFetcher struct {
	acList *aircraft.AircraftList
	err    error
}

func (m *mockFetcher) Fetch() (*aircraft.AircraftList, error) {
	return m.acList, m.err
}

func (m *mockFetcher) SetFilters(_ string, _ int, _ int, _ bool, _ string, _ string) {}
func (m *mockFetcher) SetLocation(_ float64, _ float64, _ float64)                   {}
func (m *mockFetcher) SetAuth(_ string, _ string)                                    {}

var _ = Describe("Monitor", func() {
	var (
		cfg    *config.Config
		logger *zap.Logger
	)

	BeforeEach(func() {
		cfg = &config.Config{
			Server:       config.ServerConfig{URL: "http://test"},
			Filters:      config.FilterConfig{},
			Location:     config.LocationConfig{Latitude: 51.5, Longitude: 0.0, MaxDistance: 100},
			Notification: config.NotificationConfig{Enabled: true, Duration: time.Second, CleanupInterval: time.Minute},
			Monitoring:   config.MonitoringConfig{PollInterval: 10 * time.Second},
		}
		var err error
		logger, err = zap.NewDevelopment()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		_ = logger.Sync()
	})

	It("should construct a Monitor", func() {
		mon, err := NewMonitor(cfg, logger)
		Expect(err).ToNot(HaveOccurred())
		Expect(mon).ToNot(BeNil())
	})

	It("should process aircraft and call notifier", func() {
		ac := aircraft.Aircraft{Call: "TEST1", Lat: 51.6, Long: 0.1, Alt: 10000, Type: "A320", Spd: 400}
		acList := &aircraft.AircraftList{Aircraft: []aircraft.Aircraft{ac}}
		fetcher := &mockFetcher{acList: acList}
		n := notification.NewMockNotificationSender()
		notifier := notification.NewNotifierWithSender(true, time.Second, logger, n)
		mon, _ := NewMonitorWithDeps(cfg, logger, fetcher, notifier)
		err := mon.processAircraft(ac)
		Expect(err).ToNot(HaveOccurred())
		calls := n.GetNotifications()
		Expect(len(calls)).To(Equal(1))
		Expect(calls[0].Title).To(ContainSubstring("Aircraft Detected: TEST1"))
	})

	It("should handle notifier error gracefully", func() {
		ac := aircraft.Aircraft{Call: "ERR1", Lat: 51.6, Long: 0.1, Alt: 10000, Type: "A320", Spd: 400}
		fetcher := &mockFetcher{}
		n := notification.NewMockNotificationSender()
		n.SetShouldError(true, "test error")
		notifier := notification.NewNotifierWithSender(true, time.Second, logger, n)
		mon, _ := NewMonitorWithDeps(cfg, logger, fetcher, notifier)
		err := mon.processAircraft(ac)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("mock error: test error"))
	})

	It("should clean up aircraft history", func() {
		fetcher := &mockFetcher{}
		notifier := notification.NewNotifier(cfg.Notification.Enabled, cfg.Notification.Duration, logger)
		mon, _ := NewMonitorWithDeps(cfg, logger, fetcher, notifier)
		mon.aircraftHistory = map[string]*AircraftTracker{
			"old":    {LastSeen: time.Now().Add(-20 * time.Minute)},
			"recent": {LastSeen: time.Now()},
		}
		mon.cleanupAircraftHistory()
		Expect(mon.aircraftHistory).To(HaveKey("recent"))
		Expect(mon.aircraftHistory).ToNot(HaveKey("old"))
	})
})
