package chart

import (
	"math"
)

type Range struct {
	Min    float64
	Max    float64
	Domain int
}

func (r *Range) GetDelta() float64 {
	return r.Max - r.Min
}

func (r *Range) Translate(value float64) int {
	normalized := value - r.Min
	ratio := normalized / r.GetDelta()

	return int(math.Ceil(ratio * float64(r.Domain)))
}
