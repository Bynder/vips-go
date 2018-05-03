package vips

// #cgo pkg-config: vips
// #include "bridge.h"
import "C"

import "unsafe"

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func freeCString(s *C.char) {
	C.free(unsafe.Pointer(s))
}

func toGboolean(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}
