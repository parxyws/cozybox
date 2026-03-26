package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core/entity"
	"github.com/johnfercher/maroto/v2/pkg/props"
	xdraw "golang.org/x/image/draw"
)

func createWatermarkImage(logoBytes []byte) ([]byte, error) {
	logoImg, err := png.Decode(bytes.NewReader(logoBytes))
	if err != nil {
		return nil, err
	}

	// Canvas dimensions (A4 proportioned at 72 DPI)
	canvasW, canvasH := 595, 842

	// Calculate proportional size fitting within a 250x150 bounding box
	bounds := logoImg.Bounds()
	ratio := float64(bounds.Dx()) / float64(bounds.Dy())
	fmt.Println(float64(bounds.Dx()))
	fmt.Println(float64(bounds.Dy()))
	fmt.Println(ratio)

	maxW, maxH := 250.0, 150.0
	targetRatio := maxW / maxH
	fmt.Println(targetRatio)

	var targetW, targetH int
	// If image is proportionally wider than the bounding box, constrain by width
	if ratio >= targetRatio {
		targetW = int(maxW)
		targetH = int(maxW / ratio)
	} else {
		// If image is proportionally taller than the bounding box, constrain by height
		targetH = int(maxH)
		targetW = int(maxH * ratio)
	}

	canvas := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))

	// 20% opacity mask
	opacity := uint8(51) // 20%
	// opacity := uint8(77) // 30%
	mask := image.NewUniform(color.Alpha{A: opacity})

	// Scale the logo
	scaledLogo := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	xdraw.BiLinear.Scale(scaledLogo, scaledLogo.Bounds(), logoImg, logoImg.Bounds(), xdraw.Src, nil)

	// Calculate center position
	offsetX := (canvasW - targetW) / 2
	offsetY := (canvasH - targetH) / 2
	dp := image.Pt(offsetX, offsetY)
	targetRect := image.Rectangle{Min: dp, Max: dp.Add(scaledLogo.Bounds().Size())}

	// Draw the scaled logo onto the canvas WITH the transparency mask
	draw.DrawMask(canvas, targetRect, scaledLogo, image.Point{}, mask, image.Point{}, draw.Over)

	// Encode to PNG buffer
	var buf bytes.Buffer
	err = png.Encode(&buf, canvas)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func BuildConfig() *entity.Config {
	logoBytes, err := os.ReadFile("../../../../cozybox.png")
	if err != nil {
		panic(err)
	}

	bgBytes, err := createWatermarkImage(logoBytes)
	if err != nil {
		// fallback to original if processing fails
		bgBytes = logoBytes
	}

	return config.NewBuilder().WithDebug(false).WithPageSize(pagesize.A4).WithOrientation(orientation.Vertical).WithDefaultFont(&props.Font{}).WithBackgroundImage(bgBytes, extension.Png).Build()
}
