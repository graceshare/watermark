package main

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/golang/freetype/truetype"

	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/vgimg"
)

type WaterMarker struct {
	Text              string
	HorizontalSpacing float64
	VerticalSpacing   float64
	FontSize          float64
	OutputDPI         int
	FontName          string
	FontPath          string
	Color             color.Color
	Resize            Size
}

type Size struct {
	Width  int
	Height int
}

func l(v float64) vg.Length {
	return vg.Length(v)
}

func (w *WaterMarker) Mark(inputFilename, outputFilename string) error {
	input, err := os.Open(inputFilename)
	if err != nil {
		return err
	}

	img, _, err := image.Decode(input)
	if err != nil {
		return err
	}

	input.Close()

	output, err := os.Create(outputFilename)
	if err != nil {
		return err
	}

	defer output.Close()

	ext := filepath.Ext(outputFilename)

	return w.mark(img, ext, output)
}

func (w *WaterMarker) mark(img image.Image, format string, out io.Writer) error {
	bounds := img.Bounds()

	width := l(float64(bounds.Max.X))
	height := l(float64(bounds.Max.Y))

	c := vgimg.NewWith(
		vgimg.UseWH(width, height),
		vgimg.UseDPI(w.OutputDPI),
	)

	c.DrawImage(vg.Rectangle{Max: vg.Point{X: width, Y: height}}, img)

	c.SetColor(w.Color)

	fontName := "HiraginoSansGB"
	ttfBytes, err := ioutil.ReadFile(w.FontPath)
	if err != nil {
		log.Println("读取字体文件失败", err.Error())
	}
	font, _ := truetype.Parse(ttfBytes)
	vg.AddFont(fontName, font)

	// fontStyle, err := vg.MakeFont(w.FontName, l(w.FontSize))
	fontStyle, err := vg.MakeFont(fontName, l(w.FontSize))
	if err != nil {
		log.Println(err)
		return err
	}

	textWidth := fontStyle.Width(w.Text)
	textHeight := w.FontSize

	textBoxWidth := float64(textWidth) + w.HorizontalSpacing*2
	textBoxHeight := textHeight + w.VerticalSpacing*2

	c.Rotate(-math.Pi / 4)

	xOffsetMin := -float64(height) / math.Sqrt(2)
	yOffsetMax := float64(width) * math.Sqrt(2)

	for xOffset := xOffsetMin; xOffset < float64(width); xOffset += textBoxWidth {
		for yOffset := 0.0; yOffset < yOffsetMax; yOffset += textBoxHeight {
			c.FillString(fontStyle, vg.Point{
				X: l(xOffset),
				Y: l(yOffset),
			}, w.Text)
		}
	}

	w.resize(&c)

	return writeCanvas(c, format, out)
}

func (w *WaterMarker) resize(c **vgimg.Canvas) {
	width := w.Resize.Width
	height := w.Resize.Height
	if width == 0 && height == 0 {
		return
	}

	dstImg := imaging.Resize((*c).Image(), width, height, imaging.Lanczos)

	width = dstImg.Rect.Max.X
	height = dstImg.Rect.Max.Y

	lWidth := l(float64(width))
	lHeight := l(float64(height))

	*c = vgimg.NewWith(
		vgimg.UseWH(lWidth, lHeight),
		vgimg.UseDPI(w.OutputDPI),
	)

	(*c).DrawImage(vg.Rectangle{Max: vg.Point{X: lWidth, Y: lHeight}}, dstImg)
}

func writeCanvas(c *vgimg.Canvas, format string, out io.Writer) error {
	switch format {
	case ".jpeg", ".jpg":
		_, err := vgimg.JpegCanvas{Canvas: c}.WriteTo(out)
		return err
	case ".png":
		_, err := vgimg.PngCanvas{Canvas: c}.WriteTo(out)
		return err
	default:
		return fmt.Errorf("unsupported file format")
	}
}
