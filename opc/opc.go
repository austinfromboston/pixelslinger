package opc

import (
	"bufio"
	"fmt"
	"github.com/davecheney/profile"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const FILLING_LED = 0
const SENDING_LED = 1

const SPI_MAGIC_WORD = "SPI"

const SPI_FN = "/dev/spidev1.0"

const CONNECTION_TRIES = 1

// times in ms
const WAIT_TO_RETRY = 1000
const WAIT_BETWEEN_RETRIES = 1

func helpAndQuit() {
	fmt.Println("--------------------------------------------------------------------------------\\")
	fmt.Println("")
	fmt.Println("Usage:  program-name  <layout.json>  [ip:port  [fps  [seconds-to-run]]]")
	fmt.Println("")
	fmt.Println("    layout.json       A layout json file")
	fmt.Println("    ip:port           Server to connect to.  Port is optional and defaults to 7890.")
	fmt.Println("                        You can use a hostname instead of an ip address.")
	fmt.Println("                        Or set it to \"SPI\" to send data out to the SPI port.")
	fmt.Println("    fps               Desired frames per second.  User 0 for no limit.")
	fmt.Println("    seconds-to-run    Quit after this many seconds.  Use 0 for forever.")
	fmt.Println("                        If nonzero, the profiler will be turned on.")
	fmt.Println("                        Use negative numbers to benchmark your pixelThread function.")
	fmt.Println("")
	fmt.Println("--------------------------------------------------------------------------------/")
	os.Exit(0)
}

func ParseFlags() (layoutPath, ipPort string, fps float64, timeToRun float64) {
	layoutPath = "layouts/freespace.json"
	ipPort = "127.0.0.1:7890"
	fps = 40
	timeToRun = 0
	var err error

	if len(os.Args) >= 2 {
		if os.Args[1] == "-h" || os.Args[1] == "--help" {
			helpAndQuit()
		}
		layoutPath = os.Args[1]
	}
	if len(os.Args) >= 3 {
		ipPort = os.Args[2]
		if ipPort != SPI_MAGIC_WORD && !strings.Contains(ipPort, ":") {
			ipPort += ":7890"
		}
	}
	if len(os.Args) >= 4 {
		fps, err = strconv.ParseFloat(os.Args[3], 64)
		if err != nil {
			helpAndQuit()
		}
	}
	if len(os.Args) >= 5 {
		timeToRun, err = strconv.ParseFloat(os.Args[4], 64)
		if err != nil {
			helpAndQuit()
		}
	}
	if len(os.Args) >= 6 || len(os.Args) <= 1 {
		helpAndQuit()
	}
	return
}

// Set one of the on-board LEDs on the Beaglebone.
//    ledNum: between 0 and 3 inclusive
//    val: 0 or 1.
func SetOnboardLED(ledNum int, val int) {
    return
	ledFn := fmt.Sprintf("/sys/class/leds/beaglebone:green:usr%d/brightness", ledNum)
	fmt.Println(ledFn)

	// open output file
	ledFile, err := os.Create(ledFn)
	if err != nil {
		panic(err)
	}
	// close ledFile on exit and check for its returned error
	defer func() {
		if err := ledFile.Close(); err != nil {
			panic(err)
		}
	}()

	if _, err := ledFile.WriteString(strconv.Itoa(val)); err != nil {
		panic(err)
	}
}

// Read locations from JSON file into a slice of floats
func ReadLocations(fn string) []float64 {
	locations := make([]float64, 0)
	var file *os.File
	var err error
	if file, err = os.Open(fn); err != nil {
		panic(fmt.Sprintf("could not open layout file: %s", fn))
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '[' || line[0] == ']' {
			continue
		}
		line = strings.Split(line, "[")[1]
		line = strings.Split(line, "]")[0]
		coordStrings := strings.Split(line, ", ")
		var x, y, z float64
		x, err = strconv.ParseFloat(coordStrings[0], 64)
		y, err = strconv.ParseFloat(coordStrings[1], 64)
		z, err = strconv.ParseFloat(coordStrings[2], 64)
		locations = append(locations, x, y, z)
	}
	fmt.Printf("[opc.ReadLocations] Read %v pixel locations from %s\n", len(locations), fn)
	return locations
}

// Try to connect a couple of times
// If we fail after several tries, return nil
func getConnection(ipPort string) net.Conn {
	fmt.Printf("[opc.getConnection] connecting to %v...\n", ipPort)
	triesLeft := CONNECTION_TRIES
	var conn net.Conn
	var err error
	for {
		conn, err = net.Dial("tcp", ipPort)
		if err == nil {
			break
		}
		fmt.Println("[opc.getConnection", triesLeft, err)
		time.Sleep(WAIT_BETWEEN_RETRIES * time.Millisecond)
		triesLeft -= 1
		if triesLeft == 0 {
			return nil
		}
	}
	fmt.Println("[opc.getConnection]    connected")
	return conn
}

func SendingDummyThread(sendThisSlice chan []byte, sliceIsSent chan int) {
	flipper := 0
    for _ = range sendThisSlice {
		// toggle onboard LED
		SetOnboardLED(SENDING_LED, flipper)
		flipper = 1 - flipper

        sliceIsSent <- 1
    }
}

// Recieve byte slices over the pixelsToSend channel.
// When we get one, write it to the SPI file descriptor and toggle one
//  of the Beaglebone's onboard LEDs.
// After sending the frame, send 1 over the sliceIsSent channel.
// The byte slice should hold values from 0 to 255 in [r g b  r g b  r g b  ... ] order.
func SendingToLPD8806Thread(sendThisSlice chan []byte, sliceIsSent chan int) {
    fmt.Println("[opc.SendingToLPD8806Thread] booting")

	// open output file and keep the file descriptor around
	spiFile, err := os.Create(SPI_FN)
	if err != nil {
        fmt.Println("[opc.SendingToLPD8806Thread] Error opening SPI file:")
        fmt.Println(err)
        os.Exit(1)
	}
	// close spiFile on exit and check for its returned error
	defer func() {
		if err := spiFile.Close(); err != nil {
			panic(err)
		}
	}()

	flipper := 0
	// as we get byte slices over the channel...
	for values := range sendThisSlice {
		fmt.Println("[opc.SendingToLPD8806Thread] starting to send", len(values), "values")

		// toggle onboard LED
		SetOnboardLED(SENDING_LED, flipper)
		flipper = 1 - flipper

		// build a new slice of bytes in the format the LED strand wants
        // TODO: avoid allocating these bytes over and over
		bytes := make([]byte, 0)

		// leading zeros to begin a new frame of values
		numZeroes := (len(values) / 32) + 2
		for ii := 0; ii < numZeroes*5; ii++ {
			bytes = append(bytes, 0)
		}

		// values
		for _, v := range values {
			// high bit must be always on, remaining seven bits are data
			v2 := 128 | (v >> 1)
			bytes = append(bytes, v2)
		}

		// final zero to latch the last pixel
		bytes = append(bytes, 0)

        // actually send bytes over the wire
        if _, err := spiFile.Write(bytes); err != nil {
            panic(err)
        }

		sliceIsSent <- 1
	}
}

// Initiate and Maintain a connection to ipPort.
// When a slice comes in through sendThisSlice, send it with an OPC header.
// Loop forever.
func SendingToOpcThread(sendThisSlice chan []byte, sliceIsSent chan int, ipPort string) {
    fmt.Println("[opc.SendingToOpcThread] booting")

	var conn net.Conn
	var err error

    flipper := 0
	for values := range sendThisSlice {
		// toggle onboard LED
		SetOnboardLED(SENDING_LED, flipper)
		flipper = 1 - flipper

		// if the connection has gone bad, make a new one
		if conn == nil {
			conn = getConnection(ipPort)
		}
		// if that didn't work, wait a second and restart the loop
		if conn == nil {
			sliceIsSent <- 1
            fmt.Println("[opc.SendingToOpcThread] waiting to retry")
			time.Sleep(WAIT_TO_RETRY * time.Millisecond)
			continue
		}

		// ok, at this point the connection is good

		// make and send OPC header
		channel := byte(0)
		command := byte(0)
		lenLowByte := byte(len(values) % 256)
		lenHighByte := byte(len(values) / 256)
		header := []byte{channel, command, lenHighByte, lenLowByte}
		_, err = conn.Write(header)
		if err != nil {
			// net error -- set conn to nil so we can try to make a new one
			fmt.Println("[opc.SendingToOpcThread]", err)
			conn = nil
			sliceIsSent <- 1
			continue
		}

		// send actual pixel values
		_, err = conn.Write(values)
		if err != nil {
			// net error -- set conn to nil so we can try to make a new one
			fmt.Println("[opc.SendingToOpcThread]", err)
			conn = nil
			sliceIsSent <- 1
			continue
		}
		sliceIsSent <- 1
	}
}

// Launch the pixelThread and suck pixels out of it
// Also launch the SendingToOpcThread and feed the pixels to it
// Run until timeToRun seconds have passed
// Set timeToRun to 0 to run forever
// Set timeToRun to a negative to benchmark your pixelThread function by itself.
// Set fps to the number of frames per second you want, or 0 for unlimited.
func MainLoop(pixelThread func(chan []byte, chan int, []float64), layoutPath, ipPort string, fps float64, timeToRun float64) {
	fmt.Println("--------------------------------------------------------------------------------\\")
	defer fmt.Println("--------------------------------------------------------------------------------/")

	if timeToRun != 0 {
		if timeToRun > 0 {
			fmt.Printf("[opc.MainLoop] Running for %f seconds with profiling turned on, pixels and network\n", timeToRun)
		} else if timeToRun < 0 {
			fmt.Printf("[opc.MainLoop] Running for %f seconds with profiling turned on, pixels only\n", -timeToRun)
		}
		defer profile.Start(profile.CPUProfile).Stop()
	} else {
		fmt.Println("[opc.MainLoop] Running forever")
	}

	frame_budget_ms := 1000.0 / fps

	// load location and build initial slices
	locations := ReadLocations(layoutPath)
	n_pixels := len(locations) / 3
	fillingSlice := make([]byte, n_pixels*3)
	sendingSlice := make([]byte, n_pixels*3)

	fillThisSlice := make(chan []byte, 0)
	sliceIsFilled := make(chan int, 0)
	sendThisSlice := make(chan []byte, 0)
	sliceIsSent := make(chan int, 0)

	// start threads
    if ipPort == SPI_MAGIC_WORD {
        go SendingToLPD8806Thread(sendThisSlice, sliceIsSent)
    } else {
        go SendingToOpcThread(sendThisSlice, sliceIsSent, ipPort)
    }
    go pixelThread(fillThisSlice, sliceIsFilled, locations)

	// main loop
	startTime := float64(time.Now().UnixNano()) / 1.0e9
	lastPrintTime := startTime
	framesSinceLastPrint := int(0)
	var t, t2 float64
	for {
		// fps reporting and bookkeeping
		t = float64(time.Now().UnixNano()) / 1.0e9
		framesSinceLastPrint += 1
		if t > lastPrintTime+1 {
			lastPrintTime = t
			fmt.Printf("[opc.MainLoop] %f ms (%d fps)\n", 1000.0/float64(framesSinceLastPrint), framesSinceLastPrint)
			framesSinceLastPrint = 0
		}

		// quit after a while, for profiling purposes
		if timeToRun != 0 && t > startTime+math.Abs(timeToRun) {
			return
		}

		// start filling and sendingSlice
		fillThisSlice <- fillingSlice
		if timeToRun >= 0 {
			sendThisSlice <- sendingSlice
		}

		// wait until both are ready
		<-sliceIsFilled
		if timeToRun >= 0 {
			<-sliceIsSent
		}

		// control framerate
		if timeToRun >= 0 && fps > 0 {
			// sleep if we still have frame budget left
			t2 = float64(time.Now().UnixNano()) / 1.0e9
			timeRemaining := float64(frame_budget_ms)/1000 - (t2 - t)
			if timeRemaining > 0 {
				time.Sleep(time.Duration(timeRemaining*1000*1000) * time.Microsecond)
			}
		}

		// swap
		fillingSlice, sendingSlice = sendingSlice, fillingSlice
	}
}
