package chart

import (
	"github.com/caarlos0/starcharts/internal/chart/svg"
	"io"
	"math"
)

type XAxis struct {
	Name string
}

func (xa XAxis) Measure(canvas Box, ra *Range, ticks []Tick) Box {
	var ltx, rtx int
	var tx, ty int
	var left, right, bottom = math.MaxInt32, 0, 0
	for _, t := range ticks {
		v := t.Value
		tb := measureText(t.Label, AxisFontSize, nil)

		tx = canvas.Left + ra.Translate(v)
		ty = canvas.Bottom + XAxisMargin + tb.Height()
		ltx = tx - tb.Width()>>1
		rtx = tx + tb.Width()>>1

		left = min(left, ltx)
		right = max(right, rtx)
		bottom = max(bottom, ty)
	}

	tb := measureText(xa.Name, AxisFontSize, nil)
	bottom += XAxisMargin + tb.Height()

	return Box{
		Top:    canvas.Bottom,
		Left:   left,
		Right:  right,
		Bottom: bottom,
	}
}

func (xa XAxis) Render(w io.Writer, canvasBox Box, ra *Range, ticks []Tick) {
	svg.Path().
		MoveTo(canvasBox.Left, canvasBox.Bottom).
		LineTo(canvasBox.Right, canvasBox.Bottom).
		Render(w)

	var tx, ty int
	var maxTextHeight int
	for _, t := range ticks {
		v := t.Value
		lx := ra.Translate(v)

		tx = canvasBox.Left + lx

		svg.Path().
			MoveTo(tx, canvasBox.Bottom).
			LineTo(tx, canvasBox.Bottom+VerticalTickHeight).
			Render(w)

		tb := measureText(t.Label, AxisFontSize, nil)

		tx = tx - tb.Width()>>1
		ty = canvasBox.Bottom + XAxisMargin + tb.Height()

		svg.Text().
			Content(t.Label).
			Attr("x", svg.Point(tx)).
			Attr("y", svg.Point(ty)).
			Render(w)

		maxTextHeight = max(maxTextHeight, tb.Height())
	}

	tb := measureText(xa.Name, AxisFontSize, nil)
	tx = canvasBox.Right - (canvasBox.Width()>>1 + tb.Width()>>1)
	ty = canvasBox.Bottom + XAxisMargin + maxTextHeight + XAxisMargin + tb.Height()

	svg.Text().
		Content(xa.Name).
		Attr("x", svg.Point(tx)).
		Attr("y", svg.Point(ty)).
		Render(w)
}
