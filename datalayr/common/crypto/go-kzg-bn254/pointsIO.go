package kzg

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	bls "github.com/Layr-Labs/datalayr/common/crypto/go-kzg-bn254/bn254"
)

func ReadG1Points(filepath string, n uint64, numWorker int) []bls.G1Point {
	g1f, err := os.Open(filepath)
	if err != nil {
		log.Println("Cannot ReadG1Points", filepath)
		log.Fatal(err)
	}

	defer func() {
		if err := g1f.Close(); err != nil {
			panic(err)
		}
	}()

	startTimer := time.Now()
	g1r := bufio.NewReaderSize(g1f, int(n*64))

	if n < uint64(numWorker) {
		numWorker = int(n)
	}

	buf, _, err := g1r.ReadLine()
	if err != nil {
		panic(err)
	}

	if uint64(len(buf)) < 64*n {
		log.Fatalf("Error. Insufficient G1 points. Only contains %v. Requesting %v\n", len(buf)/64, n)
	}

	// measure reading time
	t := time.Now()
	elapsed := t.Sub(startTimer)
	fmt.Printf("    Reading G1 points (%v bytes) takes %v\n", (n * 64), elapsed)
	startTimer = time.Now()

	s1Outs := make([]bls.G1Point, n, n)

	var wg sync.WaitGroup
	wg.Add(numWorker)

	start := uint64(0)
	end := uint64(0)
	size := n / uint64(numWorker)

	for i := uint64(0); i < uint64(numWorker); i++ {
		start = i * size

		if i == uint64(numWorker)-1 {
			end = n
		} else {
			end = (i + 1) * size
		}
		//fmt.Printf("worker %v start %v end %v. size %v\n", i, start, end, end - start)
		go readG1Worker(buf, s1Outs, start, end, 64, &wg)
	}
	wg.Wait()

	// measure parsing time
	t = time.Now()
	elapsed = t.Sub(startTimer)
	fmt.Println("    Parsing takes", elapsed)
	return s1Outs
}

func readG1Worker(
	buf []byte,
	outs []bls.G1Point,
	start uint64, // in element, not in byte
	end uint64,
	step uint64,
	wg *sync.WaitGroup,
) {
	for i := start; i < end; i++ {
		g1 := buf[i*step : (i+1)*step]
		err := outs[i].UnmarshalText(g1[:])
		if err != nil {
			panic(err)
		}
	}
	wg.Done()
}

func readG2Worker(
	buf []byte,
	outs []bls.G2Point,
	start uint64, // in element, not in byte
	end uint64,
	step uint64,
	wg *sync.WaitGroup,
) {
	for i := start; i < end; i++ {
		g1 := buf[i*step : (i+1)*step]
		err := outs[i].UnmarshalText(g1[:])
		if err != nil {
			panic(err)
		}
	}
	wg.Done()
}

func ReadG2Points(filepath string, n uint64, numWorker int) []bls.G2Point {
	g1f, err := os.Open(filepath)
	if err != nil {
		log.Println("Cannot ReadG2Points", filepath)
		log.Fatal(err)
	}

	defer func() {
		if err := g1f.Close(); err != nil {
			panic(err)
		}
	}()

	startTimer := time.Now()
	g1r := bufio.NewReaderSize(g1f, int(n*128))

	if n < uint64(numWorker) {
		numWorker = int(n)
	}

	buf, _, err := g1r.ReadLine()
	if err != nil {
		panic(err)
	}

	if uint64(len(buf)) < 128*n {
		log.Fatalf("Error. Insufficient G1 points. Only contains %v. Requesting %v\n", len(buf)/128, n)
	}

	// measure reading time
	t := time.Now()
	elapsed := t.Sub(startTimer)
	fmt.Printf("    Reading G2 points (%v bytes) takes %v\n", (n * 128), elapsed)

	startTimer = time.Now()

	s2Outs := make([]bls.G2Point, n, n)

	var wg sync.WaitGroup
	wg.Add(numWorker)

	start := uint64(0)
	end := uint64(0)
	size := n / uint64(numWorker)
	for i := uint64(0); i < uint64(numWorker); i++ {
		start = i * size

		if i == uint64(numWorker)-1 {
			end = n
		} else {
			end = (i + 1) * size
		}

		go readG2Worker(buf, s2Outs, start, end, 128, &wg)
	}
	wg.Wait()

	// measure parsing time
	t = time.Now()
	elapsed = t.Sub(startTimer)
	fmt.Println("    Parsing takes", elapsed)
	return s2Outs
}
