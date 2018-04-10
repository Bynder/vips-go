package vips

// #cgo pkg-config: vips
// #include "bridge.h"
import "C"
import (
	"log"
	"strings"
	"sync"
)

// ExportParams are options when exporting an image to file or buffer
type ExportParams struct {
	Format        ImageType
	Quality       int
	Compression   int
	Interlaced    bool
	Lossless      bool
	StripMetadata bool
}

// ImageType represents an image type
type ImageType int

// ImageType enum
const (
	ImageTypeUnknown ImageType = C.UNKNOWN
	ImageTypeJPEG    ImageType = C.JPEG
	ImageTypePNG     ImageType = C.PNG
	ImageTypeTIFF    ImageType = C.TIFF
	ImageTypeWEBP    ImageType = C.WEBP
)

// Kernel represents VipsKernel type
type Kernel int

// Kernel enum
const (
	KernelNearest  Kernel = C.VIPS_KERNEL_NEAREST
	KernelLinear   Kernel = C.VIPS_KERNEL_LINEAR
	KernelCubic    Kernel = C.VIPS_KERNEL_CUBIC
	KernelLanczos2 Kernel = C.VIPS_KERNEL_LANCZOS2
	KernelLanczos3 Kernel = C.VIPS_KERNEL_LANCZOS3
)

// CompassDirection represents VipsCompassDirection type
type CompassDirection int

// CompassDirection enum
const (
	CompassDirectionCentre    CompassDirection = C.VIPS_COMPASS_DIRECTION_CENTRE
	CompassDirectionNorth     CompassDirection = C.VIPS_COMPASS_DIRECTION_NORTH
	CompassDirectionEast      CompassDirection = C.VIPS_COMPASS_DIRECTION_EAST
	CompassDirectionSouth     CompassDirection = C.VIPS_COMPASS_DIRECTION_SOUTH
	CompassDirectionWest      CompassDirection = C.VIPS_COMPASS_DIRECTION_WEST
	CompassDirectionNorthEast CompassDirection = C.VIPS_COMPASS_DIRECTION_NORTH_EAST
	CompassDirectionSouthEast CompassDirection = C.VIPS_COMPASS_DIRECTION_SOUTH_EAST
	CompassDirectionSouthWest CompassDirection = C.VIPS_COMPASS_DIRECTION_SOUTH_WEST
	CompassDirectionNorthWest CompassDirection = C.VIPS_COMPASS_DIRECTION_NORTH_WEST
)

// Extend represents VipsExtend type
type Extend int

// Extend enum
const (
	ExtendBlack      Extend = C.VIPS_EXTEND_BLACK      // Black pixels
	ExtendCopy       Extend = C.VIPS_EXTEND_COPY       // Copies the image edges
	ExtendRepeat     Extend = C.VIPS_EXTEND_REPEAT     // Repeats the whole image
	ExtendMirror     Extend = C.VIPS_EXTEND_MIRROR     // Mirrors the whole image
	ExtendWhite      Extend = C.VIPS_EXTEND_WHITE      // White pixels
	ExtendBackground Extend = C.VIPS_EXTEND_BACKGROUND // Selects color from background property
	ExtendLast       Extend = C.VIPS_EXTEND_LAST       // Extends with last pixel
)

var imageTypes = map[ImageType]string{
	ImageTypeJPEG: "jpeg",
	ImageTypePNG:  "png",
	ImageTypeTIFF: "tiff",
	ImageTypeWEBP: "webp",
}

var (
	once                sync.Once
	typeLoaders         = make(map[string]ImageType)
	supportedImageTypes = make(map[ImageType]bool)
)

func isTypeSupported(imageType ImageType) bool {
	return supportedImageTypes[imageType]
}

// InitTypes initializes caches and figures out which image types are supported
func initTypes() {
	once.Do(func() {
		cType := C.CString("VipsOperation")
		defer freeCString(cType)

		for k, v := range imageTypes {
			name := strings.ToLower("VipsForeignLoad" + v)
			typeLoaders[name] = k
			typeLoaders[name+"buffer"] = k

			cFunc := C.CString(v + "load")
			defer freeCString(cFunc)

			ret := C.vips_type_find(
				cType,
				cFunc)
			log.Printf("Registered image type loader type=%s", v)
			supportedImageTypes[k] = int(ret) != 0
		}
	})
}
