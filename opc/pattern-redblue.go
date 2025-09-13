package opc

// RedBlue
//   Set first half of pixels to red and second half to blue.

import (
    "github.com/austinfromboston/pixelslinger/midi"
)

func MakePatternRedBlue(locations []float64) ByteThread {
    return func(bytesIn chan []byte, bytesOut chan []byte, midiState *midi.MidiState) {
        _ = midiState // not used, but kept for signature compatibility
        for bytes := range bytesIn {
            nPixels := len(bytes) / 3
            half := nPixels / 2

            // First half: red
            for i := 0; i < half; i++ {
                base := i * 3
                bytes[base+0] = 255 // R
                bytes[base+1] = 0   // G
                bytes[base+2] = 0   // B
            }

            // Second half: blue
            for i := half; i < nPixels; i++ {
                base := i * 3
                bytes[base+0] = 0   // R
                bytes[base+1] = 0   // G
                bytes[base+2] = 255 // B
            }

            bytesOut <- bytes
        }
    }
}

