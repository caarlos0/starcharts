package svg

import "fmt"

type Number interface {
	int | int64 | float32 | float64
}

func Px[T Number](value T) string {
	return fmt.Sprintf("%vpx", value)
}

func Point[T Number](value T) string {
	return fmt.Sprintf("%v", value)
}
