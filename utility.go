package main

// Maxi returns the largest of two integers.
func Maxi(a, b int) int {
	if a < b {
		return b
	}

	return a
}

// Mini returns the smallest of two integers.
func Mini(a, b int) int {
	if a < b {
		return a
	}

	return b
}