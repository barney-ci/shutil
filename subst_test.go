// Copyright © 2022 Arista Networks, Inc. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be found
// in the LICENSE file.

package shutil

import (
	"testing"
)

func TestSubstitute(t *testing.T) {

	t.Run("Simple", func(t *testing.T) {

		tcases := []struct {
			In, Expected string
		}{
			{`${variable}`, "value"},
			{`${undefined:-default}`, "default"},
			{`${variable:+default}`, "default"},
			{`${variable/ue/or/}`, "valor"},
			{`${variable/^v(al)(u)e$/g\1li\2m}`, "gallium"},

			// These are unterminated variables and are not substituted
			{`${variable`, "${variable"},
			{`${variable:-`, "${variable:-"},
			{`${variable/foo/bar`, "${variable/foo/bar"},
		}

		vals := SimpleVariableMap{}
		vals["variable"] = "value"

		for _, tc := range tcases {
			t.Run(tc.In, func(t *testing.T) {
				actual, err := Substitute(tc.In, vals)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if actual != tc.Expected {
					t.Fatalf("expected %q, got %q", tc.Expected, actual)
				}
			})
		}
	})

	t.Run("Malformed", func(t *testing.T) {

		tcases := []struct {
			In string
		}{
			{`${undefined}`},
			{`${variable:invalid}`},
			{`${variable/invalid}`},
			{`${variable/}`},
		}

		vals := SimpleVariableMap{}
		vals["variable"] = "value"

		for _, tc := range tcases {
			t.Run(tc.In, func(t *testing.T) {
				actual, err := Substitute(tc.In, vals)
				if err == nil {
					t.Fatalf("unexpected success: subtituted to %q", actual)
				}
				t.Log(err)
			})
		}
	})

}
