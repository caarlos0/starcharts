package svg

type StyleBuilder struct {
	TagBuilder
}

func Style() *StyleBuilder {
	return &StyleBuilder{
		TagBuilder{
			tag:        "style",
			attributes: map[string]string{},
		},
	}
}
