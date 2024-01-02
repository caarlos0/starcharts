package chart

import (
	"math"
	"time"
)

type Number interface {
	int | int64 | float32 | float64
}

func getRoundToForDelta(delta float64) float64 {
	startingDeltaBound := math.Pow(10.0, 10.0)
	for cursor := startingDeltaBound; cursor > 0; cursor /= 10.0 {
		if delta > cursor {
			return cursor / 10.0
		}
	}

	return 0.0
}

func roundUp(value, roundTo float64) float64 {
	d1 := math.Ceil(value / roundTo)
	return d1 * roundTo
}

func roundDown(value, roundTo float64) float64 {
	d1 := math.Floor(value / roundTo)
	return d1 * roundTo
}

func abs[T Number](value T) T {
	if value < 0 {
		return -value
	}
	return value
}

func mean[T Number](values ...T) T {
	return sum(values...) / T(len(values))
}

func sum[T Number](values ...T) (total T) {
	for _, v := range values {
		total += v
	}

	return total
}

func degreesToRadians(degrees float64) float64 {
	return degrees * (math.Pi / 180)
}

func rotateCoordinate(cx, cy, x, y int, thetaRadians float64) (rx, ry int) {
	tempX, tempY := float64(x-cx), float64(y-cy)
	rotatedX := tempX*math.Cos(thetaRadians) - tempY*math.Sin(thetaRadians)
	rotatedY := tempX*math.Sin(thetaRadians) + tempY*math.Cos(thetaRadians)
	rx = int(rotatedX) + cx
	ry = int(rotatedY) + cy
	return
}

func toFloat64(t time.Time) float64 {
	return float64(t.UnixNano())
}

func pointsToPixels(dpi, points float64) float64 {
	return (points * dpi) / 72.0
}
