package opc

import (
	"fmt"
	"gonum.org/v1/gonum/mat"
	//"github.com/austinfromboston/pixelslinger/config"
	//"github.com/austinfromboston/pixelslinger/colorutils"
	"github.com/gordonklaus/portaudio"
	"bytes"
	"encoding/binary"
	//"os"
)

// Sound represents a sound stream implementing the io.Reader interface
// that provides the microphone data.
type Sound struct {
	stream *portaudio.Stream
	data   []int16
}

type ExpFilter struct {
	alphaDecay float64
	alphaRise float64
	value mat.Matrix
}

func (filter *ExpFilter) ProduceValueMask(value mat.Matrix) (mat.Matrix, mat.Matrix) {
	var alpha *mat.Dense
	var alphaInverted *mat.Dense

	alpha.Sub(value, filter.value)
	alphaInverted.Apply( filter.AdjustInverted, alpha)
	alpha.Apply( filter.Adjust, alpha)
	return alpha, alphaInverted

	//r, c := filter.value.Dims()
	//alpha := mat.NewDense(r, c, []float64)
	//alphaInverted := mat.NewDense(r, c, []float64)

	//for i := 0; i < r; i++ {
	//	for j := 0; j < c; j++ {
	//		if alpha.At(i, j) > 0 {
	//			alpha.Set(i, j, filter.alphaRise)
	//			alphaInverted.Set(i, j, 1 - filter.alphaRise)
	//		} else {
	//			alpha.Set(i, j, filter.alphaDecay)
	//			alphaInverted.Set(i, j, 1 - filter.alphaDecay)
	//		}
	//	}
	//}
	//return alpha, alphaInverted
}

func (filter *ExpFilter) Adjust(i int, j int, value float64) float64 {
	if value > 0 {
		return filter.alphaRise
	}
	return filter.alphaDecay
}

func (filter *ExpFilter) AdjustInverted(i int, j int, value float64) float64 {
	if value > 0 {
		return 1 - filter.alphaRise
	}
	return 1 - filter.alphaDecay
}


func (filter *ExpFilter) Update(value mat.Matrix) mat.Matrix {
	var c *mat.Dense
	alpha, alphaInverted := filter.ProduceValueMask(value)
	c.Mul(value, alpha)
	var d *mat.Dense
	d.Mul(filter.value, alphaInverted)
	d.Add(c, d)
	filter.value = d
	return filter.value
}

// Init initializes the Sound's PortAudio stream.
func (s *Sound) Init() {
	inputChannels := 1
	outputChannels := 0
	sampleRate := 16000
	s.data = make([]int16, 1024)

	// initialize the audio recording interface
	err := portaudio.Initialize()
	if err != nil {
		fmt.Errorf("Error initialize audio interface: %s", err)
		return
	}

	// open the sound input stream for the microphone
	stream, err := portaudio.OpenDefaultStream(inputChannels, outputChannels, float64(sampleRate), len(s.data), s.data)
	if err != nil {
		fmt.Errorf("Error open default audio stream: %s", err)
		return
	}

	err = stream.Start()
	if err != nil {
		fmt.Errorf("Error on stream start: %s", err)
		return
	}

	s.stream = stream
}

// Close closes down the Sound's PortAudio connection.
func (s *Sound) Close() {
	s.stream.Close()
	portaudio.Terminate()
}

// Read is the Sound's implementation of the io.Reader interface.
func (s *Sound) Read(p []byte) (int, error) {
	s.stream.Read()

	buf := &bytes.Buffer{}
	for _, v := range s.data {
		binary.Write(buf, binary.LittleEndian, v)
	}

	copy(p, buf.Bytes())
	return len(p), nil
}

