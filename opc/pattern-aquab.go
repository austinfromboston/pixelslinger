package opc

// Spatial Stripes
//   Creates spatial sine wave stripes: x in the red channel, y--green, z--blue
//   Also makes a white dot which moves down the strip non-spatially in the order
//   that the LEDs are indexed.

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	"math"
	"math/rand"
	"time"
	//"fmt"
)


func MakePatternAquaB(locations []float64) ByteThread {

	var (
		// number of colour, weight to give each color component
		nc = 3.5
		)

	// get bounding box
	n_pixels := len(locations) / 3
	var max_coord_x, max_coord_y, max_coord_z float64
	var min_coord_x, min_coord_y, min_coord_z float64
	for ii := 0; ii < n_pixels; ii++ {
		x := locations[ii*3+0]
		y := locations[ii*3+1]
		z := locations[ii*3+2]
				if ii == 0 || x > max_coord_x { max_coord_x = x }
				if ii == 0 || y > max_coord_y { max_coord_y = y }
				if ii == 0 || z > max_coord_z { max_coord_z = z }
				if ii == 0 || x < min_coord_x { min_coord_x = x }
				if ii == 0 || y < min_coord_y { min_coord_y = y }
				if ii == 0 || z < min_coord_z { min_coord_z = z }
	}

	// make persistant random values
	rng := rand.New(rand.NewSource(9))
	randomValues := make([]float64, len(locations)/3)
	for ii := range randomValues {
		randomValues[ii] = math.Pow(rng.Float64(), 30.0)
		// fmt.Printf("%v\n",randomValues[ii])
	}


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

				s1 := colorutils.Cos((z*0.3)*(y*0.3)*(x*0.3)    , t/8, 1, 0.1, 0.7)
				s3 := colorutils.Cos(x*0.2 + y*0.8, t/4, 1, 0.1, 0.6)
				s2 := colorutils.Cos(z*0.2, t/10, 1, 0.0, 1.0)
				s4 := 0.3*s1 + 0.7*s3

				r:= 0.1
				g:= 0.1
				b:= 0.1

				// number of colors


				//lightcyan
				r += (0.875*s1) / nc
				g += (1.000*s1) / nc
				b += (1.000*s1) / nc

				//cyan
				r += (0.000*s2) / nc
				g += (1.000*s2) / nc
				b += (1.000*s2) / nc

				// aquamarine
				r += (0.000 *s3) /nc
				g += (0.350 *s3) /nc
				b += (0.790 *s3) /nc

				// teal
				r += (0.790 *s4) /nc
				g += (1.000 *s4) /nc
				b += (0.890 *s4) /nc


				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
