package picker

import "strings"

// Match reports whether query is a case-insensitive subsequence of the
// joined parts (or a contiguous substring). An empty query matches.
func Match(query string, parts ...string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return true
	}
	hay := strings.ToLower(strings.Join(parts, " "))
	if strings.Contains(hay, q) {
		return true
	}
	qi := 0
	qr := []rune(q)
	for _, r := range hay {
		if qi < len(qr) && r == qr[qi] {
			qi++
		}
	}
	return qi == len(qr)
}
