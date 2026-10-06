package job

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
	"viral-game-network/src/cache"
	"viral-game-network/src/database/repository"
	dbtype "viral-game-network/src/database/type"
	"viral-game-network/src/k8"
)

func Job_Monitoring() { RUN_Monitoring_ClusterEvents() }

func RUN_Monitoring_ClusterEvents() {
	_, err := cache.WithLock(context.Background(), "monitoring-cluster-events", 30*time.Second, runMonitoringClusterEvents)
	if err != nil {
		log.Printf("[Job_Monitoring]: @ClusterEvents: %v", err)
	}
}

func runMonitoringClusterEvents(ctx context.Context) error {
	// Kubernetes retains events for a limited time. Reconcile every retained
	// event by UID: timestamps cannot identify events or safely fence retries.
	events, err := k8.GetLogsClusterWithContext(ctx, nil)
	if err != nil {
		return err
	}
	var failures error
	for _, event := range events {
		if err := ctx.Err(); err != nil {
			return errors.Join(failures, err)
		}
		severity := "info"
		if event.Type == "Warning" {
			severity = "warning"
		}
		_, err := repository.PopSystemevent(&dbtype.SystemEvent{
			Ref_ID: string(event.UID), Ref_Source: "cluster",
			Ref_Target: event.InvolvedObject.Name, Severity: severity, Message: event.Message,
		})
		if err != nil {
			failures = errors.Join(failures, fmt.Errorf("event %s: %w", event.Name, err))
		}
	}
	return failures
}
