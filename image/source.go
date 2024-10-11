package image

import (
	"bytes"
	"image"
	"image/color"
	_ "image/gif"
	"log"
	"os"
	"strings"
	"sync/atomic"

	_ "golang.org/x/image/webp"

	"gioui.org/op/paint"
	"golang.org/x/image/draw"
)

type Quality uint8

const (
	// scale source image using the nearest neighbor interpolator.
	Low Quality = iota
	// scale using ApproxBiLinear interpolator.
	Medium
	// scale using BiLinear interpolator. It is slow but gives high quality.
	High
	// scale using Catmull-Rom interpolator. It is very slow but gives the
	// best quality among the four.
	Highest
)

// ImageSource wraps a local or remote image. Only jpeg, png and gif format for
// is supported. When displaying, image scaled by the specified size and cached.
type ImageSource struct {
	// location has the value of "memory" for byte buffer. And it holds the url
	// of the image for network image. It's the local file path for local image.
	location string
	// src is the buffer of the source image.
	src     []byte
	srcSize image.Point
	// The name of the registered image format, like "jpeg", "gif" or "png".
	format string

	// for local and network image
	isLoading atomic.Bool
	loadErr   error

	// cache the last scaled image
	cache *paint.ImageOp
	// onLoaded is a callback called when image data is loaded.
	onLoaded func()
	// Select the quality of the scaled image.
	ScaleQuality Quality
	// Choose whether to buffer src image or not. Buffering reduces frequent
	// image loading, but at the price of higher memory usage. Has no effect
	// for image reading from bytes.
	UseSrcBuf bool
}

// ImageFromBuf loads an image from bytes buffer.
func ImageFromBuf(src []byte) *ImageSource {
	img := &ImageSource{location: "memory", ScaleQuality: Medium}
	img.loadData(src)
	return img
}

// ImageFromFile load an image from local filesystem or from network lazily.
// For eager image loading, use ImageFromBuf instead.
func ImageFromFile(src string) *ImageSource {
	return &ImageSource{location: src, ScaleQuality: Medium}
}

func (img *ImageSource) loadData(src []byte) error {
	srcReader := bytes.NewReader(src)
	imgConfig, format, err := image.DecodeConfig(srcReader)
	if err != nil {
		img.loadErr = err
		return err
	}

	img.srcSize = image.Point{X: imgConfig.Width, Y: imgConfig.Height}
	img.format = format
	img.src = src
	return nil
}

// IsNetworkImg check if this image is loaded/to be loaded from network.
func (img *ImageSource) IsNetworkImg() bool {
	return strings.HasPrefix(img.location, "http://") || strings.HasPrefix(img.location, "https://")
}

// loads the img from network asynchronously.
func (img *ImageSource) load() {
	if img.location == "memory" || img.src != nil {
		return
	}

	if !img.isLoading.CompareAndSwap(false, true) {
		return
	}

	go func() {
		defer func() {
			if img.isLoading.CompareAndSwap(true, false) {
				if img.onLoaded != nil {
					img.onLoaded()
				}
			}
		}()
		if !img.IsNetworkImg() {
			// Try to load from the file system.
			imgBuf, err := os.ReadFile(img.location)
			if err != nil {
				img.loadErr = err
				return
			}

			img.loadErr = img.loadData(imgBuf)
			return
		}

		// load from remote server.
		httpClient := newClient()
		imgBuf, err := httpClient.Download(img.location)
		if err != nil {
			img.loadErr = err
			return
		}

		img.loadErr = img.loadData(imgBuf)
	}()
}

func (img *ImageSource) Error() error {
	return img.loadErr
}

func (img *ImageSource) ScaleBySize(size image.Point) (*paint.ImageOp, error) {
	return img.scale(size)
}

func (img *ImageSource) ScaleByRatio(ratio float32) (*paint.ImageOp, error) {
	if ratio <= 0 {
		ratio = 1.0
	}

	width, height := img.srcSize.X, img.srcSize.Y
	size := image.Point{X: int(float32(width) * ratio), Y: int(float32(height) * ratio)}

	return img.scale(size)
}

func (img *ImageSource) scale(size image.Point) (*paint.ImageOp, error) {
	if size == (image.Point{}) {
		size = img.srcSize
	}

	if img.cache != nil && size == img.cache.Size() {
		return img.cache, nil
	}

	srcImg, _, err := image.Decode(bytes.NewReader(img.src))
	if err != nil {
		return nil, err
	}

	defer func() {
		if img.location == "memory" || img.UseSrcBuf {
			return
		}
		img.src = nil
	}()

	dest := image.NewRGBA(image.Rectangle{Max: size})
	var interpolator draw.Interpolator
	switch img.ScaleQuality {
	case Low:
		interpolator = draw.NearestNeighbor
	case Medium:
		interpolator = draw.ApproxBiLinear
	case High:
		interpolator = draw.BiLinear
	case Highest:
		interpolator = draw.CatmullRom
	default:
		interpolator = draw.ApproxBiLinear
	}

	interpolator.Scale(dest, dest.Bounds(), srcImg, srcImg.Bounds(), draw.Src, nil)
	op := paint.NewImageOp(dest)
	img.cache = &op
	return img.cache, nil
}

var emptyImg = paint.NewImageOp(image.NewUniform(color.Opaque))

// ImageOp scales the src image dynamically or scales to the expected size if size if set.
// If the passed size is the zero value of image Point, image is not scaled.
func (img *ImageSource) ImageOp(size image.Point) *paint.ImageOp {
	img.load()
	if img.isLoading.Load() {
		if img.cache != nil {
			return img.cache
		}
		return &emptyImg
	}
	if img.loadErr != nil {
		return &emptyImg
	}

	width, height := img.srcSize.X, img.srcSize.Y
	ratio := min(float32(size.X)/float32(width), float32(size.Y)/float32(height))
	if ratio > 1.0 {
		// Do not scale up, do it in Gio image.
		ratio = 1.0
	}
	scaledImg, err := img.ScaleByRatio(ratio)
	if err != nil {
		log.Printf("scale image failed: %v", err)
		return &emptyImg
	}

	return scaledImg
}

func (img *ImageSource) Size() image.Point {
	return img.srcSize
}

func (img *ImageSource) Format() string {
	return img.format
}
