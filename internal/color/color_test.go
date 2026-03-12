package color

import "testing"

func TestSetAndGetNoColor(t *testing.T) {
	// Reset to default state after test
	defer SetNoColor(false)

	// Default: false
	SetNoColor(false)
	if NoColor() {
		t.Error("expected NoColor() == false after SetNoColor(false)")
	}

	// Set to true
	SetNoColor(true)
	if !NoColor() {
		t.Error("expected NoColor() == true after SetNoColor(true)")
	}

	// Toggle back to false
	SetNoColor(false)
	if NoColor() {
		t.Error("expected NoColor() == false after second SetNoColor(false)")
	}
}
