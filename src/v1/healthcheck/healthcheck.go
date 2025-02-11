package healthcheck

import (
	"time"

	"github.com/sirupsen/logrus"
)

// MonitorHealth runs the health check function at the given interval in the background.
// It logs INFO if healthy, ERROR otherwise.
func MonitorHealth(logger *logrus.Logger, intervalSeconds int, healthCheck func() (bool, error)) {
	go func() {
		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			healthy, err := healthCheck()
			if healthy {
				logger.Info("OK!")
			} else {
				logger.Errorf("Health check failed: %v", err)
			}
		}
	}()
}
