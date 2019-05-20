package opc

import (
	"fmt"
	"github.com/austinfromboston/pixelslinger/colorutils"
	"github.com/austinfromboston/pixelslinger/config"
	"github.com/austinfromboston/pixelslinger/midi"
	"image"
	_ "image/color"
	_ "image/png"
	"math"
	"math/rand"
	"os"
	"time"
)

func handleErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
}

//================================================================================
// IMAGE

type MyColor struct {
	r float64
	g float64
	b float64
}

type MyImage struct {
	xres   int
	yres   int
	pixels [][]*MyColor // 2d array, [x][y]
}

func interpolateSunLocation(timeOfDay float64, sunriseTime float64, locations []float64) []float64 {
	sunArcLocations := []int{2340, 2460, 2580, 2700, 2920, 3040, 3160, 3280, 3400, 3520, 3640, 3760, 3880, 4000, 4120, 4240};
	//sunArcLocations := []int{2459, 2579, 2699, 2819, 2939, 3059, 3179, 3299, 3319, 3480, 3600, 3720, 3840, 3960, 4080, 4200, 4320};
	rawLocation := int(colorutils.EaseRemapAndClamp(timeOfDay, sunriseTime + 0.2, 0.55+sunriseTime, float64(sunArcLocations[0]), float64(sunArcLocations[len(sunArcLocations)-1])))
	for i, l := range sunArcLocations[:len(sunArcLocations)-1] {
		nextLocation := sunArcLocations[i + 1]
		if rawLocation > l && rawLocation <= nextLocation {
			priorX := locations[l*3+0] / 2
			priorY := locations[l*3+1] / 2
			priorZ := locations[l*3+2] / 2
			nextX := locations[nextLocation*3+0] / 2
			nextY := locations[nextLocation*3+1] / 2
			nextZ := locations[nextLocation*3+2] / 2
			interpolatedX := colorutils.EaseRemapAndClamp(float64(rawLocation), float64(l), float64(nextLocation), priorX, nextX)
			interpolatedY := colorutils.EaseRemapAndClamp(float64(rawLocation), float64(l), float64(nextLocation), priorY, nextY)
			interpolatedZ := colorutils.EaseRemapAndClamp(float64(rawLocation), float64(l), float64(nextLocation), priorZ, nextZ)
			return []float64{interpolatedX, interpolatedY, interpolatedZ};
		}

	}
	return []float64{0, 0, 0};
}

// Init the MyImage pixel array, creating MyColor objects
// from the data in the given image (from the built-in image package).
// HSV is computed here also for each pixel.
func (mi *MyImage) populateFromImage(imgFn string) {
	// read and decode image
	file, err := os.Open(imgFn)
	handleErr(err)
	defer file.Close()
	img, _, err := image.Decode(file)
	handleErr(err)

	// copy and convert pixels
	mi.xres = img.Bounds().Max.X
	mi.yres = img.Bounds().Max.Y
	mi.pixels = make([][]*MyColor, mi.xres)
	for x := 0; x < mi.xres; x++ {
		mi.pixels[x] = make([]*MyColor, mi.yres)
		for y := 0; y < mi.yres; y++ {
			r, g, b, _ := img.At(x, y).RGBA()
			c := &MyColor{float64(r) / 256 / 256, float64(g) / 256 / 256, float64(b) / 256 / 256}
			mi.pixels[x][y] = c
		}
	}
}

func (mi *MyImage) String() string {
	return fmt.Sprintf("<image %v x %v>", mi.xres, mi.yres)
}

// given x and y as floats between 0 and 1,
// return r,g,b as floats between 0 and 1
func (mi *MyImage) getInterpolatedColor(x, y float64, wrapMethod string) (r, g, b float64) {

	switch wrapMethod {
	case "tile":
		// keep x and y between 0 and 1
		_, x = math.Modf(x)
		if x < 0 {
			x += 1
		}
		_, y = math.Modf(y)
		if y < 0 {
			y += 1
		}
	case "extend":
		x = colorutils.Clamp(x, 0, 1)
		y = colorutils.Clamp(y, 0, 1)
	case "mirror":
		x = colorutils.PosMod2(x, 2)
		if x > 1 {
			x = 2 - x
		}
		y = colorutils.PosMod2(y, 2)
		if y > 1 {
			y = 2 - y
		}
	}

	// float pixel coords
	xp := x * float64(mi.xres-1) * 0.999999
	yp := y * float64(mi.yres-1) * 0.999999

	// integer pixel coords
	x0 := int(xp)
	x1 := x0 + 1
	y0 := int(yp)
	y1 := y0 + 1

	// subpixel fractional coords for interpolation
	_, xPct := math.Modf(xp)
	_, yPct := math.Modf(yp)

	// retrieve colors from image array
	c00 := mi.pixels[x0][y0]
	c10 := mi.pixels[x1][y0]
	c01 := mi.pixels[x0][y1]
	c11 := mi.pixels[x1][y1]

	// interpolate
	r = (c00.r*(1-xPct)+c10.r*xPct)*(1-yPct) + (c01.r*(1-xPct)+c11.r*xPct)*yPct
	g = (c00.g*(1-xPct)+c10.g*xPct)*(1-yPct) + (c01.g*(1-xPct)+c11.g*xPct)*yPct
	b = (c00.b*(1-xPct)+c10.b*xPct)*(1-yPct) + (c01.b*(1-xPct)+c11.b*xPct)*yPct

	return r, g, b
}

//================================================================================
// PIXEL PATTERN

