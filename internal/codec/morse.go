package codec

import (
	"strings"
	"unicode"
)

var morseTable = map[rune]string{
	'A': ".-", 'B': "-...", 'C': "-.-.", 'D': "-..", 'E': ".", 'F': "..-.",
	'G': "--.", 'H': "....", 'I': "..", 'J': ".---", 'K': "-.-", 'L': ".-..",
	'M': "--", 'N': "-.", 'O': "---", 'P': ".--.", 'Q': "--.-", 'R': ".-.",
	'S': "...", 'T': "-", 'U': "..-", 'V': "...-", 'W': ".--", 'X': "-..-",
	'Y': "-.--", 'Z': "--..",
	'0': "-----", '1': ".----", '2': "..---", '3': "...--", '4': "....-",
	'5': ".....", '6': "-....", '7': "--...", '8': "---..", '9': "----.",
	'.': ".-.-.-", ',': "--..--", '?': "..--..", '\'': ".----.", '!': "-.-.--",
	'/': "-..-.", '(': "-.--.", ')': "-.--.-", '&': ".-...", ':': "---...",
	';': "-.-.-.", '=': "-...-", '+': ".-.-.", '-': "-....-", '_': "..--.-",
	'"': ".-..-.", '$': "...-..-", '@': ".--.-.",
}

var reverseMorseTable = buildReverseMorseTable()

// MorseEncode encodes supported ASCII letters, digits, and punctuation as Morse code.
func MorseEncode(s string) (string, error) {
	words := strings.Fields(s)
	encodedWords := make([]string, 0, len(words))
	for _, word := range words {
		codes := make([]string, 0, len(word))
		for _, r := range word {
			code, ok := morseTable[unicode.ToUpper(r)]
			if !ok {
				return "", invalidCodecInput("encode morse", errInvalidAlphabetChar)
			}
			codes = append(codes, code)
		}
		encodedWords = append(encodedWords, strings.Join(codes, " "))
	}
	return strings.Join(encodedWords, " / "), nil
}

// MorseDecode decodes Morse code using spaces between letters and "/" between words.
func MorseDecode(s string) (string, error) {
	if strings.TrimSpace(s) == "" {
		return "", nil
	}
	words := strings.Split(s, "/")
	decodedWords := make([]string, 0, len(words))
	for _, word := range words {
		letters := strings.Fields(word)
		var b strings.Builder
		for _, letter := range letters {
			r, ok := reverseMorseTable[letter]
			if !ok {
				return "", invalidCodecInput("decode morse", errInvalidAlphabetChar)
			}
			b.WriteRune(r)
		}
		decodedWords = append(decodedWords, b.String())
	}
	return strings.Join(decodedWords, " "), nil
}

func buildReverseMorseTable() map[string]rune {
	out := make(map[string]rune, len(morseTable))
	for r, code := range morseTable {
		out[code] = r
	}
	return out
}
