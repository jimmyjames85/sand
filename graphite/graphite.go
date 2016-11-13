package graphite

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
	"image"
	"image/color"
	"math/rand"
)

type fPoint struct {
	X float64
	Y float64
}

type metric struct {
	TargetName string        `json:"target"`
	Start      time.Time     `json:"start"`
	End        time.Time     `json:"end"`

	
	Data       []fPoint     `json:"data"`

}

type RawData struct {
	TargetName string        `json:"target"`
	Start      time.Time     `json:"start"`
	End        time.Time     `json:"end"`
	Step       time.Duration `json:"step"`
	Data       []float64     `json:"data"`
	DataMin    float64
	DataMax    float64
}

func randColor() color.RGBA {
	return color.RGBA{
		uint8(rand.Uint32() % 256),
		uint8(rand.Uint32() % 256),
		uint8(rand.Uint32() % 256),
		255,
	}
}

func drawLine(m *image.RGBA, min, max image.Point, c color.Color) {

	if max.X < min.X {
		min, max = max, min
	}

	if max.X == min.X {
		//vertical line
		if min.Y <= max.Y {
			for y := min.Y; y <= max.Y; y++ {
				m.Set(min.X, y, c)
			}
		} else {
			for y := max.Y; y <= min.Y; y++ {
				m.Set(min.X, y, c)
			}
		}
		return
	}

	slope := float64(max.Y - min.Y) / float64(max.X - min.X)
	b := float64(min.Y) - slope * float64(min.X)

	for x := min.X; x <= max.X; x++ {
		y := slope * float64(x) + b
		m.Set(x, int(y), c)
	}

	if min.Y <= max.Y {
		for y := min.Y; y <= max.Y; y++ {
			x := (float64(y) - b) / slope
			m.Set(int(x), y, c)
		}
	} else {
		for y := max.Y; y <= min.Y; y++ {
			x := (float64(y) - b) / slope
			m.Set(int(x), y, c)
		}
	}

}

func (rd *RawData) Image(r image.Rectangle) image.Image {

	m := image.NewRGBA(r)
	xFactor := float64(r.Dx()) / float64(rd.End.Unix() - rd.Start.Unix())
	yFactor := float64(r.Dy()) / float64(rd.DataMax - rd.DataMin)
	if (rd.DataMax - rd.DataMin) == 0 {
		yFactor = 1
	}

	cur := image.Point{X:0, Y:0}

	t := rd.Start
	for _, v := range rd.Data {
		///draw x,y

		next := image.Point{
			X: int(float64(int64(t.Sub(rd.Start))) * xFactor),
			Y:int((v - rd.DataMin) * yFactor),
		}
		drawLine(m, cur, next, color.RGBA{255, 0, 0, 255})
		cur = next
		t = t.Add(rd.Step)
	}

	return m
}

func (rd *RawData) ToJSON() string {
	b, err := json.Marshal(rd)
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
}

func (rd *RawData) String() string {
	return rd.ToJSON()
}

func ParseRawData(r *bufio.Reader) ([]RawData, error) {

	ret := make([]RawData, 0)

	line, err := r.ReadString('\n')
	for err == nil {

		rd, perr := parseSingleRawData(line)
		if perr != nil {
			return ret, perr
		}

		ret = append(ret, *rd)

		line, err = r.ReadString('\n')
	}

	if err != io.EOF {
		return ret, err
	}
	return ret, nil

}

// parseSingleRawData parses one line of data which represents a single target
//
// @see graphite.readthedocs.io/en/latest/render_api.html
func parseSingleRawData(data string) (*RawData, error) {

	data = strings.Trim(data, " \n\t")
	rd := &RawData{
		Data: make([]float64, 0),
	}

	fields := strings.Split(data, "|")
	if len(fields) != 2 {
		return nil, fmt.Errorf("incorrect format")
	}
	dataPoints := strings.Split(fields[1], ",")
	fields = strings.Split(fields[0], ",")

	for i, p := range dataPoints {

		if strings.Index(p, "None") == 0 {
			rd.Data = append(rd.Data, 0.0) //TODO
			continue
		}

		fp, err := strconv.ParseFloat(p, 64)
		if err != nil {
			return nil, err
		}

		if i == 0 {
			rd.DataMin = fp
			rd.DataMax = fp
		}

		if rd.DataMin > fp {
			rd.DataMin = fp
		}
		if rd.DataMax < fp {
			rd.DataMax = fp
		}
		rd.Data = append(rd.Data, fp)
	}

	for i, f := range fields {
		switch i {
		case 0:
			//todo error if not all fields get populated
			rd.TargetName = f
		case 1:
			timestamp, err := strconv.ParseInt(f, 10, 64)
			if err != nil {
				return nil, err
			}
			rd.Start = time.Unix(timestamp, 0)
		case 2:
			timestamp, err := strconv.ParseInt(f, 10, 64)
			if err != nil {
				return nil, err
			}
			rd.End = time.Unix(timestamp, 0)
		case 3:
			timestep, err := strconv.ParseInt(f, 10, 64)
			if err != nil {
				return nil, err
			}
			rd.Step = time.Duration(timestep)
		}
	}

	return rd, nil
}
