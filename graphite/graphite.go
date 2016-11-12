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
)

type RawData struct {
	TargetName string        `json:"target"`
	Start      time.Time     `json:"start"`
	End        time.Time     `json:"end"`
	Step       time.Duration `json:"step"`
	Data       []float64     `json:"data"`
	DataMin    float64
	DataMax    float64
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

		rd, perr := ParseRawDataLine(line)
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
func ParseRawDataLine(data string) (*RawData, error) {

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
