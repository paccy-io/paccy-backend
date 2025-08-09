package stats

import (
	"fmt"
	"runtime"
	"time"
)

// Başka yerden sadece bu fonksiyon çağrılacak
func StartMemoryLogger(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		fmt.Printf("[RAM] %s | Alloc: %d KB | TotalAlloc: %d KB | Sys: %d KB | GC: %d\n",
			time.Now().Format("2006-01-02 15:04:05"),
			m.Alloc/1024,
			m.TotalAlloc/1024,
			m.Sys/1024,
			m.NumGC,
		)
	}
}