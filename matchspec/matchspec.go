package matchspec

import (
	"fmt"
	"regexp"
	"strings"

	"barney.ci/shutil"
)

type Spec []string

func (s Spec) Match(candidate string) (bool, error) {
	return Match(s, candidate)
}

func (s Spec) Validate() error {
	return Validate(s)
}

// Match returns true if the candidate string matches the list of patterns.
// Patterns may be globs (case-insensitive) or regexes delimited by slashes,
// optionally followed by 'i' for case-insensitive matching. Patterns prefixed
// with '!' are negations: a matched negation immediately returns false.
// If only negative patterns are present and none match, the result is true.
func Match(patterns Spec, candidate string) (bool, error) {
	hasPositivePatterns := false
	matchedAnyPositive := false

	for _, pattern := range patterns {
		isNegative := false

		if strings.HasPrefix(pattern, "!") {
			isNegative = true
			pattern = pattern[1:]
		}

		if pattern == "" {
			continue
		}

		if !isNegative {
			hasPositivePatterns = true
		}

		var matched bool
		var err error
		if strings.HasPrefix(pattern, "/") {
			matched, err = matchRegex(pattern, candidate)
		} else {
			matched, err = matchGlob(pattern, candidate)
		}
		if err != nil {
			return false, err
		}

		if isNegative && matched {
			// Negative match found: reject immediately
			return false, nil
		}
		if !isNegative && matched {
			// Positive matches can't be reported until the end (in case
			// there are also negative matches later in the list).
			matchedAnyPositive = true
		}
	}

	// If we only have negative patterns and none matched, return true.
	// If we have positive patterns, at least one must have matched.
	if hasPositivePatterns {
		return matchedAnyPositive, nil
	}

	return true, nil
}

func matchGlob(pattern, candidate string) (bool, error) {
	// Glob patterns are always case-insensitive.
	g, err := shutil.CompileGlob(strings.ToLower(pattern))
	if err != nil {
		return false, err
	}
	return g.Match(strings.ToLower(candidate)), nil
}

func matchRegex(pattern, candidate string) (bool, error) {
	trimPattern := pattern
	caseInsensitive := false

	// Check for case-insensitive flag '/i'
	if strings.HasSuffix(trimPattern, "/i") {
		trimPattern = strings.TrimSuffix(trimPattern, "/i")
		trimPattern = strings.TrimPrefix(trimPattern, "/")
		caseInsensitive = true
	} else if strings.HasSuffix(trimPattern, "/") {
		trimPattern = strings.TrimSuffix(trimPattern, "/")
		trimPattern = strings.TrimPrefix(trimPattern, "/")
	} else {
		return false, fmt.Errorf("invalid regex pattern: %q", pattern)
	}

	if caseInsensitive {
		trimPattern = "(?i)" + trimPattern
	}

	re, err := regexp.Compile(trimPattern)
	if err != nil {
		return false, err
	}

	return re.MatchString(candidate), nil
}

// Validate checks if the provided patterns are syntactically correct
// (valid regex or valid glob).
func Validate(patterns Spec) error {
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "!") {
			pattern = pattern[1:]
		}

		if pattern == "" {
			// Empty patterns are technically valid (just ignored)
			continue
		}

		var err error
		// We pass an empty candidate (""). We only care about the error return.
		if strings.HasPrefix(pattern, "/") {
			_, err = matchRegex(pattern, "")
		} else {
			_, err = matchGlob(pattern, "")
		}
		if err != nil {
			return err
		}
	}
	return nil
}
