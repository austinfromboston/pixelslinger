package potty

import (
	"math"
	"math/rand"

	"github.com/longears/pixelslinger/config"
	"github.com/longears/pixelslinger/midi"
	colorful "github.com/lucasb-eyer/go-colorful"
)

const (
	CSpeed    = 0.004 // How fast they go up
	CSpeedVar = 0.004 // Speed variation each time a circle starts over
	CLifeSpan = 0.75

	CircleButton = config.BLINK_CIRCLE_PAD
)

type ColorDanceEffect struct {
	space            *PixelSpace
	circles          map[float64]*Circle
	resetTime        float64
	currentCLifeSpan float64
	buttonPressed    bool
	fakeButtonPress  float64
}

func NewColorDanceEffect(space *PixelSpace) *ColorDanceEffect {
	return &ColorDanceEffect{
		space:   space,
		circles: make(map[float64]*Circle),
	}
}

func (e *ColorDanceEffect) Render(midiState *midi.MidiState, t float64) {
	for _, circle := range e.circles {
		circle.Move(t)
		if circle.Dead {
			delete(e.circles, circle.ID())
		}
	}
	/* fake button
	if t > e.fakeButtonPress+0.5 {
		e.fakeButtonPress = t
		midiState.KeyVolumes[CircleButton] = 100
	} else {
		midiState.KeyVolumes[CircleButton] = 0
	}
	*/
	if midiState.KeyVolumes[CircleButton] == 0 {
		e.buttonPressed = false
	}

	if !e.buttonPressed && midiState.KeyVolumes[CircleButton] > 0 {
		e.buttonPressed = true
		circle := NewCircle(e.space, t)
		e.circles[circle.ID()] = circle
	}

	for _, pixel := range e.space.Pixels {
		//pixel.Color = Black
		for _, circle := range e.circles {
			pixel.Color = circle.Blend(pixel)
		}
	}
}

type Circle struct {
	space       *PixelSpace
	colorPicker colorPicker
	Pixel
	Speed    float64
	Strength float64
	EndTime  float64
	Dead     bool
}

func NewCircle(space *PixelSpace, t float64) *Circle {
	p := space.RandomPixel()
	cp := NextColorPicker()
	p.Color = cp()
	return &Circle{
		space:       space,
		colorPicker: cp,
		Pixel:       *p,
		Speed:       CSpeed + RandGen.Float64()*CSpeedVar,
		EndTime:     t + CLifeSpan,
	}

}

func (c *Circle) ID() float64 {
	return c.EndTime
}

func (c *Circle) Blend(pixel *Pixel) colorful.Color {
	falloff := math.Pow(1.0-c.space.NormalDistance(pixel, &c.Pixel), 1)
	strength := falloff * c.Strength
	//color := c.colorPicker()
	color := c.Color
	return pixel.Color.BlendLab(color, strength).Clamped()
}

func (c *Circle) Move(t float64) {
	if t > c.EndTime {
		c.Dead = true
		return
	}

	c.Strength = (c.EndTime - t) / CLifeSpan

	c.Z += c.Speed
	if c.Z >= c.space.MaxZ {
		c.Z = c.space.MinZ
	}
}

type colorPicker func() colorful.Color

var colorPickers = []colorPicker{
	//RandBlue,
	RandGreen,
	RandRed,
}

var nextColorPicker = 0

func NextColorPicker() colorPicker {
	nextColorPicker++
	if nextColorPicker == len(colorPickers) {
		nextColorPicker = 0
	}
	return colorPickers[nextColorPicker]
}

func RandBlue() colorful.Color {
	return colorful.Hsv(RandVal(340, 360), 1.0, 1.0)
}

func RandGreen() colorful.Color {
	return colorful.Hsv(RandVal(30, 40), 1.0, 1.0)
}

func RandRed() colorful.Color {
	return colorful.Hsv(RandVal(0, 20), 1.0, 1.0)
}

func RandVal(low, high float64) float64 {
	return math.Floor(rand.Float64()*high) + low
}
