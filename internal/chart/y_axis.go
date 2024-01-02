package chart

import (
	"io"
	"math"

	"github.com/caarlos0/starcharts/internal/chart/svg"
)

type YAxis struct {
	Name        string
	StrokeWidth float64
	Color       string
}

func (ya *YAxis) Measure(canvas *Box, ra *Range, ticks []Tick) *Box {
	tx := canvas.Right + YAxisMargin

	minX, maxX, minY, maxY := math.MaxInt32, 0, math.MaxInt32, 0
	maxTextHeight := 0
	for _, t := range ticks {
		ly := canvas.Bottom - ra.Translate(t.Value)

		tb := measureText(t.Label, AxisFontSize)
		maxTextHeight = max(tb.Height(), maxTextHeight)

		minX = canvas.Right
		maxX = max(maxX, tx+tb.Width())

		tbh2 := tb.Height() >> 1
		minY = min(minY, ly-tbh2)
		maxY = max(maxY, ly+tbh2)
	}

	maxX += YAxisMargin + maxTextHeight

	return &Box{
		Top:    minY,
		Left:   minX,
		Right:  maxX,
		Bottom: maxY,
	}
}

func (ya *YAxis) Render(w io.Writer, canvasBox *Box, ra *Range, ticks []Tick) {
	lx := canvasBox.Right
	tx := lx + YAxisMargin
	strokeStyle := styles("stroke", ya.Color)
	fillStyle := styles("fill", ya.Color)

	strokeWidth := normaliseStrokeWidth(ya.StrokeWidth)

	svg.Path().
		Attr("stroke-width", strokeWidth).
		Attr("style", strokeStyle).
		MoveTo(lx, canvasBox.Bottom).
		LineToF(float64(lx), float64(canvasBox.Top)-ya.StrokeWidth/2).
		Render(w)

	var maxTextWidth int
	var finalTextY int
	for _, t := range ticks {
		ly := canvasBox.Bottom - ra.Translate(t.Value)
		tb := measureText(t.Label, AxisFontSize)

		if tb.Width() > maxTextWidth {
			maxTextWidth = tb.Width()
		}

		finalTextY = ly + tb.Height()>>1

		svg.Path().
			Attr("stroke-width", strokeWidth).
			Attr("style", strokeStyle).
			MoveTo(lx, ly).
			LineTo(lx+HorizontalTickWidth, ly).
			Render(w)

		svg.Text().
			Content(t.Label).
			Attr("style", fillStyle).
			Attr("x", svg.Point(tx)).
			Attr("y", svg.Point(finalTextY)).
			Render(w)
	}

	tb := measureText(ya.Name, AxisFontSize)
	tx = canvasBox.Right + YAxisMargin + maxTextWidth + YAxisMargin
	ty := canvasBox.Top + (canvasBox.Height()>>1 - tb.Height()>>1)

	svg.Text().
		Content(ya.Name).
		Attr("x", svg.Point(tx)).
		Attr("y", svg.Point(ty)).
		Attr("style", fillStyle).
		Attr("transform", rotate(90, tx, ty)).
		Render(w)
}
