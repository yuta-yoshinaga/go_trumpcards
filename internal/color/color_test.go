package color

import "testing"

func TestSetAndGetNoColor(t *testing.T) {
	original := NoColor()
	defer SetNoColor(original)

	tests := []struct {
		name  string
		value bool
	}{
		{"set false", false},
		{"set true", true},
		{"set false again", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetNoColor(tt.value)
			if got := NoColor(); got != tt.value {
				t.Errorf("NoColor() = %v, want %v", got, tt.value)
			}
		})
	}
}
