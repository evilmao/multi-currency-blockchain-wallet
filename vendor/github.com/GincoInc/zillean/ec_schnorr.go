package zillean

import (
	"bytes"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"math/big"

	crypto "github.com/GincoInc/go-crypto"
)

/*
ECSchnorr represents a EC-Schnorr object.

See https://www.bsi.bund.de/SharedDocs/Downloads/EN/BSI/Publications/TechGuidelines/TR03111/BSI-TR-03111_pdf.pdf?__blob=publicationFile&v=1 for more information about the EC-Schnorr signature implementation.
*/
type ECSchnorr struct {
	Curve elliptic.Curve
}

// NewECSchnorr returns a new ECSchnorr.
func NewECSchnorr() *ECSchnorr {
	return &ECSchnorr{
		Curve: crypto.Secp256k1(),
	}
}

// GeneratePrivateKey generates a new private key for Schnorr signature.
func (ecs *ECSchnorr) GeneratePrivateKey() []byte {
	b := make([]byte, 32)
	for {
		rand.Read(b)
		_b := new(big.Int).SetBytes(b)
		if _b.Cmp(big.NewInt(0)) == 1 && _b.Cmp(ecs.Curve.Params().N) == -1 {
			break
		}
	}
	return b
}

// GetPublicKey returns the public key of a corresponding private key.
func (ecs *ECSchnorr) GetPublicKey(privKey []byte, compress bool) []byte {
	pubx, puby := ecs.Curve.ScalarBaseMult(privKey)

	return crypto.Marshal(ecs.Curve, pubx, puby, compress)
}

// Sign returns the signature (r, s) on a given message.
func (ecs *ECSchnorr) Sign(privKey, pubKey, msg []byte) ([]byte, []byte) {
	for {
		k, err := generateDRN(msg)
		if err != nil {
			continue
		}
		r, s, err := ecs.trySign(privKey, pubKey, k, msg)
		if err == nil {
			return r, s
		}
	}
}

// trySign tries to return the signature (r, s) on a given message.
// The algorithm takes the following step:
// 1. Take a radom k as an input
// 2. Compute the commitment Q = kG, where  G is the base point
// 3. Compute the challenge r = H(Q, pubKey, msg)
// 4. Compute s = k - r * privKey mod n
// 5. Signature on m is (r, s)
func (ecs *ECSchnorr) trySign(privKey, pubKey, k, msg []byte) ([]byte, []byte, error) {
	// 1. Take a radom k as an input
	_k := new(big.Int).SetBytes(k)
	if _k.Cmp(big.NewInt(0)) == 0 || _k.Cmp(ecs.Curve.Params().N) >= 0 {
		return nil, nil, errors.New("Invalid k")
	}

	// 2. Compute the commitment Q = kG, where  G is the base point
	Qx, Qy := ecs.Curve.ScalarBaseMult(k)
	Q := crypto.Compress(ecs.Curve, Qx, Qy)

	// 3. Compute the challenge r = H(Q, pubKey, msg)
	r := new(big.Int).SetBytes(hash(Q, pubKey, msg))
	r = r.Mod(r, ecs.Curve.Params().N)
	sk := new(big.Int).SetBytes(privKey)
	_r := *r

	// 4. Compute s = k - r * privKey mod n
	s := new(big.Int).Mod(_r.Sub(new(big.Int).SetBytes(k), _r.Mul(&_r, sk)), ecs.Curve.Params().N)
	if s.Cmp(big.NewInt(0)) == 0 {
		return nil, nil, errors.New("Invalid s")
	}

	return r.Bytes(), s.Bytes(), nil
}

// Verify returns a boolean that implies whether a given signature is successfully verified or not.
// The algorithm takes the following steps:
// 1. Compute Q = sG + r * pubKey
// 2. r' = H(Q, kpub, m)
// 3. return r' == r
func (ecs *ECSchnorr) Verify(r, s, pubKey, msg []byte) bool {
	pkx, pky := elliptic.Unmarshal(ecs.Curve, pubKey)
	rpkx, rpky := ecs.Curve.ScalarMult(pkx, pky, r)
	sGx, sGy := ecs.Curve.ScalarBaseMult(s)
	Qx, Qy := ecs.Curve.Add(sGx, sGy, rpkx, rpky)
	Q := crypto.Compress(ecs.Curve, Qx, Qy)

	return bytes.Equal(r, hash(Q, pubKey, msg))
}
