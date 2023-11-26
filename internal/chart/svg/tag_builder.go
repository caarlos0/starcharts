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

func (t *TagBuilder) WriteTo(io io.Writer) (n int, err error) {
	return fmt.Fprintf(io, "<%s %s>%s</%s>", t.tag, t.attrString(), t.content.String(), t.tag)
}

func (t *TagBuilder) Attr(key, value string) *TagBuilder {
	t.attributes[key] = value
	return t
}

func (t *TagBuilder) Content(content string) *TagBuilder {
	t.content.WriteString(content)

	return t
}

func (t *TagBuilder) ContentFunc(contentFunc func() string) *TagBuilder {
	t.content.WriteString(contentFunc())

	return t
}

func (t *TagBuilder) String() string {
	builder := strings.Builder{}

	t.WriteTo(&builder)

	return builder.String()
}

func (t *TagBuilder) attrString() string {
	var attrs []string
	for key, value := range t.attributes {
		attrs = append(attrs, fmt.Sprintf(`%s="%s"`, key, value))
	}

	return strings.Join(attrs, " ")
}
