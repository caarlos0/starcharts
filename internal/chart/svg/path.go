package svg

import (
	"fmt"
	"math"
)

type PathBuilder struct {
	TagBuilder
	path []string
}

// MoveTo implements the interface method.
func (vr *PathBuilder) MoveTo(x, y int) {
	vr.path = append(vr.path, fmt.Sprintf("M %d %d", x, y))
}

// LineTo implements the interface method.
func (vr *PathBuilder) LineTo(x, y int) {
	vr.path = append(vr.path, fmt.Sprintf("L %d %d", x, y))
}

func (vr *PathBuilder) ArcTo(cx, cy int, rx, ry, startAngle, delta float64) {
	startAngle = RadianAdd(startAngle, _pi2)
	endAngle := RadianAdd(startAngle, delta)

	startx := cx + int(rx*math.Sin(startAngle))
	starty := cy - int(ry*math.Cos(startAngle))

	if len(vr.path) > 0 {
		vr.path = append(vr.path, fmt.Sprintf("L %d %d", startx, starty))
	} else {
		vr.path = append(vr.path, fmt.Sprintf("M %d %d", startx, starty))
	}

	endx := cx + int(rx*math.Sin(endAngle))
	endy := cy - int(ry*math.Cos(endAngle))

	dd := RadiansToDegrees(delta)

	largeArcFlag := 0
	if delta > _pi {
		largeArcFlag = 1
	}

	vr.path = append(vr.path, fmt.Sprintf("A %d %d %0.2f %d 1 %d %d", int(rx), int(ry), dd, largeArcFlag, endx, endy))
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
