#include "bridge.h"

int init_image(void *buf, size_t len, int imageType, VipsImage **out) {
    int code = 1;

    switch(imageType) {
        case JPEG:
        code = vips_jpegload_buffer(buf, len, out, "access", VIPS_ACCESS_RANDOM, NULL);
        break;
        case PNG:
        code = vips_pngload_buffer(buf, len, out, "access", VIPS_ACCESS_RANDOM, NULL);
        break;
        case WEBP:
        code = vips_webpload_buffer(buf, len, out, "access", VIPS_ACCESS_RANDOM, NULL);
        break;
        case TIFF:
        code = vips_tiffload_buffer(buf, len, out, "access", VIPS_ACCESS_RANDOM, NULL);
        break;
    }

    return code;
}

int load_jpeg_buffer(void *buf, size_t len, VipsImage **out, int shrink) {
    if (shrink > 0) {
        return vips_jpegload_buffer(buf, len, out, "shrink", shrink, NULL);
    } else {
        return vips_jpegload_buffer(buf, len, out, NULL);
    }
}

int save_jpeg_buffer(VipsImage *in, void **buf, size_t *len, int strip, int quality, int interlace) {
    return vips_jpegsave_buffer(in, buf, len,
        "strip", strip,
        "Q", quality,
        "optimize_coding", TRUE,
        "interlace", interlace,
        NULL);
}

int save_png_buffer(VipsImage *in, void **buf, size_t *len, int strip, int compression, int quality, int interlace) {
    return vips_pngsave_buffer(in, buf, len,
        "strip", strip,
        "compression", compression,
        "interlace", interlace,
        "filter", VIPS_FOREIGN_PNG_FILTER_NONE,
        NULL);
}

int save_webp_buffer(VipsImage *in, void **buf, size_t *len, int strip, int quality, int lossless) {
    return vips_webpsave_buffer(in, buf, len,
        "strip", strip,
        "Q", quality,
        "lossless", lossless,
        NULL);
}

int save_tiff_buffer(VipsImage *in, void **buf, size_t *len) {
    return vips_tiffsave_buffer(in, buf, len, NULL);
}

int resize_image(VipsImage *in, VipsImage **out, double scale, double vscale, int kernel) {
    if (vscale > 0) {
        return vips_resize(in, out, scale, "vscale", vscale, "kernel", kernel, NULL);
    }
    return vips_resize(in, out, scale, "kernel", kernel, NULL);
}

int extract_image_area(VipsImage *in, VipsImage **out, int left, int top, int width, int height) {
    return vips_extract_area(in, out, left, top, width, height, NULL);
}

int gravity_image(VipsImage *in, VipsImage **out, int direction, int width, int height, int extend, double r, double g, double b) {
    if (extend == VIPS_EXTEND_BACKGROUND) {
        double background[3] = {r, g, b};
        VipsArrayDouble *vipsBackground = vips_array_double_new(background, 3);
        return vips_gravity(in, out, direction, width, height, "extend", extend, "background", vipsBackground, NULL);
    }
    return vips_gravity(in, out, direction, width, height, "extend", extend, NULL);
}

void gobject_set_property(VipsObject *object, const char *name, const GValue *value) {
    VipsObjectClass *object_class = VIPS_OBJECT_GET_CLASS(object);
    GType type = G_VALUE_TYPE(value);

    GParamSpec *pspec;
    VipsArgumentClass *argument_class;
    VipsArgumentInstance *argument_instance;

    if (vips_object_get_argument(object, name, &pspec, &argument_class, &argument_instance)) {
        vips_warn(NULL, "gobject warning: %s", vips_error_buffer());
        vips_error_clear();
        return;
    }

    if (G_IS_PARAM_SPEC_ENUM(pspec) && type == G_TYPE_STRING) {
        GType pspec_type = G_PARAM_SPEC_VALUE_TYPE(pspec);

        int enum_value;
        GValue value2 = {0};

        if ((enum_value = vips_enum_from_nick(object_class->nickname, pspec_type, g_value_get_string(value))) < 0) {
            vips_warn(NULL, "gobject warning: %s", vips_error_buffer());
            vips_error_clear();
            return;
        }

        g_value_init(&value2, pspec_type);
        g_value_set_enum(&value2, enum_value);
        g_object_set_property(G_OBJECT(object), name, &value2);
        g_value_unset(&value2);
    } else {
        g_object_set_property(G_OBJECT(object), name, value);
    }
}
