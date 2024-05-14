package merkzg

import (
	"bytes"

	bls "github.com/Layr-Labs/datalayr/common/crypto/go-kzg-bn254/bn254"
	"github.com/ethereum/go-ethereum/crypto"
)

type KzgMerkleTree struct {
	G1Points []bls.G1Point
	G2Points []bls.G2Point
	Tree     [][][]byte
	// [
	//        [root]
	//    [c1,       c2]
	//[c21,  c22, c23, c24]
	//        ...
	// ]
}

// func main() {
// 	numLeaves := 131072
// 	srsG1 := ReadG1Points("./g1.point.300000", uint64(numLeaves), 5)

// 	proveIndex := 89372
// 	pointHash := HashPoint(srsG1[proveIndex])
// 	mt := NewKzgMerkleTree(srsG1, numLeaves)

// 	fmt.Println(mt.VerifyProof(proveIndex, pointHash, mt.ProveIndex(proveIndex)))

// }

func NewG1MerkleTree(points []bls.G1Point, numLeaves int) *KzgMerkleTree {
	srsG1MerkleTree := [][][]byte{}
	srsG1MerkleTree = append(srsG1MerkleTree, [][]byte{})

	for i := 0; i < numLeaves; i++ {
		srsG1MerkleTree[0] = append(srsG1MerkleTree[0], HashG1Point(points[i]))
	}
	numLeaves /= 2
	for numLeaves > 0 {
		srsG1MerkleTree = append([][][]byte{[][]byte{}}, srsG1MerkleTree...)
		for j := 0; j < len(srsG1MerkleTree[1]); j += 2 {
			srsG1MerkleTree[0] = append(srsG1MerkleTree[0], crypto.Keccak256(
				append(
					srsG1MerkleTree[1][j],
					srsG1MerkleTree[1][j+1]...,
				)))
		}
		numLeaves /= 2
	}
	return &KzgMerkleTree{G1Points: points, Tree: srsG1MerkleTree}
}

func NewG2MerkleTree(points []bls.G2Point, numLeaves int) *KzgMerkleTree {
	srsG2MerkleTree := [][][]byte{}
	srsG2MerkleTree = append(srsG2MerkleTree, [][]byte{})

	for i := 0; i < numLeaves; i++ {
		srsG2MerkleTree[0] = append(srsG2MerkleTree[0], HashG2Point(points[i]))
	}
	numLeaves /= 2
	for numLeaves > 0 {
		srsG2MerkleTree = append([][][]byte{[][]byte{}}, srsG2MerkleTree...)
		for j := 0; j < len(srsG2MerkleTree[1]); j += 2 {
			srsG2MerkleTree[0] = append(srsG2MerkleTree[0], crypto.Keccak256(
				append(
					srsG2MerkleTree[1][j],
					srsG2MerkleTree[1][j+1]...,
				)))
		}
		numLeaves /= 2
	}
	return &KzgMerkleTree{G2Points: points, Tree: srsG2MerkleTree}
}

func (mt *KzgMerkleTree) ProveIndex(index int) [][]byte {
	proof := make([][]byte, 0)
	height := len(mt.Tree) - 1
	tmp := index
	for height > 0 {
		if tmp%2 == 0 {
			proof = append(proof, mt.Tree[height][tmp+1])
		} else {
			proof = append(proof, mt.Tree[height][tmp-1])
		}
		tmp /= 2
		height--
	}
	return proof
}

func VerifyProof(index int, leaf []byte, proof [][]byte, root []byte) bool {
	tmp := index
	calcHash := leaf
	for i := 0; i < len(proof); i++ {
		if tmp%2 == 0 {
			// fmt.Println("l", hex.EncodeToString(calcHash), hex.EncodeToString(proof[i]))
			calcHash = crypto.Keccak256(
				append(
					calcHash,
					proof[i]...,
				))
		} else {
			// fmt.Println("r", hex.EncodeToString(calcHash), hex.EncodeToString(proof[i]))
			calcHash = crypto.Keccak256(
				append(
					proof[i],
					calcHash...,
				))
		}
		// fmt.Println(hex.EncodeToString(calcHash))
		tmp /= 2
	}
	return bytes.Compare(calcHash, root) == 0
}

func HashG1Point(point bls.G1Point) []byte {
	pointBytes := make([]byte, 0)
	tmp := point.X.Bytes()
	for i := 0; i < len(tmp); i++ {
		pointBytes = append(pointBytes, tmp[i])
	}
	tmp = point.Y.Bytes()
	for i := 0; i < len(tmp); i++ {
		pointBytes = append(pointBytes, tmp[i])
	}

	return crypto.Keccak256(pointBytes)
}

func HashG2Point(point bls.G2Point) []byte {
	pointBytes := make([]byte, 0)
	tmp := point.X.A0.Bytes()
	for i := 0; i < len(tmp); i++ {
		pointBytes = append(pointBytes, tmp[i])
	}
	tmp = point.X.A1.Bytes()
	for i := 0; i < len(tmp); i++ {
		pointBytes = append(pointBytes, tmp[i])
	}
	tmp = point.Y.A0.Bytes()
	for i := 0; i < len(tmp); i++ {
		pointBytes = append(pointBytes, tmp[i])
	}
	tmp = point.Y.A1.Bytes()
	for i := 0; i < len(tmp); i++ {
		pointBytes = append(pointBytes, tmp[i])
	}

	return crypto.Keccak256(pointBytes)
}
