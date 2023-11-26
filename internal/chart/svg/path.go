package svg

import (
	"fmt"
	"io"
	"math"
	"strings"
)

type PathBuilder struct {
	TagBuilder
	path []string
}

func (t *PathBuilder) Attr(key, value string) *PathBuilder {
	t.attributes[key] = value
	return t
}

func (t *PathBuilder) Content(content string) *PathBuilder {
	t.content.WriteString(content)

	return t
}

func (pb *PathBuilder) MoveTo(x, y int) *PathBuilder {
	pb.path = append(pb.path, fmt.Sprintf("M %d %d", x, y))

	return pb
}

func (pb *PathBuilder) LineTo(x, y int) *PathBuilder {
	pb.path = append(pb.path, fmt.Sprintf("L %d %d", x, y))

	return pb
}

func (pb *PathBuilder) ArcTo(cx, cy int, rx, ry, startAngle, delta float64) *PathBuilder {
	startAngle = RadianAdd(startAngle, _pi2)
	endAngle := RadianAdd(startAngle, delta)

	startX := cx + int(rx*math.Sin(startAngle))
	startY := cy - int(ry*math.Cos(startAngle))

	if len(pb.path) > 0 {
		pb.path = append(pb.path, fmt.Sprintf("L %d %d", startX, startY))
	} else {
		pb.path = append(pb.path, fmt.Sprintf("M %d %d", startX, startY))
	}

	endX := cx + int(rx*math.Sin(endAngle))
	endY := cy - int(ry*math.Cos(endAngle))

	dd := RadiansToDegrees(delta)

	largeArcFlag := 0
	if delta > _pi {
		largeArcFlag = 1
	}

	pb.path = append(pb.path, fmt.Sprintf("A %d %d %0.2f %d 1 %d %d", int(rx), int(ry), dd, largeArcFlag, endX, endY))

	return pb
}

func (pb *PathBuilder) WriteTo(io io.Writer) (n int, err error) {
	pb.attributes["d"] = strings.Join(pb.path, " ")

	return pb.TagBuilder.WriteTo(io)
}

func Path() *PathBuilder {
	return &PathBuilder{
		TagBuilder: TagBuilder{
			tag:        "path",
			attributes: map[string]string{},
		},
		path: []string{},
	}
}

func (pb *PathBuilder) String() string {
	builder := &strings.Builder{}
	pb.WriteTo(builder)
	return builder.String()
}
