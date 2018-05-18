package cistern

type ZigZagStrip struct {
	height int
	length int
	offsetY int
}

func (strip *ZigZagStrip) VerticalHeight(y int) int {
	if y > strip.height {
		return strip.height - (y - strip.height)
	}
	return y
}

type StripLoc struct {
	x int
	strip ZigZagStrip
}

type Pixel struct {
	r byte
	g byte
	b byte
}


type TankCanvas struct {
	strips []StripLoc
	pixels [][]Pixel
}

func (canvas *TankCanvas) Init(qty int) {
	var strip ZigZagStrip;
	var stripLoc StripLoc;
	for i = 0; i < qty; i++ {
		strip = ZigZagStrip{148, 296, 24}
		stripLoc = StripLoc{i*48,  strip}
		canvas.strips = append(canvas.strips, stripLoc)
	}
}

func (canvas *TankCanvas) Square(x int, y int, size int) {

}