// Copyright (c) 2020 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package shutil

import (
	"fmt"
	"strings"
)

// VariableMap is the interface that wraps the Get method.
//
// Get takes a variable name, and returns the value associated to the variable
// if it has one, along with whether the variable is present in the variable map
// or not.
type VariableMap interface{
	Get(variable string) (value string, present bool)
}

// SimpleVariableMap is a thin wrapper around map[string]string that implements
// VariableMap.
type SimpleVariableMap map[string]string

func (smap SimpleVariableMap) Get(variable string) (string, bool) {
	val, ok := smap[variable]
	return val, ok
}

// Substitute expands and substitutes shell variables in s, and returns
// the fully substituted string. It errors out if s contains variables
// that do not exist in the specified variable map.
//
// The syntax for variable substitution is a restricted variant to that of
// a POSIX shell:
//
// * Variables are denoted with ${variable_name}.
// * All characters except ":" and "}" are accepted in variable names.
// * ${variable:-default} expands to "default" if the variable is not defined
//   in the variable map, or the value of the variable otherwise.
// * ${variable:+alternate} expands to "alternate" if the variable is defined
//   in the variable map, or the empty string otherwise.
func Substitute(s string, vars VariableMap) (string, error) {
	var out strings.Builder
	start := 0
	for i := 0; i < len(s); i++ {
		if strings.HasPrefix(s[i:], "${") {
			subsStart := i

			i += 2
			delim := strings.IndexAny(s[i:], ":}")
			if delim == -1 {
				break
			}

			name := s[i : i+delim]
			var def *string

			if s[i+delim] == ':' {
				i += delim + 1
				delim = strings.IndexByte(s[i:], '}')
				if delim == -1 {
					break
				}
				slice := s[i : i+delim]
				def = &slice
			}

			out.WriteString(s[start:subsStart])
			value, present := vars.Get(name)

			if def == nil {
				if !present {
					return "", fmt.Errorf("undefined variable %q", name)
				}
			} else {
				deref := *def
				if deref == "" {
					deref = "\x00"
				}
				switch deref[0] {
				case '-':
					if !present {
						value = deref[1:]
					}
				case '+':
					if present {
						value = deref[1:]
					}
				default:
					return "", fmt.Errorf("malformed variable substitution %q", s[subsStart:i+delim+1])
				}
			}

			out.WriteString(value)

			i += delim + 1
			start = i
		}
	}
	out.WriteString(s[start:])
	return out.String(), nil
}
