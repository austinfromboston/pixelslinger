package potty

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
)

type WaterEffect struct {
	space *PixelSpace
}

func NewWaterEffect(space *PixelSpace) *WaterEffect {
	return &WaterEffect{
		space: space,
	}
}

func (w *WaterEffect) Render(midiState *midi.MidiState, t float64) {
	for _, p := range w.space.Pixels {
		s1 := colorutils.Cos((p.Z*0.3)*(p.Y*0.3)*(p.X*0.3), t/8, 1, 0.1, 0.7)
		s3 := colorutils.Cos(p.X*0.2+p.Y*0.8, t/4, 1, 0.1, 0.6)
		s2 := colorutils.Cos(p.Z*0.2, t/10, 1, 0.0, 1.0)
		s4 := 0.3*s1 + 0.7*s3

		p.Blend(&LightCyan, s1)
		p.Blend(&DarkCyan, s2)
		p.Blend(&Aquamarine, s3)
		p.Blend(&Teal, s4)
	}
}
