#include <stdlib.h>
#include <vips/vips.h>

enum types {
    UNKNOWN = 0,
    JPEG,
    WEBP,
    PNG,
    TIFF
};

int init_image(void *buf, size_t len, int imageType, VipsImage **out);

int load_jpeg_buffer(void *buf, size_t len, VipsImage **out, int shrink);
int save_jpeg_buffer(VipsImage *image, void **buf, size_t *len, int strip, int quality, int interlace);
int save_png_buffer(VipsImage *in, void **buf, size_t *len, int strip, int compression, int quality, int interlace);
int save_webp_buffer(VipsImage *in, void **buf, size_t *len, int strip, int quality, int lossless);
int save_tiff_buffer(VipsImage *in, void **buf, size_t *len);

int resize_image(VipsImage *in, VipsImage **out, double scale, double vscale, int kernel);
int extract_image_area(VipsImage *in, VipsImage **out, int left, int top, int width, int height);
int gravity_image(VipsImage *in, VipsImage **out, int direction, int width, int height, int extend, double r, double g, double b);

void gobject_set_property(VipsObject *object, const char *name, const GValue *value);

static int has_alpha_channel(VipsImage *image) {
    return (
               (image->Bands == 2 && image->Type == VIPS_INTERPRETATION_B_W) ||
               (image->Bands == 4 && image->Type != VIPS_INTERPRETATION_CMYK) ||
               (image->Bands == 5 && image->Type == VIPS_INTERPRETATION_CMYK))
        ? 1
        : 0;
}

#if VIPS_MAJOR_VERSION < 8
error_requires_version_8_6
#endif

#if VIPS_MAJOR_VERSION == 8 && VIPS_MINOR_VERSION < 6
error_requires_version_8_6
#endif
