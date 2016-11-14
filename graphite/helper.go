package graphite

import (
	"image"
	"image/color"
	"fmt"
	"math/rand"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func randColor() color.RGBA {
	return color.RGBA{
		uint8(rand.Uint32() % 256),
		uint8(rand.Uint32() % 256),
		uint8(rand.Uint32() % 256),
		255,
	}
}

func FillRectangle(img *image.RGBA, r image.Rectangle, c color.Color) {
	for x := r.Min.X; x < r.Max.X; x++ {
		for y := r.Min.Y; y < r.Max.Y; y++ {
			img.Set(x, y, c)
		}
	}
}


func DrawText(img *image.RGBA, x, y int, label string, c color.Color) {

	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}

func DrawRectangle(img *image.RGBA, r image.Rectangle, c color.Color) {
	x0 := r.Min.X
	y0 := r.Min.Y
	x1 := r.Max.X
	y1 := r.Max.Y

	DrawLine(img, x0, y0, x0, y1, c)
	DrawLine(img, x0, y1, x1, y1, c)
	DrawLine(img, x1, y1, x1, y0, c)
	DrawLine(img, x1, y0, x0, y0, c)
}

func DrawLineP(img *image.RGBA, min, max image.Point, c color.Color) {
	DrawLine(img, min.X, min.Y, max.X, max.Y, c)
}

func DrawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {

	if x1 < x0 {
		//x0 should be first
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	} else if x1 == x0 {
		//vertical line
		if y0 <= y1 {
			for y := y0; y <= y1; y++ {
				img.Set(x0, y, c)
			}
		} else {
			for y := y1; y <= y0; y++ {
				img.Set(x0, y, c)
			}
		}
		return
	}

	// ignoring err case becase we take care of vertical line above
	m, b, _ := slopeIntercept(float64(x0), float64(y0), float64(x1), float64(y1))
	//slope := float64(y1 - y0) / float64(x1 - x0)
	//b := float64(y0) - slope * float64(x0)

	for x := x0; x <= x1; x++ {
		y := m * float64(x) + b
		img.Set(x, int(y), c)
	}

	if y0 <= y1 {
		for y := y0; y <= y1; y++ {
			x := (float64(y) - b) / m //todo what if m=0
			img.Set(int(x), y, c)
		}
	} else {
		for y := y1; y <= y0; y++ {
			x := (float64(y) - b) / m //todo what if m=0
			img.Set(int(x), y, c)
		}
	}
}

func slopeIntercept(x0, y0, x1, y1 float64) (float64, float64, error) {
	if x0 == x1 {
		return 0, 0, fmt.Errorf("divide by zero")
	}

	m := (y1 - y0) / (x1 - x0)
	b := y0 - m * x0
	return m, b, nil
}