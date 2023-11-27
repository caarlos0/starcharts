package chart

type ValueFormatter func(v interface{}) string

type Chart struct {
	XAxis XAxis
	YAxis YAxis

	Series Series

	Width  int
	Height int
}

const (
	DefaultChartHeight = 400
	DefaultChartWidth  = 1024
)
