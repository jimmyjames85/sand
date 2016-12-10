package main

import (
	"bufio"
	"image"
	"image/png"
	"io"
	"log"
	"os"

	"encoding/json"
	"fmt"
	"time"

	"github.com/jimmyjames85/sand/graphite"
)

//func traverseNode(node string, dest map[string]string) error {
//	//todo what is this for?
//	parms := make(map[string]string)
//	parms["query"] = fmt.Sprintf("%s.*", node)
//
//	ret, err := graphite.CURL(GRAPHITE_URL+"/metrics/expand", "GET", parms, nil)
//	if err != nil {
//		return err
//	}
//
//	dest[parms["query"]] = ret
//	return nil
//}

func myContinueWalk(m *graphite.APIMetric) bool {
	return true
}

func traverse(graphiteUrl, search string) {
	g := graphite.NewGraphite(graphiteUrl)
	metrics := make(chan *graphite.APIMetric)
	done := make(chan struct{})
	var count int
	go g.WalkMetrics(search, myContinueWalk, metrics, done)
METRICCONSUMER:
	for {
		select {
		case m := <-metrics:
			fmt.Printf("%s\n", toJSON(m))
			count++
		case <-done:
			break METRICCONSUMER
		}

	}
	log.Printf("Total: %d\n", count)
}

func usage() string {
	return fmt.Sprintf(` %s
	-t 'metric.*.to.traverse'
	-g 'metric.*.to.graph' -3d
`, os.Args[0])
}

func pargs() {
	//appName := os.Args[0]
	args := os.Args[1:]

	for len(args) > 0 {
		cmd := args[0]
		args = args[1:]

		switch cmd {
		case "walk":
			if len(args) == 0 {
				log.Fatal(usage())
			}
			cmd = args[0]
			args = args[1:]

		}

	}

}
func toJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return string(b)
}

func graphMetricsRAWperLine(graphiteUrl, query string, reader io.Reader) {

	metrics, err := graphite.ParseMetricsRAW(bufio.NewReader(reader))
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

	for _, m := range metrics {
		if size := len(m.TimeValues); size == 0 {
			log.Printf("%s has no data\n", m.Name)
		} else {
			log.Printf("%s has %d value(s)\n", m.Name, size)
		}
	}
	graphite.PaintMetrics(img, tvr, metrics)

	err = png.Encode(os.Stdout, img)
	if err != nil {
		log.Fatal(err)
	}
}

func graphMetrics(metrics []*graphite.MetricData) {

	r := image.Rect(0, 0, 2048, 1024)
	img := image.NewRGBA(r)
	tvr := graphite.CalculateBounds(metrics)

	tvr.Min.Value -= 750
	tvr.Max.Value += 750

	for _, m := range metrics {
		if size := len(m.TimeValues); size == 0 {
			log.Printf("%s has no data\n", m.Name)
		} else {
			log.Printf("%s has %d value(s)\n", m.Name, size)
		}
	}
	graphite.PaintMetrics(img, tvr, metrics)

	err := png.Encode(os.Stdout, img)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	start := time.Now()
	defer func() { log.Printf("runtime = %s \n", time.Now().Sub(start)) }()

	//traverse()
	//return
	graphiteUrl := os.Getenv("GRAPHITE_URL")
	g := graphite.NewGraphite(graphiteUrl)

	// eg  main.go -t metric.path.to.traverse
	if len(os.Args) > 2 {

		cmd := os.Args[1]
		query := os.Args[2]

		if cmd == "-t" {
			traverse(graphiteUrl, query)
			return
		} else if cmd == "-e" {

			nodes, err := g.APIExpand(query)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s\n", toJSON(nodes))
			return
		} else if cmd == "-g" {
			from := ""  // 24hrs ago
			until := "" // NOW
			if len(os.Args) > 3 {
				from = os.Args[3]
			}
			if len(os.Args) > 4 {
				until = os.Args[4]
			}

			ic := graphite.LoadMetric(graphiteUrl, query, "json", from, until)
			metrics, err := graphite.ParseMetricsJSON([]byte(ic))
			if err != nil {
				log.Fatal(err)
			}
			graphMetrics(metrics)

			var subTot float64
			var metricCount int64

			for _, m:= range metrics{
				for _, tv := range m.TimeValues {
					subTot += tv.Value
					metricCount++
				}

				//if len(m.TimeValues)

			}


			if metricCount!=0{
				log.Printf("TOTAL AVERAGE: %f", subTot/float64(metricCount))
			}


		}
	}
}
