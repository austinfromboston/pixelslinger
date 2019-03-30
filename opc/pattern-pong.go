package opc

// Polar Pong

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	"math"
	"time"
	"fmt"
	"github.com/longears/pixelslinger/config"
)

// Configs
// physics update interval (ms)
const (physicsUpdateIntervalMiliSeconds = 25
	physicsUpdateIntervalSeconds = physicsUpdateIntervalMiliSeconds / 1e3
	ballDisplayRadMin = 0.08
	paddleLength = 0.5
	halfPaddleLength = paddleLength / 2
	leftTheta=math.Pi / 20.0
	rightTheta=math.Pi * 19.0/20.0
	paddleActiveLightUpSecs = 1
	paddleControlSensitivity = 0.0001)


type aState struct {
	ballPosX  float64
	ballPosY  float64
	ballVert  float64
	ballHoriz float64
	leftPaddlePos float64 // the percentage it is from the bottom to top
	rightPaddlePos float64
	leftPaddleHitTime float64
	rightPaddleHitTime float64
	gamePhase string // are we in demo-mode, inbetween "games", in a "game" at the end of a set.
	lastPreGrameStart float64
	lastGameStart float64
	lastMatchStart float64
}

func distanceToPaddle(x float64, y float64, paddlePos float64, paddleAngle float64) (float64, float64, float64){
	paddleX, paddleY := paddleXYfromPos(paddlePos, paddleAngle)
	distToPaddle := math.Sqrt(math.Pow(x-paddleX, 2) + math.Pow(y-paddleY, 2))
	return paddleX, paddleY, distToPaddle
}

func paddleXYfromPos(paddlePos float64, paddleAngle float64) (float64, float64) {
	// the two comes from the fact that the radius of the fan is 2
	return 2 * paddlePos * math.Cos(paddleAngle), 2 * paddlePos * math.Sin(paddleAngle)
}

func paddleHitCheckRoutine(state aState, t float64, r float64, leftRight string) aState {
	paddlePos := state.leftPaddlePos
	sideTheta := leftTheta
	paddleHitTime := &state.leftPaddleHitTime
	if leftRight == "right"{
		paddlePos = state.rightPaddlePos
		sideTheta = rightTheta
		paddleHitTime = &state.rightPaddleHitTime
	}
	_, paddleY, dtp := distanceToPaddle(state.ballPosX, state.ballPosY, paddlePos, sideTheta)
	if dtp < halfPaddleLength{
		//fmt.Println("hit paddle")
		reflectUpDown := 1.0
		if paddleY > state.ballPosY { reflectUpDown = -1.0} else {reflectUpDown = 1.0}
		reflection := reflectUpDown * (dtp / halfPaddleLength)
		state.ballHoriz = state.ballHoriz * -1.0
		fmt.Println(reflection, state.ballVert)
		state.ballVert = reflection
		*paddleHitTime = t
	}else {
		// didnt hit paddle
		state.ballHoriz = state.ballHoriz * -1.0
	}

	state.ballPosX = r*math.Cos(sideTheta)
	//fmt.Println("left collision")
	return state
}

func detectBoundaryCollision(state aState, t float64) aState{
	xsq := math.Pow(state.ballPosX, 2)
	ysq := math.Pow(state.ballPosY, 2)
	r := math.Sqrt(xsq+ysq)
	theta := math.Atan2(state.ballPosY, state.ballPosX)
	//fmt.Println("theta ", theta)
	// check top arc r>=2
	if (r >=2){
		state.ballVert = state.ballVert * -1.0
		//fmt.Println("top collision")
		//fmt.Println(state)
		return state
	}
	// check at origin r<=0.01
	if (r <=0.01 || state.ballPosY < 0.0){
		state.ballVert = state.ballVert * -1.0
		//fmt.Println("bottom collision")
		//fmt.Println(state)
		state.ballPosY = 0.01
		return state
	}
	// check left side theta<=-pi/20
	if (theta < leftTheta) {
		state = paddleHitCheckRoutine(state, t, r, "left")
		return state
	}
	// check right side theta>=pi/20
	if (theta > rightTheta){
		state = paddleHitCheckRoutine(state, t, r,  "right")
		return state
	}
	return state
}

