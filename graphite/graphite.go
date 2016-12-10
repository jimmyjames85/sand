package graphite

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"io"
	"strconv"
	"strings"
	"time"

	"encoding/json"
	"log"
	"math"
	"math/rand"
)

var (
	RED    = color.RGBA{255, 0, 0, 255}
	GREEEN = color.RGBA{0, 255, 0, 255}
	BLUE   = color.RGBA{0, 0, 255, 255}
	BLACK  = color.RGBA{0, 0, 0, 255}

	BLACK50 = color.RGBA{50, 50, 50, 255}
)

type TimeValue struct {
	Time  time.Time
	Value float64
}

type TVRectangle struct {
	Min TimeValue
	Max TimeValue
}

func TVRect(t0 time.Time, v0 float64, t1 time.Time, v1 float64) TVRectangle {

	if t1.Before(t0) {
		t0, t1 = t1, t0
	}
	if v0 > v1 {
		v0, v1 = v1, v0
	}
	return TVRectangle{TimeValue{t0, v0}, TimeValue{t1, v1}}
}

type MetricData struct {
	Name       string
	Start      time.Time     `json:"start"`
	End        time.Time     `json:"end"`
	Step       time.Duration `json:"step"`
	MinValue   float64
	MaxValue   float64
	TimeValues []TimeValue
}

func (m *MetricData) TVRectangle() TVRectangle {
	return TVRect(m.Start, m.MinValue, m.End, m.MaxValue)
}

type tMapper func(time.Time) int
type vMapper func(float64) int

func hrFloor(t time.Time) time.Time {
	return t.Add(-time.Duration(t.Minute()) * time.Minute)
	return t
}

func setupImageLines(img *image.RGBA, tvr TVRectangle, fx tMapper, fy vMapper) {

	t0 := tvr.Min.Time
	t1 := tvr.Max.Time
	v0 := tvr.Min.Value
	v1 := tvr.Max.Value

	// bg
	FillRectangle(img, img.Bounds(), BLACK)

	// draw hour lines
	t := hrFloor(t0)
	tstep := time.Hour * time.Duration(int(hrFloor(t1).Sub(t).Hours()/6))
	for t.Before(t1) {
		x := fx(t)
		DrawText(img, x+5, fy(v0)-2, t.Format("01/02/2006 15:04"), RED)
		DrawLine(img, x, fy(v0), x, fy(v1), BLACK50)
		t = t.Add(tstep)
	}

	// draw val lines
	step := (v1 - v0) / 5

	v := tvr.Min.Value
	for v < v1 {
		//x := fx(t)
		y := fy(v)
		DrawLine(img, fx(t0), y, fx(t1), y, BLACK50)
		DrawText(img, fx(t0), y+5, fmt.Sprintf("%f", v), RED)
		v += step
	}
}
func genMappers(r image.Rectangle, tvr TVRectangle) (tMapper, vMapper, error) {

	fx, err := genTMapper(r, tvr)
	if err != nil {
		return nil, nil, err
	}

	fy, err := genVMapper(r, tvr)
	if err != nil {
		return nil, nil, err
	}

	return fx, fy, nil
}

func genTMapper(r image.Rectangle, tvr TVRectangle) (tMapper, error) {

	// x transformation
	mx, bx, err := slopeIntercept(float64(tvr.Min.Time.Unix()), float64(r.Min.X), float64(tvr.Max.Time.Unix()), float64(r.Max.X))
	if err != nil {
		return nil, err
	}

	return func(t time.Time) int {
		return int(mx*float64(t.Unix()) + bx)
	}, nil
}

func genVMapper(r image.Rectangle, tvr TVRectangle) (vMapper, error) {

	// y transformation
	my, by, err := slopeIntercept(tvr.Min.Value, float64(r.Min.Y), tvr.Max.Value, float64(r.Max.Y))
	if err != nil {
		return nil, err
	}

	return func(v float64) int {
		return r.Max.Y - int(my*v+by)
	}, nil
}

