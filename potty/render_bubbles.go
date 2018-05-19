package potty

import (
	"github.com/longears/pixelslinger/midi"
)

const BSpread = 9       // How spread out the buubbles are (aka few can you see)
const BSpeed = 0.008    // How fast they go up
const BSpeedVar = 0.004 // Speed variation each time a bubble starts over

type BubbleEffect struct {
	space   *PixelSpace
	bubbles map[float64]*Bubble
}

func NewBubbleEffect(space *PixelSpace) *BubbleEffect {
	b := &BubbleEffect{
		space:   space,
		bubbles: make(map[float64]*Bubble),
	}

	for _, p := range space.Pixels {
		b.bubbles[p.XFlat] = NewBubble()
	}
	return b
}

func (b *BubbleEffect) Render(midiState *midi.MidiState, t float64) {
	for _, p := range b.space.Pixels {
		bubble := b.bubbles[p.XFlat]
		pZ := b.space.ZNormal(p)
		p.Color = p.Color.BlendLab(White, bubble.Strength(pZ))
	}

	for _, bubble := range b.bubbles {
		bubble.Move()
	}
}

type Bubble struct {
	Speed float64
	Z     float64
}

func NewBubble() *Bubble {
	return &Bubble{
		Z:     RandGen.Float64()*BSpread - BSpread/2,
		Speed: BSpeed + RandGen.Float64()*BSpeedVar,
	}
}

func (b *Bubble) Move() {
	b.Z += 0.01
	if b.Z >= BSpread {
		b.Z -= BSpread
		b.Speed = BSpeed + RandGen.Float64()*BSpeedVar
	}
}

func (b *Bubble) Strength(pZ float64) float64 {
	if pZ > b.Z {
		return 0.0
	}

	return Clamp(0.6 - (b.Z-pZ)/(b.Z+pZ)*10)
}

func Clamp(s float64) float64 {
	switch {
	case s <= 0.0:
		return 0.0
	case s >= 1.0:
		return 1.0
	default:
		return s
	}
}
