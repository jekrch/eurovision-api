package auth

import (
	"time"
)

func StartCleanupJob() {
	ticker := time.NewTicker(24 * time.Hour)
	for range ticker.C {
		cleanupUnconfirmedUsers()
	}
}
