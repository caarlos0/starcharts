package chart

import (
	"fmt"
	"math"
	"strings"
)

type Tick struct {
	Value float64
	Label string
}

type Ticks []Tick

func (t Ticks) String() string {
	var values []string
	for i, tick := range t {
		values = append(values, fmt.Sprintf("[%d: %s]", i, tick.Label))
	}
	return strings.Join(values, ", ")
}

func generateTicks(rng *Range, isVertical bool, formatter ValueFormatter) []Tick {
	ticks := []Tick{
		{Value: rng.Min, Label: formatter(rng.Min)},
	}

	labelBox := measureText(formatter(rng.Min), AxisFontSize)

	var tickSize float64
	if isVertical {
		tickSize = float64(labelBox.Height() + MinimumTickVerticalSpacing)
	} else {
		tickSize = float64(labelBox.Width() + MinimumTickHorizontalSpacing)
	}

	domainRemainder := float64(rng.Domain) - (tickSize * 2)
	intermediateTickCount := int(math.Floor(domainRemainder / tickSize))

	rangeDelta := abs(rng.Max - rng.Min)
	tickStep := rangeDelta / float64(intermediateTickCount)

	roundTo := getRoundToForDelta(rangeDelta) / 10
	intermediateTickCount = min(intermediateTickCount, DefaultTickCountSanityCheck)

	for x := 1; x < intermediateTickCount; x++ {
		tickValue := rng.Min + roundUp(tickStep*float64(x), roundTo)
		ticks = append(ticks, Tick{
			Value: tickValue,
			Label: formatter(tickValue),
		})
	}

	return append(ticks, Tick{
		Value: rng.Max,
		Label: formatter(rng.Max),
	})
}
