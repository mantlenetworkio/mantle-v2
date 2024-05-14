package kzgEncoder_test

import (
	"bufio"
	"log"
	"os"
	"testing"

	rs "github.com/Layr-Labs/datalayr/lib/encoding/encoder"
	kzgRs "github.com/Layr-Labs/datalayr/lib/encoding/kzgEncoder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProveZeroPadding(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	group, _ := kzgRs.NewKzgEncoderGroup(kzgConfig)
	enc, err := group.NewKzgEncoder(numSys, numPar, uint64(len(GETTYSBURG_ADDRESS_BYTES)))
	require.Nil(t, err)

	inputFr := rs.ToFrArray(GETTYSBURG_ADDRESS_BYTES)

	_, _, _, _, err = enc.Encode(inputFr)
	require.Nil(t, err)

	assert.True(t, true, "Proof %v failed\n")
}

func writeStringToFile(outputPath string, data []byte) error {
	if err := os.Mkdir("outputs", os.ModePerm); err != nil {
		log.Println(err)
	}
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	// remember to close the File
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	//for each signature, write it to a new line in the signature file
	_, err = file.Write(data)

	return err
}
