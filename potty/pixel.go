package potty

import (
	"math"
	"sort"

	colorful "github.com/lucasb-eyer/go-colorful"
)

type PixelSpace struct {
	Pixels []*Pixel   // Pixels all in a row
	Strips [][]*Pixel // Same Pixels organized by strips

	Len int // Total number of pixels

	// Bounding Box
	MaxX float64
	MaxY float64
	MaxZ float64
	MinX float64
	MinY float64
	MinZ float64

	MaxXFlat float64
	MinXFlat float64
}

// NewPixelSpace takes a slice of xyz coordinates and returns a slice of pixels
// in a bounding box.
func NewPixelSpace(locations []float64) *PixelSpace {
	len := len(locations) / 3
	b := &PixelSpace{
		Pixels: make([]*Pixel, len),
		Strips: make([][]*Pixel, 0),
		Len:    len,
	}

	stripMap := make(map[float64][]*Pixel) //helper for sorting pixels into strips

	for i := 0; i < len; i++ {
		b.Pixels[i] = NewPixel(locations[i*3+0], locations[i*3+1], locations[i*3+2])
		pixel := b.Pixels[i]

		// Calculate Bounding Box
		if i == 0 || pixel.X > b.MaxX {
			b.MaxX = pixel.X
		}
		if i == 0 || pixel.Y > b.MaxY {
			b.MaxY = pixel.Y
		}
		if i == 0 || pixel.Z > b.MaxZ {
			b.MaxZ = pixel.Z
		}
		if i == 0 || pixel.X < b.MinX {
			b.MinX = pixel.X
		}
		if i == 0 || pixel.Y < b.MinY {
			b.MinY = pixel.Y
		}
		if i == 0 || pixel.Z < b.MinZ {
			b.MinZ = pixel.Z
		}
		if i == 0 || pixel.XFlat > b.MaxXFlat {
			b.MaxXFlat = pixel.XFlat
		}
		if i == 0 || pixel.XFlat < b.MinXFlat {
			b.MinX = pixel.XFlat
		}

		// assume all pixels in a strip have the same XFlat value
		stripMap[pixel.XFlat] = append(stripMap[pixel.XFlat], pixel)
	}

	for _, strip := range stripMap {
		sort.Slice(strip, func(i, j int) bool { return strip[i].Z < strip[j].Z })
		b.Strips = append(b.Strips, strip)
	}

	sort.Slice(b.Strips, func(i, j int) bool { return b.Strips[i][0].XFlat < b.Strips[j][0].XFlat })

	return b
}

// Z coord for pixel in [0,1] space
func (b *PixelSpace) RandomPixel() *Pixel {
	i := int(math.Floor(RandGen.Float64() * float64(b.Len)))
	return b.Pixels[i]
}

// X coord for pixel in [0,1] space
func (b *PixelSpace) XNormal(pixel *Pixel) float64 {
	return (pixel.X - b.MinX) / b.MaxX
}

// Z coord for pixel in [0,1] space
func (b *PixelSpace) ZNormal(pixel *Pixel) float64 {
	return (pixel.Z - b.MinZ) / b.MaxZ
}

// XFlat coord for pixel in [0,1] space
func (b *PixelSpace) XFlatNormal(pixel *Pixel) float64 {
	return (pixel.XFlat - b.MinXFlat) / b.MaxXFlat
}

// Distance between two pixels in [0,1] space
func (b *PixelSpace) NormalDistance(p1, p2 *Pixel) float64 {
	pX := math.Pow(b.XFlatNormal(p1)-b.XFlatNormal(p2), 2)
	pZ := math.Pow(b.ZNormal(p1)-b.ZNormal(p2), 2)
	return math.Sqrt(pX + pZ)
}

func (b *PixelSpace) SetFromBytes(bytes []byte) {
	for i := 0; i < b.Len; i++ {
		b.Pixels[i].Color = ColorFromBytes(bytes[i*3+0], bytes[i*3+1], bytes[i*3+2])
	}
}

// ToBytes writes the Pixel to the output buffer
func (b *PixelSpace) ToBytes(bytes []byte) []byte {
	for i := 0; i < b.Len; i++ {
		bytes[i*3+0], bytes[i*3+1], bytes[i*3+2] = b.Pixels[i].ToBytes()
	}
	return bytes
}

type Pixel struct {
	Color colorful.Color

	X float64
	Y float64
	Z float64

	XFlat float64 // Y dimension flattened into X
}

func NewPixel(x, y, z float64) *Pixel {
	pixel := &Pixel{
		Color: White,
		X:     x,
		Y:     y,
		Z:     z,
	}

	if pixel.X < 0.1 {
		pixel.XFlat = pixel.X - pixel.Y
	} else {
		pixel.XFlat = pixel.X + pixel.Y
	}

	return pixel
}

func (p *Pixel) ToBytes() (r, g, b byte) {
	r, g, b = p.Color.RGB255()
	return
	R, G, B := p.Color.RGB255()
	return byte(R), byte(G), byte(B)
}

func ColorFromBytes(r, g, b byte) colorful.Color {
	return colorful.Color{R: FloatToByte(r), G: FloatToByte(g), B: FloatToByte(b)}
}

func FloatToByte(x byte) float64 {
	return float64(x) / 255
}
