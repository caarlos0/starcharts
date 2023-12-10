package chart

var BoxPadding = Box{
	Top:    5,
	Left:   5,
	Right:  5,
	Bottom: 5,
}

const (
	DPI = 92.0

	DefaultTickCountSanityCheck = 1024

	AxisFontSize = 10.0

	MinimumTickHorizontalSpacing = 20
	MinimumTickVerticalSpacing   = 20

	YAxisMargin = 10
	XAxisMargin = 10

	VerticalTickHeight  = XAxisMargin >> 1
	HorizontalTickWidth = YAxisMargin >> 1

	MinStrokeWidth = 1.0
)
