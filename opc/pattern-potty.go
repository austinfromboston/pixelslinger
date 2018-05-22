package opc

import (
	"github.com/longears/pixelslinger/potty"
)

// This obnoxious function only exists because ByteThread is an opc type
// and thus cannot be referenced from potty without creating a dependecy cycle.
func MakePatternHousePotty(locations []float64) ByteThread {
	return potty.MakeWaterPattern(locations)
}
