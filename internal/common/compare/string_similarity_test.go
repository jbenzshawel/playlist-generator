package compare

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringSimilarity(t *testing.T) {
	testCases := []struct {
		s1       string
		s2       string
		expected float64
	}{
		{
			s1:       "kitten",
			s2:       "sitting",
			expected: 57.14285714285714,
		},
		{
			s1:       "hello",
			s2:       "helo",
			expected: 80.00,
		},
		{
			s1:       "golang",
			s2:       "java",
			expected: 16.666666666666664,
		},
		{
			s1:       "match",
			s2:       "match",
			expected: 100,
		},
		{
			s1:       "not empty",
			s2:       "",
			expected: 0,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s/%s", tc.s1, tc.s2), func(t *testing.T) {
			assert.Equal(t, tc.expected, StringSimilarity(tc.s1, tc.s2))
		})
	}
}
