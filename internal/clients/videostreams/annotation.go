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

type AnnotationMetadata struct {
	Width       int
	Height      int
	Timestamp   time.Time
	JPEGQuality int

	FPSCount  int64
	FPSPeriod float32
}

func (m AnnotationMetadata) String() string {
	return fmt.Sprintf(
		"[%dx%d] [%s] [%.1f fps] [q=%3d]",
		m.Width, m.Height,
		m.Timestamp.UTC().Format(timestampFormat),
		float32(m.FPSCount)/m.FPSPeriod,
		m.JPEGQuality,
	)
}

func (m AnnotationMetadata) WithFrameData(f *ImageFrame) AnnotationMetadata {
	m.Width = f.Im.Bounds().Max.X
	m.Height = f.Im.Bounds().Max.Y
	m.Timestamp = f.Meta.ReceiveTime
	m.JPEGQuality = f.Meta.Settings.JPEGEncodeQuality
	return m
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

const annotationBarMargin = 6

var LineHeight = fontFace.Metrics().Height.Round()

func AddAnnotationPadding(source image.Image, topLines, bottomLines int) (output draw.Image) {
	topHeight := 0
	if topLines > 0 {
		topHeight = topLines*LineHeight + annotationBarMargin
	}
	bottomHeight := 0
	if bottomLines > 0 {
		bottomHeight = bottomLines*LineHeight + annotationBarMargin
	}

	output = image.NewRGBA(image.Rectangle{
		Min: source.Bounds().Min,
		Max: source.Bounds().Max.Add(image.Point{
			Y: topHeight + bottomHeight,
		}),
	})
	draw.Draw(output, image.Rectangle{
		Min: source.Bounds().Min.Add(image.Point{
			Y: topHeight,
		}),
		Max: source.Bounds().Max.Add(image.Point{
			Y: topHeight + bottomHeight,
		}),
	}, source, image.Point{}, draw.Src)
	return output
}

func AnnotateTop(im draw.Image, annotations string, lines int) {
	const max = 255
	height := lines*LineHeight + annotationBarMargin
	backgroundColor := color.RGBA{A: max}
	labelColor := color.RGBA{max, max, max, max}
	draw.Draw(im, image.Rectangle{
		Min: im.Bounds().Min,
		Max: image.Point{
			X: im.Bounds().Max.X,
			Y: height,
		},
	}, &image.Uniform{backgroundColor}, image.Point{}, draw.Src)
	AddLabel(im, annotations, labelColor, annotationBarMargin, 0)
}
