package opc

// Polar Pong

import (
	"github.com/austinfromboston/pixelslinger/colorutils"
	"github.com/austinfromboston/pixelslinger/midi"
	"math"
	"time"
	"fmt"
	"github.com/austinfromboston/pixelslinger/config"
	"math/rand"
)

// Configs
// physics update interval (ms)
const (physicsUpdateIntervalMiliSeconds = 25
	physicsUpdateIntervalSeconds = physicsUpdateIntervalMiliSeconds / 1e3
	ballDisplayRadMin = 0.08
	paddleLength = 0.25
	halfPaddleLength = paddleLength / 2
	leftTheta=math.Pi / 20.0
	rightTheta=math.Pi * 19.0/20.0
	paddleActiveLightUpSecs = 1
	PLAY_TO_SCORE = 9
	DISPLAY_SCORE_SECONDS = 3.2
	DISPLAY_POST_SECONDS = 5
	SPIRAL_DISPLAY=0.6
	PADDLE_BOX_EASE=0.25
	)


type aState struct {
	ballPosX  float64
	ballPosY  float64
	ballVert  float64
	ballHoriz float64
	leftPaddlePos float64 // the percentage it is from the bottom to top
	rightPaddlePos float64
	leftPaddleHitTime float64
	rightPaddleHitTime float64
	matchPhase string// are we in pre, game, score, or post =
	lastPreGrameStart float64
	lastGameStart float64
	lastMatchStart float64
	lastDisplayStart float64
	lastPostStart float64
	score [2]int // the left score and the right score
	leftScoreAngle float64// float64
	rightScoreAngle float64// flaoat64
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
	scoringPos := 0
	if leftRight == "right"{
		paddlePos = state.rightPaddlePos
		sideTheta = rightTheta
		paddleHitTime = &state.rightPaddleHitTime
		scoringPos = 1
	}

	_, paddleY, dtp := distanceToPaddle(state.ballPosX, state.ballPosY, paddlePos, sideTheta)
	if dtp < halfPaddleLength+PADDLE_BOX_EASE{
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
		// scoring points
		state.score[scoringPos] += 1
		state.matchPhase = "score"
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

	state.leftPaddlePos = 1-float64(midiState.ControllerValues[config.MORPH_KNOB]) / 127.0 //1- value to compensate for mirror effect
	state.rightPaddlePos = float64(midiState.ControllerValues[config.PlAYER2_KNOB]) / 127.0

	return state
}

func simulateBall(state aState, t float64) aState{
	// if first time
	if state.lastGameStart == -1 { state.lastGameStart = t}

	gameTimeElapsed := t- state.lastGameStart
	//fmt.Println("gameTimeElapsed", gameTimeElapsed)
	//fmt.Print(state)
	xsq := math.Pow(state.ballPosX, 2)
	ysq := math.Pow(state.ballPosY, 2)
	r := math.Sqrt(xsq+ysq)
	theta := math.Atan2(state.ballPosY, state.ballPosX)
	rShiftBase := 0.000001
	rShiftSpeed:= 0.0000001 * gameTimeElapsed
	thetaShiftBase := 0.00005
	thetaShiftSpeed:= 0.000005 * gameTimeElapsed
	rShiftAmt := (rShiftBase+rShiftSpeed) * state.ballVert
	// i didive the shift amount by 2pi(19/40)r, so that
	// to the time to travel the segment length is proportional to the radius.
	thetaShiftAmt := (thetaShiftBase+thetaShiftSpeed) * state.ballHoriz /(2.98*r)
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

func displayScore(t float64, state aState) aState{
	// if last display is unset then we are stating to animate
	if state.lastDisplayStart == -1{
		state.lastDisplayStart = t
	}

	// end of animation
	if t - state.lastDisplayStart > DISPLAY_SCORE_SECONDS{
		state.lastDisplayStart = -1
		// end routine
		if state.score[0]>=PLAY_TO_SCORE ||  state.score[1]>=PLAY_TO_SCORE{
			state.matchPhase = "post" // victory condition
		} else{
			state.matchPhase = "game"
			state.lastGameStart = t
			// reset the ball
			state.ballPosY = 1 + rand.Float64()
			state.ballPosX = 0
		}
	}

	// calculate the score angles
	// how far away through the display phase
	displayDuration :=  (t - state.lastDisplayStart) / float64(DISPLAY_SCORE_SECONDS-1)
	if displayDuration > 1{ displayDuration = 1}

	rightScoreAngleMax := math.Pi - (math.Pi/20.0 * float64(state.score[0]))
	state.rightScoreAngle = math.Pi - ((math.Pi - rightScoreAngleMax) * displayDuration)

	leftScoreAngleMax := math.Pi/20.0 * float64(state.score[1])
	state.leftScoreAngle = leftScoreAngleMax * displayDuration

	return state
}

func displayPostGame(t float64, state aState) aState {
	if state.lastPostStart == -1{
		state.lastPostStart = t
	}
	if t - state.lastPostStart > DISPLAY_POST_SECONDS {
		state.lastPostStart = -1
		// end routine
		state.score = [2]int{0,0}
		state.matchPhase = "game"
	}

	return state
	}

func updatePhysics (t float64, state aState) aState{
	//fmt.Println("new update")
	//fmt.Print(state)
	state = detectBoundaryCollision(state, t)
	state = simulateBall(state, t)
	state = removePaddleActives(state, t)

	return state
}

func phaseDispatcher(t float64, state aState) aState{
	//fmt.Println("phase", state)
	switch {
	case state.matchPhase == "game":
		return 	updatePhysics(t, state)
	case state.matchPhase == "score":
		return displayScore(t, state)
	case state.matchPhase == "post":
		return displayPostGame(t, state)
	default:
		return  state
	}
}


func MakePatternPolarPong(locations []float64) ByteThread {

	state := aState{ballPosY:1.0,
		ballPosX:0.0,
		ballHoriz:1.0,
		ballVert:1.0,
		leftPaddlePos:0.5,
		rightPaddlePos:0.5,
		score:[2]int{0,0},
		matchPhase:"game",
		lastDisplayStart:-1,
		lastPostStart:-1,
		lastGameStart: -1,
	}

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
					state = phaseDispatcher(t, state)
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
				if state.matchPhase == "game" || state.matchPhase == "score" || state.matchPhase == "post"{
					spiralAmt := Spiral(x-state.ballPosX,y-state.ballPosY,t, 0.1, 12, 1, 0.15, 7)
					r+=spiralAmt*SPIRAL_DISPLAY
					g+=spiralAmt*SPIRAL_DISPLAY
					b+=spiralAmt*SPIRAL_DISPLAY
				}

				// floating green lattice
				if (colorutils.PosMod(x, 0.35)<=0.035 && math.Abs(x)>=0.25) ||
					(colorutils.PosMod(y, 0.35)<=0.035 && math.Abs(y)>=0.25) {
					g+=0.25
				}

				theta := math.Atan2(y, x)

				// render score
				if (theta > state.rightScoreAngle - 0.01) && (state.matchPhase == "score"){
					//fmt.Println("right left thetas", state.rightScoreAngle, state.leftScoreAngle)
					b+=0.75
				}

				if (theta < state.leftScoreAngle+0.01) && (state.matchPhase == "score"){
					g+=0.75
				}


				// render paddles
				if (theta < leftTheta+0.01){
					_, _, dtp := distanceToPaddle(x, y, state.leftPaddlePos, leftTheta)
					if dtp < paddleLength{
						if state.leftPaddleHitTime>0.0{
							g=1; b=1
						}else{
							g=1
						}
					}
				}
				if (theta > rightTheta-0.01){
					_, _, dtp := distanceToPaddle(x, y, state.rightPaddlePos, rightTheta)
					if dtp < paddleLength{
						if state.rightPaddleHitTime>0.0{
							b=1; g=1
						} else {
							b=1
						}
					}
				}

				// render post animation
				if state.matchPhase == "post" {
					winnerSideIsLeft := true
					if state.score[1]>state.score[0]{winnerSideIsLeft = false}
					xIsLeft := x < 0

					shouldFlash := (winnerSideIsLeft && xIsLeft) || (!winnerSideIsLeft && !xIsLeft)

					evenNess := colorutils.PosMod(t, 0.25) < 0.125
					if evenNess == true && shouldFlash{
						r=0.2;b=0.2;g=0.2
					}
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
