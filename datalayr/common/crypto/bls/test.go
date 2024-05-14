package bls

// import (
// 	"encoding/hex"
// 	"fmt"
// 	"math/big"

// 	"github.com/consensys/gnark-crypto/ecc/bn254"
// 	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
// )

// func main() {
// 	n := 15
// 	createRegistrationData := false
// 	publicKeys := make([]*bn254.G2Affine, 0)
// 	secretKeys := make([]*fp.Element, 0)
// 	var g2Gen bn254.G2Affine
// 	g2Gen.X.SetString("10857046999023057135944570762232829481370756359578518086990519993285655852781",
// 		"11559732032986387107991004021392285783925812861821192530917403151452391805634")
// 	g2Gen.Y.SetString("8495653923123431417604973247489272438418190587263600148770280649306958101930",
// 		"4082367875863433681332203403145435568316851327593401208105741076214120093531")
// 	// acc := g2Gen
// 	//generate n
// 	for i := 0; i < n; i++ {
// 		// sk, err := new(fp.Element).SetRandom()
// 		// if err != nil {
// 		//  fmt.Println(err)
// 		//  return
// 		// }
// 		sk := new(fp.Element).SetInt64(int64(436364636 + i*69))
// 		pk := new(bn254.G2Affine).ScalarMultiplication(&g2Gen, sk.ToBigIntRegular(new(big.Int)))
// 		publicKeys = append(publicKeys, pk)
// 		secretKeys = append(secretKeys, sk)
// 		// fmt.Println(sig.Y.A0.ToBigIntRegular(new(big.Int)))
// 		if createRegistrationData {
// 			fmt.Printf("registrationData.push(hex\"%s\");", hex.EncodeToString(makeRegistrationData(sk, pk)))
// 			// fmt.Println()
// 			// printG2(pk)
// 			// fmt.Println()
// 			// fmt.Println(acc.Add(pk, &acc))
// 		}
// 		// createRegistrationData(publicKeys[i])
// 		// fmt.Println()
// 	}
// 	//get message
// 	msgBytes, err := hex.DecodeString("0102030405060708091011121314151617181920")
// 	H := hashToCurve(msgBytes)
// 	secretKeys = append(secretKeys, new(fp.Element).SetOne())
// 	publicKeys = append(publicKeys, &g2Gen)
// 	sigs := make([]*bn254.G1Affine, 0)
// 	for i := 0; i < len(secretKeys); i++ {
// 		sigs = append(sigs, new(bn254.G1Affine).ScalarMultiplication(H, secretKeys[i].ToBigIntRegular(new(big.Int))))
// 	}
// 	//aggregate sigs
// 	aggSig := sigs[0]
// 	for i := 1; i < len(sigs); i++ {
// 		aggSig.Add(aggSig, sigs[i])
// 	}
// 	//aggregate public keys
// 	aggPk := publicKeys[0]
// 	for i := 1; i < len(publicKeys); i++ {
// 		aggPk.Add(aggPk, publicKeys[i])
// 	}
// 	//print pairings
// 	pairing1, err := bn254.Pair([]bn254.G1Affine{*aggSig, *H}, []bn254.G2Affine{*new(bn254.G2Affine).Neg(&g2Gen), *aggPk})
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	pairing2, err := bn254.Pair([]bn254.G1Affine{*aggSig, *H}, []bn254.G2Affine{*new(bn254.G2Affine).Neg(&g2Gen), *aggPk})
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	fmt.Println()
// 	fmt.Println("e(G1, Sigma) =", pairing1.String())
// 	if pairing1.Equal(pairing2.SetOne()) {
// 		fmt.Println("SIGNATURE IS VALID")
// 	} else {
// 		fmt.Println("SIGNATURE IS INVALID")
// 	}
// 	//
// 	//   input:
// 	//   input:
// 	//   input:
// 	//   input:
// 	//   input:
// 	//   input:
// 	//   input:
// 	//   input:
// 	//   input:
// 	//   input:
// 	//   input:
// 	// 11941283020102586554559929106587605041301217271111447561703124458776807385414
// 	// 17481617014045991712611664178572842151350982645517484551752615542557078246902
// 	// 18507428821816114421698399069438744284866101909563082454551586195885282320634
// 	// 20820493588973199354272631301248587752629863429201347184003644368113679196121
// 	// 3512517006108887301063578607317108977425754510174956792003926207778790018672
// 	// 1263326262781780932600377484793962587101562728383804037421955407439695092960
// 	// 7155561537864411538991615376457474334371827900888029310878886991084477170996
// 	// 10352977531892356631551102769773992282745949082157652335724669165983475588346
// 	// 11559732032986387107991004021392285783925812861821192530917403151452391805634
// 	// 10857046999023057135944570762232829481370756359578518086990519993285655852781
// 	// 17805874995975841540914202342111839520379459829704422454583296818431106115052
// 	// 13392588948715843804641432497768002650278120570034223513918757245338268106653
// 	// fmt.Println(H)
// 	fmt.Println(H.X.String())
// 	fmt.Println(H.Y.String())
// 	fmt.Println(aggPk.X.A1.String())
// 	fmt.Println(aggPk.X.A0.String())
// 	fmt.Println(aggPk.Y.A1.String())
// 	fmt.Println(aggPk.Y.A0.String())
// 	fmt.Println(aggSig.X.String())
// 	fmt.Println(aggSig.Y.String())
// 	fmt.Println(new(bn254.G2Affine).Neg(&g2Gen).X.A1.String())
// 	fmt.Println(new(bn254.G2Affine).Neg(&g2Gen).X.A0.String())
// 	fmt.Println(new(bn254.G2Affine).Neg(&g2Gen).Y.A1.String())
// 	fmt.Println(new(bn254.G2Affine).Neg(&g2Gen).Y.A0.String())
// 	fmt.Println(aggPk)
// 	fmt.Println(aggSig)
// 	/*
// 	   FULL CALLDATA FORMAT:
// 	   uint48 dumpNumber,
// 	   bytes32 headerHash,
// 	   uint32 numberOfNonSigners,
// 	   uint256[numberOfSigners][4] pubkeys of nonsigners,
// 	   uint32 apkIndex,
// 	   uint256[4] apk,
// 	   uint256[2] sigma
// 	*/
// 	// fmt.Println(aggPk.X.String(), aggPk.Y.String())
// 	// fmt.Println("e(apk, H(m)) =", pairing2.String())
// 	// byteStuff := aggPk.X.Bytes()
// 	// fmt.Println(hex.EncodeToString(byteStuff[:]))
// 	// byteStuff = aggPk.Y.Bytes()
// 	// fmt.Println(hex.EncodeToString(byteStuff[:]))
// 	// // printG2Point(&h)
// 	// byteStuff = h.X.A1.Bytes()
// 	// fmt.Println(hex.EncodeToString(byteStuff[:]))
// 	// byteStuff = h.X.A0.Bytes()
// 	// fmt.Println(hex.EncodeToString(byteStuff[:]))
// 	// byteStuff = h.Y.A1.Bytes()
// 	// fmt.Println(hex.EncodeToString(byteStuff[:]))
// 	// byteStuff = h.Y.A0.Bytes()
// 	// fmt.Println(hex.EncodeToString(byteStuff[:]))
// 	// //
// 	// byteStuff = g1Gen.X.Bytes()
// 	// fmt.Println(hex.EncodeToString(byteStuff[:]))
// 	// byteStuff = new(bn254.G1Affine).Neg(&g1Gen).Y.Bytes()
// 	// fmt.Println(hex.EncodeToString(byteStuff[:]))
// 	// // fmt.Println(g1Gen.X.String())
// 	// // fmt.Println(new(bn254.G1Affine).Neg(&g1Gen).Y.String())
// 	// // printG2Point(aggSig)
// 	// byteStuff = aggSig.X.A1.Bytes()
// 	// fmt.Println(hex.EncodeToString(byteStuff[:]))
// 	// byteStuff = aggSig.X.A0.Bytes()
// 	// fmt.Println(hex.EncodeToString(byteStuff[:]))
// 	// byteStuff = aggSig.Y.A1.Bytes()
// 	// fmt.Println(hex.EncodeToString(byteStuff[:]))
// 	// byteStuff = aggSig.Y.A0.Bytes()
// 	// fmt.Println(hex.EncodeToString(byteStuff[:]))
// }
