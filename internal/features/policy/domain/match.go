package domain

import "strings"

// MatchPackage reports whether pattern matches importPath. * matches exactly
// one path segment; ** matches zero or more segments. Matching is against
// the full import path using / as the separator.
func MatchPackage(pattern, importPath string) bool {
	return matchSegments(splitPattern(pattern), strings.Split(importPath, "/"))
}

func splitPattern(pattern string) []string {
	if pattern == "" {
		return nil
	}

	return strings.Split(pattern, "/")
}

func matchSegments(pattern, path []string) bool {
	pi, si := 0, 0
	for pi < len(pattern) {
		if pattern[pi] == "**" {
			pi++
			if pi == len(pattern) {
				return true
			}

			for si <= len(path) {
				if matchSegments(pattern[pi:], path[si:]) {
					return true
				}

				si++
			}

			return false
		}

		if si >= len(path) {
			return false
		}

		if pattern[pi] != "*" && pattern[pi] != path[si] {
			return false
		}

		pi++
		si++
	}

	return si == len(path)
}
