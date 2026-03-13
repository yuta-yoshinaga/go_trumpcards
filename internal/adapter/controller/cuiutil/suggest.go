package cuiutil

// LevenshteinDistance returns the Levenshtein edit distance between two strings.
func LevenshteinDistance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	la, lb := len(ra), len(rb)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			ins := curr[j-1] + 1
			del := prev[j] + 1
			sub := prev[j-1] + cost
			m := ins
			if del < m {
				m = del
			}
			if sub < m {
				m = sub
			}
			curr[j] = m
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

// SuggestCommand returns the closest matching command from validCommands
// whose edit distance is at most maxDistance. Returns "" if no match is found.
func SuggestCommand(input string, validCommands []string, maxDistance int) string {
	if input == "" || len(validCommands) == 0 {
		return ""
	}
	best := ""
	bestDist := maxDistance + 1
	for _, cmd := range validCommands {
		d := LevenshteinDistance(input, cmd)
		if d < bestDist {
			bestDist = d
			best = cmd
		}
	}
	if bestDist > maxDistance {
		return ""
	}
	return best
}
