//go:build test

package cuiutil

import (
	"reflect"
	"testing"
)

type fakeIF struct {
	calls []string
}

func (f *fakeIF) hit() string   { f.calls = append(f.calls, "hit"); return "HIT" }
func (f *fakeIF) stand() string { f.calls = append(f.calls, "stand"); return "STAND" }

func TestCommandMap_AddAndLookup(t *testing.T) {
	t.Parallel()

	m := NewCommandMap[*fakeIF]().
		Add((*fakeIF).hit, "h", "hit").
		Add((*fakeIF).stand, "s", "stand")

	cases := []struct {
		alias string
		want  string
	}{
		{"h", "HIT"},
		{"hit", "HIT"},
		{"s", "STAND"},
		{"stand", "STAND"},
	}
	for _, tc := range cases {
		fn, ok := m.Lookup(tc.alias)
		if !ok {
			t.Errorf("Lookup(%q) ok=false, want true", tc.alias)
			continue
		}
		if got := fn(&fakeIF{}); got != tc.want {
			t.Errorf("Lookup(%q)() = %q, want %q", tc.alias, got, tc.want)
		}
	}
}

func TestCommandMap_LookupMissing(t *testing.T) {
	t.Parallel()

	m := NewCommandMap[*fakeIF]()
	if fn, ok := m.Lookup("nope"); ok || fn != nil {
		t.Errorf("Lookup(missing) ok=%v fn!=nil=%v, want false/false", ok, fn != nil)
	}
}

func TestCommandMap_Names(t *testing.T) {
	t.Parallel()

	m := NewCommandMap[*fakeIF]().
		Add((*fakeIF).stand, "stand", "s").
		Add((*fakeIF).hit, "hit", "h")

	got := m.Names()
	want := []string{"h", "hit", "s", "stand"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Names() = %v, want %v", got, want)
	}
}

func TestCommandMap_DuplicateAliasPanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic on duplicate alias, got none")
		}
	}()

	NewCommandMap[*fakeIF]().
		Add((*fakeIF).hit, "h").
		Add((*fakeIF).stand, "h")
}

func TestCommandMap_NilFunctionPanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic when registering nil fn, got none")
		}
	}()

	NewCommandMap[*fakeIF]().Add(nil, "h")
}

func TestCommandMap_EmptyAliasesPanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic on Add with no aliases, got none")
		}
	}()

	NewCommandMap[*fakeIF]().Add((*fakeIF).hit)
}
