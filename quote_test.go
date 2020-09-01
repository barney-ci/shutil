// Copyright © 2020 Arista Networks, Inc. All rights reserved.
//
// Use of this source code is governed by the MIT license that can be found
// in the LICENSE file.

package shutil

import (
	"strconv"
	"testing"
)

func TestQuote(t *testing.T) {
	for i, tc := range []struct {
		argv   []string
		quoted string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"hello"}, "hello"},
		{[]string{"hello", "world"}, "hello world"},
		{[]string{"he llo"}, "'he llo'"},
		{[]string{"he llo", "wo rld"}, "'he llo' 'wo rld'"},
		{[]string{"he l'lo", "wo r'ld"}, "he\\ l\\'lo wo\\ r\\'ld"},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			act := Quote(tc.argv)
			if act != tc.quoted {
				t.Errorf("quoting %q: got %q, expected %q", tc.argv, act, tc.quoted)
			}
		})
	}
}
