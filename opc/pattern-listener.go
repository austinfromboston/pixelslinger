package opc

import (
	"fmt"
	"gonum.org/v1/gonum/mat"
	"github.com/austinfromboston/pixelslinger/midi"
	"time"
	"github.com/austinfromboston/pixelslinger/config"
	"github.com/austinfromboston/pixelslinger/colorutils"
	"github.com/gordonklaus/portaudio"
	"bytes"
	"encoding/binary"
)

func handleErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
}

// Sound represents a sound stream implementing the io.Reader interface
// that provides the microphone data.
type Sound struct {
	stream *portaudio.Stream
	data   []int16
}

type ExpFilter struct {
	alphaDecay float32
	alphaRise float32
	value float32
}

func (filter *ExpFilter) Update(value float32) float32 {
	alpha := filter.alphaDecay
	if value > filter.value {
		alpha = filter.alphaRise
	}
	filter.value = alpha * filter.value + (1.0 - alpha) * filter.value
	return filter.value
}

type ExpFilterMatrix struct {
	alphaDecay mat64
	alphaRise float32
	value float32
}

func (filter *ExpFilter) Update(value []float32) float32 {
	alpha := filter.alphaDecay
	if value > filter.value {
		alpha = filter.alphaRise
	}
	filter.value = alpha * filter.value + (1.0 - alpha) * filter.value
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

func MakePatternListener(locations []float64) ByteThread {

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

			// setup microphone listener
			// open the mic
			mic := &Sound{}
			mic.Init()
			defer mic.Close()

			for ii := 0; ii < n_pixels; ii++ {

			}
		}
	}
}