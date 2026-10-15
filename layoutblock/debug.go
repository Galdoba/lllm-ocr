package layoutblock

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// VisualizeBlocks renders an overlay PNG showing bounding boxes and block indices on top of the source image.
func VisualizeBlocks(srcPath string, blocks []LayoutBlock, dstPath string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open source image: %w", err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		return fmt.Errorf("decode PNG: %w", err)
	}

	rgba, ok := img.(*image.RGBA)
	if !ok {
		b := img.Bounds()
		rgba = image.NewRGBA(b)
		draw.Draw(rgba, b, img, b.Min, draw.Src)
	}

	width := rgba.Bounds().Dx()
	height := rgba.Bounds().Dy()

	for i, block := range blocks {
		x0 := block.BBox.X0 * width / 1000
		y0 := block.BBox.Y0 * height / 1000
		x1 := block.BBox.X1 * width / 1000
		y1 := block.BBox.Y1 * height / 1000

		if x0 < 0 {
			x0 = 0
		}
		if y0 < 0 {
			y0 = 0
		}
		if x1 > width {
			x1 = width
		}
		if y1 > height {
			y1 = height
		}
		if x1 <= x0 || y1 <= y0 {
			continue
		}

		rect := image.Rect(x0, y0, x1, y1)

		overlayColor := color.RGBA{R: 255, G: 165, B: 0, A: 128}
		draw.Draw(rgba, rect, &image.Uniform{C: overlayColor}, image.Point{}, draw.Over)

		text := fmt.Sprintf("%d", i)
		d := &font.Drawer{
			Dst:  rgba,
			Src:  image.NewUniform(color.RGBA{R: 0, G: 0, B: 0, A: 255}),
			Face: basicfont.Face7x13,
			Dot:  fixed.P(x0+2, y0+12),
		}
		d.DrawString(text)
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer out.Close()

	if err := png.Encode(out, rgba); err != nil {
		return fmt.Errorf("encode PNG: %w", err)
	}
	return nil
}
