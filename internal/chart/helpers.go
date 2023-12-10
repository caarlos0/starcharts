package chart

import (
	"fmt"
	"github.com/caarlos0/starcharts/internal/chart/svg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"time"
)

func measureText(body string, size float64) Box {
	drawer := &font.Drawer{
		Face: truetype.NewFace(GetFont(), &truetype.Options{
			DPI:  DPI,
			Size: size,
		}),
	}

	return Box{
		Right:  drawer.MeasureString(body).Ceil(),
		Bottom: int(pointsToPixels(DPI, size)),
	}
}

func timeValueFormatter(v interface{}) string {
	dateFormat := "2006-01-02"
	if typed, isTyped := v.(float64); isTyped {
		return time.Unix(0, int64(typed)).Format(dateFormat)
	}

	return ""
}

func intValueFormatter(v interface{}) string {
	return fmt.Sprintf("%.0f", v)
}

func rotate(ang float32, x int, y int) string {
	return fmt.Sprintf("rotate(%0.2f,%d,%d)", ang, x, y)
}

func normaliseStrokeWidth(strokeWidth float64) string {
	return svg.Point(max(MinStrokeWidth, strokeWidth))
}
