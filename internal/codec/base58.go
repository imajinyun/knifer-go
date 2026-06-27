package codec

import (
	"bytes"
	"math/big"
)

// Base58Alphabet identifies a Base58 alphabet.
type Base58Alphabet string

const (
	// Base58BitcoinAlphabet is the Bitcoin Base58 alphabet.
	Base58BitcoinAlphabet Base58Alphabet = "bitcoin"
	// Base58FlickrAlphabet is the Flickr Base58 alphabet.
	Base58FlickrAlphabet Base58Alphabet = "flickr"
)

const (
	base58Bitcoin = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	base58Flickr  = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"
)

var bigRadix58 = big.NewInt(58)

// Base58Encode encodes bytes with the Bitcoin Base58 alphabet.
func Base58Encode(data []byte) string { return Base58EncodeWithAlphabet(data, Base58BitcoinAlphabet) }

// Base58EncodeWithAlphabet encodes bytes with a supported Base58 alphabet.
func Base58EncodeWithAlphabet(data []byte, alphabet Base58Alphabet) string {
	table := base58Alphabet(alphabet)
	x := new(big.Int).SetBytes(data)
	if x.Sign() == 0 {
		return leadingBase58Zeros(data, table)
	}

	var encoded []byte
	mod := new(big.Int)
	for x.Sign() > 0 {
		x.DivMod(x, bigRadix58, mod)
		encoded = append(encoded, table[mod.Int64()])
	}
	for _, b := range data {
		if b != 0 {
			break
		}
		encoded = append(encoded, table[0])
	}
	reverseBytes(encoded)
	return string(encoded)
}

// Base58Decode decodes a Bitcoin Base58 string.
func Base58Decode(s string) ([]byte, error) {
	return Base58DecodeWithAlphabet(s, Base58BitcoinAlphabet)
}

// Base58DecodeWithAlphabet decodes a Base58 string with a supported alphabet.
func Base58DecodeWithAlphabet(s string, alphabet Base58Alphabet) ([]byte, error) {
	table := base58Alphabet(alphabet)
	index := make(map[byte]int, len(table))
	for i := range table {
		index[table[i]] = i
	}

	x := big.NewInt(0)
	for i := 0; i < len(s); i++ {
		v, ok := index[s[i]]
		if !ok {
			return nil, invalidCodecInput("decode base58", errInvalidAlphabetChar)
		}
		x.Mul(x, bigRadix58)
		x.Add(x, big.NewInt(int64(v)))
	}

	out := x.Bytes()
	for i := 0; i < len(s) && s[i] == table[0]; i++ {
		out = append([]byte{0}, out...)
	}
	return out, nil
}

func base58Alphabet(alphabet Base58Alphabet) string {
	if alphabet == Base58FlickrAlphabet {
		return base58Flickr
	}
	return base58Bitcoin
}

func leadingBase58Zeros(data []byte, table string) string {
	if len(data) == 0 {
		return ""
	}
	return string(bytes.Repeat([]byte{table[0]}, len(data)))
}
