package svg

func Text() *TagBuilder {
	styleStr := "stroke-width:0;stroke:none;fill:rgba(51,51,51,1.0);font-size:12.8px;font-family:'Roboto Medium',sans-serif"

	return &TagBuilder{
		tag: "text",
		attributes: map[string]string{
			"style": styleStr,
		},
	}
}
