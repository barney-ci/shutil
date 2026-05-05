package matchspec

import (
	"testing"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		name      string
		patterns  Spec
		candidate string
		want      bool
	}{
		{
			name:      "Simple Glob Match",
			patterns:  []string{"src/**/*.ts"},
			candidate: "src/app/utils.ts",
			want:      true,
		},
		{
			name:      "Glob Match Case Insensitive",
			patterns:  []string{"SRC/**/*.TS"},
			candidate: "src/app/utils.ts",
			want:      true,
		},
		{
			name: "Negative Match Overrides Positive",
			patterns: []string{
				"src/**/*.ts",
				"!src/app/utils.ts",
			},
			candidate: "src/app/utils.ts",
			want:      false,
		},
		{
			name:      "Regex Match Case Insensitive",
			patterns:  []string{"/^SRC\\/.*\\.ts$/i"},
			candidate: "src/app/utils.ts",
			want:      true,
		},
		{
			name:      "Regex Match Case Sensitive Failure",
			patterns:  []string{"/^SRC\\/.*\\.ts$/"},
			candidate: "src/app/utils.ts",
			want:      false,
		},
		{
			name:      "Only Negative Patterns (Implicitly Allows Others)",
			patterns:  []string{"!src/other/**"},
			candidate: "src/app/utils.ts",
			want:      true,
		},
		{
			name:      "Only Negative Patterns (Rejection)",
			patterns:  []string{"!src/app/**"},
			candidate: "src/app/utils.ts",
			want:      false,
		},
		{
			name: "Multiple Positive Patterns (OR Logic)",
			patterns: []string{
				"docs/**",
				"src/**/*.go",
			},
			candidate: "src/main.go",
			want:      true,
		},
		{
			name:      "No Patterns Match",
			patterns:  []string{"docs/**"},
			candidate: "src/main.go",
			want:      false,
		},
		{
			name:      "Empty Pattern List (Defaults to True)",
			patterns:  []string{},
			candidate: "anything",
			want:      true,
		},
		{
			name: "Select most release branches but exclude specific ones",
			patterns: []string{
				"/^(master|main|.+-release-.+)$/",
				"!/^(mfw-release-6.0|mfw-release-6.1)$/",
			},
			candidate: "mfw-release-6.0",
			want:      false,
		},
		{
			name: "Select most release branches but exclude specific ones (allowed)",
			patterns: []string{
				"/^(master|main|.+-release-.+)$/",
				"!/^(mfw-release-6.0|mfw-release-6.1)$/",
			},
			candidate: "mfw-release-6.2",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Match(tt.patterns, tt.candidate)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		wantErr  bool
	}{
		{
			name: "Valid globs and regexes",
			patterns: []string{
				"src/**",
				"!/^test$/i",
			},
			wantErr: false,
		},
		{
			name:     "Invalid Regex Delimiter",
			patterns: []string{"/missing-slash"},
			wantErr:  true,
		},
		{
			name:     "Invalid Regex Syntax",
			patterns: []string{"/unclosed(paren/"},
			wantErr:  true,
		},
		{
			name:     "Invalid Glob Syntax",
			patterns: []string{"src/[unclosed"},
			wantErr:  true,
		},
		{
			name:     "Empty Pattern in List",
			patterns: []string{"src/**", ""},
			wantErr:  false, // Ignored, not error
		},
		{
			name:     "Just Negation Operator",
			patterns: []string{"!"},
			wantErr:  false, // effectively empty after stripping '!'
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.patterns)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
