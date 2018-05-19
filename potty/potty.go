// potty is a little pixel library and set of toilet effects for pixelslinger
package potty

import (
	"math/rand"
	"time"

	"github.com/longears/pixelslinger/midi"
	colorful "github.com/lucasb-eyer/go-colorful"
)

// make persistant random values
var RandGen = rand.New(rand.NewSource(9))

var (
	White      = colorful.LinearRgb(1, 1, 1)
	LightCyan  = colorful.LinearRgb(0.875, 1, 1)
	Cyan       = colorful.LinearRgb(0.1, 1, 1)
	DarkCyan   = colorful.LinearRgb(0.1, 0.6, 0.6)
	Aquamarine = colorful.LinearRgb(0.1, 0.25, 0.65)
	Teal       = colorful.LinearRgb(0.79, 1, 0.89)
	Black      = colorful.LinearRgb(0, 0, 0)
)

type Renderer interface {
	Render(midiState *midi.MidiState, t float64)
}

func MakePattern(locations []float64) func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
	// create the pixels and coordinate space
	space := NewPixelSpace(locations)

	// create the renderers
	renderStack := []Renderer{
		NewWaterEffect(space), // water should go first to provide a base color
		NewColorDanceEffect(space),
		NewBubbleEffect(space), // tiny bubbles
		NewFlushEffect(space),  // flush should go last to mask off the shape
	}

	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		for bytes := range bytesIn {
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8

			for _, r := range renderStack {
				r.Render(midiState, t)
			}

			bytesOut <- space.ToBytes(bytes)
		}
	}
}
