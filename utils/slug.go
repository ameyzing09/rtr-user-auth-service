package utils

import (
	"errors"
	"regexp"
	"strings"
)

var (
	slugPattern          = regexp.MustCompile(`[^a-z0-9-]`)
	slugCollapsePattern  = regexp.MustCompile(`-{2,}`)
	slugValidationRegexp = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`)
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

func SuggestSlugAlternatives(base string) []string {
	suffixes := []string{"-hq", "-io", "-team"}
	suggestions := make([]string, 0, len(suffixes))
	for _, suffix := range suffixes {
		candidate := base + suffix
		if len(candidate) <= 30 {
			suggestions = append(suggestions, candidate)
		}
	}
	if len(suggestions) == 0 {
		suggestions = append(suggestions, base)
	}
	return suggestions
}
