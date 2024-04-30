package image

import (
	"bytes"
	"image"
	"image/color"
	"io"
	"log"
	"os"
	"strings"

	"gioui.org/op/paint"
	"golang.org/x/image/draw"
)

// ImageSource wraps a local or remote image. Only jpeg and png format for
// is supported. When displaying, image scaled by the specified size and cached.
type ImageSource struct {
	location string
	// image data
	src     []byte
	srcSize image.Point
	// The name of the registered image format, like "jpeg" or "png".
	format string

	// for network image
	isLoading bool
	loadErr   error

	// cache the last scaled image
	cache *paint.ImageOp
}

// ImageFromBuf loads an image from bytes buffer.
func ImageFromBuf(src []byte) *ImageSource {
	img := &ImageSource{location: "memory"}
	img.loadData(src)
	return img
}

// ImageFromFile load an image from local filesystem or from network.
func ImageFromFile(src string) *ImageSource {
	img := &ImageSource{location: src}
	img.load()
	return img
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
	if img.location == "memory" {
		return
	}

	img.isLoading = true

	go func() {
		defer func() { img.isLoading = false }()
		if !img.IsNetworkImg() {
			// Try to load from the file system.
			imgBuf, err := os.ReadFile(img.location)
			if err != nil {
				img.loadErr = err
				return
			}

			img.loadErr = img.loadData(imgBuf)
			return
		} else {
			// load from remote server.
			httpClient := newClient()
			_, resp, err := httpClient.Download(img.location)
			if err != nil {
				img.loadErr = err
				return
			}

			defer resp.Close()

			imgBuf, err := io.ReadAll(resp)
			if err != nil {
				img.loadErr = err
				return
			}

			img.loadErr = img.loadData(imgBuf)
		}
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

	dest := image.NewRGBA(image.Rectangle{Max: size})
	draw.NearestNeighbor.Scale(dest, dest.Bounds(), srcImg, srcImg.Bounds(), draw.Src, nil)
	op := paint.NewImageOp(dest)
	img.cache = &op
	return img.cache, nil
}

var emptyImg = paint.NewImageOp(image.NewUniform(color.Opaque))

// ImageOp scales the src image to make it fit within the constraint specified by size
// and convert it to Gio ImageOp. If size has zero value of image Point, image is not scaled.
func (img *ImageSource) ImageOp(size image.Point) *paint.ImageOp {
	if img.isLoading || img.loadErr != nil {
		// log.Println("load err: ", img.loadErr)
		return &emptyImg
	}

	width, height := img.srcSize.X, img.srcSize.Y
	ratio := min(float32(size.X)/float32(width), float32(size.Y)/float32(height))
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
