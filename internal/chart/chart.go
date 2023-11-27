package chart

import "time"

type ValueFormatter func(v interface{}) string

type Series struct {
	XValues []time.Time
	YValues []float64
}

func (ts Series) Len() int {
	return len(ts.XValues)
}
func (ts Series) GetValues(index int) (x, y float64) {
	x = toFloat64(ts.XValues[index])
	y = ts.YValues[index]
	return
}

func (ts Series) GetLastValues() (x, y float64) {
	x = toFloat64(ts.XValues[len(ts.XValues)-1])
	y = ts.YValues[len(ts.YValues)-1]
	return
}

type XAxis struct {
	Name           string
	ValueFormatter ValueFormatter
}

type YAxis struct {
	Name           string
	ValueFormatter ValueFormatter
}

type Chart struct {
	XAxis string
	YAxis string

	Series Series

	Width  int
	Height int
}

const (
	DefaultChartHeight = 400
	DefaultChartWidth  = 1024
)
