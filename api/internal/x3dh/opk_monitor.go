// internal/x3dh/opk_monitor.go
package x3dh

import (
	"context"
	"log"
	"time"
)

type OPKMonitor struct {
	store     KeyStore
	threshold int
	interval  time.Duration
}

func NewOPKMonitor(store KeyStore, threshold int, interval time.Duration) *OPKMonitor {
	return &OPKMonitor{
		store:     store,
		threshold: threshold,
		interval:  interval,
	}
}

func (m *OPKMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("🔁 OPK monitor stopped.")
			return
		case <-ticker.C:
			m.checkAllUsers(ctx)
		}
	}
}

func (m *OPKMonitor) checkAllUsers(ctx context.Context) {
	userIDs, err := m.store.ListUsersWithOPKs(ctx)
	if err != nil {
		log.Printf("❌ Failed to list users for OPK monitoring: %v", err)
		return
	}

	for _, userID := range userIDs {
		count, err := m.store.CountAvailableOPKs(ctx, userID)
		if err != nil {
			log.Printf("⚠️ Could not count OPKs for user %s: %v", userID, err)
			continue
		}

		if count < m.threshold {
			log.Printf("🚨 LOW OPKs: user %s has only %d OPKs left!", userID, count)
		}
	}
}
