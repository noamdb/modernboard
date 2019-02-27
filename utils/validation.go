package utils

func ValidLength(str string, min int, max int) bool {
	l := len([]rune(str))
	if l < min {
		return false
	}
	if max != -1 && l > max {
		return false
	}
	return true
}
