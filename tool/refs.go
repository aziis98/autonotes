package autonotes

import (
	"strings"
	"unicode"
)

// ExpandRefs expands a space-separated string of references, supporting bracket syntax.
// Example: "prefix-[a, b]-suffix" -> ["prefix-a-suffix", "prefix-b-suffix"]
func ExpandRefs(input string) []string {
	var results []string
	parts := splitOutsideBrackets(input)
	for _, part := range parts {
		results = append(results, expand(part)...)
	}
	return results
}

// splitOutsideBrackets splits a string by whitespace (if no delimiters are provided) 
// or by the given delimiters, but only when not inside balanced brackets [].
func splitOutsideBrackets(s string, delimiters ...rune) []string {
	var parts []string
	var current strings.Builder
	depth := 0
	
	isDelimiter := func(r rune) bool {
		if len(delimiters) == 0 {
			return unicode.IsSpace(r)
		}
		for _, d := range delimiters {
			if r == d {
				return true
			}
		}
		return false
	}

	for _, r := range s {
		if r == '[' {
			depth++
			current.WriteRune(r)
		} else if r == ']' {
			depth--
			current.WriteRune(r)
		} else if depth == 0 && isDelimiter(r) {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

// expand recursively expands a single reference part.
// It finds the first balanced bracket pair [...] and expands it by taking the 
// Cartesian product of the prefix, each item in the bracket, and the expansion 
// of the suffix.
func expand(s string) []string {
	start := strings.Index(s, "[")
	if start == -1 {
		return []string{s}
	}

	end := -1
	depth := 0
	for i := start; i < len(s); i++ {
		if s[i] == '[' {
			depth++
		} else if s[i] == ']' {
			depth--
			if depth == 0 {
				end = i
				break
			}
		}
	}

	if end == -1 {
		return []string{s} // Unbalanced brackets, treat as literal
	}

	prefix := s[:start]
	content := s[start+1 : end]
	suffix := s[end+1:]

	// Split content by commas at the top level
	items := splitOutsideBrackets(content, ',')
	
	var expanded []string
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		// Recurse to handle multiple brackets like [a,b][c,d] or nested like [a, [b,c]]
		subExp := expand(prefix + trimmed + suffix)
		expanded = append(expanded, subExp...)
	}
	
	return expanded
}

// FindSimilarUIDs returns up to 3 most similar UIDs from the declared list.
func FindSimilarUIDs(broken string, declaredUIDs map[string]bool) []string {
	type match struct {
		uid  string
		dist int
	}
	var matches []match

	for uid := range declaredUIDs {
		dist := levenshtein(broken, uid)
		// Only consider matches with reasonable distance (e.g. less than half the length)
		if dist < len(broken)/2+1 {
			matches = append(matches, match{uid, dist})
		}
	}

	// Sort by distance
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].dist > matches[j].dist {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	var results []string
	for i := 0; i < len(matches) && i < 3; i++ {
		results = append(results, matches[i].uid)
	}
	return results
}

func levenshtein(a, b string) int {
	f := make([]int, len(b)+1)
	for j := range f {
		f[j] = j
	}
	for _, ca := range a {
		j := 1
		nw := f[0]
		f[0]++
		for _, cb := range b {
			cur := f[j]
			if ca == cb {
				f[j] = nw
			} else {
				f[j] = min(nw, f[j], f[j-1]) + 1
			}
			nw = cur
			j++
		}
	}
	return f[len(b)]
}

func min(a, b, c int) int {
	if a <= b && a <= c {
		return a
	}
	if b <= c {
		return b
	}
	return c
}
