package chart

import (
	"github.com/caarlos0/starcharts/internal/chart/svg"
	"io"
	"time"
)

type Series struct {
	XValues []time.Time
	YValues []float64
}

func (ts Series) Len() int {
	return len(ts.XValues)
}
func (ts Series) GetValues(index int) (x, y float64) {
	x = toFloat64(ts.XValues[index])
	y = ts.YValues[index]
	return
}

func (ts Series) GetLastValues() (x, y float64) {
	x = toFloat64(ts.XValues[len(ts.XValues)-1])
	y = ts.YValues[len(ts.YValues)-1]
	return
}

// Render renders the series.
func (ts Series) Render(w io.Writer, canvasBox Box, xrange, yrange *Range) {
	if len(ts.XValues) == 0 {
		return
	}

	cb := canvasBox.Bottom
	cl := canvasBox.Left

	v0x, v0y := ts.GetValues(0)
	x0 := cl + xrange.Translate(v0x)
	y0 := cb - yrange.Translate(v0y)

	var vx, vy float64
	var x, y int

	path := svg.Path().MoveTo(x0, y0)

	for i := 1; i < ts.Len(); i++ {
		vx, vy = ts.GetValues(i)
		x = cl + xrange.Translate(vx)
		y = cb - yrange.Translate(vy)
		path.LineTo(x, y)
	}

	path.Render(w)
}
