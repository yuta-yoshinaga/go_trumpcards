package webutil

// ClampIntPtr returns *ptr if non-nil and within [min, max], otherwise defaultVal.
func ClampIntPtr(ptr *int, min, max, defaultVal int) int {
	if ptr == nil {
		return defaultVal
	}
	val := *ptr
	if val < min || val > max {
		return defaultVal
	}
	return val
}
