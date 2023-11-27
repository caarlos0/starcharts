package chart

import (
	"fmt"
	"github.com/caarlos0/starcharts/internal/chart/svg"
	"io"
	"math"
)

type YAxis struct {
	Name string
}

func (ya YAxis) Measure(canvas Box, ra *Range, ticks []Tick) Box {
	tx := canvas.Right + YAxisMargin

	minx, maxx, miny, maxy := math.MaxInt32, 0, math.MaxInt32, 0
	var maxTextHeight int
	for _, t := range ticks {
		v := t.Value
		ly := canvas.Bottom - ra.Translate(v)

		tb := measureText(t.Label, AxisFontSize, nil)
		tbh2 := tb.Height() >> 1
		maxTextHeight = max(tb.Height(), maxTextHeight)

		minx = canvas.Right
		maxx = max(maxx, tx+tb.Width())

		miny = min(miny, ly-tbh2)
		maxy = max(maxy, ly+tbh2)
	}

	maxx += YAxisMargin + maxTextHeight

	return Box{
		Top:    miny,
		Left:   minx,
		Right:  maxx,
		Bottom: maxy,
	}
}

func (ya YAxis) Render(w io.Writer, canvasBox Box, ra *Range, ticks []Tick) {

	sw := 2

	lx := canvasBox.Right + sw
	tx := lx + YAxisMargin

	svg.Path().
		MoveTo(lx, canvasBox.Bottom).
		LineTo(lx, canvasBox.Top).
		Render(w)

	var maxTextWidth int
	var finalTextX, finalTextY int
	for _, t := range ticks {
		v := t.Value
		ly := canvasBox.Bottom - ra.Translate(v)

		tb := measureText(t.Label, AxisFontSize, nil)

		if tb.Width() > maxTextWidth {
			maxTextWidth = tb.Width()
		}

		finalTextX = tx

		finalTextY = ly + tb.Height()>>1

		svg.Path().
			MoveTo(lx, ly).
			LineTo(lx+DefaultHorizontalTickWidth, ly).
			Render(w)

		svg.Text().
			Content(t.Label).
			Attr("x", svg.Point(finalTextX)).
			Attr("y", svg.Point(finalTextY)).
			Render(w)
	}

	tb := measureText(ya.Name, AxisFontSize, nil)
	tx = canvasBox.Right + sw + YAxisMargin + maxTextWidth + YAxisMargin
	ty := canvasBox.Top + (canvasBox.Height()>>1 - tb.Height()>>1)

	svg.Text().
		Content(ya.Name).
		Attr("x", svg.Point(tx)).
		Attr("y", svg.Point(ty)).
		Attr("transform", fmt.Sprintf("rotate(%0.2f,%d,%d)", radiansToDegrees(degreesToRadians(90)), tx, ty)).
		Render(w)
}