func MakePatternSunset(locations []float64) ByteThread {

	var (
		IMG_PATH            = "images/sky4_square.png"
		DAY_LENGTH          = 20.0 // seconds
		SUN_SOFT_EDGE       = 0.2
		STAR_BRIGHTNESS_EXP = 1.8 // higher number means fewer bright stars
		STAR_THRESH         = 0.70
		STAR_CONTRAST       = 3.0
		STAR_FADE_EXP       = 3.0 // higher numbers keep stars from showing during sunrise/sunset
	)

	// make persistant random values
	rng := rand.New(rand.NewSource(19))
	randomValues := make([]float64, len(locations)/3)
	for ii := range randomValues {
		randomValues[ii] = math.Pow(rng.Float64(), STAR_BRIGHTNESS_EXP)
	}

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

	// load image
	myImage := &MyImage{}
	myImage.populateFromImage(IMG_PATH)

	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		last_t := 0.0
		t := 0.0
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3

			// time and speed knob bookkeeping
			this_t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8
			speedKnob := float64(midiState.ControllerValues[config.SPEED_KNOB]) / 127.0
			if speedKnob < 0.5 {
				speedKnob = colorutils.RemapAndClamp(speedKnob, 0, 0.4, 0, 1)
			} else {
				speedKnob = colorutils.RemapAndClamp(speedKnob, 0.6, 1, 1, 4)
			}
			if midiState.KeyVolumes[config.SLOWMO_PAD] > 0 {
				speedKnob *= 0.25
			}
			if last_t != 0 {
				t += (this_t - last_t) * speedKnob
			}
			last_t = this_t

			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------

				x := locations[ii*3+0] / 2
				y := locations[ii*3+1] / 2
				z := locations[ii*3+2] / 2
				_ = x
				_ = y
				_ = z

				zp := colorutils.Remap(z, min_coord_z, max_coord_z, 0, 1)

				// time of day, cycles through range 0 to 1.  0 is midnight, 0.5 is noon
				// sunrise at 0.25, sunset at 0.75
				timeOfDay := colorutils.PosMod2(t/DAY_LENGTH, 1)

				// compute sun height in range 0 to 1
				sunHeight := 0.0
				SUNRISE_TIME := 0.2 // range 0 to 0.25
				SUNSET_TIME := 0.75 + SUNRISE_TIME
				switch {
				case timeOfDay < 0.25-SUNRISE_TIME:
					sunHeight = 0
				case timeOfDay < 0.25+SUNRISE_TIME:
					sunHeight = colorutils.EaseRemapAndClamp(timeOfDay, 0.25-SUNRISE_TIME, 0.25+SUNRISE_TIME, 0, 1)
				case timeOfDay < 0.75-SUNRISE_TIME:
					sunHeight = 3
				case timeOfDay < SUNSET_TIME:
					sunHeight = colorutils.EaseRemapAndClamp(timeOfDay, 0.75-SUNRISE_TIME, 0.75+SUNRISE_TIME, 1, 0)
				default:
					sunHeight = 0
				}

				// sky color
				r, g, b := myImage.getInterpolatedColor(timeOfDay+0.5, 1-zp, "tile")

				// stars
				if (timeOfDay < SUNRISE_TIME + 0.1 || timeOfDay > SUNSET_TIME - 0.2) {
					// day/night
					starAmt := math.Pow(1-sunHeight, STAR_FADE_EXP)
					// fade at horizon
					starAmt *= math.Pow(colorutils.RemapAndClamp(zp, 0.35, 0.48, 0, 1), 2)
					// individual stars
					starAmt *= colorutils.ContrastAndClamp(randomValues[ii], STAR_THRESH, STAR_CONTRAST, 0, 1)
					// twinkle
					starAmt *= colorutils.Cos(t, randomValues[ii], 0.3+2*colorutils.PosMod2(randomValues[ii]*7, 1), 0.6, 1)
					r += starAmt
					g += starAmt
					b += starAmt
				}

				// sun circle
				SUN_RADIUS := float64(0.3)
				if timeOfDay > SUNRISE_TIME && timeOfDay < SUNSET_TIME {
					sunAngle := colorutils.RemapAndClamp(timeOfDay, SUNRISE_TIME,  SUNSET_TIME, 0 - (math.Pi * 0.25), math.Pi + (math.Pi * 0.5))
					sunX := math.Cos(sunAngle) * max_coord_x / 4
					//sunY := min_coord_y / 2
					sunZ := math.Sin(sunAngle) * max_coord_z / 4

					distance := math.Sqrt(math.Pow(sunX - x,2 ) + math.Pow(sunZ - z, 2))
					if distance < SUN_RADIUS {
						pct := float64(ii) / 160.0
						pct = pct * 2
						if pct > 1 {
							pct = 2 - pct
						}
						//val := colorutils.Contrast(pct, colorutils.Remap(sunHeight, 0, 1, -SUN_SOFT_EDGE*2, 1+SUN_SOFT_EDGE*2), 1/SUN_SOFT_EDGE)
						//val = colorutils.Clamp(1-val, 0, 1)
						val := colorutils.Remap(distance, 0, SUN_RADIUS, -SUN_SOFT_EDGE*2, 1+SUN_SOFT_EDGE*2)
						//val := colorutils.Remap(distance, 0, SUN_RADIUS,1+SUN_SOFT_EDGE*2, -SUN_SOFT_EDGE*2)
						r = val * 1.13
						g = val * 0.85
						b = val * 0.65
					}
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
