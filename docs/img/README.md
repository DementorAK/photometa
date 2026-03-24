# Test Images for PhotoMeta

This directory contains sample images used for testing PhotoMeta's metadata extraction capabilities.

## License and Attribution

The sample images in this directory are sourced from various authors and are provided under their respective licenses (Creative Commons or Public Domain), as specified in the individual file descriptions below.

### Summary of Licenses Used

- **CC BY-NC-ND 2.0**: Used for all JPEG sample images.
- **CC BY 4.0**: Used for the first PNG sample.
- **CC BY-SA 4.0**: Used for the second PNG sample and the WebP sample.
- **Public Domain**: Used for the TIFF sample image.

### Required Attribution

When using these images, you must provide attribution as required by the specific license of each image. Please refer to each image's section for source links and author details.

---

## Sample Images

### sample_jpg_1.jpg

| Property | Value |
|----------|-------|
| **Author** | [Steve Walser](https://www.flickr.com/photos/bassbro/) |
| **Source** | [Flickr](https://flic.kr/p/2nZhajG) |
| **License** | CC BY-NC-ND 2.0 |

---

### sample_jpg_2.jpg

| Property | Value |
|----------|-------|
| **Author** | [Steve Walser](https://www.flickr.com/photos/bassbro/) |
| **Source** | [Flickr](https://flic.kr/p/2qdNgcK) |
| **License** | CC BY-NC-ND 2.0 |

---

### sample_jpg_3.jpg

| Property | Value |
|----------|-------|
| **Author** | [Steve Walser](https://www.flickr.com/photos/bassbro/) |
| **Source** | [Flickr](https://flic.kr/p/2nWa53y) |
| **License** | CC BY-NC-ND 2.0 |

---

### sample_jpg_4.jpg

| Property | Value |
|----------|-------|
| **Author** | [Steve Walser](https://www.flickr.com/photos/bassbro/) |
| **Source** | [Flickr](https://flic.kr/p/2nWvUCc) |
| **License** | CC BY-NC-ND 2.0 |

---

### sample_png_1.png

| Property | Value |
|----------|-------|
| **Author** | [Trougnouf (Benoit Brummer)](https://commons.wikimedia.org/wiki/User:Trougnouf) |
| **Source** | [Wikimedia Commons](https://commons.wikimedia.org/wiki/File:Untitled_sculpture_by_Barry_Flanagan_(DSCF7040).png) |
| **License** | CC BY 4.0, via Wikimedia Commons |

---

### sample_png_2.png

| Property | Value |
|----------|-------|
| **Author** | [Charlie27it](https://meta.wikimedia.org/wiki/Special:CentralAuth/Charlie27it) |
| **Source** | [Wikimedia Commons](https://commons.wikimedia.org/wiki/File:Demetra_2.png) |
| **License** | CC BY-SA 4.0, via Wikimedia Commons |

---

### sample_tiff.tif

| Property | Value |
|----------|-------|
| **Author** | [Carol M. Highsmith](https://www.loc.gov/pictures/collection/highsm/) |
| **Source** | [Wikimedia Commons](https://commons.wikimedia.org/wiki/File:Scene_from_the_252-acre_Shangri_La_Botanical_Center_and_Nature_Gardens_in_Orange,_Texas_LCCN2014630741.tif) |
| **License** | Public domain, via Wikimedia Commons |

---

### sample_webp.webp

| Property | Value |
|----------|-------|
| **Author** | [Vandomacielbr](https://commons.wikimedia.org/wiki/User:Vandomacielbr) |
| **Source** | [Wikimedia Commons](https://commons.wikimedia.org/wiki/File:SIDNEY_NO_LOUVRE.webp) |
| **License** | CC BY-SA 4.0, via Wikimedia Commons |

---

### photometa_logo.jpg

Project logo - not a test sample.

---

## Usage in Tests

These images are used by:
- `integration/integration_test.go` - For end-to-end testing
- Manual testing of metadata extraction features

## License Summary (for reference)

| Sample File | License Type | Commercial Use | Derivatives | Attribution |
|-------------|--------------|----------------|-------------|-------------|
| `sample_jpg_*.jpg` | CC BY-NC-ND 2.0 | ❌ No | ❌ No | ✅ Yes |
| `sample_png_1.png` | CC BY 4.0 | ✅ Yes | ✅ Yes | ✅ Yes |
| `sample_png_2.png` | CC BY-SA 4.0 | ✅ Yes | 🔄 ShareAlike | ✅ Yes |
| `sample_tiff.tif` | Public Domain | ✅ Yes | ✅ Yes | ❌ Optional |
| `sample_webp.webp` | CC BY-SA 4.0 | ✅ Yes | 🔄 ShareAlike | ✅ Yes |

For full license texts, visit:
- [CC BY-NC-ND 2.0](https://creativecommons.org/licenses/by-nc-nd/2.0/)
- [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)
- [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/)
- [Public Domain](https://creativecommons.org/publicdomain/mark/1.0/)