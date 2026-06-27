package str

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	knifer "github.com/imajinyun/knifer-go"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

// ToUTF8 converts data from the named charset to UTF-8.
func ToUTF8(data []byte, from string) ([]byte, error) {
	enc, err := lookupEncoding(from)
	if err != nil {
		return nil, err
	}
	if enc == nil {
		return append([]byte(nil), StripBOM(data)...), nil
	}
	out, err := io.ReadAll(transform.NewReader(bytes.NewReader(StripBOM(data)), enc.NewDecoder()))
	if err != nil {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "charset decode failed", Cause: err}
	}
	return out, nil
}

// FromUTF8 converts UTF-8 data to the named charset.
func FromUTF8(data []byte, to string) ([]byte, error) {
	enc, err := lookupEncoding(to)
	if err != nil {
		return nil, err
	}
	if enc == nil {
		return append([]byte(nil), data...), nil
	}
	out, err := io.ReadAll(transform.NewReader(bytes.NewReader(data), enc.NewEncoder()))
	if err != nil {
		return nil, &knifer.Error{Code: knifer.ErrCodeInvalidInput, Message: "charset encode failed", Cause: err}
	}
	return out, nil
}

func lookupEncoding(name string) (encoding.Encoding, error) {
	normalized := strings.ToLower(strings.NewReplacer("_", "-", " ", "").Replace(strings.TrimSpace(name)))
	switch normalized {
	case "", "utf8", "utf-8":
		return nil, nil
	case "gbk", "cp936":
		return simplifiedchinese.GBK, nil
	case "gb18030":
		return simplifiedchinese.GB18030, nil
	case "big5", "big-5":
		return traditionalchinese.Big5, nil
	case "shift-jis", "shiftjis", "sjis", "cp932":
		return japanese.ShiftJIS, nil
	case "euc-kr", "euckr":
		return korean.EUCKR, nil
	case "iso-8859-1", "latin1", "latin-1":
		return charmap.ISO8859_1, nil
	default:
		return nil, &knifer.Error{
			Code:    knifer.ErrCodeUnsupported,
			Message: fmt.Sprintf("charset %q is unsupported", name),
		}
	}
}
