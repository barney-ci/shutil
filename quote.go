// Copyright (c) 2020 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package shutil

import (
	"strings"
)

// needsQuote returns true if the given character in the given position in a word needs quoting.
func needsQuote(i int, ch rune) bool {
	if i == 0 && ch == '~' {
		return true
	}
	return ch == '`' || ch == '$' || ch == '&' || ch == '*' || ch == '(' || ch == ')' ||
		ch == '{' || ch == '[' || ch == '\\' || ch == '|' || ch == ' ' ||
		ch == ';' || ch == '\'' || ch == '"' || ch == '<' || ch == '>' || ch == '?'
}

// Quote returns a bash command approximating what a human would probably type to invoke the
// given argv array.
// For example, Quote([]string{"rm", "abc def", "hij"}) returns "rm 'abc def' hij".
func Quote(argv []string) string {
	var b strings.Builder
	first := true
	for _, arg := range argv {
		if first {
			first = false
		} else {
			b.WriteByte(' ')
		}
		sawApostrophe := false
		needQuoting := false
		for i, ch := range arg {
			if ch == '\'' {
				sawApostrophe = true
				break
			}
			if needsQuote(i, ch) {
				needQuoting = true
			}
		}
		if sawApostrophe {
			// Word contains an apostrophe.  Just backslash-escape everything.
			for i, ch := range arg {
				if needsQuote(i, ch) {
					b.WriteByte('\\')
				}
				b.WriteRune(ch)
			}
		} else {
			// Word contains no apostrophe.  So, use apostrophes to quote, if any
			// quoting is needed.
			if needQuoting {
				b.WriteByte('\'')
			}
			b.WriteString(arg)
			if needQuoting {
				b.WriteByte('\'')
			}
		}
	}
	return b.String()
}
