package svg

func Rect() *TagBuilder {
	return &TagBuilder{tag: "rect", attributes: map[string]string{}}
}
