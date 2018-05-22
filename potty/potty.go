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
	White = colorful.LinearRgb(1, 1, 1)
	Black = colorful.LinearRgb(0, 0, 0)
)

type Renderer interface {
	Render(midiState *midi.MidiState, t float64)
}

func MakeWaterPattern(locations []float64) func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
	space := NewPixelSpace(locations)

	return makePattern(space, []Renderer{
		NewWaterEffect(space),  // water should go first to provide a base color
		NewBubbleEffect(space), // tiny bubbles
	})
}

func MakeEffectFaderPattern(locations []float64) func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
	space := NewPixelSpace(locations)

	return makePattern(space, []Renderer{
		NewColorDanceEffect(space), // banging colors to the midi beat
		NewFlushEffect(space),      // flush masks off the shape
	})
}

func makePattern(space *PixelSpace, renderStack []Renderer) func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		for bytes := range bytesIn {
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8
			space.SetFromBytes(bytes)

			for _, r := range renderStack {
				r.Render(midiState, t)
			}

			bytesOut <- space.ToBytes(bytes)
		}
	}
}
