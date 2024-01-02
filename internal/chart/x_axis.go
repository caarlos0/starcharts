package chart

import (
	"github.com/caarlos0/starcharts/internal/chart/svg"
	"io"
	"math"
)

type XAxis struct {
	Name        string
	StrokeWidth float64
	Color       string
}

func (xa *XAxis) Measure(canvas *Box, ra *Range, ticks []Tick) *Box {
	var ltx, rtx int
	var tx, ty int
	var left, right, bottom = math.MaxInt32, 0, 0
	for _, t := range ticks {
		v := t.Value
		tb := measureText(t.Label, AxisFontSize)

		tx = canvas.Left + ra.Translate(v)
		ty = canvas.Bottom + XAxisMargin + tb.Height()
		ltx = tx - tb.Width()>>1
		rtx = tx + tb.Width()>>1

		left = min(left, ltx)
		right = max(right, rtx)
		bottom = max(bottom, ty)
	}

	tb := measureText(xa.Name, AxisFontSize)
	bottom += XAxisMargin + tb.Height()

	return &Box{
		Top:    canvas.Bottom,
		Left:   left,
		Right:  right,
		Bottom: bottom,
	}
}

func (xa *XAxis) Render(w io.Writer, canvasBox *Box, ra *Range, ticks []Tick) {
	strokeWidth := normaliseStrokeWidth(xa.StrokeWidth)
	strokeStyle := styles("stroke", xa.Color)
	fillStyle := styles("fill", xa.Color)

	svg.Path().
		Attr("stroke-width", strokeWidth).
		Attr("style", strokeStyle).
		MoveToF(float64(canvasBox.Left)-xa.StrokeWidth/2, float64(canvasBox.Bottom)).
		LineTo(canvasBox.Right, canvasBox.Bottom).
		Render(w)

	var tx, ty int
	var maxTextHeight int
	for _, t := range ticks {
		v := t.Value
		lx := ra.Translate(v)

		tx = canvasBox.Left + lx

		svg.Path().
			Attr("stroke-width", strokeWidth).
			Attr("style", strokeStyle).
			MoveTo(tx, canvasBox.Bottom).
			LineTo(tx, canvasBox.Bottom+VerticalTickHeight).
			Render(w)

		tb := measureText(t.Label, AxisFontSize)

		tx = tx - tb.Width()>>1
		ty = canvasBox.Bottom + XAxisMargin + tb.Height()

		svg.Text().
			Content(t.Label).
			Attr("style", fillStyle).
			Attr("x", svg.Point(tx)).
			Attr("y", svg.Point(ty)).
			Render(w)

		maxTextHeight = max(maxTextHeight, tb.Height())
	}

	tb := measureText(xa.Name, AxisFontSize)
	tx = canvasBox.Right - (canvasBox.Width()>>1 + tb.Width()>>1)
	ty = canvasBox.Bottom + XAxisMargin + maxTextHeight + XAxisMargin + tb.Height()

	svg.Text().
		Content(xa.Name).
		Attr("style", fillStyle).
		Attr("x", svg.Point(tx)).
		Attr("y", svg.Point(ty)).
		Render(w)
}
