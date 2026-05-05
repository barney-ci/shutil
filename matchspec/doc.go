// Package matchspec implements a string pattern matching specification for
// filtering lists of strings using glob and regular expression patterns,
// including negation support.
//
// # Overview
//
// This package is a Go implementation of the string pattern matching concept
// defined by the Renovate project:
//
//	https://docs.renovatebot.com/string-pattern-matching/
//
// The [Spec] type represents an ordered list of pattern strings. Callers build
// a Spec from a []string and then call [Match] (or [Spec.Match]) to test
// whether a candidate string satisfies the spec.
//
// # Pattern Types
//
// Each element of a [Spec] is one of two pattern types:
//
//   - Glob pattern — any string that does not begin with '/'. Evaluated with
//     the glob rules provided by [barney.ci/shutil.CompileGlob]. Glob matching
//     is always case-insensitive: both the pattern and the candidate are
//     lowercased before comparison.
//
//   - Regex pattern — a string delimited by forward slashes, e.g. /^abc/.
//     The closing delimiter may be followed by 'i' (/^abc/i) to make the match
//     case-insensitive. Regex patterns are case-sensitive by default. Patterns
//     that begin with '/' but do not end with '/' or '/i' are rejected as
//     invalid.
//
// # Negation
//
// Any pattern may be prefixed with '!' to negate it. A negated pattern that
// matches the candidate causes [Match] to return false immediately, regardless
// of any positive patterns that may have already matched. Positive patterns
// have OR semantics: if any positive pattern matches and no negative pattern
// matches, the result is true.
//
// When the Spec contains only negative patterns, the implicit default is true:
// the candidate is accepted unless at least one negative pattern matches.
//
// An empty Spec (or a Spec containing only empty strings) always returns true.
//
// # Examples
//
// Simple glob match:
//
//	s := matchspec.Spec{"src/**/*.go"}
//	ok, _ := s.Match("src/pkg/foo.go") // true
//	ok, _ = s.Match("vendor/pkg/foo.go") // false
//
// Negation — include everything except a subtree:
//
//	s := matchspec.Spec{"!vendor/**"}
//	ok, _ := s.Match("src/main.go") // true  (no positive patterns; negation didn't fire)
//	ok, _ = s.Match("vendor/lib/x.go") // false (negation fired)
//
// Mixed positive and negative patterns:
//
//	s := matchspec.Spec{"src/**", "!src/generated/**"}
//	ok, _ := s.Match("src/api/handler.go") // true
//	ok, _ = s.Match("src/generated/pb.go") // false
//
// Regex with case-insensitive flag:
//
//	s := matchspec.Spec{"/^v\\d+\\.\\d+/i"}
//	ok, _ := s.Match("V1.2-alpha") // true
//
// # Differences from the Renovate Specification
//
// This implementation diverges from Renovate's reference implementation in a
// few areas:
//
//   - Regex engine: Renovate uses JavaScript's re2 package. This package uses
//     Go's standard [regexp] package, which also implements RE2 syntax. The
//     practical difference is that Go's RE2 enforces stricter syntax in some
//     edge cases; patterns that are valid in one engine may need adjustment in
//     the other.
//
//   - Glob engine: Renovate evaluates globs with the JavaScript minimatch
//     library. This package uses [barney.ci/shutil.CompileGlob]. The two
//     engines share the common glob conventions (*, **, ?) but may differ on
//     unusual edge cases such as brace expansion or character classes.
//
//   - The '*' special case: Renovate documents '*' as a standalone "match
//     everything" token that cannot be combined with other patterns. This
//     implementation treats '*' as an ordinary glob that happens to match any
//     string without a path separator. Combining '*' with other patterns is
//     not restricted, though the result is the same in practice.
package matchspec
