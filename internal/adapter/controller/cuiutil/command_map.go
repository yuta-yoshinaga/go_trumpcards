package cuiutil

import "sort"

// CommandMap is a registry of no-arg CUI commands keyed by alias.
//
// Many CUI controllers register the same action under multiple aliases (e.g.,
// "h"/"hit", "s"/"stand"). Hand-writing one map entry per alias and then a
// separate validCommands list duplicates the alias set in two places. CommandMap
// collapses both: Add takes a function plus all of its aliases in one call,
// and Names returns the alias set for use as the validCommands suggestion list.
type CommandMap[T any] struct {
	m map[string]func(T) string
}

// NewCommandMap returns an empty CommandMap parameterised on the interactor type T.
func NewCommandMap[T any]() *CommandMap[T] {
	return &CommandMap[T]{m: make(map[string]func(T) string)}
}

// Add binds fn to one or more aliases. Panics on:
//   - nil fn (would surface only as a nil pointer dereference at exec time);
//   - empty alias list (almost always a typo — the call would silently no-op);
//   - duplicate alias registration (would silently overwrite a previous binding).
//
// All three are init-time programmer errors, so failing fast at startup is
// preferred over surfacing them later at command-exec time.
func (c *CommandMap[T]) Add(fn func(T) string, aliases ...string) *CommandMap[T] {
	if fn == nil {
		panic("cuiutil.CommandMap: function cannot be nil")
	}
	if len(aliases) == 0 {
		panic("cuiutil.CommandMap: at least one alias is required")
	}
	for _, a := range aliases {
		if _, exists := c.m[a]; exists {
			panic("cuiutil.CommandMap: duplicate alias " + a)
		}
		c.m[a] = fn
	}
	return c
}

// Lookup returns the function registered for cmd, or (nil, false) if unknown.
func (c *CommandMap[T]) Lookup(cmd string) (func(T) string, bool) {
	fn, ok := c.m[cmd]
	return fn, ok
}

// Names returns all registered aliases sorted alphabetically. Useful for
// passing to execCuiCommand as the validCommands suggestion list.
func (c *CommandMap[T]) Names() []string {
	out := make([]string, 0, len(c.m))
	for k := range c.m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
