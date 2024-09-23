package main

// TODO: figure out how to handle varying numbers of pixels
// when we're getting pixels via our OPC server source

import (
	"fmt"
	"github.com/austinfromboston/pixelslinger/config"
	"github.com/austinfromboston/pixelslinger/midi"
	"github.com/austinfromboston/pixelslinger/opc"
	oscpixels "github.com/austinfromboston/pixelslinger/osc"
	"github.com/austinfromboston/pixelslinger/potty"
	"github.com/droundy/goopt"
	"github.com/pkg/profile"
	"github.com/rakyll/portmidi"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"
)

//const ONBOARD_LED_HEARTBEAT = 0
//const ONBOARD_LED_MIDI = 1

const SPI_MAGIC_WORD = "spi"
const ARTNET_MAGIC_WORD = "artnet"
const PRINT_MAGIC_WORD = "print"
const DEVNULL_MAGIC_WORD = "/dev/null"
const LOCALHOST = "localhost"
const SPI_FN = "/dev/spidev1.0"

func init() {
	runtime.GOMAXPROCS(2)
}

// these are pointers to the actual values from the command line parser
var LAYOUT_FN = goopt.String([]string{"-l", "--layout"}, "...", "layout file (required)")
var SOURCE = goopt.String([]string{"-s", "--source"}, "spatial-stripes", "pixel source (either a pattern name or "+LOCALHOST+"[:port])")
var DEST = goopt.String([]string{"-d", "--dest"}, "localhost", "destination (one of "+PRINT_MAGIC_WORD+", "+SPI_MAGIC_WORD+", "+DEVNULL_MAGIC_WORD+", "+ARTNET_MAGIC_WORD+"or hostname[:port])")
var DEST2 = goopt.String([]string{"-D", "--dest2"}, "", "secondary destination ("+ARTNET_MAGIC_WORD+")")
var FPS = goopt.Int([]string{"-f", "--fps"}, 40, "max frames per second")
var SECONDS = goopt.Int([]string{"-n", "--seconds"}, 0, "quit after this many seconds")
var ONCE = goopt.Flag([]string{"-o", "--once"}, []string{}, "quit after one frame", "")
var MIDI_SOURCE = goopt.String([]string{"-M", "--midi-source"}, "/dev/midi1", "midi device buffer (linux) or 'socket' (to use socket listener)")
var NET_PROTO = goopt.String([]string{"-P", "--protocol"}, "opc", "opc (openpixelcontrol) or 'artnet'")

