package webutil

// BoundedIntPtr returns *ptr if non-nil and within [lo, hi], otherwise defaultVal.
func BoundedIntPtr(ptr *int, lo, hi, defaultVal int) int {
	if ptr == nil {
		return defaultVal
	}
	val := *ptr
	if val < lo || val > hi {
		return defaultVal
	}
	return val
}

// BoolPtrOr returns *ptr if non-nil, otherwise defaultVal.
func BoolPtrOr(ptr *bool, defaultVal bool) bool {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}
