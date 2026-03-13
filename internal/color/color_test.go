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

func TestColorFunctions(t *testing.T) {
	original := NoColor()
	defer SetNoColor(original)

	funcs := []struct {
		name   string
		fn     func(string) string
		prefix string
	}{
		{"Red", Red, "\033[31m"},
		{"Green", Green, "\033[32m"},
		{"Yellow", Yellow, "\033[33m"},
		{"Bold", Bold, "\033[1m"},
		{"BoldYellow", BoldYellow, "\033[1;33m"},
	}

	for _, f := range funcs {
		t.Run(f.name+"_color_on", func(t *testing.T) {
			SetNoColor(false)
			got := f.fn("hello")
			want := f.prefix + "hello" + "\033[0m"
			if got != want {
				t.Errorf("%s(\"hello\") = %q, want %q", f.name, got, want)
			}
		})
		t.Run(f.name+"_color_off", func(t *testing.T) {
			SetNoColor(true)
			got := f.fn("hello")
			if got != "hello" {
				t.Errorf("%s(\"hello\") with NoColor = %q, want %q", f.name, got, "hello")
			}
		})
	}
}

func TestColorFunctionsEmptyString(t *testing.T) {
	original := NoColor()
	defer SetNoColor(original)

	SetNoColor(false)
	if got := Red(""); got != "" {
		t.Errorf("Red(\"\") = %q, want %q", got, "")
	}

	SetNoColor(true)
	if got := Red(""); got != "" {
		t.Errorf("Red(\"\") with NoColor = %q, want %q", got, "")
	}
}
