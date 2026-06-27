package codec

// ROT13 applies the ROT13 substitution to ASCII letters.
func ROT13(s string) string { return ROTN(s, 13) }

// ROT47 applies the ROT47 substitution to printable ASCII characters.
func ROT47(s string) string {
	out := []rune(s)
	for i, r := range out {
		if r >= 33 && r <= 126 {
			out[i] = 33 + ((r-33)+47)%94
		}
	}
	return string(out)
}

// ROTN applies a Caesar shift to ASCII letters.
func ROTN(s string, n int) string {
	n %= 26
	if n < 0 {
		n += 26
	}
	out := []rune(s)
	for i, r := range out {
		switch {
		case r >= 'a' && r <= 'z':
			out[i] = 'a' + ((r-'a')+rune(n))%26
		case r >= 'A' && r <= 'Z':
			out[i] = 'A' + ((r-'A')+rune(n))%26
		}
	}
	return string(out)
}
