package graphite

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var (
	RED   = color.RGBA{255, 0, 0, 255}
	BLACK = color.RGBA{0, 0, 0, 255}
)

type TimeValue struct {
	Time  time.Time
	Value float64
}

type Metric struct {
	Name       string
	Start      time.Time     `json:"start"`
	End        time.Time     `json:"end"`
	Step       time.Duration `json:"step"`
	MinValue   float64
	MaxValue   float64
	TimeValues []TimeValue
}

func randColor() color.RGBA {
	return color.RGBA{
		uint8(rand.Uint32() % 256),
		uint8(rand.Uint32() % 256),
		uint8(rand.Uint32() % 256),
		255,
	}
}
func DrawRectangle(img *image.RGBA, r image.Rectangle, c color.Color) {
	x0 := r.Min.X
	y0 := r.Min.Y
	x1 := r.Max.X
	y1 := r.Max.Y

	log.Printf("%d %d %d %d\n", x0, y0, x0, y1)
	log.Printf("%d %d %d %d\n", x0, y1, x1, y1)
	log.Printf("%d %d %d %d\n", x1, y1, x1, y0)
	log.Printf("%d %d %d %d\n", x1, y0, x0, y0)

	drawLine(img, x0, y0, x0, y1, c)
	drawLine(img, x0, y1, x1, y1, c)
	drawLine(img, x1, y1, x1, y0, c)
	drawLine(img, x1, y0, x0, y0, c)
}

func drawLineP(img *image.RGBA, min, max image.Point, c color.Color) {
	drawLine(img, min.X, min.Y, max.X, max.Y, c)
}

func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {

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

	slope := float64(y1-y0) / float64(x1-x0)
	b := float64(y0) - slope*float64(x0)

	for x := x0; x <= x1; x++ {
		y := slope*float64(x) + b
		img.Set(x, int(y), c)
	}

	if y0 <= y1 {
		for y := y0; y <= y1; y++ {
			x := (float64(y) - b) / slope
			img.Set(int(x), y, c)
		}
	} else {
		for y := y1; y <= y0; y++ {
			x := (float64(y) - b) / slope
			img.Set(int(x), y, c)
		}
	}
}

func (m *Metric) Image(r image.Rectangle, t0 time.Time, v0 float64, t1 time.Time, v1 float64) *image.RGBA {

	if t1.Before(t0) {
		log.Fatal("t1 < t0")
	}

	img := image.NewRGBA(r)
	rr := r
	rr.Max.X = rr.Max.X - 1
	rr.Max.Y = rr.Max.Y - 1
	DrawRectangle(img, rr, BLACK)


	m.Start.Sub(t0)

	//xFactor := float64(r.Dx()) / float64(m.End.Unix()- m.Start.Unix())
	xFactor := float64(r.Dx()) / float64(t1.Sub(t0))
	yFactor := float64(r.Dy()) / float64(m.MaxValue-m.MinValue)
	if (m.MaxValue - m.MinValue) == 0 {
		yFactor = 1
	}

	var cur image.Point

	tOffset := m.Start.Sub(t0)

	for i, tv := range m.TimeValues {

		//x := xFactor * float64(tOffset+tv.Time.Sub(m.Start))
		x := xFactor * float64(tOffset+tv.Time.Sub(m.Start))
		next := image.Point{
			X: int(x),
			//X: int(float64(int64(tv.Time.Sub(m.Start) -tOffset)) * xFactor),
			Y: int((tv.Value - m.MinValue) * yFactor),
			//X: int(float64(int64(t.Sub(rd.Start))) * xFactor),
			//Y: int((v.Value - rd.MinValue) * yFactor),
		}

		if i > 0 {
			///draw x,y
			drawLineP(img, cur, next, RED)
		}
		cur = next

	}


	return img
}

func ParseMetrics(r *bufio.Reader) ([]*Metric, error) {

	ret := make([]*Metric, 0)

	line, err := r.ReadString('\n')
	for err == nil {
		m, perr := parseMetric(line)
		if perr != nil {
			return ret, perr
		}
		ret = append(ret, m)
		line, err = r.ReadString('\n')
	}

	if err != io.EOF {
		return ret, err
	}
	return ret, nil
}

func parseMetric(data string) (*Metric, error) {

	data = strings.Trim(data, " \n\t")

	ret := &Metric{
		TimeValues: make([]TimeValue, 0),
	}

	fields := strings.Split(data, "|")
	if len(fields) != 2 {
		return nil, fmt.Errorf("incorrect format")
	}
	dataPoints := strings.Split(fields[1], ",")
	fields = strings.Split(fields[0], ",")

	if len(fields) != 4 {
		return nil, fmt.Errorf("unexpected format")
	}

	ret.Name = fields[0]

	timestamp, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return nil, err
	}
	ret.Start = time.Unix(timestamp, 0)

	timestamp, err = strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return nil, err
	}
	ret.End = time.Unix(timestamp, 0)

	timestep, err := strconv.ParseInt(fields[3], 10, 64)
	if err != nil {
		return nil, err
	}
	ret.Step = time.Second * time.Duration(timestep)

	t := ret.Start
	for i, p := range dataPoints {
		if strings.Index(p, "None") < 0 {
			fp, err := strconv.ParseFloat(p, 64)
			if err != nil {
				return nil, err
			}

			ret.TimeValues = append(ret.TimeValues, TimeValue{
				Time:  t,
				Value: fp,
			})

			if i == 0 {
				ret.MinValue = fp
				ret.MaxValue = fp
			}

			if ret.MinValue > fp {
				ret.MinValue = fp
			}
			if ret.MaxValue < fp {
				ret.MaxValue = fp
			}
		}
		t = t.Add(ret.Step)
	}

	return ret, nil
}
