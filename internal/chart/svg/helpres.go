package svg

import "fmt"

func Px(value int) string {
	return fmt.Sprintf("%dpx", value)
}

func Point(value int) string {
	return fmt.Sprintf("%d", value)
}
