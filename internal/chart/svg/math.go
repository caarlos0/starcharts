package svg

import "math"

const (
	_2pi = 2 * math.Pi
	_pi2 = math.Pi / 2
	_r2d = 180 / math.Pi
)

func RadianAdd(base, delta float64) float64 {
	value := base + delta
	if value > _2pi {
		return math.Mod(value, _2pi)
	}

	if value < 0 {
		return math.Mod(_2pi+value, _2pi)
	}

	return value
}

func RadiansToDegrees(value float64) float64 {
	return math.Mod(value, _2pi) * _r2d
}
