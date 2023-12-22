package chart

type ValueFormatter func(v interface{}) string

type Chart struct {
	XAxis XAxis
	YAxis YAxis

	Series Series
	Styles string

	Width  int
	Height int
}
