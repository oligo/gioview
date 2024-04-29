package image

import (
	"bytes"
	"errors"
	"image"
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
	// image data
	src     []byte
	srcSize image.Point
	// The name of the registered image format, like "jpeg" or "png".
	format string

	// cache the last scaled image
	cache *paint.ImageOp
}

// ImageFromReader loads an image from a io.Reader.
// Image bytes buffer can be wrapped using a bytes.Reader to get an
// ImageSource.
func ImageFromReader(src []byte) (*ImageSource, error) {
	imgConfig, format, err := image.DecodeConfig(bytes.NewReader(src))
	if err != nil {
		return nil, err
	}

	return &ImageSource{
		src:     src,
		srcSize: image.Point{X: imgConfig.Width, Y: imgConfig.Height},
		format:  format,
	}, nil
}

// ImageFromFile load an image from local filesystem or from network.
func ImageFromFile(src string) (*ImageSource, error) {
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		// load from remote server.
		httpClient := newClient()
		_, resp, err := httpClient.Download(src)
		if err != nil {
			return nil, err
		}

		defer resp.Close()

		imgFile, err := io.ReadAll(resp)
		if err != nil {
			return nil, err
		}

		return ImageFromReader(imgFile)
	}

	// Try to load from the file system.
	imgFile, err := os.ReadFile(src)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	return ImageFromReader(imgFile)

}

func (img *ImageSource) ScaleBySize(size image.Point) (*paint.ImageOp, error) {
	return img.scale(size)
}

func (img *ImageSource) ScaleByRatio(ratio float32) (*paint.ImageOp, error) {
	if ratio <= 0 {
		return nil, errors.New("negative scaling ratio")
	}

	width, height := img.srcSize.X, img.srcSize.Y
	size := image.Point{X: int(float32(width) * ratio), Y: int(float32(height) * ratio)}

	return img.scale(size)
}

func (img *ImageSource) scale(size image.Point) (*paint.ImageOp, error) {
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

// ImageOp scales the src image to make it fit within the constraint specified by size
// and convert it to Gio ImageOp. If size has zero value of image Point, image is not scaled.
func (img *ImageSource) ImageOp(size image.Point) (*paint.ImageOp, error) {
	if img.cache != nil {
		return img.cache, nil
	}

	if size == (image.Point{}) || size.X <= 0 || size.Y <= 0 {
		return img.ScaleBySize(size)
	}

	width, height := img.srcSize.X, img.srcSize.Y
	ratio := min(float32(size.X)/float32(width), float32(size.Y)/float32(height))
	scaledImg, err := img.ScaleByRatio(ratio)
	if err != nil {
		log.Println("scale image failed:", err)
		return &paint.ImageOp{}, err
	}

	return scaledImg, nil
}

func (img *ImageSource) Size() image.Point {
	return img.srcSize
}
