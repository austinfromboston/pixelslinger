package potty

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
)

type SlowColorEffect struct {
	space *PixelSpace
}

func NewSlowColorEffect(space *PixelSpace) *WaterEffect {
	return &WaterEffect{
		space: space,
	}
}

func (w *SlowColorEffect) Render(midiState *midi.MidiState, t float64) {
	for _, p := range w.space.Pixels {
		s1 := colorutils.Cos((p.Z*0.3)*(p.Y*0.3)*(p.X*0.3), t/8, 1, 0.1, 0.7)
		s3 := colorutils.Cos(p.X*0.2+p.Y*0.8, t/4, 1, 0.1, 0.6)
		s2 := colorutils.Cos(p.Z*0.2, t/10, 1, 0.0, 1.0)
		s4 := 0.3*s1 + 0.7*s3

		// p.Color = White
		p.Color = p.Color.BlendLab(LightCyan, s1)
		p.Color = p.Color.BlendLab(DarkCyan, s2)
		p.Color = p.Color.BlendLab(Aquamarine, s3)
		p.Color = p.Color.BlendLab(Teal, s4)
	}
}
