package template

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mmonterroca/docxgo/v2/domain"
)

// MergeTemplate replaces all placeholders in doc with values from data.
// The document is modified in place.
//
// By default, placeholders use {{key}} syntax. Unmatched placeholders are left
// as-is unless StrictMode is enabled in options.
func MergeTemplate(doc domain.Document, data MergeData, opts ...MergeOptions) error {
	opt := DefaultMergeOptions()
	if len(opts) > 0 {
		opt = opts[0]
	}

	pattern := buildPattern(opt)
	var missingKeys []string

	err := walkParagraphs(doc, func(para domain.Paragraph, ctx paragraphContext) error {
		ConsolidateRuns(para)
		missing := replaceParagraph(para, data, pattern, opt)
		missingKeys = append(missingKeys, missing...)
		return nil
	})
	if err != nil {
		return err
	}

	if opt.StrictMode && len(missingKeys) > 0 {
		// Deduplicate missing keys
		seen := make(map[string]struct{})
		var unique []string
		for _, k := range missingKeys {
			if _, ok := seen[k]; !ok {
				seen[k] = struct{}{}
				unique = append(unique, k)
			}
		}
		return fmt.Errorf("template: missing keys: %s", strings.Join(unique, ", "))
	}

	return nil
}

// ValidateTemplate checks that all placeholders have corresponding keys in data
// and reports unused data keys.
func ValidateTemplate(doc domain.Document, data MergeData, opts ...MergeOptions) []ValidationError {
	opt := DefaultMergeOptions()
	if len(opts) > 0 {
		opt = opts[0]
	}

	names := placeholderNamesWithOptions(doc, opt)
	var errors []ValidationError

	// Check for missing keys (placeholders without data)
	for _, name := range names {
		if _, ok := data[name]; !ok {
			errors = append(errors, ValidationError{
				Severity: SeverityError,
				Key:      name,
				Message:  "placeholder has no corresponding data key",
			})
		}
	}

	// Check for unused keys (data without placeholders)
	nameSet := make(map[string]struct{})
	for _, n := range names {
		nameSet[n] = struct{}{}
	}
	for k := range data {
		if _, ok := nameSet[k]; !ok {
			errors = append(errors, ValidationError{
				Severity: SeverityWarning,
				Key:      k,
				Message:  "data key has no corresponding placeholder in template",
			})
		}
	}

	return errors
}

// replaceParagraph replaces all placeholders in a single paragraph.
// Returns a list of placeholder keys for which no data was found.
func replaceParagraph(para domain.Paragraph, data MergeData, pattern *regexp.Regexp, opt MergeOptions) []string {
	var missing []string
	runs := para.Runs()

	for _, run := range runs {
		text := run.Text()
		if !pattern.MatchString(text) {
			continue
		}

		newText := pattern.ReplaceAllStringFunc(text, func(match string) string {
			key := extractKey(match, opt)
			if val, ok := data[key]; ok {
				return val
			}
			missing = append(missing, key)
			return match // leave unreplaced
		})

		if newText != text {
			run.SetText(newText)
		}
	}

	return missing
}

// buildPattern creates a regex pattern from the configured delimiters.
func buildPattern(opt MergeOptions) *regexp.Regexp {
	open := regexp.QuoteMeta(opt.OpenDelimiter)
	close := regexp.QuoteMeta(opt.CloseDelimiter)
	return regexp.MustCompile(open + `\.?(\w+(?:\.\w+)*)` + close)
}

// extractKey extracts the placeholder key from a full match string.
func extractKey(match string, opt MergeOptions) string {
	// Remove delimiters
	key := match
	key = strings.TrimPrefix(key, opt.OpenDelimiter)
	key = strings.TrimSuffix(key, opt.CloseDelimiter)
	// Remove optional leading dot
	key = strings.TrimPrefix(key, ".")
	return key
}

// placeholderNamesWithOptions returns deduplicated placeholder names using custom options.
func placeholderNamesWithOptions(doc domain.Document, opt MergeOptions) []string {
	pattern := buildPattern(opt)
	placeholders := findPlaceholdersWithPattern(doc, pattern)

	seen := make(map[string]struct{})
	var names []string
	for _, p := range placeholders {
		if _, ok := seen[p.Name]; !ok {
			seen[p.Name] = struct{}{}
			names = append(names, p.Name)
		}
	}
	return names
}
