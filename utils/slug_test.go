package utils

import (
	"os"
	"strings"
	"testing"
)

func TestNormalizeSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "valid company name",
			input:    "Acme Corporation",
			expected: "acme-corporation",
			hasError: false,
		},
		{
			name:     "name with underscores",
			input:    "my_company_name",
			expected: "my-company-name",
			hasError: false,
		},
		{
			name:     "name with special characters",
			input:    "Company@#$%Name",
			expected: "companyname",
			hasError: false,
		},
		{
			name:     "name with multiple spaces",
			input:    "My   Company   Name",
			expected: "my-company-name",
			hasError: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
			hasError: true,
		},
		{
			name:     "too short",
			input:    "ab",
			expected: "",
			hasError: true,
		},
		{
			name:     "too long",
			input:    "this-is-a-very-long-company-name-that-exceeds-limit",
			expected: "",
			hasError: true,
		},
		{
			name:     "starts with dash",
			input:    "-company",
			expected: "company",
			hasError: false,
		},
		{
			name:     "ends with dash",
			input:    "company-",
			expected: "company",
			hasError: false,
		},
		{
			name:     "valid minimum length",
			input:    "abc",
			expected: "abc",
			hasError: false,
		},
		{
			name:     "valid maximum length",
			input:    "this-is-exactly-thirty-chars",
			expected: "this-is-exactly-thirty-chars",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeSlug(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input %q, but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

func TestSuggestSlugAlternatives_Default(t *testing.T) {
	// Clear any environment variable
	os.Unsetenv("SLUG_SUGGESTION_SUFFIXES")

	base := "acme"
	suggestions := SuggestSlugAlternatives(base)

	// Should have suggestions with default suffixes
	expectedCount := len(defaultSlugSuffixes)
	if len(suggestions) != expectedCount {
		t.Errorf("Expected %d suggestions, got %d", expectedCount, len(suggestions))
	}

	// Check that all suggestions are valid
	for _, suggestion := range suggestions {
		if len(suggestion) > 30 {
			t.Errorf("Suggestion %q exceeds 30 character limit", suggestion)
		}
		if !strings.HasPrefix(suggestion, base) {
			t.Errorf("Suggestion %q doesn't start with base %q", suggestion, base)
		}
	}
}

func TestSuggestSlugAlternatives_CustomEnvironment(t *testing.T) {
	tests := []struct {
		name             string
		envValue         string
		expectedSuffixes []string
	}{
		{
			name:             "custom suffixes with dashes",
			envValue:         "-corp,-inc,-ltd",
			expectedSuffixes: []string{"-corp", "-inc", "-ltd"},
		},
		{
			name:             "custom suffixes without dashes",
			envValue:         "corp,inc,ltd",
			expectedSuffixes: []string{"-corp", "-inc", "-ltd"},
		},
		{
			name:             "mixed format",
			envValue:         "-corp,inc,-ltd",
			expectedSuffixes: []string{"-corp", "-inc", "-ltd"},
		},
		{
			name:             "with spaces",
			envValue:         " -corp , inc , -ltd ",
			expectedSuffixes: []string{"-corp", "-inc", "-ltd"},
		},
		{
			name:             "empty values filtered",
			envValue:         "-corp,,inc,,-ltd",
			expectedSuffixes: []string{"-corp", "-inc", "-ltd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv("SLUG_SUGGESTION_SUFFIXES", tt.envValue)
			defer os.Unsetenv("SLUG_SUGGESTION_SUFFIXES")

			base := "test"
			suggestions := SuggestSlugAlternatives(base)

			// Should have suggestions with custom suffixes
			expectedCount := len(tt.expectedSuffixes)
			if len(suggestions) != expectedCount {
				t.Errorf("Expected %d suggestions, got %d: %v", expectedCount, len(suggestions), suggestions)
			}

			// Check that suggestions match expected suffixes
			for i, suggestion := range suggestions {
				expected := base + tt.expectedSuffixes[i]
				if suggestion != expected {
					t.Errorf("Expected suggestion %q, got %q", expected, suggestion)
				}
			}
		})
	}
}

func TestSuggestSlugAlternatives_EnvironmentFallback(t *testing.T) {
	// Test with invalid environment variable (should fall back to defaults)
	os.Setenv("SLUG_SUGGESTION_SUFFIXES", "")
	defer os.Unsetenv("SLUG_SUGGESTION_SUFFIXES")

	base := "acme"
	suggestions := SuggestSlugAlternatives(base)

	// Should fall back to default suffixes
	expectedCount := len(defaultSlugSuffixes)
	if len(suggestions) != expectedCount {
		t.Errorf("Expected %d suggestions (default), got %d", expectedCount, len(suggestions))
	}
}

func TestSuggestSlugAlternatives_LongBase(t *testing.T) {
	// Test with a base that's already close to the 30-character limit
	longBase := "this-is-a-very-long-base-name"
	suggestions := SuggestSlugAlternatives(longBase)

	// Should return at least the base slug
	if len(suggestions) == 0 {
		t.Error("Expected at least one suggestion (the base slug)")
	}

	// First suggestion should be the base slug
	if suggestions[0] != longBase {
		t.Errorf("Expected first suggestion to be base slug %q, got %q", longBase, suggestions[0])
	}

	// All suggestions should be within 30 character limit
	for _, suggestion := range suggestions {
		if len(suggestion) > 30 {
			t.Errorf("Suggestion %q exceeds 30 character limit", suggestion)
		}
	}
}

func TestGetSlugSuffixes_EnvironmentVariable(t *testing.T) {
	// Test the internal function directly
	tests := []struct {
		name     string
		envValue string
		expected []string
	}{
		{
			name:     "valid environment variable",
			envValue: "-corp,-inc,-ltd",
			expected: []string{"-corp", "-inc", "-ltd"},
		},
		{
			name:     "empty environment variable",
			envValue: "",
			expected: defaultSlugSuffixes,
		},
		{
			name:     "invalid environment variable",
			envValue: ",,,",
			expected: defaultSlugSuffixes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("SLUG_SUGGESTION_SUFFIXES", tt.envValue)
			} else {
				os.Unsetenv("SLUG_SUGGESTION_SUFFIXES")
			}
			defer os.Unsetenv("SLUG_SUGGESTION_SUFFIXES")

			result := getSlugSuffixes()

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d suffixes, got %d", len(tt.expected), len(result))
			}

			for i, suffix := range result {
				if suffix != tt.expected[i] {
					t.Errorf("Expected suffix %q at index %d, got %q", tt.expected[i], i, suffix)
				}
			}
		})
	}
}
