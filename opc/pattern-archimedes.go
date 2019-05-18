package opc

// Spatial Stripes
//   Creates spatial sine wave stripes: x in the red channel, y--green, z--blue
//   Also makes a white dot which moves down the strip non-spatially in the order
//   that the LEDs are indexed.

import (
	"github.com/austinfromboston/pixelslinger/colorutils"
	"github.com/austinfromboston/pixelslinger/midi"
	"github.com/austinfromboston/pixelslinger/config"
	"github.com/lucasb-eyer/go-colorful"
	"math"
	"math/rand"
	"time"
    //"fmt"
)

func Spiral(x, y, t, SPIRAL_tightness, SPIRAL_speed, SPIRAL_thickness, SPIRAL_thickness_gradient float64, SPIRAL_rings int) float64 {
	lhs := math.Sqrt((x*x + y*y))
	// theta = r // sqrt()
	tt := t * SPIRAL_speed
	rhs_num := y*math.Cos(tt) + x*math.Sin(tt)
	rhs_den := x*math.Cos(tt) - y*math.Sin(tt)
	rhs_canonical := math.Atan2(rhs_num, rhs_den) * SPIRAL_tightness

	on_amount := 0.0
	for ring := 0; ring <= SPIRAL_rings; ring++ {
		SPIRAL_offset := SPIRAL_tightness * (2 * math.Pi)
		multSign := +1.0
		if SPIRAL_tightness < 0 {
			multSign = -1.0
		}
		rhs := rhs_canonical + (multSign * (float64(ring) * SPIRAL_offset))
		diff := lhs - rhs
		abs_diff := math.Abs(diff)
		if abs_diff < SPIRAL_thickness {
			on_amount += 1 - math.Pow(abs_diff, SPIRAL_thickness_gradient)
		}
	}
	return on_amount
}

