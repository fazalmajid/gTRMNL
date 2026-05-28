package main

import (
	"image"
	"image/color"
	"image/draw"
)

// Spectra6 primaries as used by the Seeed ReTerminal E1002 firmware.
// The firmware does Euclidean RGB nearest-neighbour matching against these
// six colours, then maps each to its panel waveform code:
//   Black(0,0,0)→0x000f  White(255,255,255)→0x0000  Red(255,0,0)→0x0006
//   Green(0,255,0)→0x0002  Blue(0,0,255)→0x000d  Yellow(255,255,0)→0x000b
// Six entries produce a 4-bpp indexed PNG, which takes the correct case-4
// branch in ReduceBpp (the case-2 branch has a pointer-advance bug).
var spectra6 = color.Palette{
	color.RGBA{R: 0, G: 0, B: 0, A: 255},
	color.RGBA{R: 255, G: 255, B: 255, A: 255},
	color.RGBA{R: 255, G: 0, B: 0, A: 255},
	color.RGBA{R: 0, G: 255, B: 0, A: 255},
	color.RGBA{R: 0, G: 0, B: 255, A: 255},
	color.RGBA{R: 255, G: 255, B: 0, A: 255},
}

func ditherToSpectra6(src image.Image) *image.Paletted {
	dst := image.NewPaletted(src.Bounds(), spectra6)
	draw.FloydSteinberg.Draw(dst, dst.Bounds(), src, image.Point{})
	return dst
}
