package utils

import (
	"strings"
)

func NormalizeDomain(d string) string {
	d = strings.TrimSpace(d)
	d = strings.ToLower(d)
	d = strings.TrimSuffix(d, ".")
	return d
}

func IsSubdomainOf(candidate, root string) bool {
	candidate = NormalizeDomain(candidate)
	root = NormalizeDomain(root)

	if candidate == root {
		return true
	}

	return strings.HasSuffix(candidate, "."+root)
}
