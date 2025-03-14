package svg

func Text() *TagBuilder {
	return &TagBuilder{
		tag: "text",
		attributes: map[string]string{},
	}
}
