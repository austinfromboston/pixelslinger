// potty is a little pixel library and set of toilet effects for pixelslinger
package potty

import (
	"math/rand"
	"time"

	"github.com/longears/pixelslinger/midi"
)

// make persistant random values
var RandGen = rand.New(rand.NewSource(9))

type Renderer interface {
	Render(midiState *midi.MidiState, t float64)
}

func MakePattern(locations []float64) func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
	// create the pixels and coordinate space
	space := NewPixelSpace(locations)

	// create the renderers
	renderers := []Renderer{
		NewWaterEffect(space),
		NewBubbleEffect(space),
		NewFlushEffect(space),
	}

	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		for bytes := range bytesIn {
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8
			for _, r := range renderers {
				r.Render(midiState, t)
			}
			bytesOut <- space.ToBytes(bytes)
		}
	}
}