func handlePlayerInput(state aState, midiState *midi.MidiState) aState {
	// Was thinking about having left and right buttons rather than scrollers.
	//
	//leftPlayerLeftwardAmt := float64(midiState.KeyVolumes[midi.LPD8_PAD1]) / 127.0 * paddleControlSensitivity
	//leftPlayerRightwardAmt := float64(midiState.KeyVolumes[midi.LPD8_PAD2]) / 127.0 * paddleControlSensitivity
	//rightPlayerLeftwardAmt := float64(midiState.KeyVolumes[midi.LPD8_PAD3]) / 127.0 * paddleControlSensitivity
	//rightPlayerRightwardAmt := float64(midiState.KeyVolumes[midi.LPD8_PAD4]) / 127.0 * paddleControlSensitivity
	//
	////fmt.Println(midiState.KeyVolumes[midi.LPD8_PAD1])
	//state.rightPaddlePos =  state.rightPaddlePos - rightPlayerLeftwardAmt + rightPlayerRightwardAmt
	//state.leftPaddlePos = state.leftPaddlePos - leftPlayerLeftwardAmt + leftPlayerRightwardAmt
	//
	//if state.rightPaddlePos > 1 {state.rightPaddlePos = 1}
	//if state.leftPaddlePos > 1 {state.leftPaddlePos = 1}
	//if state.rightPaddlePos < 0 {state.rightPaddlePos = 0}
	//if state.leftPaddlePos < 0 {state.leftPaddlePos = 0}

	state.leftPaddlePos = float64(midiState.ControllerValues[config.MORPH_KNOB]) / 127.0
	state.rightPaddlePos = float64(midiState.ControllerValues[config.PlAYER2_KNOB]) / 127.0

	return state
}

func simulateBall(state aState) aState{
	//fmt.Print(state)
	xsq := math.Pow(state.ballPosX, 2)
	ysq := math.Pow(state.ballPosY, 2)
	r := math.Sqrt(xsq+ysq)
	theta := math.Atan2(state.ballPosY, state.ballPosX)
	rShiftBase := 0.000001
	thetaShiftBase := 0.00005
	rShiftAmt := rShiftBase * state.ballVert
	// i didive the shift amount by 2pi(19/40)r, so that
	// to the time to travel the segment length is proportional to the radius.
	thetaShiftAmt := thetaShiftBase * state.ballHoriz /(2.98*r)
	//fmt.Println("thetaShiftAmt", thetaShiftAmt)
	r_prime := r+ rShiftAmt
	theta_prime := theta+ thetaShiftAmt


	state.ballPosX = r_prime * math.Cos(theta_prime)
	state.ballPosY = r_prime * math.Sin(theta_prime)
	//fmt.Print(state)

	return state
}

func removePaddleActives(state aState, t float64) aState{
	if t - state.leftPaddleHitTime > paddleActiveLightUpSecs{
		state.leftPaddleHitTime = -1
	}
	if t - state.rightPaddleHitTime> paddleActiveLightUpSecs{
		state.rightPaddleHitTime = -1
	}

	return state
}

func updatePhysics (t float64, state aState) aState{
	//fmt.Println("new update")
	//fmt.Print(state)
	state = detectBoundaryCollision(state, t)
	state = simulateBall(state)
	state = removePaddleActives(state, t)

	return state
}
func MakePatternPolarPong(locations []float64) ByteThread {

	state := aState{ballPosY:0.1,
		ballPosX:1.0,
		ballHoriz:1.0,
		ballVert:1.0,
	leftPaddlePos:0.5,
	rightPaddlePos:0.5}

	return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
		for bytes := range bytesIn {
			n_pixels := len(bytes) / 3
			// time comes in nano second
			// one nano second is 1e9(th) seconds
			// so fraction becomes seconds
			// subtract
			t := float64(time.Now().UnixNano())/1.0e9 - 9.4e8
			// fill in bytes slice
			for ii := 0; ii < n_pixels; ii++ {
				//--------------------------------------------------------------------------------
				frameness := colorutils.PosMod2(t, physicsUpdateIntervalSeconds)
				//fmt.Print("***", t, physicsUpdateInterval, frameness)
				if frameness <= 1.0 {
					//fmt.Print("UPDATE")
					state = updatePhysics(t, state)
					}
				//fmt.Println(float64(midiState.ControllerValues[config.MORPH_KNOB]) / 127.0)
				state = handlePlayerInput(state, midiState)

				x := locations[ii*3+0]
				y := locations[ii*3+1]
				z := locations[ii*3+2]
				y = z //actually need to target x-z plane

				r := 0.0
				g := 0.0
				b := 0.0

				// render ball
				ysq := math.Pow(y, 2)
				xsq := math.Pow(x, 2)
				rad := math.Sqrt(xsq+ysq)
				distToBall := math.Sqrt(math.Pow(x-state.ballPosX,2) + math.Pow(y-state.ballPosY,2))
				if (distToBall < ballDisplayRadMin*rad+0.05){
					r=1;g=1;b=1
				}

				//spiral emanating
				spiralAmt := Spiral(x-state.ballPosX,y-state.ballPosY,t, 0.1, 12, 1, 0.15, 5)
				r+=spiralAmt
				g+=spiralAmt
				b+=spiralAmt

				// floating green lattice
				if (colorutils.PosMod(x, 0.35)<=0.035 && math.Abs(x)>=0.25) || (colorutils.PosMod(y, 0.35)<=0.035 && math.Abs(y)>=0.25){
					g+=0.25
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
