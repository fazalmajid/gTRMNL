package main

import (
	"image"
	"image/color"
	"image/draw"
)

var spectra6 = color.Palette{
	color.RGBA{R: 0, G: 0, B: 0, A: 255},       // Black
	color.RGBA{R: 255, G: 255, B: 255, A: 255}, // White
	color.RGBA{R: 255, G: 255, B: 0, A: 255},   // Yellow
	color.RGBA{R: 255, G: 0, B: 0, A: 255},     // Red
	color.RGBA{R: 0, G: 0, B: 255, A: 255},     // Blue
	color.RGBA{R: 0, G: 255, B: 0, A: 255},     // Green
}

func ditherToSpectra6(src image.Image) *image.Paletted {
	dst := image.NewPaletted(src.Bounds(), spectra6)
	draw.FloydSteinberg.Draw(dst, dst.Bounds(), src, image.Point{})
	return dst
}
