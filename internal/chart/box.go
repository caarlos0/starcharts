package chart

type Box struct {
	Top    int
	Left   int
	Right  int
	Bottom int
}

func (b *Box) Width() int {
	return abs(b.Right - b.Left)
}

func (b *Box) Height() int {
	return abs(b.Bottom - b.Top)
}

func (b *Box) Center() (x, y int) {
	w2, h2 := b.Width()>>1, b.Height()>>1
	return b.Left + w2, b.Top + h2
}

func (b *Box) Clone() *Box {
	return &Box{
		Top:    b.Top,
		Left:   b.Left,
		Right:  b.Right,
		Bottom: b.Bottom,
	}
}

func (b *Box) Grow(other *Box) *Box {
	return &Box{
		Top:    min(b.Top, other.Top),
		Left:   min(b.Left, other.Left),
		Right:  max(b.Right, other.Right),
		Bottom: max(b.Bottom, other.Bottom),
	}
}

func (b *Box) Corners() *BoxCorners {
	return &BoxCorners{
		TopLeft:     Point{b.Left, b.Top},
		TopRight:    Point{b.Right, b.Top},
		BottomRight: Point{b.Right, b.Bottom},
		BottomLeft:  Point{b.Left, b.Bottom},
	}
}

func (b *Box) OuterConstrain(bounds, other *Box) *Box {
	newBox := b.Clone()
	if other.Top < bounds.Top {
		delta := bounds.Top - other.Top
		newBox.Top = b.Top + delta
	}

	if other.Left < bounds.Left {
		delta := bounds.Left - other.Left
		newBox.Left = b.Left + delta
	}

	if other.Right > bounds.Right {
		delta := other.Right - bounds.Right
		newBox.Right = b.Right - delta
	}

	if other.Bottom > bounds.Bottom {
		delta := other.Bottom - bounds.Bottom
		newBox.Bottom = b.Bottom - delta
	}
	return newBox
}

type BoxCorners struct {
	TopLeft, TopRight, BottomRight, BottomLeft Point
}

func (bc *BoxCorners) Box() *Box {
	return &Box{
		Top:    min(bc.TopLeft.Y, bc.TopRight.Y),
		Left:   min(bc.TopLeft.X, bc.BottomLeft.X),
		Right:  max(bc.TopRight.X, bc.BottomRight.X),
		Bottom: max(bc.BottomLeft.Y, bc.BottomRight.Y),
	}
}

func (bc *BoxCorners) Center() (x, y int) {
	left := mean(bc.TopLeft.X, bc.BottomLeft.X)
	right := mean(bc.TopRight.X, bc.BottomRight.X)
	x = ((right - left) >> 1) + left

	top := mean(bc.TopLeft.Y, bc.TopRight.Y)
	bottom := mean(bc.BottomLeft.Y, bc.BottomRight.Y)
	y = ((bottom - top) >> 1) + top

	return
}

func (bc *BoxCorners) Rotate(thetaDegrees float64) *BoxCorners {
	cx, cy := bc.Center()

	thetaRadians := degreesToRadians(thetaDegrees)

	tlx, tly := rotateCoordinate(cx, cy, bc.TopLeft.X, bc.TopLeft.Y, thetaRadians)
	trx, try := rotateCoordinate(cx, cy, bc.TopRight.X, bc.TopRight.Y, thetaRadians)
	brx, bry := rotateCoordinate(cx, cy, bc.BottomRight.X, bc.BottomRight.Y, thetaRadians)
	blx, bly := rotateCoordinate(cx, cy, bc.BottomLeft.X, bc.BottomLeft.Y, thetaRadians)

	return &BoxCorners{
		TopLeft:     Point{tlx, tly},
		TopRight:    Point{trx, try},
		BottomRight: Point{brx, bry},
		BottomLeft:  Point{blx, bly},
	}
}

type Point struct {
	X, Y int
}