func (m *MetricData) Paint(img *image.RGBA, tvr TVRectangle, c color.RGBA) error {

	resolution := 1 //must be >0 and btw this is inaccurately named (paint every rth value)

	fx, fy, err := genMappers(img.Bounds(), tvr)
	if err != nil {
		return err
	}

	var p0 image.Point
	for i, tv := range m.TimeValues {
		p1 := image.Point{fx(tv.Time), fy(tv.Value)}

		if i > 0 {
			if i%resolution == 0 {
				DrawLineP(img, p0, p1, c)
				p0 = p1
			}
		} else {
			DrawText(img, p1.X, p1.Y, m.Name, RED)
			p0 = p1
		}

	}
	return nil
}

func PaintMetrics(img *image.RGBA, tvr TVRectangle, metrics []*MetricData) error {

	rand.Seed(time.Now().Unix())
	fx, fy, err := genMappers(img.Bounds(), tvr)
	if err != nil {
		return err
	}
	setupImageLines(img, tvr, fx, fy)

	for _, m := range metrics {
		err := m.Paint(img, tvr, randColor())
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func (m *MetricData) Image(r image.Rectangle, tvr TVRectangle, c color.RGBA) (*image.RGBA, error) {
	img := image.NewRGBA(r)

	err := m.Paint(img, tvr, c)
	return img, err
}

func CalculateBounds(metrics []*MetricData) TVRectangle {
	var ret TVRectangle

	if len(metrics) == 0 {
		return ret
	}

	for i, m := range metrics {
		if i == 0 {
			ret = metrics[0].TVRectangle()
			continue
		}

		ret.Min.Value = math.Min(ret.Min.Value, m.MinValue)
		ret.Max.Value = math.Max(ret.Max.Value, m.MaxValue)

		if m.Start.Before(ret.Min.Time) {
			ret.Min.Time = m.Start
		}

		if ret.Max.Time.Before(m.End) {
			ret.Max.Time = m.End
		}
	}
	return ret
}

func ParseMetricsRAW(r *bufio.Reader) ([]*MetricData, error) {

	ret := make([]*MetricData, 0)

	line, err := r.ReadString('\n')
	for err == nil {
		m, perr := parseMetricRAW(line)
		if perr != nil {
			log.Printf("%s\n", perr)
		} else {
			ret = append(ret, m)
		}
		line, err = r.ReadString('\n')
	}

	if err != io.EOF {
		return ret, err
	}
	return ret, nil
}

func ParseMetricsJSON(data []byte) ([]*MetricData, error) {

	ret := make([]*MetricData, 0)

	parsed := make ( []GMetricDataJSON, 0)
	err := json.Unmarshal(data, &parsed)
	if err != nil {
		return ret, err
	}

	for _, gm := range parsed {
		m := &MetricData{
			TimeValues: make([]TimeValue, 0),
		}
		m.Name = gm.Target
		m.Start = time.Now()
		m.End = time.Unix(0, 0)

		for _, tv := range gm.Datapoints {
			if len(tv) != 2 || tv[0] == nil || tv[1] == nil {
				continue
			}
			v := *tv[0]
			t := time.Unix(int64(*tv[1]), 0)

			if t.Before(m.Start) {
				m.Start = t
			}
			if t.After(m.End) {
				m.End = t
			}

			if len(tv) == 0 {
				m.MinValue = v
				m.MaxValue = v
			}

			if m.MinValue > v {
				m.MinValue = v
			}
			if m.MaxValue < v {
				m.MaxValue = v
			}

			m.TimeValues = append(m.TimeValues, TimeValue{
				Time:  t,
				Value: v,
			})
		}
		ret = append(ret, m)
	}
	return ret, nil
}

type GMetricDataJSON struct {
	Datapoints [][]*float64 `json:"datapoints"`
	Target     string       `json:"target"`
}

func parseMetricRAW(data string) (*MetricData, error) {

	data = strings.Trim(data, " \n\t")

	ret := &MetricData{
		TimeValues: make([]TimeValue, 0),
	}

	fields := strings.Split(data, "|")
	if len(fields) != 2 {
		return nil, fmt.Errorf("%d: incorrect format: %s", len(fields), data)
	}
	dataPoints := strings.Split(fields[1], ",")
	fields = strings.Split(fields[0], ",")

	if len(fields) != 4 {
		return nil, fmt.Errorf("%d: unexpected format %v", len(fields), fields)
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
