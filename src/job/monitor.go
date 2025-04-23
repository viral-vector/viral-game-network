package job

import (
	"log"
	"time"
	"viral-game-network/src/cache"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/database/repository"
	"viral-game-network/src/k8"
)

/**
 * Monitor System
 */
func Job_Monitoring() {
	// Check if we have a monitoring lock
	if cache_lock, _ := cache.Get[string]("monitoring-cluster-events"); cache_lock == "" {
		go func() {
			// Lock monitoring
			cache.Set[string]("monitoring-cluster-events", "true", time.Second * 5)
			// RUN
			RUN_Monitoring_ClusterEvents()
		}()
	}
}

// Monitor the cluster
func RUN_Monitoring_ClusterEvents() {
	log.Println("[Job_Monitoring]: @ClusterEvents")

	cachedLastEventStr, err := cache.Get[string]("monitoring-cluster:last_event")
	var cachedLastEvent time.Time
	if cachedLastEventStr != "" {
		cachedLastEvent, err = time.Parse(time.RFC3339, cachedLastEventStr)
		if err != nil {
			log.Println("[Job_Monitoring]: @ClusterEvents: %v", err)
		}
	}
	ctlines := int64(100)
	cevents, _ := k8.GetLogsCluster(&ctlines)
	var latestEventTime time.Time
	if cevents != nil {
		for _, event := range cevents {
			// Skip events that occurred before the cached last event.
			if !cachedLastEvent.IsZero() && (
				event.FirstTimestamp.Time.Before(cachedLastEvent) || event.FirstTimestamp.Time.Equal(cachedLastEvent)) {
				continue
			}

			eventType := "info"
			switch event.Type {
			case "Warning":
				eventType = "warning"
			}
			systemEvent := dbtype.SystemEvent{
				Ref_ID: string(event.ObjectMeta.UID),
				Ref_Source: "cluster",
				Date_Created: event.FirstTimestamp.Time.String(),
				Ref_Target: event.InvolvedObject.Name,
				Severity: eventType,
				Message:  event.Message,
			}
			repository.PopSystemevent(&systemEvent)

			// Track the most recent event time.
			if event.FirstTimestamp.Time.After(latestEventTime) {
				latestEventTime = event.FirstTimestamp.Time
			}
		}
	}
	// Update the cache with the latest event timestamp if available.
	if !latestEventTime.IsZero() {
		cache.Set("monitoring-cluster:last_event", latestEventTime.Format(time.RFC3339), time.Second * 30)
	}
}