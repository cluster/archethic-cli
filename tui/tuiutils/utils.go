package tuiutils

import archethic "github.com/archethic-foundation/libgo"

func GetHashAlgorithmName(h archethic.HashAlgo) string {
	switch h {
	case archethic.SHA256:
		return "SHA256"
	case archethic.SHA512:
		return "SHA512"
	case archethic.SHA3_256:
		return "SHA3_256"
	case archethic.SHA3_512:
		return "SHA3_512"
	case archethic.BLAKE2B:
		return "BLAKE2B"
	}
	panic("Unknown hash algorithm")
}

func GetCurveName(h archethic.Curve) string {
	switch h {
	case archethic.ED25519:
		return "ED25519"
	case archethic.P256:
		return "P256"
	case archethic.SECP256K1:
		return "SECP256K1"
	}
	panic("Unknown curve")
}
