package potty

import (
	"math"

	"github.com/longears/pixelslinger/config"
	"github.com/longears/pixelslinger/midi"
	colorful "github.com/lucasb-eyer/go-colorful"
)

const (
	CSpread   = 9     // How spread out the buubbles are (aka few can you see)
	CSpeed    = 0.004 // How fast they go up
	CSpeedVar = 0.004 // Speed variation each time a circle starts over
	CLifeSpan = 1.0

	CircleButton = config.BLINK_CIRCLE_PAD
)

type ColorDanceEffect struct {
	space            *PixelSpace
	circles          map[float64]*Circle
	resetTime        float64
	currentCLifeSpan float64
	buttonPressed    bool
}

func NewColorDanceEffect(space *PixelSpace) *ColorDanceEffect {
	return &ColorDanceEffect{
		space:   space,
		circles: make(map[float64]*Circle),
	}
}

func (e *ColorDanceEffect) Render(midiState *midi.MidiState, t float64) {
	if !e.buttonPressed && midiState.KeyVolumes[CircleButton] > 0 {
		e.buttonPressed = true
		circle := NewCircle(e.space, t)
		e.circles[circle.ID()] = circle
	}

	if midiState.KeyVolumes[CircleButton] == 0 {
		e.buttonPressed = false
	}

	for _, pixel := range e.space.Pixels {
		for _, circle := range e.circles {
			circle.Blend(pixel)
		}
	}

	for _, circle := range e.circles {
		circle.Move(t)
		if circle.Dead {
			delete(e.circles, circle.ID())
		}
	}
}

type Circle struct {
	space *PixelSpace
	Pixel
	Speed    float64
	Strength float64
	EndTime  float64
	Dead     bool
}

func NewCircle(space *PixelSpace, t float64) *Circle {
	p := space.RandomPixel()
	p.Color = colorful.HappyColor()
	return &Circle{
		space:   space,
		Pixel:   *p,
		Speed:   CSpeed + RandGen.Float64()*CSpeedVar,
		EndTime: t + CLifeSpan,
	}

}

func (c *Circle) ID() float64 {
	return c.EndTime
}

func (c *Circle) Blend(pixel *Pixel) {
	dist := c.space.NormalDistance(pixel, &c.Pixel)
	falloff := math.Pow(1.0-dist, 8)
	strength := falloff * c.Strength
	color := c.Color.BlendHcl(White, falloff*0.9).Clamped()

	pixel.Color = pixel.Color.BlendRgb(color, strength).Clamped()
}

func (c *Circle) Move(t float64) {
	if t > c.EndTime {
		c.Dead = true
		return
	}

	c.Strength = (c.EndTime - t) / CLifeSpan

	c.Z += c.Speed
	if c.Z >= CSpread {
		c.Z -= CSpread
	}
}
