package gobox_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcPercent(t *testing.T) {
	// Define a slice of test cases using a struct.
	// Each test case has a name (for clarity in output),
	// input values i and total, and the expected output.
	tests := []struct {
		name     string
		i, total int
		expected int
	}{
		{"Zero total", 0, 0, 100}, // Test when total is zero (edge case)
		{"Halfway", 49, 100, 50},  // Test approximately half progress
		{"Full", 99, 100, 100},    // Test full progress (almost last index)
	}

	// Loop over each test case.
	for _, tt := range tests {
		// Run each test case as a subtest with the name provided.
		// This isolates tests and gives clear output per case.
		t.Run(tt.name, func(t *testing.T) {
			// Call the function with the inputs from the test case.
			result := calcPercent(tt.i, tt.total)

			// Use the assert package to compare expected vs actual result.
			// If they don't match, the test will fail and report the mismatch.
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkCalcPercent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		calcPercent(50, 100)
	}
}
