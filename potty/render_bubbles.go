package potty

import (
	"github.com/longears/pixelslinger/midi"
)

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
		b.bubbles[p.X] = NewBubble()
	}
	return b
}

func (b *BubbleEffect) Render(midiState *midi.MidiState, t float64) {
	for i, p := range b.space.Pixels {
		// don't render bubbles on the side, too lazy..
		if p.Y > 0 {
			continue
		}
		bubble := b.bubbles[p.X]
		pZ := b.space.ZNormal(i)
		p.Blend(&White, bubble.Strength(pZ))
	}

	for _, bubble := range b.bubbles {
		bubble.Move()
	}
}

// Bubble Strength
const BSpread = 9
const BSpeed = 0.008
const BSpeedVar = 0.004

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
		return 0
	}
	return 0.6 - (b.Z-pZ)/(b.Z+pZ)*10
}
