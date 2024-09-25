package oscpixels

import (
	"github.com/austinfromboston/pixelslinger/patterns"
	"github.com/hypebeast/go-osc/osc"
	"math/rand"
	"sort"
	"time"
)

type RhythmState struct {
	BPM               float32
	RecentOSCMessages []*osc.Message
	LastBeat          int
	BeatStartTime     int64
	Timing            struct {
		InitialBeat int64
		LastOneBeat int64
	}
	CurrentTrackTitle string
	NextTrackTitle    string
	Scrambles         map[string]Scramble
	CurrentSpeed      int
}

type Scramble struct {
	Pattern    string
	Gain       int
	Hue        int
	Saturation int
}

const BeatsPerMeasure = 4

func (rhythmState *RhythmState) UpdateStateFromMessage(msg *osc.Message) {
	currentTime := time.Now().UnixMilli()
	//println("BPM", rhythmState.BPM, "lastBeat", rhythmState.LastBeat, "current", currentTime)
	switch msg.Address {
	case "/lx/tempo/beat":
		if rhythmState.Timing.InitialBeat == 0 {
			rhythmState.Timing.InitialBeat = currentTime
		}
		currentBeat := msg.Arguments[0].(int32)
		rhythmState.LastBeat = int(currentBeat)
		rhythmState.BeatStartTime = currentTime
		//println("Beat:", rhythmState.BeatStartTime, " - ", currentBeat)
		if currentBeat == 1 {
			rhythmState.Timing.LastOneBeat = currentTime
		} else {
			rhythmState.Timing.LastOneBeat = int64(float32(currentTime) - (float32(currentBeat) * (60000.0 / rhythmState.BPM)))
		}
	case "/lx/tempo/setBPM":
		rhythmState.BPM = float32(msg.Arguments[0].(float64))
		rhythmState.CurrentSpeed = getSpeed(int(rhythmState.BPM))
		//println("new speed", rhythmState.CurrentSpeed)
	case "/lx/track/fadeout/track_title":
		rhythmState.CurrentTrackTitle = rhythmState.NextTrackTitle
		// foo
	case "/lx/track/fadein/track_title":
		rhythmState.NextTrackTitle = msg.Arguments[0].(string)
		if _, ok := rhythmState.Scrambles[rhythmState.NextTrackTitle]; ok {
			// do nothing
		} else {
			rhythmState.Scrambles[rhythmState.NextTrackTitle] = newScramble()
		}

	}
}

type BeatInfo struct {
	Beat          int
	BeatStartTime int64
	BeatEndTime   int64
}

func newScramble() Scramble {
	rand.Seed(time.Now().UnixNano()) // seed or it will be set to 1
	patternIndex := rand.Intn(len(patterns.PATTERN_LIST))
	hue := rand.Intn(127)
	gain := 64 + rand.Intn(64)
	saturation := rand.Intn(60)
	return Scramble{
		patterns.PATTERN_LIST[patternIndex],
		gain,
		hue,
		saturation,
	}
}

var bpmRanges = []int{-1, 60, 80, 100, 120, 140, 190, 355}
var speeds = []int{30, 40, 65, 80, 95, 110, 127}

func getSpeed(n int) int {
	return speeds[sort.SearchInts(bpmRanges, n)]
}

func (rhythmState *RhythmState) CurrentBeat() BeatInfo {
	currentTime := time.Now().UnixMilli()
	if rhythmState.Timing.LastOneBeat > 0 {
		elapsedTime := currentTime - rhythmState.Timing.LastOneBeat
		beatDuration := (60.0 / rhythmState.BPM) * 1000.0
		beatsElapsed := int(float32(elapsedTime) / beatDuration)
		currentBeat := beatsElapsed % BeatsPerMeasure
		beatStartTime := rhythmState.Timing.LastOneBeat + int64(beatDuration*float32(currentBeat))
		info := BeatInfo{
			int(currentBeat),
			beatStartTime,
			beatStartTime + int64(beatDuration),
		}
		//println("time", currentTime, "beatstart", beatStartTime, "elapsed-since-post", elapsedTime)
		//println("beat", info.Beat, "start", info.BeatStartTime, "elapsed", info.BeatStartTime-rhythmState.Timing.InitialBeat)
		return info
	}
	return BeatInfo{1, 0, 0}
}

// Given a slice of MidiMessages, update the MidiState object.
// Note that the MidiState object will keep a pointer to the midiMessages
// object you provide here (as MidiState.RecentMidiMessages).
func (rhythmState *RhythmState) UpdateStateFromOSCSlice(oscMessages []*osc.Message) {
	rhythmState.RecentOSCMessages = oscMessages
	for _, m := range rhythmState.RecentOSCMessages {
		//osc.PrintMessage(m)
		rhythmState.UpdateStateFromMessage(m)

	}
}

func (rhythmState *RhythmState) UpdateStateFromChannel(oscMessageChan chan *osc.Message) {
	rhythmState.UpdateStateFromOSCSlice(GetAvailableOSCMessages(oscMessageChan))
}

// Pull all the available MidiMessages out of the channel without blocking.  Requires a channel
// with a buffer length greater than zero.
// This can deadlock if used by more than one goroutine at a time pulling on the same channel.
func GetAvailableOSCMessages(oscMessageChan chan *osc.Message) []*osc.Message {
	result := make([]*osc.Message, 0)
	for {
		if len(oscMessageChan) == 0 {
			break
		}
		result = append(result, <-oscMessageChan)
	}
	return result
}

//
//func (rhythmState *RhythmState) UpdateStateFromRhythm(state midi.MidiState, frameStartTime float64) float64 {
//	if rhythmState.CurrentBeat(frameStartTime) == 0 {
//		midi.MidiState["KeyVolumes"][config.FLASH_PAD] = 127
//	} else {
//		midi.MidiState.KeyVolumes[config.FLASH_PAD] = 0
//	}
//}
