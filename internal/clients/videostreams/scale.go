package videostreams

import (
	"image"

	"golang.org/x/image/draw"
)

func CopyRGBA(im image.Image) draw.Image {
	result := image.NewRGBA(im.Bounds())
	draw.Draw(result, result.Bounds(), im, im.Bounds().Min, draw.Src)
	return result
}

func ResizeToHeight(im image.Image, height int) draw.Image {
	aspectRatio := float32(im.Bounds().Max.X) / float32(im.Bounds().Max.Y)
	width := int(aspectRatio * float32(height))
	result := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.ApproxBiLinear.Scale(result, result.Rect, im, im.Bounds(), draw.Over, nil)
	return result
}
