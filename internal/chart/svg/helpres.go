package svg

import "fmt"

func Px(value int) string {
	return fmt.Sprintf("%dpx", value)
}
