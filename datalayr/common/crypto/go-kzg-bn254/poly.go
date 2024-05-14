package kzg

//import "github.com/protolambda/go-kzg/bls"
import (
	"math"

	bls "github.com/Layr-Labs/datalayr/common/crypto/go-kzg-bn254/bn254"
)

// invert the divisor, then multiply
func polyFactorDiv(dst *bls.Fr, a *bls.Fr, b *bls.Fr) {
	// TODO: use divmod instead.
	var tmp bls.Fr
	bls.InvModFr(&tmp, b)
	bls.MulModFr(dst, &tmp, a)
}

// Long polynomial division for two polynomials in coefficient form
func polyLongDiv(dividend []bls.Fr, divisor []bls.Fr) []bls.Fr {
	a := make([]bls.Fr, len(dividend), len(dividend))
	for i := 0; i < len(a); i++ {
		bls.CopyFr(&a[i], &dividend[i])
	}
	aPos := len(a) - 1
	bPos := len(divisor) - 1
	diff := aPos - bPos
	out := make([]bls.Fr, diff+1, diff+1)
	for diff >= 0 {
		quot := &out[diff]
		polyFactorDiv(quot, &a[aPos], &divisor[bPos])
		var tmp, tmp2 bls.Fr
		for i := bPos; i >= 0; i-- {
			// In steps: a[diff + i] -= b[i] * quot
			// tmp =  b[i] * quot
			bls.MulModFr(&tmp, quot, &divisor[i])
			// tmp2 = a[diff + i] - tmp
			bls.SubModFr(&tmp2, &a[diff+i], &tmp)
			// a[diff + i] = tmp2
			bls.CopyFr(&a[diff+i], &tmp2)
		}
		aPos -= 1
		diff -= 1
	}
	return out
}

// poly multiplication using FFT
func PolyMul(f []bls.Fr, g []bls.Fr) []bls.Fr {

	// needed for doing DFT
	degf := len(f) - 1
	degg := len(g) - 1
	fftLen := CeilIntPowerOf2Num(uint64(float64(degf+1) + float64(degg)))

	// FFT of f and g
	fs := NewFFTSettings(uint8(math.Log2(float64(fftLen))))
	fext := make([]bls.Fr, fftLen)
	gext := make([]bls.Fr, fftLen)
	for i := 0; i <= degf; i++ {
		bls.CopyFr(&fext[i], &f[i])
	}

	for i := 0; i <= degg; i++ {
		bls.CopyFr(&gext[i], &g[i])
	}

	fdft := make([]bls.Fr, len(fext))
	fs.InplaceFFT(fext, fdft, false)

	fext1 := make([]bls.Fr, len(fext))
	fs.InplaceFFT(fdft, fext1, true)

	gdft := make([]bls.Fr, len(gext))
	fs.InplaceFFT(gext, gdft, false)

	// get product polynomial in frequency domain
	proddft := make([]bls.Fr, fftLen)
	for i := 0; i <= int(fftLen-1); i++ {
		bls.MulModFr(&proddft[i], &fdft[i], &gdft[i])
	}
	// get poly coeffs via inverse DFT
	prod := make([]bls.Fr, len(proddft))
	fs.InplaceFFT(proddft, prod, true)

	deg := degf + degg
	return prod[:(deg + 1)]
	// return prod, err

}

// /*
// returns the power of 2 which is immediately bigger than the input
// */
// func CeilIntPowerOf2Num(d uint64) uint64 {
// 	nextPower := math.Ceil(math.Log2(float64(d)))
// 	return uint64(math.Pow(2.0, nextPower))
// }
