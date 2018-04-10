package vips

// #cgo pkg-config: vips
// #include "vips/vips.h"
import "C"
import (
	"fmt"
	"log"
	"runtime"
	"sync"
)

const (
	defaultConcurrencyLevel = 1
	defaultMaxCacheMem      = 100 * 1024 * 1024
	defaultMaxCacheSize     = 500
)

var (
	running           = false
	initLock          sync.Mutex
	statCollectorDone chan struct{}
)

// StartupConfig allows fine-tuning of libvips library
type StartupConfig struct {
	ConcurrencyLevel int
	MaxCacheFiles    int
	MaxCacheMem      int
	MaxCacheSize     int
	ReportLeaks      bool
	CacheTrace       bool
}

// Startup sets up the vips support and ensures the versions are correct. Pass in nil for
// default configuration.
func Startup(config *StartupConfig) {
	initLock.Lock()
	runtime.LockOSThread()
	defer initLock.Unlock()
	defer runtime.UnlockOSThread()

	if running {
		log.Print("warning libvips already started")
		return
	}

	if C.VIPS_MAJOR_VERSION < 8 {
		panic("Requires libvips version 8.3+")
	}

	if C.VIPS_MINOR_VERSION < 3 {
		panic("Requires libvips version 8.3+")
	}

	cName := C.CString("govips")
	defer freeCString(cName)

	err := C.vips_init(cName)
	if err != 0 {
		panic(fmt.Sprintf("Failed to start vips code=%d", err))
	}

	running = true

	C.vips_concurrency_set(defaultConcurrencyLevel)
	C.vips_cache_set_max(defaultMaxCacheSize)
	C.vips_cache_set_max_mem(defaultMaxCacheMem)

	if config != nil {
		C.vips_leak_set(toGboolean(config.ReportLeaks))

		if config.ConcurrencyLevel > 0 {
			C.vips_concurrency_set(C.int(config.ConcurrencyLevel))
		}
		if config.MaxCacheFiles > 0 {
			C.vips_cache_set_max_files(C.int(config.MaxCacheFiles))
		}
		if config.MaxCacheMem > 0 {
			C.vips_cache_set_max_mem(C.size_t(config.MaxCacheMem))
		}
		if config.MaxCacheSize > 0 {
			C.vips_cache_set_max(C.int(config.MaxCacheSize))
		}
		if config.CacheTrace {
			C.vips_cache_set_trace(toGboolean(true))
		}
	}

	log.Printf("Vips started with concurrency=%d cache_max_files=%d cache_max_mem=%d cache_max=%d",
		int(C.vips_concurrency_get()),
		int(C.vips_cache_get_max_files()),
		int(C.vips_cache_get_max_mem()),
		int(C.vips_cache_get_max()))

	initTypes()
}
