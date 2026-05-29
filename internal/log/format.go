package log

import (
	"fmt"
	"strings"
)

// renderLogMessage supports Hutool-style "{}" placeholders.
// When template contains "{}", arguments are replaced positionally;
// otherwise it falls back to fmt.Sprintf.
//
//	renderLogMessage("hello {}", "world") // "hello world"
//	renderLogMessage("a=%d", 1)           // "a=1"
func renderLogMessage(template string, args ...any) string {
	if template == "" {
		return concatArgs(args...)
	}
	if strings.Contains(template, "{}") {
		return replacePlaceholders(template, args...)
	}
	if len(args) == 0 {
		return template
	}
	return sprintfLogMessage(template, args)
}

func sprintfLogMessage(template string, args []any) string {
	return fmt.Sprintf(template, args...)
}

// replacePlaceholders replaces "{}" placeholders in argument order.
func replacePlaceholders(format string, args ...any) string {
	var b strings.Builder
	b.Grow(len(format))
	idx := 0
	for i := 0; i < len(format); i++ {
		if i+1 < len(format) && format[i] == '{' && format[i+1] == '}' {
			if idx < len(args) {
				fmt.Fprint(&b, args[idx])
				idx++
			} else {
				b.WriteString("{}")
			}
			i++
			continue
		}
		b.WriteByte(format[i])
	}
	return b.String()
}

// concatArgs concatenates arguments into a single string, matching fmt.Sprint.
func concatArgs(args ...any) string {
	if len(args) == 0 {
		return ""
	}
	return fmt.Sprint(args...)
}
