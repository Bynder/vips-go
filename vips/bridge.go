package vips

// #cgo pkg-config: vips
// #include "bridge.h"
import "C"
import (
	"errors"
	"fmt"
	"log"
	"math"
	"runtime"
	dbg "runtime/debug"
	"unsafe"
)

const (
	defaultQuality     = 90
	defaultCompression = 6
	maxScaleFactor     = 10
)

var (
	// ErrUnsupportedImageFormat when image type is unsupported
	ErrUnsupportedImageFormat = errors.New("Unsupported image format")
)

// ImageRef contains a libvips image and manages its lifecycle. You should
// close an image when done or it will leak until the next GC
type ImageRef struct {
	Image  *C.VipsImage
	Format ImageType

	// NOTE(d): We keep a reference to this so that the input buffer is
	// never garbage collected during processing. Some image loaders use random
	// access transcoding and therefore need the original buffer to be in memory.
	Buf []byte
}

// Close closes an image and frees internal memory associated with it
func (ref *ImageRef) Close() {
	ref.Image = nil
	ref.Buf = nil
}

// newImageRef creates an image reference that should be closed when not used anymore.
func newImageRef(vipsImage *C.VipsImage, format ImageType) *ImageRef {
	stream := &ImageRef{
		Image:  vipsImage,
		Format: format,
	}
	runtime.SetFinalizer(stream, finalizeImage)
	return stream
}

func finalizeImage(ref *ImageRef) {
	ref.Close()
}

// InitImage opens image in vips
func InitImage(buf []byte) (*ImageRef, error) {
	// Reference buf here so it's not garbage collected during image initialization.
	defer runtime.KeepAlive(buf)

	var image *C.VipsImage
	imageType := vipsDetermineImageType(buf)

	if imageType == ImageTypeUnknown {
		if len(buf) > 2 {
			log.Printf("Failed to understand image format size=%d %x %x %x", len(buf), buf[0], buf[1], buf[2])
		} else {
			log.Printf("Failed to understand image format size=%d", len(buf))
		}
		return nil, ErrUnsupportedImageFormat
	}

	len := C.size_t(len(buf))
	imageBuf := unsafe.Pointer(&buf[0])

	err := C.init_image(imageBuf, len, C.int(imageType), &image)
	if err != 0 {
		return nil, handleVipsError()
	}

	imageRef := newImageRef(image, imageType)
	imageRef.Buf = buf
	return imageRef, nil
}

func vipsPrepareForExport(params *ExportParams) {
	if params.Quality == 0 {
		params.Quality = defaultQuality
	}

	if params.Compression == 0 {
		params.Compression = defaultCompression
	}
}

// SaveBuffer saves vips image to bytes
func SaveBuffer(imageRef *ImageRef, params *ExportParams) ([]byte, error) {
	vipsPrepareForExport(params)

	cLen := C.size_t(0)
	var cErr C.int
	interlaced := C.int(boolToInt(params.Interlaced))
	quality := C.int(params.Quality)
	lossless := C.int(boolToInt(params.Lossless))
	stripMetadata := C.int(boolToInt(params.StripMetadata))
	format := params.Format

	if format != ImageTypeUnknown && !isTypeSupported(format) {
		return nil, fmt.Errorf("cannot save to %#v", imageTypes[format])
	}

	var ptr unsafe.Pointer

	switch format {
	case ImageTypeWEBP:
		cErr = C.save_webp_buffer(imageRef.Image, &ptr, &cLen, stripMetadata, quality, lossless)
	case ImageTypePNG:
		cErr = C.save_png_buffer(imageRef.Image, &ptr, &cLen, stripMetadata, C.int(params.Compression), quality, interlaced)
	case ImageTypeTIFF:
		cErr = C.save_tiff_buffer(imageRef.Image, &ptr, &cLen)
	default:
		format = ImageTypeJPEG
		cErr = C.save_jpeg_buffer(imageRef.Image, &ptr, &cLen, stripMetadata, quality, interlaced)
	}

	if int(cErr) != 0 {
		return nil, handleVipsError()
	}

	buf := C.GoBytes(ptr, C.int(cLen))
	C.g_free(C.gpointer(ptr))
	return buf, nil
}

func vipsDetermineImageType(buf []byte) ImageType {
	if len(buf) < 12 {
		return ImageTypeUnknown
	}
	if buf[0] == 0xFF && buf[1] == 0xD8 && buf[2] == 0xFF {
		return ImageTypeJPEG
	}
	if buf[0] == 0x89 && buf[1] == 0x50 && buf[2] == 0x4E && buf[3] == 0x47 {
		return ImageTypePNG
	}
	if isTypeSupported(ImageTypeTIFF) &&
		((buf[0] == 0x49 && buf[1] == 0x49 && buf[2] == 0x2A && buf[3] == 0x0) ||
			(buf[0] == 0x4D && buf[1] == 0x4D && buf[2] == 0x0 && buf[3] == 0x2A)) {
		return ImageTypeTIFF
	}
	if isTypeSupported(ImageTypeWEBP) && buf[8] == 0x57 && buf[9] == 0x45 && buf[10] == 0x42 && buf[11] == 0x50 {
		return ImageTypeWEBP
	}
	return ImageTypeUnknown
}

// ResizeImage resizes input image with one of the specified interpolation algorithms
func ResizeImage(imageRef *ImageRef, scale, vscale float64, kernel Kernel) error {
	var output *C.VipsImage

	// Let's not be insane
	scale = math.Min(scale, maxScaleFactor)
	vscale = math.Min(vscale, maxScaleFactor)

	defer C.g_object_unref(C.gpointer(imageRef.Image))
	if err := C.resize_image(imageRef.Image, &output, C.double(scale), C.double(vscale), C.int(kernel)); err != 0 {
		return handleVipsError()
	}
	imageRef.Image = output
	return nil
}

// HasAlphaChannel checks is image have alpha channel or not
func HasAlphaChannel(imageRef *ImageRef) bool {
	return int(C.has_alpha_channel(imageRef.Image)) > 0
}

// ExtractImageArea crops image to specified size
func ExtractImageArea(imageRef *ImageRef, left, top, width, height int) error {
	var output *C.VipsImage
	defer C.g_object_unref(C.gpointer(imageRef.Image))
	err := C.extract_image_area(imageRef.Image, &output, C.int(left), C.int(top), C.int(width), C.int(height))
	if err != 0 {
		return handleVipsError()
	}
	imageRef.Image = output
	return nil
}

// GravityImage changes canvas size with specified direction and canvas fill logic
func GravityImage(imageRef *ImageRef, direction CompassDirection, width, height int, extend Extend, red, green, blue float64) error {
	var output *C.VipsImage
	defer C.g_object_unref(C.gpointer(imageRef.Image))
	err := C.gravity_image(imageRef.Image, &output, C.int(direction), C.int(width), C.int(height), C.int(extend), C.double(red), C.double(green), C.double(blue))
	if err != 0 {
		return handleVipsError()
	}
	imageRef.Image = output
	return nil
}

func handleVipsError() error {
	defer C.vips_thread_shutdown()
	defer C.vips_error_clear()

	s := C.GoString(C.vips_error_buffer())
	stack := string(dbg.Stack())
	return fmt.Errorf("%s\nStack:\n%s", s, stack)
}
