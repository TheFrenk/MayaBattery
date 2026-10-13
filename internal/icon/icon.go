package icon

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"strconv"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	ModeNumeric = "numeric"
	ModeBattery = "battery"
)

var sizes = []int{16, 20, 24, 32, 48}

var defaultFont = mustParse(gobold.TTF)

func mustParse(data []byte) *opentype.Font {
	f, err := opentype.Parse(data)
	if err != nil {
		panic(err)
	}
	return f
}

type Renderer struct{ font *opentype.Font }

func NewRenderer() *Renderer { return &Renderer{font: defaultFont} }

func (r *Renderer) LoadFont(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	f, err := opentype.Parse(data)
	if err != nil {
		return fmt.Errorf("нужен .ttf или .otf: %w", err)
	}
	if _, err := opentype.NewFace(f, &opentype.FaceOptions{Size: 16, DPI: 72}); err != nil {
		return err
	}
	r.font = f
	return nil
}

func (r *Renderer) ResetFont() { r.font = defaultFont }

func (r *Renderer) face(px float64) font.Face {
	f, _ := opentype.NewFace(r.font, &opentype.FaceOptions{
		Size: px, DPI: 72, Hinting: font.HintingFull,
	})
	return f
}

func levelColor(percent int, charging bool) color.NRGBA {
	switch {
	case charging:
		return color.NRGBA{R: 80, G: 190, B: 255, A: 255}
	case percent <= 20:
		return color.NRGBA{R: 255, G: 80, B: 80, A: 255}
	case percent <= 40:
		return color.NRGBA{R: 255, G: 195, B: 60, A: 255}
	default:
		return color.NRGBA{R: 90, G: 225, B: 140, A: 255}
	}
}

var grey = color.NRGBA{R: 150, G: 150, B: 150, A: 255}

func (r *Renderer) renderText(text string, size int, c color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	var face font.Face
	var bounds fixed.Rectangle26_6
	for px := float64(size) * 1.15; px >= 6; px -= 0.5 {
		face = r.face(px)
		bounds, _ = font.BoundString(face, text)
		w := (bounds.Max.X - bounds.Min.X).Ceil()
		h := (bounds.Max.Y - bounds.Min.Y).Ceil()
		if w <= size && h <= size {
			break
		}
	}

	w := bounds.Max.X - bounds.Min.X
	h := bounds.Max.Y - bounds.Min.Y
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot: fixed.Point26_6{
			X: (fixed.I(size)-w)/2 - bounds.Min.X,
			Y: (fixed.I(size)-h)/2 - bounds.Min.Y,
		},
	}
	d.DrawString(text)
	return img
}

func renderBattery(percent, size int, c color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	s := float64(size) / 32.0
	px := func(v float64) int { return int(v*s + 0.5) }
	white := color.NRGBA{R: 235, G: 240, B: 245, A: 255}

	rect := func(x0, y0, x1, y1 int, col color.NRGBA) {
		draw.Draw(img, image.Rect(x0, y0, x1, y1), image.NewUniform(col), image.Point{}, draw.Src)
	}
	rect(px(3), px(8), px(27), px(25), white)
	rect(px(27), px(13), px(30), px(20), white)
	rect(px(3)+1, px(8)+1, px(27)-1, px(25)-1, color.NRGBA{})
	fillW := (px(27) - px(3) - 4) * percent / 100
	rect(px(3)+2, px(8)+2, px(3)+2+fillW, px(25)-2, c)
	return img
}

func build(fn func(size int) *image.NRGBA) []byte {
	imgs := make([]*image.NRGBA, 0, len(sizes))
	for _, s := range sizes {
		imgs = append(imgs, fn(s))
	}
	return encodeICO(imgs)
}

func (r *Renderer) Make(percent int, mode string, charging bool) []byte {
	percent = max(0, min(100, percent))
	c := levelColor(percent, charging)
	return build(func(size int) *image.NRGBA {
		if mode == ModeBattery {
			return renderBattery(percent, size, c)
		}
		return r.renderText(strconv.Itoa(percent), size, c)
	})
}

func (r *Renderer) Disconnected() []byte {
	return build(func(size int) *image.NRGBA {
		return r.renderText("--", size, grey)
	})
}
