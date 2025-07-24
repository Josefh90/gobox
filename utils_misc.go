package gobox_utils

// calcPercent calculates the percentage progress based on current index and total count.
// Parameters:
//   - i: current zero-based index (e.g., 0 for first item).
//   - total: total number of items.
// Returns:
//   - int percentage value representing progress from 1 to 100.
//     If total is zero, returns 100 to avoid division by zero.
func calcPercent(i, total int) int {
	if total == 0 {
		// Avoid division by zero; assume 100% when total is zero
		return 100

	}
	// Calculate progress percentage: (current index + 1) / total * 100
	return int(float64(i+1) / float64(total) * 100)
}
