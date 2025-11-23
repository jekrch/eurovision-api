package auth

import (
	"time"
)

/*
 * StartCleanupJob starts a cleanup job that runs every 24 hours to remove
 * unconfirmed users that have not confirmed their email address within 24 hours.
 */
func StartCleanupJob() {
	ticker := time.NewTicker(24 * time.Hour)
	for range ticker.C {
		cleanupUnconfirmedUsers()
	}
}
