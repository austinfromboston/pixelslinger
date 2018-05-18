package potty

import "github.com/longears/pixelslinger/colorutils"

var (
	White      = NewColor(1, 1, 1)
	LightCyan  = NewColor(0.875, 1, 1)
	Cyan       = NewColor(0.1, 1, 1)
	DarkCyan   = NewColor(0.1, 0.3, 0.3)
	Aquamarine = NewColor(0.1, 0.25, 0.65)
	Teal       = NewColor(0.79, 1, 0.89)
	Black      = NewColor(0, 0, 0)
)

type Color struct {
	R float64
	G float64
	B float64
}

func NewColor(r, g, b float64) Color {
	return Color{r, g, b}
}

func (p *Color) Blend(p2 *Color, weight float64) {
	if weight <= 0 {
		return
	}
	p.R = p.R + weight*(p2.R-p.R)
	p.G = p.G + weight*(p2.G-p.G)
	p.B = p.B + weight*(p2.B-p.B)
}

func (p *Color) ToBytes() (r, g, b byte) {
	return colorutils.FloatToByte(p.R), colorutils.FloatToByte(p.G), colorutils.FloatToByte(p.B)
}
