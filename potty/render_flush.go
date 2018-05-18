package potty

import (
	"math"

	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
)

const (
	FLUSH_LENGTH = 4.0 //seconds
	FLUSH_REFILL = 8.0 //seconds
	FLUSH_REST   = 4.0 //seconds
	FLUSH_CYCLE  = FLUSH_LENGTH + FLUSH_REST + FLUSH_REFILL

	// degree to which the stips don't just go up and down in unison
	// higher is less jitter
	FLUSH_JITTER = 0.15
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

	// TODO: check midiState for flush signal...
	flushSignalRecieved := t > f.startTime+FLUSH_CYCLE+10

	switch {
	case !f.isFlushing && flushSignalRecieved:
		f.isFlushing = true
		f.startTime = t
	case t > f.startTime+FLUSH_CYCLE:
		f.isFlushing = false
	}
}

func (f *FlushEffect) Render(midiState *midi.MidiState, t float64) {
	f.SetFlushState(midiState, t)

	flushTime := t - f.startTime

	for i, p := range f.space.Pixels {
		waterLevel := 1.0 - colorutils.Cos(p.X*0.2+p.Y*0.8, t/4, 1, 0.1, 0.6)/5.0
		switch {
		//flushing
		case flushTime < FLUSH_LENGTH:
			waterLevel -= flushTime / FLUSH_LENGTH
			waterLevel += (FLUSH_JITTER * f.random[i])
		//drained
		case flushTime < FLUSH_LENGTH+FLUSH_REST:
			waterLevel -= 1.0
		//refilling
		case flushTime < FLUSH_CYCLE:
			refillTime := flushTime - (FLUSH_LENGTH + FLUSH_REST)
			waterLevel -= 1.0 - refillTime/FLUSH_REFILL
		}

		// If pixel is above the waterlevel, paint it black
		if waterLevel < f.space.ZNormal(i) {
			p.Color = Black
		}
	}
}
