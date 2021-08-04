package components

import (
	"context"
	"time"

	"github.com/2gis/loggo/logging"
)

// RetrievePeriodic retrieves retriever initially and further, every given time interval
func RetrievePeriodic(ctx context.Context, retriever Retriever, interval time.Duration, logger logging.Logger) {
	retrieve(retriever, logger)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			retrieve(retriever, logger)
		case <-ctx.Done():
			return
		}
	}
}

func retrieve(retriever Retriever, logger logging.Logger) {
	if err := retriever.Retrieve(); err != nil {
		logger.Error(err)
	}
}
