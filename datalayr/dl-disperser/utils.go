package disperser

import (
	"fmt"
	"math/rand"
	"time"
)

func uint32ToByteSlice(x uint32) []byte {
	res := make([]byte, 4)
	res[0] = byte(x >> 24)
	res[1] = byte((x >> 16) & 255)
	res[2] = byte((x >> 8) & 255)
	res[3] = byte(x & 255)
	return res
}

func packTo(x []byte, n int) []byte {
	for i := len(x); i < n; i++ {
		x = append([]byte{byte(0)}, x...)
	}
	return x
}

func shuffleRegistrantIdsInPlace(ids []uint64) {
	for i := 0; i < len(ids); i++ {
		r := rand.Intn(i + 1)
		if i != r {
			ids[i], ids[r] = ids[r], ids[i]
		}
	}
}

func shuffleBytesInPlace(ids []byte) {
	for i := 0; i < len(ids); i++ {
		r := rand.Intn(i + 1)
		if i != r {
			ids[i], ids[r] = ids[r], ids[i]
		}
	}
}

func chooseRandomK(aggSig [][]byte, k uint8) [][]byte {
	if k == 0 {
		return nil
	}

	if uint8(len(aggSig)) < k {
		return aggSig
	}

	out := make([][]byte, 0)
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for i, d := range r.Perm(len(aggSig)) {
		out = append(out, aggSig[d])
		if uint8(i) == k-1 {
			break
		}
	}
	return out
}

func make32ByteArray(a []byte) ([32]byte, error) {
	var b [32]byte
	if len(a) != 32 {
		return b, fmt.Errorf("input is not 32 byte cannot copy to array")
	}

	copy(b[:], a[:])
	return b, nil
}
