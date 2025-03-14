package svg

import (
	"fmt"
	"io"
	"strings"
)

type TagBuilder struct {
	tag        string
	attributes map[string]string
	content    strings.Builder
}

func (t *TagBuilder) Write(p []byte) (n int, err error) {
	return t.content.Write(p)
}

func (t *TagBuilder) Render(io io.Writer) {
	if t.content.Len() == 0 {
		_, err := fmt.Fprintf(io, "<%s %s />", t.tag, t.attrString())
		if err != nil {
			panic(err)
		}
	} else {
		_, err := fmt.Fprintf(io, "<%s %s>%s</%s>", t.tag, t.attrString(), t.content.String(), t.tag)
		if err != nil {
			panic(err)
		}
	}
}

func (t *TagBuilder) Attr(key, value string) *TagBuilder {
	if value == "" {
		delete(t.attributes, key)
		return t
	}
	t.attributes[key] = value
	return t
}

func (t *TagBuilder) Content(content string) *TagBuilder {
	t.content.WriteString(content)
	return t
}

func (t *TagBuilder) ContentFunc(contentFunc func(writer io.Writer)) *TagBuilder {
	contentFunc(&t.content)

	return t
}

func (t *TagBuilder) String() string {
	builder := strings.Builder{}

	t.Render(&builder)

	return builder.String()
}

func (t *TagBuilder) attrString() string {
	var attrs []string
	for key, value := range t.attributes {
		attrs = append(attrs, fmt.Sprintf(`%s="%s"`, key, value))
	}

	return strings.Join(attrs, " ")
}
