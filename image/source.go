package image

import (
	"bytes"
	"errors"
	"image"
	"io"
	"log"
	"os"
	"strings"

	"image/jpeg"
	//_ "image/jpeg"
	"image/png"
	//_ "image/png"

	"gioui.org/op/paint"
	"golang.org/x/image/draw"
)

// ImageSource wraps a local or remote image, and handle things like
// scaling and converting. Only jpeg and png format for the source
// image is supported.
type ImageSource struct {
	// image data
	src     []byte
	srcSize image.Point
	// The name of the registered image format, like "jpeg" or "png".
	format string

	// cache the last scaled image
	destImg image.Image
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
		return nil, errors.New("not implemented")
	}

	// Try to load from the file system.
	imgFile, err := os.ReadFile(src)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	return ImageFromReader(imgFile)

}

func (img *ImageSource) ScaleBySize(size image.Point) (image.Image, error) {
	return img.scale(size)
}

func (img *ImageSource) ScaleByRatio(ratio float32) (image.Image, error) {
	if ratio <= 0 {
		return nil, errors.New("negative scaling ratio")
	}

	width, height := img.srcSize.X, img.srcSize.Y
	size := image.Point{X: int(float32(width) * ratio), Y: int(float32(height) * ratio)}

	return img.scale(size)
}

func (img *ImageSource) scale(size image.Point) (image.Image, error) {
	if img.destImg != nil && size == img.destImg.Bounds().Size() {
		return img.destImg, nil
	}

	srcImg, _, err := image.Decode(bytes.NewReader(img.src))
	if err != nil {
		log.Println("err: ", err)
		return nil, err
	}

	dest := image.NewRGBA(image.Rectangle{Max: size})
	draw.NearestNeighbor.Scale(dest, dest.Bounds(), srcImg, srcImg.Bounds(), draw.Src, nil)
	img.destImg = dest
	return dest, nil
}

// Save scale the image if required, and encode image.Image to PNG/JPEG image, finally write
// to the provided writer. Format must be value of "jpeg" or "png".
func (img *ImageSource) Save(out io.Writer, format string, size image.Point) error {
	if img.destImg == nil && size != img.destImg.Bounds().Size() {
		img.ScaleBySize(size)
	}

	if format == "" {
		format = img.format
	}

	if format == "jpeg" {
		return jpeg.Encode(out, img.destImg, &jpeg.Options{Quality: 100})
	}

	if format == "png" {
		return png.Encode(out, img.destImg)
	}

	return errors.New("unknown image format: " + format)
}

// ImageOp scales the src image to make it fit within the constraint specified by size
// and convert it to Gio ImageOp. If size has zero value of image Point, image is not scaled.
func (img *ImageSource) ImageOp(size image.Point) (paint.ImageOp, error) {
	if size == (image.Point{}) || size.X <= 0 || size.Y <= 0 {
		img.ScaleBySize(size)
		return paint.NewImageOp(img.destImg), nil
	}

	width, height := img.srcSize.X, img.srcSize.Y
	ratio := min(float32(size.X)/float32(width), float32(size.Y)/float32(height))
	scaledImg, err := img.ScaleByRatio(ratio)
	if err != nil {
		log.Println("scale image failed:", err)
		return paint.ImageOp{}, err
	}

	return paint.NewImageOp(scaledImg), nil
}

func (img *ImageSource) Size() image.Point {
	return img.srcSize
}
