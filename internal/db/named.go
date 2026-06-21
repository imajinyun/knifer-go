package db

import (
	"strings"
	"unicode"
)

// NamedSQL is a parsed named-parameter SQL statement.
type NamedSQL struct {
	SQL    string
	Params []any
	Names  []string
}

// ParseNamed replaces :name parameters with dialect placeholders.
func ParseNamed(query string, args map[string]any, dialect Dialect) (NamedSQL, error) {
	var b strings.Builder
	params := make([]any, 0)
	names := make([]string, 0)
	inSingle := false
	inDouble := false
	for i := 0; i < len(query); i++ {
		ch := query[i]
		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			b.WriteByte(ch)
			continue
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
			b.WriteByte(ch)
			continue
		}
		if ch != ':' || inSingle || inDouble || i+1 >= len(query) || query[i+1] == ':' || (i > 0 && query[i-1] == ':') || !isNameStart(rune(query[i+1])) {
			b.WriteByte(ch)
			continue
		}
		j := i + 2
		for j < len(query) && isNamePart(rune(query[j])) {
			j++
		}
		name := query[i+1 : j]
		value, ok := args[name]
		if !ok {
			return NamedSQL{}, invalidInputf("db: missing named parameter %q", name)
		}
		params = append(params, value)
		names = append(names, name)
		b.WriteString(dialect.placeholder(len(params)))
		i = j - 1
	}
	return NamedSQL{SQL: b.String(), Params: params, Names: names}, nil
}

func isNameStart(r rune) bool { return r == '_' || unicode.IsLetter(r) }

func isNamePart(r rune) bool { return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) }
