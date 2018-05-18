package potty

type PixelSpace struct {
	Pixels []*Pixel
	Len    int
	MaxX   float64
	MaxY   float64
	MaxZ   float64
	MinX   float64
	MinY   float64
	MinZ   float64
}

func NewPixelSpace(locations []float64) *PixelSpace {
	len := len(locations) / 3
	b := &PixelSpace{
		Pixels: make([]*Pixel, len),
		Len:    len,
	}

	for i := 0; i < len; i++ {
		b.Pixels[i] = NewPixel(locations[i*3+0], locations[i*3+1], locations[i*3+2])
		pixel := b.Pixels[i]

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
	}

	return b
}

func (b *PixelSpace) Pixel(pos int) *Pixel {
	return b.Pixels[pos]
}

func (b *PixelSpace) XNormal(pos int) float64 {
	return (b.Pixels[pos].X - b.MinX) / b.MaxX
}

func (b *PixelSpace) YNormal(pos int) float64 {
	return (b.Pixels[pos].Y - b.MinY) / b.MaxY
}

func (b *PixelSpace) ZNormal(pos int) float64 {
	return (b.Pixels[pos].Z - b.MinZ) / b.MaxZ
}

func (b *PixelSpace) ToBytes(bytes []byte) []byte {
	for i := 0; i < b.Len; i++ {
		bytes[i*3+0], bytes[i*3+1], bytes[i*3+2] = b.Pixels[i].ToBytes()
	}
	return bytes
}

type Pixel struct {
	Color
	X float64
	Y float64
	Z float64
}

func NewPixel(x, y, z float64) *Pixel {
	return &Pixel{
		Color: NewColor(1, 1, 1),
		X:     x,
		Y:     y,
		Z:     z,
	}
}
