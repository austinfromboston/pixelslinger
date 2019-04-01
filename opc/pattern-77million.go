package opc

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	_ "image/color"
	_ "image/png"
	"io/ioutil"
	"log"
	"fmt"
	"math/rand"
	"strings"
	"time"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/austinfromboston/pixelslinger/config"
	"math"
)
const (
	IMG_DIR = "images/77m"
	NUM_LAYERS = 4
	LAYER_CHANGE_SECS_MAX=120
	)

var (
	layerToChange = 0
	lastTransition = 0.0 //t in seconds
	LAYER_CHANGE_SECS = 15.0
)

func inList(imageName string, currImages [NUM_LAYERS]string) bool {
	for currPos := range currImages {
		currFile := currImages[currPos]
		if currFile == imageName {
			return true
		}
	}
	return false
}

func getNextImage(allImageFiles []string, currImages [NUM_LAYERS]string) string {
	nextFile := ""
	for len(nextFile) == 0 {
		nextTryPos := rand.Intn(len(allImageFiles))
		//fmt.Println("lenallimages nexttrypos", len(allImageFiles), nextTryPos)
		candNextFile := allImageFiles[nextTryPos]
		if !inList(candNextFile, currImages){
			nextFile = candNextFile
		}
	}
	return nextFile
}

func MakePattern77Million(locations []float64) ByteThread {
	rand.Seed(time.Now().UnixNano())
	//rand.Seed(44)
	allImageFiles := []string{}
	files, err := ioutil.ReadDir(IMG_DIR)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".png"){
			allImageFiles = append(allImageFiles, file.Name())
		}
	}
	fmt.Println("Found ", len(allImageFiles), " images in directory, they are ", allImageFiles)
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
	// holds the filenames
	currImageFiles := [NUM_LAYERS]string{}
	// holds the file data
	layerImages := [NUM_LAYERS]MyImage{}
	// initially populate the images
	for li:=0; li<NUM_LAYERS; li++ {
		nextImage := getNextImage(allImageFiles, currImageFiles)
		nextImageFull := IMG_DIR+"/"+nextImage
		layerImages[li].populateFromImage(nextImageFull)
		currImageFiles[li] = nextImage
	}

	currAlphas := [NUM_LAYERS]float64{}
	lastImageSwaps :=  [NUM_LAYERS]float64{}

	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3

			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------
				t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8

				x := locations[ii*3+0]
				y := locations[ii*3+1]
				z := locations[ii*3+2]
				// note in the fan y and z are switched
				y = z

				x_denom := max_coord_x-min_coord_x
				y_denom := max_coord_z-min_coord_z
				x_num := x-min_coord_x
				y_num := y-min_coord_y
				x_norm := x_num/x_denom
				y_norm := y_num/y_denom

				r := 0.0
				g := 0.0
				b := 0.0

				speedKnob := float64(midiState.ControllerValues[config.SPEED_KNOB]) / 127.0
				LAYER_CHANGE_SECS = speedKnob*LAYER_CHANGE_SECS_MAX
				xyColor := colorful.Color{r, g, b}
				for li:=0; li<NUM_LAYERS; li++{
					// the image seems to be flipped on both axes so mutliplying by -1
					offset := float64(li)/float64(NUM_LAYERS)
					//fmt.Println("Offeset for ", li ,"is", offset)


					// want to clip this slightly so that we have transitions in and out but also remains constant for a while
					alphaCos := colorutils.Cos(t, offset, LAYER_CHANGE_SECS, 0, 1.0)
					if alphaCos >= 0.5 {alphaCos=0.5}
					currAlphas[li] = alphaCos
					lr, lg, lb := layerImages[li].getInterpolatedColor(-1*x_norm, -1*y_norm, "tile")
					// TODO blend these layers better
					layerColor := colorful.Color{lr, lg, lb}
					xyColor = xyColor.BlendRgb(layerColor, currAlphas[li]).Clamped()
					//fmt.Println("li, currAlpha, layer color, xycolor", li, currAlphas[li], layerColor, xyColor)
					//r += currAlphas[li]*lr + LAYER_GAIN_BASELINE
					//g += currAlphas[li]*lg + LAYER_GAIN_BASELINE
					//b += currAlphas[li]*lb + LAYER_GAIN_BASELINE

					// switch out the image if the alpha is low and we haven't changed it recently
					lastLayerTransition := t-lastImageSwaps[li]
					if (currAlphas[li] < 0.001)  && ( lastLayerTransition> LAYER_CHANGE_SECS){
							//fmt.Println("last layer trans", lastLayerTransition)
							//fmt.Println(currAlphas)
							nextImage := getNextImage(allImageFiles, currImageFiles)
							fmt.Println("curr images, next image,", currImageFiles, nextImage)
							nextImageFull := IMG_DIR+"/"+nextImage
							layerImages[li].populateFromImage(nextImageFull)
							currImageFiles[li] = nextImage
							lastImageSwaps[li]=t


					}
				}
				if math.Mod(t, 1.0)<0.01{
					fmt.Println("acurrAlphas", currAlphas)
				}

				bytes[ii*3+0] = colorutils.FloatToByte(xyColor.R)
				bytes[ii*3+1] = colorutils.FloatToByte(xyColor.G)
				bytes[ii*3+2] = colorutils.FloatToByte(xyColor.B)

				//--------------------------------------------------------------------------------
			}

			bytesOut <- bytes
		}
	}
}
