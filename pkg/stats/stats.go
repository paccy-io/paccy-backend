package stats

import (
	"runtime"
	"time"

	"github.com/paccy-io/paccy-backend/pkg/logger"
)

// Başka yerden sadece bu fonksiyon çağrılacak
func StartMemoryLogger(interval time.Duration) {
	log := logger.Server()
	ticker := time.NewTicker(interval)
	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		log.WithField("alloc_kb", m.Alloc/1024).
			WithField("total_alloc_kb", m.TotalAlloc/1024).
			WithField("sys_kb", m.Sys/1024).
			WithField("gc_count", m.NumGC).
			Info("Memory statistics")
	}
}
