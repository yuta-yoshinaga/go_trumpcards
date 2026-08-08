package color

import "testing"

func resetColorState(t *testing.T) {
	t.Helper()
	origStdout := NoColorStdout()
	origStderr := NoColorStderr()
	t.Cleanup(func() {
		SetStdoutColor(!origStdout)
		SetStderrColor(!origStderr)
	})
}

func TestSetAndGetNoColor(t *testing.T) {
	resetColorState(t)

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
			if got := NoColorStdout(); got != tt.value {
				t.Errorf("NoColorStdout() = %v, want %v", got, tt.value)
			}
			if got := NoColorStderr(); got != tt.value {
				t.Errorf("NoColorStderr() = %v, want %v", got, tt.value)
			}
		})
	}
}

func TestSetStdoutAndStderrColorIndependent(t *testing.T) {
	resetColorState(t)

	tests := []struct {
		name       string
		stdoutOn   bool
		stderrOn   bool
		wantStdout bool // expected NoColorStdout()
		wantStderr bool // expected NoColorStderr()
	}{
		{"both on", true, true, false, false},
		{"stdout on, stderr off", true, false, false, true},
		{"stdout off, stderr on", false, true, true, false},
		{"both off", false, false, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetStdoutColor(tt.stdoutOn)
			SetStderrColor(tt.stderrOn)
			if got := NoColorStdout(); got != tt.wantStdout {
				t.Errorf("NoColorStdout() = %v, want %v", got, tt.wantStdout)
			}
			if got := NoColorStderr(); got != tt.wantStderr {
				t.Errorf("NoColorStderr() = %v, want %v", got, tt.wantStderr)
			}
		})
	}
}

func TestColorFunctions(t *testing.T) {
	resetColorState(t)

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
	resetColorState(t)

	SetNoColor(false)
	if got := Red(""); got != "" {
		t.Errorf("Red(\"\") = %q, want %q", got, "")
	}

	SetNoColor(true)
	if got := Red(""); got != "" {
		t.Errorf("Red(\"\") with NoColor = %q, want %q", got, "")
	}
}

func TestStdoutColorFuncsIgnoreStderrFlag(t *testing.T) {
	resetColorState(t)

	SetStdoutColor(true)
	SetStderrColor(false)
	if got := Red("x"); got != "\033[31mx\033[0m" {
		t.Errorf("Red should follow stdout flag; got %q", got)
	}

	SetStdoutColor(false)
	SetStderrColor(true)
	if got := Red("x"); got != "x" {
		t.Errorf("Red should follow stdout flag (off); got %q", got)
	}
}