// Parse the command line flags.  If invalid, show help and quit.
// Add default ports if needed.
// Read the layout file.
// Return the number of pixels in the layout, the source and dest thread methods.
func parseFlags() (nPixels int, sourceThread, effectThread, pottyEffectThread, destThread opc.ByteThread) {

	// get sorted pattern names
	patternNames := make([]string, len(opc.PATTERN_REGISTRY))
	ii := 0
	for k, _ := range opc.PATTERN_REGISTRY {
		patternNames[ii] = k
		ii++
	}
	sort.Strings(patternNames)

	goopt.Summary = "Available source patterns:\n"
	for _, patternName := range patternNames {
		goopt.Summary += "          " + patternName + "\n"
	}
	goopt.Parse(nil)

	// layout is required
	if *LAYOUT_FN == "..." {
		fmt.Println(goopt.Usage())
		fmt.Println("--------------------------------------------------------------------------------/")
		os.Exit(1)
	}

	// read locations
	locations := opc.ReadLocations(*LAYOUT_FN)
	nPixels = len(locations) / 3

	// choose source thread method
	if strings.Contains(*SOURCE, LOCALHOST) {
		// source is localhost, so we will start an OPC server.
		// add default port if needed
		if !strings.Contains(*SOURCE, ":") {
			*SOURCE += ":7890"
		}
		sourceThread = opc.MakeOpcServerThread(*SOURCE)
	} else if (*SOURCE)[0] == ':' {
		// source is ":4908"
		*SOURCE = "localhost" + *SOURCE
		sourceThread = opc.MakeOpcServerThread(*SOURCE)
	} else {
		// source is a pattern name
		sourceThreadMaker, ok := opc.PATTERN_REGISTRY[*SOURCE]
		if !ok {
			fmt.Printf("Error: unknown source or pattern \"%s\"\n", *SOURCE)
			fmt.Println("--------------------------------------------------------------------------------/")
			os.Exit(1)
		}
		sourceThread = sourceThreadMaker(locations)
	}

	// choose effect thread method
	effectThread = opc.MakeEffectFader(locations)
	pottyEffectThread = potty.MakeEffectFaderPattern(locations)

	// choose dest thread method
	switch *DEST {
	case DEVNULL_MAGIC_WORD:
		destThread = opc.MakeSendToDevNullThread()
	case PRINT_MAGIC_WORD:
		destThread = opc.MakeSendToScreenThread()
	case SPI_MAGIC_WORD:
		destThread = opc.MakeSendToLPD8806Thread(SPI_FN)

	default:
		if *NET_PROTO == ARTNET_MAGIC_WORD {
			if *DEST2 != "" {
				destThread = opc.MakeSendToArtnetThreadMultiple(*DEST, *DEST2)
			} else {
				destThread = opc.MakeSendToArtnetThread(*DEST)
			}
		} else {
			// add default port if needed
			if !strings.Contains(*DEST, ":") {
				*DEST += ":7890"
			}
			destThread = opc.MakeSendToOpcThread(*DEST)
		}
	}

	return // returns nPixels, sourceThread, destThread
}

