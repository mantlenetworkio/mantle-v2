package kzgEncoder

import (
	"log"

	rs "github.com/Layr-Labs/datalayr/lib/encoding/encoder"
)

func (g *KzgEncoder) DecodeSafe(frames []Frame, indices []uint64, inputSize uint64) ([]byte, error) {
	if g.Verbose {
		log.Println("Entering DecodeSafe function")
		defer log.Println("Exiting DecodeSafe function")
	}

	rsFrames := make([]rs.Frame, len(frames))
	for ind, frame := range frames {
		rsFrames[ind] = rs.Frame{Coeffs: frame.Coeffs}
	}

	return g.Encoder.DecodeSafe(rsFrames, indices, inputSize)
}
