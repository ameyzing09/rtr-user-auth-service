package utils

import (
	"errors"
	"os"
	"regexp"
	"strings"
)

var (
	slugPattern          = regexp.MustCompile(`[^a-z0-9-]`)
	slugCollapsePattern  = regexp.MustCompile(`-{2,}`)
	slugValidationRegexp = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`)

	// Default slug suggestion suffixes
	defaultSlugSuffixes = []string{"-hq", "-io", "-team", "-app", "-co"}
)

func NormalizeSlug(name string) (string, error) {
	value := strings.TrimSpace(strings.ToLower(name))
	if value == "" {
		return "", errors.New("slug source cannot be empty")
	}

	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	value = slugPattern.ReplaceAllString(value, "")
	value = slugCollapsePattern.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")

	if len(value) < 3 || len(value) > 30 {
		return "", errors.New("slug must be between 3 and 30 characters")
	}
	if !slugValidationRegexp.MatchString(value) {
		return "", errors.New("slug contains invalid characters")
	}

	return value, nil
}

// getSlugSuffixes returns the configured slug suggestion suffixes
func getSlugSuffixes() []string {
	// Check for environment variable configuration
	envSuffixes := os.Getenv("SLUG_SUGGESTION_SUFFIXES")
	if envSuffixes != "" {
		// Parse comma-separated suffixes from environment
		suffixes := strings.Split(envSuffixes, ",")
		validSuffixes := make([]string, 0, len(suffixes))
		for _, suffix := range suffixes {
			suffix = strings.TrimSpace(suffix)
			if suffix != "" {
				// Ensure suffix starts with dash if not already
				if !strings.HasPrefix(suffix, "-") {
					suffix = "-" + suffix
				}
				validSuffixes = append(validSuffixes, suffix)
			}
		}
		if len(validSuffixes) > 0 {
			return validSuffixes
		}
	}

	// Fall back to default suffixes
	return defaultSlugSuffixes
}

func SuggestSlugAlternatives(base string) []string {
	suffixes := getSlugSuffixes()
	suggestions := make([]string, 0, len(suffixes))

	for _, suffix := range suffixes {
		candidate := base + suffix
		if len(candidate) <= 30 {
			suggestions = append(suggestions, candidate)
		}
	}

	// If no suggestions were generated, return the base slug
	if len(suggestions) == 0 {
		suggestions = append(suggestions, base)
	}

	return suggestions
}