// Launch the sourceThread and destThread methods and coordinate the transfer of bytes from one to the other.
// Run until timeToRun seconds have passed and return.  If timeToRun is 0, run forever.
// Turn on the CPU profiler if timeToRun seconds > 0.
// Limit the framerate to a max of fps unless fps is 0.
func mainLoop(nPixels int, sourceThread, effectThread, pottyEffectThread, destThread opc.ByteThread, fps float64, timeToRun float64) {
	if timeToRun > 0 {
		fmt.Printf("[mainLoop] Running for %f seconds with profiling turned on, pixels and network\n", timeToRun)
		defer profile.Start(profile.CPUProfile).Stop()
	} else {
		fmt.Println("[mainLoop] Running forever")
	}

	// prepare the byte slices and channels that connect the source and dest threads
	fillingSlice := make([]byte, nPixels*3)
	sendingSlice := make([]byte, nPixels*3)

	bytesToFillChan := make(chan []byte, 0)
	toEffectChan := make(chan []byte, 0)
	toPottyEffectChan := make(chan []byte, 0)
	bytesFilledChan := make(chan []byte, 0)
	bytesToSendChan := make(chan []byte, 0)
	bytesSentChan := make(chan []byte, 0)

	// set up midi
	midiPath := ""
	if *MIDI_SOURCE == "socket" {
		midiPath = "socket"
	} else if _, err := os.Stat(*MIDI_SOURCE); err == nil {

		// path/to/whatever exists
		midiPath = *MIDI_SOURCE
	} else if os.IsNotExist(err) {
		portmidi.Initialize()
		fmt.Println("device count", portmidi.Info(portmidi.DefaultInputDeviceID()))
		defer portmidi.Terminate()
		//path/to/whatever does *not* exist
		midiPath = "/dev/midi2"
	}
	fmt.Println("midiPath is", midiPath)
	midiMessageChan := midi.GetMidiMessageStream(midiPath) // this launches the midi thread
	oscMessageChan := oscpixels.GetOSCMessageStream("localhost:8765")

	rhythmState := oscpixels.RhythmState{}
	midiState := midi.MidiState{Rhythm: rhythmState}

	// set initial values for controller knobs
	//  (because the midi hardware only sends us values when the knobs move)
	for knob, defaultVal := range config.DEFAULT_KNOB_VALUES {
		midiState.ControllerValues[knob] = defaultVal
	}
	fmt.Println(midiState)

	// launch the threads
	go sourceThread(bytesToFillChan, toEffectChan, &midiState)
	go effectThread(toEffectChan, toPottyEffectChan, &midiState)
	go pottyEffectThread(toPottyEffectChan, bytesFilledChan, &midiState)
	go destThread(bytesToSendChan, bytesSentChan, &midiState)

	// main loop
	frame_budget_ms := 1000.0 / fps
	startTime := float64(time.Now().UnixNano()) / 1.0e9
	lastPrintTime := startTime
	frameStartTime := startTime
	frameEndTime := startTime
	framesSinceLastPrint := 0
	firstIteration := true
	flipper := 0
	//beaglebone.SetOnboardLED(0, 1)
	for {
		// if we have any frame budget left from last time around, sleep to control the framerate
		if fps > 0 {
			frameEndTime = float64(time.Now().UnixNano()) / 1.0e9
			timeRemaining := float64(frame_budget_ms)/1000 - (frameEndTime - frameStartTime)
			if timeRemaining > 0 {
				time.Sleep(time.Duration(timeRemaining*1000*1000) * time.Microsecond)
			}
		}

		// fps reporting and bookkeeping
		// print framerate occasionally
		frameStartTime = float64(time.Now().UnixNano()) / 1.0e9
		framesSinceLastPrint += 1
		secondsSinceLastPrint := 5
		if frameStartTime > (lastPrintTime + float64(secondsSinceLastPrint)) {
			lastPrintTime = frameStartTime
			fmt.Printf("[%s] [mainLoop] %f ms/frame (%d fps)\n", time.Now().Format("03:04:05"), (1000.0*float64(secondsSinceLastPrint))/float64(framesSinceLastPrint), framesSinceLastPrint/secondsSinceLastPrint)
			framesSinceLastPrint = 0
			// toggle LED
			//beaglebone.SetOnboardLED(ONBOARD_LED_HEARTBEAT, flipper)
			flipper = 1 - flipper
		}

		// if profiling, quit after a while
		if timeToRun > 0 && frameStartTime > startTime+timeToRun {
			return
		}

		// get midi
		midiState.UpdateStateFromChannel(midiMessageChan)
		midiState.Rhythm.UpdateStateFromChannel(oscMessageChan)
		//if len(midiState.RecentMidiMessages) > 0 {
		//	beaglebone.SetOnboardLED(ONBOARD_LED_MIDI, 1)
		//} else {
		//	beaglebone.SetOnboardLED(ONBOARD_LED_MIDI, 0)
		//}

		// start the threads filling and sending slices in parallel.
		// if this is the first time through the loop we have to skip
		//  the sending stage or we'll send out a whole bunch of zeros.
		bytesToFillChan <- fillingSlice
		if !firstIteration {
			bytesToSendChan <- sendingSlice
		}

		// if only sending one frame, let's just get it all over with now
		//  or we'd have to compute two frames worth of pixels because of
		//  the double buffering effect of the two parallel threads
		if *ONCE {
			// get filled bytes and send them
			bytesToSendChan <- <-bytesFilledChan
			// wait for sending to complete
			<-bytesSentChan
			fmt.Println("[mainLoop] just running once.  quitting now.")
			return
		}

		// wait until both filling and sending threads are done
		<-bytesFilledChan
		if !firstIteration {
			<-bytesSentChan
		}

		// swap the slices
		sendingSlice, fillingSlice = fillingSlice, sendingSlice

		firstIteration = false
	}
}

func main() {
	fmt.Println("--------------------------------------------------------------------------------\\")
	defer fmt.Println("--------------------------------------------------------------------------------/")

	nPixels, sourceThread, effectThread, pottyEffectThread, destThread := parseFlags()
	mainLoop(nPixels, sourceThread, effectThread, pottyEffectThread, destThread, float64(*FPS), float64(*SECONDS))
}
