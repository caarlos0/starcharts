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

	minX, maxX, minY, maxY := math.MaxInt32, 0, math.MaxInt32, 0
	maxTextHeight := 0
	for _, t := range ticks {
		ly := canvas.Bottom - ra.Translate(t.Value)


		tb := measureText(t.Label, AxisFontSize, nil)
		maxTextHeight = max(tb.Height(), maxTextHeight)

		minX = canvas.Right
		maxX = max(maxX, tx+tb.Width())

		tbh2 := tb.Height() >> 1
		minY = min(minY, ly-tbh2)
		maxY = max(maxY, ly+tbh2)
	}

	maxX += YAxisMargin + maxTextHeight

	return Box{
		Top:    minY,
		Left:   minX,
		Right:  maxX,
		Bottom: maxY,
	}
}

func (ya YAxis) Render(w io.Writer, canvasBox Box, ra *Range, ticks []Tick) {
	lx := canvasBox.Right
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
			LineTo(lx+HorizontalTickWidth, ly).
			Render(w)

		svg.Text().
			Content(t.Label).
			Attr("x", svg.Point(finalTextX)).
			Attr("y", svg.Point(finalTextY)).
			Render(w)
	}

	tb := measureText(ya.Name, AxisFontSize, nil)
	tx = canvasBox.Right + YAxisMargin + maxTextWidth + YAxisMargin
	ty := canvasBox.Top + (canvasBox.Height()>>1 - tb.Height()>>1)

	svg.Text().
		Content(ya.Name).
		Attr("x", svg.Point(tx)).
		Attr("y", svg.Point(ty)).
		Attr("transform", rotate(90, tx, ty)).
		Render(w)
}

func rotate(ang float32, x int, y int) string {
	return fmt.Sprintf("rotate(%0.2f,%d,%d)", ang, x, y)
}
