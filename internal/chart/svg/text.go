package svg

func Text() *TagBuilder {
	return &TagBuilder{
		tag: "text",
		attributes: map[string]string{
			"xmlns": "http://www.w3.org/2000/svg",
		},
	}
}
