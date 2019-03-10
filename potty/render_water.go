package potty

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	colorful "github.com/lucasb-eyer/go-colorful"
)

var (
	LightCyan  = colorful.LinearRgb(0.6, 0.7, 0.8)
	DarkCyan   = colorful.LinearRgb(0.1, 0.4, 0.8)
	Aquamarine = colorful.LinearRgb(0.1, 0.25, 0.75)
	Teal       = colorful.LinearRgb(0.4, 0.7, 0.55)
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

		p.Color = Black
		p.Color = p.Color.BlendHsv(LightCyan, s1).Clamped()
		p.Color = p.Color.BlendHsv(DarkCyan, s2).Clamped()
		p.Color = p.Color.BlendHsv(Aquamarine, s3).Clamped()
		p.Color = p.Color.BlendHsv(Teal, s4).Clamped()
	}
}
