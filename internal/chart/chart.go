package chart

import "time"

type ValueFormatter func(v interface{}) string

type Series struct {
	XValues []time.Time
	YValues []float64
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
	XAxis  XAxis
	YAxis  YAxis
	Series Series
}

const (
	DefaultChartHeight = 400
	DefaultChartWidth  = 1024
)
