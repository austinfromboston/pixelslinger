package opc

// Colorbox
// Every pixel's r,g,b  is linearly related to its x,y,z.

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	"math"
	"time"
)

func MakePatternSpatialColorBox(locations []float64) ByteThread {
	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
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
				r := (x - min_coord_x) / max_coord_x
				g := (y - min_coord_y) / max_coord_y
				b := (z - min_coord_z) / max_coord_z
				// r, g, b = colorutils.ContrastRgb(r, g, b, 0.5, 2)

				// make a moving white dot showing the order of the pixels in the layout file
				spark_ii := colorutils.PosMod2(t*80, float64(n_pixels))
				spark_rad := float64(8)
				spark_val := math.Max(0, (spark_rad-colorutils.ModDist2(float64(ii), float64(spark_ii), float64(n_pixels)))/spark_rad)
				spark_val = math.Min(1, spark_val*2)
				r += spark_val
				g += spark_val
				b += spark_val

				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
