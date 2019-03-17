package opc

// Polar Pong

import (
	"github.com/longears/pixelslinger/colorutils"
	"github.com/longears/pixelslinger/midi"
	"math"
	"time"

)

// Configs
// physics update interval (ms)
const (physicsUpdateIntervalMiliSeconds = 25
	physicsUpdateIntervalSeconds = physicsUpdateIntervalMiliSeconds / 1e3
	ballDisplayRadMin = 0.08
	paddleLength = 0.1
	halfPaddleLength = paddleLength / 2
	leftTheta=math.Pi / 20.0
	rightTheta=math.Pi * 19.0/20.0)


type aState struct {
	ballPosX  float64
	ballPosY  float64
	ballVert  float64
	ballHoriz float64
	leftPaddlePos float64 // the percentage it is from the bottom to top
	rightPaddlePos float64
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

func detectBoundaryCollision(state aState) aState{
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
	if (theta <= leftTheta) {
		_, paddleY, dtp := distanceToPaddle(state.ballPosX, state.ballPosY, state.leftPaddlePos, leftTheta)
		if dtp < paddleLength {
			reflectUpDown := 1
			if paddleY > state.ballPosY { reflectUpDown = -1}
			reflection := reflectUpDown * (dtp / halfPaddleLength)
			state.ballHoriz = state.ballHoriz * -1.0
			state.ballVert += reflection
		}
		state.ballPosX = r*math.Cos(math.Pi/20)
		//fmt.Println("left collision")
		return state
	}
	// check right side theta>=pi/20
	if (theta >= rightTheta){
		state.ballHoriz = state.ballHoriz * -1.0
		state.ballPosX = r*math.Cos(math.Pi* 19.0/20.0)
		state.ballPosY = r*math.Sin(math.Pi* 19.0/20.0)
		//fmt.Println("right collision")
		//fmt.Println(state)
		return state
	}
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


func updatePhysics (t float64, state aState) aState{
	//fmt.Println("new update")
	//fmt.Print(state)
	state = detectBoundaryCollision(state)
	state = simulateBall(state)
	return state
}



func MakePatternPolarPong(locations []float64) ByteThread {


	// get bounding box

	state := aState{ballPosY:1.0,
		ballPosX:1.0,
		ballHoriz:1.0,
		ballVert:1.0,
	leftPaddlePos:0.5,
	rightPaddlePos:0.5}
	// paddleShapefn
	//const (halfPaddleWidth = 0.2
	//       angleBetweenBlades = math.Pi / 20)
	//var angleBetweenBlades float64 = math.Pi / 20
	//func paddleShape(paddleX float64, paddleY float64, leftSide bool) func(x float64, y float64) bool{
	//	angle := angleBetweenBlades
	//	if leftSide {
	//		angle = -1 * angle
	//	}
	//	return -1*halfPaddleWidth < (math.Tan(angle) * x) - y
	//}
	//
	// ballShapefn

	// State
	// ballRenderlocations
	// paddlePosition (1 and 2)
	// paddleRenderlocations

	// Functions
	// ballPath
	// collision detection (changes ball path)
	// calculate PaddlePostions

	// "Main loop"
	// getMidiPositions
	// updatePhysics (every pixel, or every X nanosec)
	// // function of (midiPositons, ballPath, t)
	////// update paddlePositions
	///////update ballPosition
	///////do colisiondection
	////////// if collidedchangeBallpath
	///////update ballandpaddlelocations
	///// renderPaddles and Ball via lookup

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

				// render paddle
				theta := math.Atan2(y, x)
				if (theta < leftTheta+0.01){
					_, _, dtp := distanceToPaddle(x, y, state.rightPaddlePos, rightTheta)
					if dtp < paddleLength{
						g=1
					}
				}
				if (theta > rightTheta-0.01){
					_, _, dtp := distanceToPaddle(x, y, state.rightPaddlePos, rightTheta)
					if dtp < paddleLength{
						b=1
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
