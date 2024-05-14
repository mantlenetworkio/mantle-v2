package header

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"

	"golang.org/x/crypto/sha3"
)

const (
	DataStoreHeaderLength = 164
)

var ErrHeaderInconsistentLength = errors.New("ErrHeaderInconsistentLength")

// header as CallData
type DataStoreHeader struct {
	KzgCommit      [64]byte
	Degree         uint32 // ToDo replace with fraud proof. Now it is size of padded data per node, 'l'
	NumSys         uint32
	NumPar         uint32
	OrigDataSize   uint32 // this size is claimed by disperser, retriever uses it
	Disperser      [20]byte
	LowDegreeProof [64]byte
}

func (h *DataStoreHeader) Encode() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0))
	if err := binary.Write(buf, binary.BigEndian, h.KzgCommit); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, h.Degree); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, h.NumSys); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, h.NumPar); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, h.OrigDataSize); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, h.Disperser); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, h.LowDegreeProof); err != nil {
		return nil, err
	}
	//if err := binary.Write(buf, binary.BigEndian, h.OrigDataSize); err != nil {
	//log.Fatal(err)
	//}
	b := buf.Bytes()
	if len(b) != DataStoreHeaderLength {
		log.Printf("header encode to %v bytes. Inconsistant to StaticLen %v\n", len(b), DataStoreHeaderLength)
		return nil, ErrHeaderInconsistentLength
	}
	return b, nil
}

func DecodeDataStoreHeader(h []byte) (DataStoreHeader, error) {
	header := DataStoreHeader{}
	buf := bytes.NewReader(h)
	if err := binary.Read(buf, binary.BigEndian, &header.KzgCommit); err != nil {
		return DataStoreHeader{}, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.Degree); err != nil {
		return DataStoreHeader{}, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.NumSys); err != nil {
		return DataStoreHeader{}, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.NumPar); err != nil {
		return DataStoreHeader{}, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.OrigDataSize); err != nil {
		return DataStoreHeader{}, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.Disperser); err != nil {
		return DataStoreHeader{}, err
	}
	if err := binary.Read(buf, binary.BigEndian, &header.LowDegreeProof); err != nil {
		return DataStoreHeader{}, err
	}
	//if err := binary.Read(buf, binary.BigEndian, &header.Duration); err != nil {
	//log.Fatal(err)
	//}
	return header, nil
}
func CreateUploadHeader(header DataStoreHeader) ([]byte, [32]byte, error) {
	headerByte, err := header.Encode()
	if err != nil {
		return nil, [32]byte{}, err
	}
	var headerHash [32]byte
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write(headerByte)
	copy(headerHash[:], hasher.Sum(nil)[:32])
	return headerByte, headerHash, nil
}
