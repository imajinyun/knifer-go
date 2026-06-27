package codec

import "math/big"

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var bigRadix62 = big.NewInt(62)

// Base62Encode encodes bytes with a URL-friendly Base62 alphabet.
func Base62Encode(data []byte) string {
	x := new(big.Int).SetBytes(data)
	if x.Sign() == 0 {
		return leadingBase62Zeros(data)
	}

	var encoded []byte
	mod := new(big.Int)
	for x.Sign() > 0 {
		x.DivMod(x, bigRadix62, mod)
		encoded = append(encoded, base62Alphabet[mod.Int64()])
	}
	for _, b := range data {
		if b != 0 {
			break
		}
		encoded = append(encoded, base62Alphabet[0])
	}
	reverseBytes(encoded)
	return string(encoded)
}

// Base62Decode decodes a Base62 string.
func Base62Decode(s string) ([]byte, error) {
	index := make(map[byte]int, len(base62Alphabet))
	for i := range base62Alphabet {
		index[base62Alphabet[i]] = i
	}

	x := big.NewInt(0)
	for i := 0; i < len(s); i++ {
		v, ok := index[s[i]]
		if !ok {
			return nil, invalidCodecInput("decode base62", errInvalidAlphabetChar)
		}
		x.Mul(x, bigRadix62)
		x.Add(x, big.NewInt(int64(v)))
	}

	out := x.Bytes()
	for i := 0; i < len(s) && s[i] == base62Alphabet[0]; i++ {
		out = append([]byte{0}, out...)
	}
	return out, nil
}

func leadingBase62Zeros(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	out := make([]byte, len(data))
	for i := range out {
		out[i] = base62Alphabet[0]
	}
	return string(out)
}

func reverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}
