package moneroutil

import (
	"fmt"
)

func CreateKeyImage(pubSpendKey, secSpendKey, secViewKey, txPubKey *Key, outIndex uint64) (*Key, error) {
	derivation, ok := GenerateKeyDerivation(txPubKey, secViewKey)
	if !ok {
		return nil, fmt.Errorf("generate key derivation failed")
	}

	derivedPubKey, ok := DerivePublicKey(&derivation, outIndex, pubSpendKey)
	if !ok {
		return nil, fmt.Errorf("derive public key failed")
	}

	derivedPriKey := DeriveSecretKey(&derivation, outIndex, secSpendKey)
	if *derivedPriKey.PubKey() != derivedPubKey {
		return nil, fmt.Errorf("derived secret key doesn't match derived public key")
	}

	keyImage := GenerateKeyImage(&derivedPriKey)
	return &keyImage, nil
}

func GenerateKeyDerivation(pubKey, secKey *Key) (keyDerivation Key, ok bool) {
	point := new(ExtendedGroupElement)
	ok = point.FromBytes(pubKey)
	if !ok {
		return
	}

	point2 := new(ProjectiveGroupElement)
	GeScalarMult(point2, secKey, point)

	point3 := new(CompletedGroupElement)
	GeMul8(point3, point2)
	point3.ToProjective(point2)
	point2.ToBytes(&keyDerivation)
	ok = true
	return
}

func derivationToScalar(derivation *Key, outIndex uint64) (scalar Key) {
	data := append((*derivation)[:], Uint64ToBytes(outIndex)...)
	scalar = Key(Keccak256(data))
	ScReduce32(&scalar)
	return
}

func DerivePublicKey(derivation *Key, outIndex uint64, base *Key) (derivedKey Key, ok bool) {
	point1 := new(ExtendedGroupElement)
	ok = point1.FromBytes(base)
	if !ok {
		return
	}

	scalar := derivationToScalar(derivation, outIndex)
	point2 := new(ExtendedGroupElement)
	GeScalarMultBase(point2, &scalar)

	point3 := new(CachedGroupElement)
	point2.ToCached(point3)

	point4 := new(CompletedGroupElement)
	geAdd(point4, point1, point3)

	point5 := new(ProjectiveGroupElement)
	point4.ToProjective(point5)
	point5.ToBytes(&derivedKey)
	ok = true
	return
}

func DeriveSecretKey(derivation *Key, outIndex uint64, base *Key) (derivedKey Key) {
	scalar := derivationToScalar(derivation, outIndex)
	ScAdd(&derivedKey, base, &scalar)
	return
}

func GenerateKeyImage(privKey *Key) (keyImage Key) {
	point := privKey.PubKey().HashToEC()
	keyImagePoint := new(ProjectiveGroupElement)
	GeScalarMult(keyImagePoint, privKey, point)
	// convert key Image point from Projective to Extended
	// in order to precompute
	keyImagePoint.ToBytes(&keyImage)
	return
}
