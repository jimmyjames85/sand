package main

import (
	"fmt"
	"log"
	"os"

	"bufio"

	"image"
	"image/png"
	"time"

	"github.com/jimmyjames85/sand/graphite"
)

func main() {
	metrics, err := graphite.ParseMetrics(bufio.NewReader(os.Stdin))

	if err != nil {
		log.Fatal(err)
	}

	if len(metrics) > 0 {
		m := metrics[0]
		img:=m.Image(image.Rect(0, 0, 2048, 512), m.Start.Add(-7660*time.Second), m.MinValue, m.End, m.MaxValue)
		//graphite.DrawRectangle(img, image.Rect(0,0,50,50),graphite.BLACK)
		err = png.Encode(os.Stdout, img)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("nope!\n")
	}

}
