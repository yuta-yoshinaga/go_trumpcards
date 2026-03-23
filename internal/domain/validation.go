package domain

import "fmt"

// ValidateRange は値が [min, max] の範囲内であることを検証する。
func ValidateRange(name string, value, min, max int) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be %d-%d, got %d", name, min, max, value)
	}
	return nil
}

// ValidateMin は値が min 以上であることを検証する。
func ValidateMin(name string, value, min int) error {
	if value < min {
		return fmt.Errorf("%s must be >= %d, got %d", name, min, value)
	}
	return nil
}
