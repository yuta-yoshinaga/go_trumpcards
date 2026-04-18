// Package games is the single source of truth for the set of games exposed
// by the HTTP server and the Cloudflare Workers. Adding or removing a game
// must happen in exactly one place: the registry defined in games.go (Web
// factory) and games_wasm.go (Worker factory, built only under js && wasm).
//
// Consumers:
//   - internal/infrastructure/web.TrumpCardsWeb iterates All() and calls
//     NewWebController to register each game's HTTP handler.
//   - cmd/workers/{casino,classic,solo}/main.go iterates ByCategory(cat) and
//     invokes RegisterWorker to register each game's KV-backed handler.
//
// Category metadata pins each game to exactly one Cloudflare Worker, removing
// the prior drift risk of maintaining three hand-edited worker registrations.
package games

import (
	"fmt"
	"net/http"
)

// Category identifies which Cloudflare Worker a game is deployed to.
type Category int

const (
	// CategoryCasino covers table and poker games.
	CategoryCasino Category = iota
	// CategoryClassic covers trick-taking, matching, and family card games.
	CategoryClassic
	// CategorySolo covers solitaire and rummy variants.
	CategorySolo
)

// String returns the lowercase worker name (casino/classic/solo).
func (c Category) String() string {
	switch c {
	case CategoryCasino:
		return "casino"
	case CategoryClassic:
		return "classic"
	case CategorySolo:
		return "solo"
	}
	return fmt.Sprintf("Category(%d)", int(c))
}

// WebController is the minimal handler interface implemented by every
// generated *GameWebController. It is what TrumpCardsWeb.register consumes.
type WebController interface {
	Exec(http.ResponseWriter, *http.Request)
	Stop()
}

// Game holds the metadata and factories for a single game. The zero value is
// not useful; always construct via the package-level registry in games.go.
type Game struct {
	// Name is the canonical short name (e.g. "blackjack"); also the URL path
	// segment used for the /<name>/exec endpoint.
	Name string

	// Category selects which Cloudflare Worker hosts the game.
	Category Category

	// NewWebController returns a fresh controller bound to an in-memory
	// session provider. Used by the HTTP server (non-WASM builds) and also
	// compiled under js && wasm so the registry is trivially iterable in both
	// build modes.
	NewWebController func() WebController

	// RegisterWorker registers the game's KV-backed handler on mux. Populated
	// only under //go:build js && wasm; nil on other builds. Callers that may
	// run on both sides must nil-check or guard via ByCategory from wasm code.
	RegisterWorker func(mux *http.ServeMux) error
}

// All returns a fresh copy of the registry in canonical order (matching the
// CLI game registry in internal/infrastructure/ui/GameManager.go).
func All() []*Game {
	out := make([]*Game, len(registry))
	copy(out, registry)
	return out
}

// ByCategory returns the games assigned to cat, preserving canonical order.
func ByCategory(cat Category) []*Game {
	var out []*Game
	for _, g := range registry {
		if g.Category == cat {
			out = append(out, g)
		}
	}
	return out
}
