package chart

import (
	"fmt"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"time"
)

func measureText(body string, size float64, textTheta *float64) Box {
	fc := &font.Drawer{
		Face: truetype.NewFace(GetFont(), &truetype.Options{
			DPI:  DPI,
			Size: size,
		}),
	}

	w := fc.MeasureString(body).Ceil()

	box := Box{
		Right:  w,
		Bottom: int(pointsToPixels(DPI, size)),
	}

	if textTheta == nil {
		return box
	}

	return box.Corners().Rotate(radiansToDegrees(*textTheta)).Box()
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
