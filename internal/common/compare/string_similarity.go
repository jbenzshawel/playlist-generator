package compare

import "math"

// StringSimilarity calculates the similarity between two strings as a percentage.
// It uses the Levenshtein distance algorithm to determine the edit distance.
func StringSimilarity(s1, s2 string) float64 {
	// Convert strings to rune slices to correctly handle multi-byte characters (like emojis or non-ASCII letters)
	r1 := []rune(s1)
	r2 := []rune(s2)

	len1 := len(r1)
	len2 := len(r2)

	// Handle edge cases
	if len1 == 0 && len2 == 0 {
		return 100.0 // Two empty strings are 100% similar
	}

	// Determine the maximum possible length, which is the denominator for the similarity calculation.
	maxLen := math.Max(float64(len1), float64(len2))

	// If one string is empty, the similarity is 0%
	// (This also prevents division by zero if maxLen was 0,
	// though the case above already handled len1=0 && len2=0)
	if maxLen == 0 {
		return 0.0
	}

	// Calculate the Levenshtein distance
	distance := levenshteinDistance(r1, r2)

	// Calculate similarity as a percentage
	// (1.0 - (distance / max_length)) * 100
	similarity := (1.0 - (float64(distance) / maxLen)) * 100.0

	return similarity
}

// levenshteinDistance calculates the edit distance between two rune slices.
// This is the classic dynamic programming implementation.
func levenshteinDistance(s1, s2 []rune) int {
	n := len(s1)
	m := len(s2)

	// Create a 2D slice (matrix) to store distances.
	// dp[i][j] will hold the distance between the first i runes of s1
	// and the first j runes of s2.
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}

	// Initialize the first row and column.
	// Cost of transforming an empty string to a non-empty string is
	// just the number of insertions.
	for i := 0; i <= n; i++ {
		dp[i][0] = i // Deletions
	}
	for j := 0; j <= m; j++ {
		dp[0][j] = j // Insertions
	}

	// Fill the rest of the matrix
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			// Cost of substitution (0 if characters are equal, 1 otherwise)
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			// Find the minimum cost from the three possible operations:
			deletion := dp[i-1][j] + 1          // Cost of deleting char from s1
			insertion := dp[i][j-1] + 1         // Cost of inserting char into s1
			substitution := dp[i-1][j-1] + cost // Cost of substituting char

			dp[i][j] = min(deletion, insertion, substitution)
		}
	}

	// The final distance is in the bottom-right cell
	return dp[n][m]
}

// min helper function to find the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
	} else {
		if b < c {
			return b
		}
	}
	return c
}
