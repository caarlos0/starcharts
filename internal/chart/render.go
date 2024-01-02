package chart

import (
	"io"
	"math"

	"github.com/caarlos0/starcharts/internal/chart/svg"
)

func (c *Chart) Render(w io.Writer) {
	canvas := c.Box()

	xRange, yRange := c.getRanges(canvas)

	xTicks := generateTicks(xRange, false, timeValueFormatter)
	yTicks := generateTicks(yRange, true, intValueFormatter)

	axesOuterBox := canvas.Clone().
		Grow(c.XAxis.Measure(canvas, xRange, xTicks)).
		Grow(c.YAxis.Measure(canvas, yRange, yTicks))

	plot := canvas.OuterConstrain(c.Box(), axesOuterBox)

	xRange.Domain = plot.Width()
	yRange.Domain = plot.Height()

	background := svg.Rect().
		Attr("x", svg.Point(0)).
		Attr("y", svg.Point(0)).
		Attr("width", svg.Px(c.Width)).
		Attr("height", svg.Px(c.Height)).
		Attr("class", "background").
		Attr("style", styles("fill", c.Background)).
		Attr("rx", "8")

	cssStyles := c.Styles
	if cssStyles == "" {
		cssStyles = LightStyles
	}

	style := svg.Style().
		Attr("type", "text/css").
		Content(cssStyles)

	svgElement := svg.SVG().
		Attr("width", svg.Px(c.Width)).
		Attr("height", svg.Px(c.Height)).
		ContentFunc(func(w io.Writer) {
			style.Render(w)
			background.Render(w)
			c.Series.Render(w, plot, xRange, yRange)
			c.YAxis.Render(w, plot, yRange, yTicks)
			c.XAxis.Render(w, plot, xRange, xTicks)
		})

	svgElement.Render(w)
}

func (c *Chart) getRanges(canvas *Box) (*Range, *Range) {
	minX, maxX := math.MaxFloat64, -math.MaxFloat64
	minY, maxY := math.MaxFloat64, -math.MaxFloat64

	seriesLength := c.Series.Len()
	for index := 0; index < seriesLength; index++ {
		vX, vY := c.Series.GetValues(index)

		minX = min(minX, vX)
		maxX = max(maxX, vX)

		minY = min(minY, vY)
		maxY = max(maxY, vY)
	}

	delta := maxY - minY
	roundTo := getRoundToForDelta(delta)

	yRange := &Range{
		Min:    roundDown(minY, roundTo),
		Max:    roundUp(maxY, roundTo),
		Domain: canvas.Height(),
	}

	xRange := &Range{
		Min:    minX,
		Max:    maxX,
		Domain: canvas.Width(),
	}

	return xRange, yRange
}

func (c *Chart) Box() *Box {
	return &Box{
		Top:    BoxPadding.Top,
		Left:   BoxPadding.Left,
		Right:  c.Width - BoxPadding.Right,
		Bottom: c.Height - BoxPadding.Bottom,
	}
}
