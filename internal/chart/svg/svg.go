package svg

func SVG() *TagBuilder {
	return &TagBuilder{tag: "svg", attributes: map[string]string{
		"xmlns":       "http://www.w3.org/2000/svg",
		"xmlns:xlink": "http://www.w3.org/1999/xlink",
	}}
}
