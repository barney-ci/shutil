// Copyright © 2020-2022 Arista Networks, Inc. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be found
// in the LICENSE file.

package shutil

import (
	"fmt"
	"regexp"
	"strings"
)

// VariableMap is the interface that wraps the Get method.
//
// Get takes a variable name, and returns the value associated to the variable
// if it has one, along with whether the variable is present in the variable map
// or not.
type VariableMap interface {
	Get(variable string) (value string, present bool)
}

// SimpleVariableMap is a thin wrapper around map[string]string that implements
// VariableMap.
type SimpleVariableMap map[string]string

func (smap SimpleVariableMap) Get(variable string) (string, bool) {
	val, ok := smap[variable]
	return val, ok
}

var reGroup = regexp.MustCompile(`\\([0-9]+)`)

// Substitute expands and substitutes shell variables in s, and returns
// the fully substituted string. It errors out if s contains variables
// that do not exist in the specified variable map.
//
// The syntax for variable substitution is a restricted variant to that of
// a POSIX shell:
//
//  - Variables are denoted with ${variable_name}.
//  - All characters except ":" and "}" are accepted in variable names.
//  - ${variable:-default} expands to "default" if the variable is not defined
//    in the variable map, or the value of the variable otherwise.
//  - ${variable:+alternate} expands to "alternate" if the variable is defined
//    in the variable map, or the empty string otherwise.
//  - ${variable/re/subst/} expands to the variable, with a regexp replacement.
//    for instance, ${variable/^([^:]*):/\1/}, where variable=foo:bar, expands
//    to foo.
//
// If the passed VariableMap implements CanSubstitute(key string) bool, then
// the method is called to determine whether the variable is to be substituted.
// If the method returns false, the variable is left untouched and is output
// as-is into the result.
func Substitute(s string, vars VariableMap) (string, error) {

	type CanSubstitute interface {
		CanSubstitute(key string) bool
	}

	cansubst := func(key string) bool { return true }
	if cs, ok := vars.(CanSubstitute); ok {
		cansubst = cs.CanSubstitute
	}

	var out strings.Builder
	start := 0
outer:
	for i := 0; i < len(s); i++ {
		if strings.HasPrefix(s[i:], "${") {
			subsStart := i

			i += 2
			delim := strings.IndexAny(s[i:], ":/}")
			if delim == -1 {
				break
			}

			name := s[i : i+delim]
			var def *string

			switch s[i+delim] {
			case ':':
				i += delim + 1
				delim = strings.IndexByte(s[i:], '}')
				if delim == -1 {
					break outer
				}
				slice := s[i : i+delim]
				def = &slice
			case '/':
				i += delim
				j := i

				count := 1
				for ; j < len(s) && count < 3; j++ {
					switch s[j] {
					case '\\':
						j++
					case '/':
						count++
					}
				}
				if count != 3 {
					return "", fmt.Errorf("malformed regexp substitution %q: must be of the form ${variable/regexp/replace}", s[subsStart:j])
				}
				d := strings.IndexByte(s[j:], '}')
				if d == -1 {
					break outer
				}
				j += d
				slice := s[i:j]
				def = &slice

				i = j
				delim = 0
			case '}':
			default:
				break outer
			}

			if !cansubst(name) {
				i += delim + 1
				continue
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
				case '/':
					// This is a regexp substitution

					i := 0
					parts := strings.FieldsFunc(*def, func(r rune) bool {
						if r != '/' {
							return false
						}
						return i == 0 || (*def)[i-1] != '\\'
					})

					if len(parts) != 2 {
						return "", fmt.Errorf("malformed regexp substitution %q: must be of the form /regexp/replace", *def)
					}

					re, err := regexp.Compile(parts[0])
					if err != nil {
						return "", err
					}

					value = re.ReplaceAllString(value, reGroup.ReplaceAllString(parts[1], `${$1}`))
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
