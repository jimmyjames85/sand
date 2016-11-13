package graphite

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/andlabs/ui"
)

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

type RawData struct {
	TargetName string        `json:"target"`
	Start      time.Time     `json:"start"`
	End        time.Time     `json:"end"`
	Step       time.Duration `json:"step"`
	Data       []float64     `json:"data"`
	DataMin    float64
	DataMax    float64
}

func (rd *RawData) Image(r image.Rectangle) image.Image {

	m := image.NewRGBA(r)
	xFactor := float64(r.Dx()) / float64(rd.End.Unix()-rd.Start.Unix())
	yFactor := float64(r.Dy()) / float64(rd.DataMax-rd.DataMin)
	if (rd.DataMax - rd.DataMin) == 0 {
		yFactor = 1
	}

	cur := image.Point{X: 0, Y: 0}

	t := rd.Start
	for _, v := range rd.Data {
		///draw x,y

		next := image.Point{
			X: int(float64(int64(t.Sub(rd.Start))) * xFactor),
			Y: int((v - rd.DataMin) * yFactor),
		}
		drawLineP(m, cur, next, color.RGBA{255, 0, 0, 255})
		cur = next
		t = t.Add(rd.Step)
	}

	return m
}

func cat() {

	var err error
	var n int
	b := make([]byte, 0)

	for err == nil {
		n, err = os.Stdin.Read(b)
		if err == nil {
			fmt.Printf("%d bytes read: %s\n", n, b)
		}

	}
	if err != io.EOF {
		log.Fatal(err)
	}

}

type AH struct {
	RawData []RawData
}

// Draw is sent when a part of the Area needs to be drawn.
// dp will contain a drawing context to draw on, the rectangle
// that needs to be drawn in, and (for a non-scrolling area) the
// size of the area. The rectangle that needs to be drawn will
// have been cleared by the system prior to drawing, so you are
// always working on a clean slate.
//
// If you call Save on the drawing context, you must call Release
// before returning from Draw, and the number of calls to Save
// and Release must match. Failure to do so results in undefined
// behavior.
func (ah *AH) Draw(a *ui.Area, dp *ui.AreaDrawParams) {

	var b *ui.Brush
	s := &ui.StrokeParams{
		Cap:        ui.LineCap(3),
		Join:       ui.LineJoin(1),
		Thickness:  1.0,
		MiterLimit: 2.0,
		Dashes:     []float64{3, 2},
		DashPhase:  0.2,
	}

	if s != nil {
		s.DashPhase = 0.2

	}

	for i, r := range ah.RawData {
		path := ui.NewPath(ui.Winding)
		path.NewFigure(0, dp.AreaHeight)
		b = &ui.Brush{
			R: 1.0, //rand.Float64(),
			G: 0.0, //rand.Float64(),
			B: 0.0, //rand.Float64(),
			A: 1.0, //rand.Float64(),
		}

		fmt.Printf("%d, %f %f\n", i, r.DataMin, r.DataMax)
		xFactor := dp.AreaWidth / float64(r.End.Unix()-r.Start.Unix())

		yFactor := dp.AreaHeight / float64(r.DataMax-r.DataMin)
		if (r.DataMax - r.DataMin) == 0 {
			yFactor = 1
		}

		t := r.Start
		for _, v := range r.Data {
			///draw x,y
			x := float64(int64(t.Sub(r.Start))) * xFactor
			y := (v - r.DataMin) * yFactor

			path.LineTo(x, dp.AreaHeight-y)
			t = t.Add(r.Step)
		}
		path.End()

		dp.Context.Fill(path, b)
		//dp.Context.Stroke(path, b, s)
	}

	//path.AddRectangle(dp.ClipX, dp.ClipY, dp.AreaWidth, dp.AreaHeight)

}

// MouseEvent is called when the mouse moves over the Area
// or when a mouse button is pressed or released. See
// AreaMouseEvent for more details.
//
// If a mouse button is being held, MouseEvents will continue to
// be generated, even if the mouse is not within the area. On
// some systems, the system can interrupt this behavior;
// see DragBroken.
func (*AH) MouseEvent(a *ui.Area, me *ui.AreaMouseEvent) {

}

// MouseCrossed is called when the mouse either enters or
// leaves the Area. It is called even if the mouse buttons are being
// held (see MouseEvent above). If the mouse has entered the
// Area, left is false; if it has left the Area, left is true.
//
// If, when the Area is first shown, the mouse is already inside
// the Area, MouseCrossed will be called with left=false.
// TODO what about future shows?
func (*AH) MouseCrossed(a *ui.Area, left bool) {
	fmt.Printf("MoueC")
}

// DragBroken is called if a mouse drag is interrupted by the
// system. As noted above, when a mouse button is held,
// MouseEvent will continue to be called, even if the mouse is
// outside the Area. On some systems, this behavior can be
// stopped by the system itself for a variety of reasons. This
// method is provided to allow your program to cope with the
// loss of the mouse in this case. You should cope by cancelling
// whatever drag-related operation you were doing.
//
// Note that this is only generated on some systems under
// specific conditions. Do not implement behavior that only
// takes effect when DragBroken is called.
func (*AH) DragBroken(a *ui.Area) {
	fmt.Printf("Dragbroken")
}

// KeyEvent is called when a key is pressed while the Area has
// keyboard focus (if the Area has been tabbed into or if the
// mouse has been clicked on it). See AreaKeyEvent for specifics.
//
// Because some keyboard events are handled by the system
// (for instance, menu accelerators and global hotkeys), you
// must return whether you handled the key event; return true
// if you did or false if you did not. If you wish to ignore the
// keyboard outright, the correct implementation of KeyEvent is
// 	func (h *MyHandler) KeyEvent(a *ui.Area, ke *ui.AreaKeyEvent) (handled bool) {
// 		return false
// 	}
// DO NOT RETURN TRUE UNCONDITIONALLY FROM THIS
// METHOD. BAD THINGS WILL HAPPEN IF YOU DO.
func (*AH) KeyEvent(a *ui.Area, ke *ui.AreaKeyEvent) (handled bool) {
	return true
}

func gui() {

	rd, err := ParseRawData(bufio.NewReader(os.Stdin))
	if err != nil {
		log.Fatal(err)
	}
	areahandler := &AH{rd}

	err = ui.Main(func() {

		name := ui.NewEntry()
		button := ui.NewButton("Greet")
		greeting := ui.NewLabel("")
		box := ui.NewVerticalBox()
		box.Append(ui.NewLabel("Enter your name:"), false)
		box.Append(name, false)
		box.Append(button, false)
		box.Append(greeting, false)
		window := ui.NewWindow("Hello", 500, 800, false)
		window.SetChild(box)

		area := ui.NewArea(areahandler)
		box.Append(area, false)

		button.OnClicked(func(*ui.Button) {
			greeting.SetText("Hello, " + name.Text() + "!")
		})
		window.OnClosing(func(*ui.Window) bool {
			ui.Quit()
			return true
		})
		window.Show()
	})
	if err != nil {
		panic(err)
	}
}

func DrawRawData() {
	rd, err := ParseRawData(bufio.NewReader(os.Stdin))

	if err != nil {
		log.Fatal(err)
	}

	if len(rd) > 0 {

		err = png.Encode(os.Stdout, rd[0].Image(image.Rect(0, 0, 2048, 1024)))
		if err != nil {
			log.Fatal(err)
		}
	}
}
