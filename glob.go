// Copyright (c) 2020 Arista Networks, Inc.  All rights reserved.
// Arista Networks, Inc. Confidential and Proprietary.

package shutil

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	ErrUnterminatedClass = errors.New("unterminated character class")
)

// GlobError represents a syntax error for a specific glob pattern.
type GlobError struct {

	// Pattern is the erroneous glob pattern.
	Pattern string

	// Index is the index where the invalid character resides.
	Index   int

	// Err is the concrete underlying error.
	Err     error
}

func (err *GlobError) Error() string {
	return fmt.Sprintf("glob error: in %q at index %d: %v", err.Pattern, err.Index, err.Err)
}

func (err *GlobError) Unwrap() error {
	return err.Err
}

const (
	eof rune = 0
)

type parseFunc func(*globParser) parseFunc

type globParser struct {
	in string
	index, width int
	neg bool
	err error
	out strings.Builder
	choiceNest int
}

func (l *globParser) next() (r rune) {
	r, l.width = utf8.DecodeRuneInString(l.in[l.index:])
	if l.width == 0 {
		return eof
	}
	l.index += l.width
	return r
}

func (l *globParser) back() {
	l.index -= l.width
	l.width = 0
}

func (l *globParser) peek() rune {
	r := l.next()
	l.back()
	return r
}

func parseMain(p *globParser) parseFunc {
	r := p.next()

	switch r {
	case eof:
		return nil
	case '\\':
		if next := p.next(); next == eof {
			goto literal
		} else {
			p.out.WriteRune(next)
		}
	case '!':
		if p.index - p.width != 0 {
			goto literal
		}
		p.neg = !p.neg
	case '.', '(', ')', '^', '$', '|', '+':
		p.out.WriteRune('\\')
		goto literal
	case '{':
		p.out.WriteRune('(')
		p.choiceNest++
	case ',':
		if p.choiceNest == 0 {
			goto literal
		}
		p.out.WriteRune('|')
	case '}':
		if p.choiceNest == 0 {
			goto literal
		}
		p.out.WriteRune(')')
		p.choiceNest--
	case '[':
		return parseClass
	case '?':
		p.out.WriteString(`[^/]`)
	case '*':
		if strings.HasPrefix(p.in[p.index:], `*/`) {
			// we either have **/ or /**/ -- this means match zero or more
			// leading directories.
			p.out.WriteString(`(|[^\0]*/)`)
			p.index += len(`*/`)
		} else if p.peek() == '*' {
			// we either have /** or ** -- the former means "anything under X",
			// while the latter means "everything", both including nothing.
			p.out.WriteString(`[^\0]*/?`)
			p.next()
		} else if p.peek() == '/' {
			p.out.WriteString(`([^/]*/)?`)
			p.next()
		} else {
			p.out.WriteString(`[^/]*`)
		}
	default:
		goto literal
	}
	return parseMain

literal:
	p.out.WriteRune(r)
	return parseMain
}

func parseClass(p *globParser) parseFunc {
	p.out.WriteRune('[')
	start := p.index
	for {
		r := p.next()

		switch r {
		case eof:
			p.err = &GlobError{Pattern: p.in, Index: p.index, Err: ErrUnterminatedClass}
			return nil
		case '\\':
			switch next := p.next(); next {
			case eof:
				goto literal
			case '\\', '-', '^', '[', ']':
				// We still need to escape these
				p.out.WriteRune('\\')
				p.out.WriteRune(next)
			default:
				p.out.WriteRune(next)
			}
		case '!':
			if p.index - start - p.width == 0 {
				p.out.WriteRune('^')
			} else {
				goto literal
			}
		case '[':
			p.out.WriteRune('\\')
			p.out.WriteRune(r)
		case ']':
			p.out.WriteRune(r)
			return parseMain
		default:
			goto literal
		}
		continue

	literal:
		p.out.WriteRune(r)
	}
}

// Glob represents a compiled glob pattern. The supported syntax is mostly the
// same as glob(7), with the following extensions:
//
// * Curly brace expansion is supported. "{a,b,c}" matches the strings "a", "b", and "c".
// * A double star ("**") is supported to match any pathname component and their children.
//   For instance, "dir/*" matches "dir/file" but not "dir/dir/file", while "dir/**" matches both.
// * If the pattern starts with "!", the whole pattern is negated. If "!" appears later in the
//   pattern, it is treated as a literal "!".
type Glob struct {
	pattern string
	re      *regexp.Regexp
	negated bool
}

// CompileGlob compiles the specified pattern into a Glob object.
//
// See the documentation of the Glob type for more details on the supported syntax.
func CompileGlob(pattern string) (*Glob, error) {
	p := globParser{in: pattern}
	p.out.WriteString(`^(?s)`)
	for state := parseMain; state != nil; state = state(&p) {
		continue
	}
	if p.err != nil {
		return nil, p.err
	}
	p.out.WriteRune('$')
	re, err := regexp.Compile(p.out.String())
	if err != nil {
		return nil, err
	}
	return &Glob{pattern, re, p.neg}, nil
}

// MustCompileGlob is like CompileGlob, but panics if the function returned an error.
func MustCompileGlob(pattern string) *Glob {
	glob, err := CompileGlob(pattern)
	if err != nil {
		panic(err)
	}
	return glob
}

// Match returns whether data matches the glob pattern.
func (g *Glob) Match(data string) bool {
	return g.re.MatchString(data)
}

// Match returns whether the specified FileInfo matches the glob pattern.
//
// Generally, the name of the FileInfo is checked against the pattern. If the FileInfo represents
// a directory, the name followed by "/" is also checked against the pattern.
//
// This behaviour allows for a pattern like "*/" to *only* match directories.
// If this is not desirable, use MatchName instead.
func (g *Glob) MatchInfo(info os.FileInfo) bool {
	match := g.Match(info.Name())
	if info.IsDir() {
		match = match || g.Match(info.Name() + "/")
	}
	return match
}

// A Namer represents types that have a Name. Notable types that implement
// this interface are *os.File and os.FileInfo.
type Namer interface {
	Name() string
}

// Match returns whether the specified Namer matches the glob pattern. It is equivalent to
// Match(namer.Name()).
//
// This function is there for convenience as both *os.File and os.FileInfo implement Name().
func (g *Glob) MatchName(namer Namer) bool {
	return g.Match(namer.Name())
}

func (g *Glob) String() string {
	return g.pattern
}

// GlobMatch compiles pattern, and then returns Glob.Match(data).
func GlobMatch(pattern, data string) (bool, error) {
	g, err := CompileGlob(pattern)
	if err != nil {
		return false, err
	}
	return g.Match(data), nil
}

// GlobMatch compiles pattern, and then returns Glob.MatchName(namer).
func GlobMatchName(pattern string, namer Namer) (bool, error) {
	g, err := CompileGlob(pattern)
	if err != nil {
		return false, err
	}
	return g.MatchName(namer), nil
}

// GlobMatch compiles pattern, and then returns Glob.MatchInfo(info).
func GlobMatchInfo(pattern string, info os.FileInfo) (bool, error) {
	g, err := CompileGlob(pattern)
	if err != nil {
		return false, err
	}
	return g.MatchInfo(info), nil
}