func MakePatternArchimedes(locations []float64) ByteThread {
	const(
		SPEED_BASE = 2
	)

	// get bounding box
	n_pixels := len(locations) / 3
	var max_coord_x, max_coord_y, max_coord_z float64
	var min_coord_x, min_coord_y, min_coord_z float64
	for ii := 0; ii < n_pixels; ii++ {
		x := locations[ii*3+0]
		y := locations[ii*3+1]
		z := locations[ii*3+2]
		if ii == 0 || x > max_coord_x {
			max_coord_x = x
		}
		if ii == 0 || y > max_coord_y {
			max_coord_y = y
		}
		if ii == 0 || z > max_coord_z {
			max_coord_z = z
		}
		if ii == 0 || x < min_coord_x {
			min_coord_x = x
		}
		if ii == 0 || y < min_coord_y {
			min_coord_y = y
		}
		if ii == 0 || z < min_coord_z {
			min_coord_z = z
		}
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
				y = z //actually need to target x-z plane

				//noi := math.Abs(rand.Float64() *0.000000000000001)

                dirKnob := (float64(midiState.ControllerValues[config.MORPH_KNOB]) - 64.0)/8.0  

				//noi := 1.0
				//reverse_periodicity := 20.0
				//reverse_dur := 5.0
				//reverse_mod := math.Mod(t, reverse_periodicity)
				// first half
				//if reverse_mod < reverse_dur {
				//	noi = -1 * reverse_mod
				//}
				//second half
				//if reverse_mod > reverse_dur && reverse_mod < 2*reverse_dur {
				//	noi = -1 * (2*reverse_dur - reverse_mod)
				//}
				//if reverse_mod > 2*reverse_dur && reverse_mod < 2*reverse_dur+1 {
				//	noi = math.Pow(reverse_mod-2*reverse_dur, 0.2)
				//}
				
                noi := dirKnob
				//fmt.Println(noi)
				noii := math.Abs(rand.Float64() * 0.0000000000001)
				//fmt.Println(noii)
				//noii = 0.0
			    speedKnob := float64(midiState.ControllerValues[config.SPEED_KNOB]) / 127.0
			    s1 := float64(midiState.ControllerValues[config.DESAT_KNOB]) / 127.0
			    player2Knob := float64(midiState.ControllerValues[config.PlAYER2_KNOB]) / 127.0
				hueKnob := float64(midiState.ControllerValues[config.HUE_KNOB]) / 127.0

				xShift := player2Knob*max_coord_x - (max_coord_x/2)
				yShift := hueKnob*max_coord_z - (max_coord_z/2)

				speed1 := speedKnob * SPEED_BASE * 2    + noii
				speed2 := speedKnob * SPEED_BASE * 4    + noii
				speed3 := speedKnob * SPEED_BASE * 8    + noii
				speed4 := speedKnob * SPEED_BASE * 0.5  + noii
				speed5 := speedKnob * SPEED_BASE * 0.25 + noii

			    spiral1 := Spiral(x+xShift, y-yShift, t, 0.1*noi,  speed1, 0.05, 0.9, 5)
				spiral2 := Spiral(x+xShift, y-yShift, t, -0.1*noi, speed2 , 0.05, 0.5, 5)
				spiral3 := Spiral(x+xShift, y-yShift, t, -0.05*noi,speed3 , 0.1, 0.3, 8)
				spiral4 := Spiral(x+xShift, y-yShift, t, 0.5*noi,  speed4 , 0.5, 0.4, 5)
				spiral5 := Spiral(x+xShift, y-yShift, t, -0.5*noi, speed5 , 0.5, 0.4, 5)

				var (
					//White = colorful.LinearRgb(1, 1, 1)
					//Black = colorful.LinearRgb(0, 0, 0)
					deepRed      = colorful.LinearRgb(1.0, 0.2, 0.2)
					flesh        = colorful.LinearRgb(1.0, 0.35, 0.25)
					peachSherbet = colorful.LinearRgb(1.0, 0.5, 0.3)
					cantaloupe   = colorful.LinearRgb(1.0, 0.65, 0.35)
					dimPeach     = colorful.LinearRgb(1.0, 0.8, 0.4)
					//alts
					are           = colorful.LinearRgb(0.95, 0.54, 0.89)
					driftingPetal = colorful.LinearRgb(0.78, 0.41, 0.74)
					influences    = colorful.LinearRgb(0.77, 0.16, 0.94)
					kleinBlue     = colorful.LinearRgb(0.14, 0.26, 0.64)
					afekCouch     = colorful.LinearRgb(0.17, 0.22, 0.37)
					// alt alts
					coolmint   = colorful.LinearRgb(0.33, 0.67, 0.58)
					freshmint  = colorful.LinearRgb(0.6, 0.89, 0.87)
					spearmint  = colorful.LinearRgb(0.39, 0.61, 0.6)
					guava      = colorful.LinearRgb(0.58, 0.87, 0.73)
					mintCoolee = colorful.LinearRgb(0.27, 0.49, 0.45)
				)

				altColors := [5]colorful.Color{deepRed, flesh, peachSherbet, cantaloupe, dimPeach}
				colors := [5]colorful.Color{are, driftingPetal, influences, kleinBlue, afekCouch}
				colors = [5]colorful.Color{coolmint, freshmint, spearmint, guava, mintCoolee}
				spirals := [5]float64{spiral1, spiral2, spiral3, spiral4, spiral5}

				linear_weight := 0.6
				r := 0.0
				g := 0.0
				b := 0.0
				//s1 := math.Pow(colorutils.Cos(t, 0, 20.0, 0, 1), 0.2)
				//fmt.Println(s1)
				r1 := 1 - s1
				//fmt.Println(altColors[1])
				for cs := 0; cs < len(spirals); cs++ {
					rr := s1*(colors[cs].R) + r1*(altColors[cs].R)
					gg := s1*(colors[cs].G) + r1*(altColors[cs].G)
					bb := s1*(colors[cs].B) + r1*(altColors[cs].B)

					r += linear_weight * rr * spirals[cs]
					g += linear_weight * gg * spirals[cs]
					b += linear_weight * bb * spirals[cs]
				}

				//fmt.Println(r)
				bytes[ii*3+0] = colorutils.FloatToByte(r)
				bytes[ii*3+1] = colorutils.FloatToByte(g)
				bytes[ii*3+2] = colorutils.FloatToByte(b)

				//--------------------------------------------------------------------------------
			}
			bytesOut <- bytes
		}
	}
}
