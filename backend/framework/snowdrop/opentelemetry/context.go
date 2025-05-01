package opentelemetry

import "strings"

func BuildSpanName(name string, args ...string) string {
	if name == "" && len(args) == 0 {
		panic("BuildSpanName: no name or arguments provided")
	}

	var b strings.Builder

	if name != "" {
		b.WriteString(name)
	}

	for _, arg := range args {
		if arg == "" {
			continue
		}

		if b.Len() > 0 {
			b.WriteByte('.')
		}

		b.WriteString(arg)
	}

	if b.Len() == 0 {
		panic("BuildSpanName: all parts are empty")
	}

	return b.String()
}
