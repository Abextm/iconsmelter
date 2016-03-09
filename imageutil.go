package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
)

func LoadAndCrop(finame string) *image.RGBA {
	fi, err := os.Open(finame)
	if err != nil {
		fmt.Println(err)
	}
	defer fi.Close()

	img, err := png.Decode(fi)
	if err != nil {
		fmt.Println(finame, err)
	}

	return crop(img, false)
}

func crop(img image.Image, skipTr bool) *image.RGBA {
	xs := 100
	xe := 0
	ys := 100
	ye := 0
	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				xs = min(xs, x)
				xe = max(xe, x)
				ys = min(ys, y)
				ye = max(ye, y)
			}
		}
	}
	rect := image.Rectangle{
		Min: image.Point{X: xs, Y: ys},
		Max: image.Point{X: xe + 1, Y: ye + 1},
	}
	if skipTr {
		rect.Min = image.ZP
	}
	nimg := image.NewRGBA(image.Rectangle{
		Min: image.ZP,
		Max: image.Point{X: rect.Dx(), Y: rect.Dy()},
	})
	draw.Draw(nimg, nimg.Bounds(), img, rect.Min, draw.Src)
	return nimg
}

func HasBG(img *image.RGBA, bgImg *image.RGBA) bool {
	if img.Rect.Dx() != bgImg.Rect.Dx() || img.Rect.Dy() != bgImg.Rect.Dy() {
		return false
	}
	samePix := 0
	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			ar, ag, ab, aa := img.At(x, y).RGBA()
			br, bg, bb, ba := bgImg.At(x, y).RGBA()
			if aa != ba {
				return false
			}
			if ar == br && ag == bg && ab == bb {
				samePix++
			}
		}
	}
	return samePix > 2
}