func audioMap(vs []byte, f func(byte) float64) []float64 {
	vsm := make([]float64, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

// Number of frequency bins to use when transforming audio to frequency domain
const FftBinCount = 24
// Sampling frequency of the microphone in Hz
const MicRate = 48000
// Desired refresh rate of the visualization (frames per second)
const Fps = 60
// Number of past audio frames to include in the rolling window
const RollingHistoryFrames = 2

func initialFilterValue(decay float64, rise float64) ExpFilter {
	fftPlotFilterSource := mat.NewDense(1, FftBinCount, make([]float64, 1, FftBinCount))
	fftPlotFilterSource.Apply(func(_ int, _ int, _ float64) float64 {
		return 1e-1;
	}, fftPlotFilterSource)
	return ExpFilter{decay, rise, fftPlotFilterSource}
}

func AbsMax(v []float64) []float64 {
	return make([]float64, 2,20)
}

//func MakePatternListener(locations []float64) func {
//
//
//	// get bounding box
//	n_pixels := len(locations) / 3
//	var max_coord_x, max_coord_y, max_coord_z float64
//	var min_coord_x, min_coord_y, min_coord_z float64
//	for ii := 0; ii < n_pixels; ii++ {
//		x := locations[ii*3+0]
//		y := locations[ii*3+1]
//		z := locations[ii*3+2]
//		if ii == 0 || x > max_coord_x { max_coord_x = x }
//		if ii == 0 || y > max_coord_y { max_coord_y = y }
//		if ii == 0 || z > max_coord_z { max_coord_z = z }
//		if ii == 0 || x < min_coord_x { min_coord_x = x }
//		if ii == 0 || y < min_coord_y { min_coord_y = y }
//		if ii == 0 || z < min_coord_z { min_coord_z = z }
//	}
//
//
//	rand.Seed(time.Now().UnixNano())
//	fftPlotFilter := initialFilterValue(0.5, 0.99)
//	melGain := initialFilterValue(0.01, 0.99)
//	melSmoothing := initialFilterValue(0.5, 0.99)
//	volume := initialFilterValue(0.02, 0.02)
//	samplesPerFrame := int(MicRate/Fps)
//	fftWindow := window.Hamming(samplesPerFrame * RollingHistoryFrames)
//	prevFpsUpdate := time.Now()
//	rawYRoll := make([]byte, samplesPerFrame * RollingHistoryFrames, samplesPerFrame * RollingHistoryFrames)
//	rand.Read(rawYRoll)
//	yRoll := make([][]float64, samplesPerFrame * RollingHistoryFrames, samplesPerFrame * RollingHistoryFrames)
//	yRoll[0] = audioMap(rawYRoll, func(b byte) float64 {
//		return float64(b);
//	})
//
//
//	// setup microphone listener
//	// open the mic
//	mic := &Sound{}
//	mic.Init()
//	defer mic.Close()
//
//	var previousRmsUpdate *[]float64
//	var previousExp *[]float64
//	var rawAudioSamples []byte;
//
//	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
//		last_t := 0.0
//		t := 0.0
//
//		_, readError := mic.Read(rawAudioSamples)
//
//		if readError != nil {
//			print("error reading from mic")
//			//bytesOut <- bytesIn
//			return
//		}
//		y := audioMap(rawAudioSamples, func(b byte) float64 {
//			return float64(b) * math.Pow(2.0, 15)
//		})
//		yRollLen := len(yRoll)
//		for i := yRollLen - 1; i <= 0; i-- {
//			yRoll[i] = yRoll[i+1]
//		}
//		//yRoll[:(yRollLen - 1)] = yRoll[1:]
//		yRoll[yRollLen] = y
//
//		yData := append(yRoll[0], yRoll[1]...)
//
//		vol := AbsMax(yData)
//		for bytes := range bytesIn {
//			n_pixels := len(bytes) / 3
//
//			for ii := 0; ii < n_pixels; ii++ {
//				x := locations[ii*3+0] / 2
//				y := locations[ii*3+1] / 2
//				z := locations[ii*3+2] / 2
//				_ = x
//				_ = y
//				_ = z
//			}
//
//
//
//			bytesOut <- bytes
//		}
//	}
//}

