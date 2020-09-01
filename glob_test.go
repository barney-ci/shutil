// Copyright © 2020 Arista Networks, Inc. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be found
// in the LICENSE file.

package shutil

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestGlobString(t *testing.T) {
	tcases := []struct{
		Pattern, File string
		Match bool
	}{
		{"", "", true},
		{"", "file", false},
		{"file", "", false},
		{"file", "file", true},

		{"?", "", false},
		{"?", "a", true},
		{"?", "ab", false},

		{"*", "", true},
		{"*", "file", true},
		{"*", "dir/file", false},
		{"*/", "dir", false},
		{"*/", "dir/", true},
		{"x/*/y", "x/z/y", true},
		{"x/*/y", "x/a/b/y", false},
		{"x/*/y", "x/y", true},
		{"*.ext", "file.ext", true},
		{"*.alt", "file.ext", false},
		{"file.*", "file.ext", true},
		{"dir.*", "file.ext", false},

		{"**", "", true},
		{"**", "file", true},
		{"**", "dir/file", true},
		{"dir/**", "dir", false},
		{"dir/**", "dir/", true},
		{"dir/**", "dir/x", true},
		{"dir/**", "dir/x/y", true},
		{"**/file", "file", true},
		{"**/file", "/file", true},
		{"**/file", "y/file", true},
		{"**/file", "x/y/file", true},
	}

	t.Run("Simple", func(t *testing.T) {
		for _, tc := range tcases {
			t.Run(tc.Pattern, func(t *testing.T) {
				ok, err := GlobMatch(tc.Pattern, tc.File)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if ok != tc.Match {
					if tc.Match {
						t.Fatalf("expected %q to match %q, but it didn't", tc.File, tc.Pattern)
					} else {
						t.Fatalf("expected %q to not match %q, but it did", tc.File, tc.Pattern)
					}
				}
			})
		}
	})

	rangeCases := []struct{
		Pattern, Accepted string
		Negated bool
	}{
		{"[a]", "a", false},
		{"[az]", "az", false},
		{"[a-z]", "abcdefghijklmnopqrstuvwxyz", false},
		{"[a-z0-9]", "abcdefghijklmnopqrstuvwxyz0123456789", false},
		{"[-z]", "-z", false},
		{"[z-]", "-z", false},
		{"[]]", "]", false},

		{"[!a]", "a", true},
		{"[!az]", "az", true},
		{"[!a-z]", "abcdefghijklmnopqrstuvwxyz", true},
		{"[!a-z0-9]", "abcdefghijklmnopqrstuvwxyz0123456789", true},
		{"[!-z]", "-z", true},
		{"[!z-]", "-z", true},
		{"[!]]", "]", true},
	}

	t.Run("Ranges", func(t *testing.T) {
		for _, tc := range rangeCases {
			t.Run(tc.Pattern, func(t *testing.T) {
				g, err := CompileGlob(tc.Pattern)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				for r := rune(1); r <= utf8.MaxRune; r++ {
					match := g.Match(string(r))
					expected := strings.IndexRune(tc.Accepted, r) != -1
					if tc.Negated {
						expected = !expected
					}

					if match != expected {
						if expected {
							t.Fatalf("expected %q to match %q, but it didn't", r, tc.Pattern)
						} else {
							t.Fatalf("expected %q to not match %q, but it did", r, tc.Pattern)
						}
					}
				}
			})
		}
	})
}
