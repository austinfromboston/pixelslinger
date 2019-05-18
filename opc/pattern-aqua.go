package opc

// Spatial Stripes
//   Creates spatial sine wave stripes: x in the red channel, y--green, z--blue
//   Also makes a white dot which moves down the strip non-spatially in the order
//   that the LEDs are indexed.

import (
	"github.com/austinfromboston/pixelslinger/colorutils"
	"github.com/austinfromboston/pixelslinger/midi"
	"math"
	"time"
)


func MakePatternAqua(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8
			// fill in bytes slice
			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------

				// make moving stripes for x, y, and z
				x := locations[ii*3+0]
				y := locations[ii*3+1]
				z := locations[ii*3+2]
				//r := colorutils.Cos(x, t/4, 1, 0, 0.7) // offset, period, min, max

				s1 := colorutils.Cos((z*0.3)*(y*0.3)*(x*0.3)    , t/8, 1, 0.05, 0.7)
				s3 := colorutils.Cos(x*0.2 + y*0.8, t/16, 1, 0.05, 0.5)
				s2 := colorutils.Cos(z*0.2, t/20, 1, 0.05, 0.8)
				s4 := colorutils.Cos(y+0.1, t/40, 1, 0.05, 0.4)


				r:= 0.0
				g:= 0.0
				b:= 0.0

				// number of colors
				nc := 4.0
				// bubble strength (inverse, greater is less bubbles)
				bs := 2.0

				//lightcyan
				r += (0.875*s1) / nc
				g += (1.000*s1) / nc
				b += (1.000*s1) / nc

				//cyan
				r += (0.000*s2) / nc
				g += (1.000*s2) / nc
				b += (1.000*s2) / nc

				// aquamarine
				r += (0.495 *s3) / nc
				g += (1.000 *s3) / nc
				b += (0.831 *s3) / nc

				// teal
				r += (0.000 *s4) / nc
				g += (0.600 *s4) / nc
				b += (0.700 *s4) / nc


				// bluebubbles
                pow1 := math.Pow((z*t), 0.99)
                pow2 := math.Pow((z*t), 0.985)
				b += colorutils.Cos(z, pow1 , 10, 0.0, 0.5) / bs


				// whitebubbles
                sb := colorutils.Cos(z, pow2, 50, 0.0, 0.5) / bs
				r += sb 
				g += sb
				b += sb

                if z<0.0 {
                // aquamarine
				r = ((0.495 *s3) *2) +0.3
				g = ((1.000 *s3) *2) +0.3
				b = ((0.831 *s3) *2) +0.3
                }

				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
