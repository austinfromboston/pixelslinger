package potty

import (
	"math"

	"github.com/longears/pixelslinger/colorutils"
	"github.com/austinfromboston/pixelslinger/config"
	"github.com/longears/pixelslinger/midi"
)

const (
	DrainingComplete = 4.0                             // seconds
	RefillStarting   = 8.0                             // seconds
	RefillComplete   = 16.0                            // seconds
	RefillDuration   = RefillComplete - RefillStarting // seconds

	// Size of falling water streams when draining
	FlushStreamerSize = 0.15

	FlushControlPad = config.RIPPLE_PAD
)

type FlushEffect struct {
	space      *PixelSpace
	random     []float64
	isFlushing bool
	startTime  float64
}

func NewFlushEffect(space *PixelSpace) *FlushEffect {
	f := &FlushEffect{
		space:  space,
		random: make([]float64, space.Len),
	}

	for i := range f.random {
		f.random[i] = math.Pow(RandGen.Float64(), 30.0)
	}

	return f
}

func (f *FlushEffect) SetFlushState(midiState *midi.MidiState, t float64) {
	flushPad := midiState.KeyVolumes[FlushControlPad]

	switch {
	/* Fake flush
	case t > f.startTime+RefillComplete+3.25:
		f.isFlushing = true
		f.startTime = t
	*/
	case !f.isFlushing && flushPad > 0:
		f.isFlushing = true
		f.startTime = t
	case t > f.startTime+RefillComplete:
		f.isFlushing = false
	}
}

func (f *FlushEffect) Render(midiState *midi.MidiState, t float64) {
	f.SetFlushState(midiState, t)

	flushTime := t - f.startTime

	for i, p := range f.space.Pixels {

		// calculate wave pattern
		waterLevel := 1.0 - colorutils.Cos(p.X*0.2+p.Y*0.8, t/4, 1, 0.1, 0.6)/5.0

		switch {

		//Draining
		case flushTime < DrainingComplete:
			// lower the water level
			waterLevel -= flushTime / DrainingComplete
			// make falling streamers
			waterLevel += (FlushStreamerSize * f.random[i])

		//Resting
		case flushTime < RefillStarting:
			waterLevel -= 1.0

		//Refilling
		case flushTime < RefillComplete:
			refillTime := flushTime - RefillStarting
			waterLevel -= 1.0 - refillTime/RefillDuration
		}

		// If pixel is above the waterlevel, paint it black
		if waterLevel < f.space.ZNormal(p) {
			p.Color = Black
		}
	}
}
