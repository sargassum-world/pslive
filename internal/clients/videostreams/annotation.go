package videostreams

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

const timestampFormat = "2006-01-02 15:04:05.000 UTC"

type Metadata struct {
	Width     int
	Height    int
	Timestamp time.Time

	FPSCount  int64
	FPSPeriod float32
	Quality   int
}

func (m Metadata) String() string {
	return fmt.Sprintf(
		"[%dx%d] [%s] [%.1f fps] [q=%3d]",
		m.Width, m.Height,
		m.Timestamp.UTC().Format(timestampFormat),
		float32(m.FPSCount)/m.FPSPeriod,
		m.Quality,
	)
}

// Drawing

var fontFace = inconsolata.Regular8x16

func AddLabel(im draw.Image, label string, co color.Color, x, y int) {
	d := &font.Drawer{
		Dst:  im,
		Src:  image.NewUniform(co),
		Face: fontFace,
		Dot: fixed.Point26_6{
			X: fixed.I(x),
			Y: fixed.I(y) + fontFace.Metrics().Height,
		},
	}
	d.DrawString(label)
}

const annotationBarMargin = 4

var (
	annotationBarHeight = fontFace.Metrics().Height.Round() + annotationBarMargin
	annotationOffset    = image.Point{
		X: 0,
		Y: annotationBarHeight,
	}
)

func CopyForAnnotation(source image.Image) (output draw.Image) {
	output = image.NewRGBA(image.Rectangle{
		Min: source.Bounds().Min,
		Max: source.Bounds().Max.Add(annotationOffset),
	})
	draw.Draw(output, image.Rectangle{
		Min: source.Bounds().Min.Add(annotationOffset),
		Max: source.Bounds().Max.Add(annotationOffset),
	}, source, image.Point{}, draw.Src)
	return output
}

func Annotate(im draw.Image, timestamp time.Time, metadata fmt.Stringer) {
	const max = 255
	backgroundColor := color.RGBA{A: max}
	labelColor := color.RGBA{max, max, max, max}
	draw.Draw(im, image.Rectangle{
		Min: im.Bounds().Min,
		Max: image.Point{
			X: im.Bounds().Max.X,
			Y: annotationBarHeight,
		},
	}, &image.Uniform{backgroundColor}, image.Point{}, draw.Src)
	AddLabel(im, metadata.String(), labelColor, annotationBarMargin, 0)
}
