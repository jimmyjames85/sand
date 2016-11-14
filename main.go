package main

import (
	"log"
	"os"
	"bufio"
	"image"
	"image/png"

	"github.com/jimmyjames85/sand/graphite"
)

func main() {
	metrics, err := graphite.ParseMetrics(bufio.NewReader(os.Stdin))
	if err != nil {
		log.Fatal(err)
	} else if len(metrics) == 0 {
		log.Fatalf("no metrics supplied\n")
	}

	r := image.Rect(0, 0, 2048, 1024)
	img := image.NewRGBA(r)
	tvr := graphite.CalculateBounds(metrics)

	tvr.Min.Value -= 750
	tvr.Max.Value += 750

	graphite.PaintMetrics(img,tvr,metrics)


	err = png.Encode(os.Stdout, img)
	if err != nil {
		log.Fatal(err)
	}

}
