// Package games is the single source of truth for the set of games exposed
// by the HTTP server and the Cloudflare Workers.
//
// The registry is split across build-tagged files so that Cloudflare Worker
// binaries (TinyGo / WASM) stay under the 1 MB gzipped free-tier limit:
//
//   - registry.go (this file, no tag)  — types and bare metadata (Name +
//     Category) for all 55 games. Cheap; no references to game code.
//   - games_server.go (!js || !wasm)   — installs Web-server factories for
//     every game via BindWebController. Imported by TrumpCardsWeb.
//   - casino/, classic/, solo/ (js && wasm) — per-category worker bindings.
//     Each worker blank-imports only its own sub-package so TinyGo dead-code
//     elimination can drop the other two categories' domain/usecase code.
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
// *GameWebController. Consumed by TrumpCardsWeb.
type WebController interface {
	Exec(http.ResponseWriter, *http.Request)
	Stop()
}

// Game holds the metadata and factories for a single game. Construct only via
// the package-level registry below. Fields are populated lazily from
// build-tagged init() functions so that each binary links only what it uses.
type Game struct {
	// Name is the canonical short name (e.g. "blackjack") and the URL path
	// segment for the /<name>/exec endpoint.
	Name string

	// Category selects which Cloudflare Worker hosts the game.
	Category Category

	// NewWebController returns a fresh controller bound to the default
	// in-memory session provider. Populated only on non-WASM builds via
	// games_server.go; nil in Cloudflare Worker binaries.
	NewWebController func() WebController

	// RegisterWorker registers the game's KV-backed handler on mux.
	// Populated only in the one WASM sub-package (casino/classic/solo) that
	// owns this game; nil otherwise.
	RegisterWorker func(mux *http.ServeMux) error
}

// registry is the canonical game set. Order mirrors the CLI registry in
// internal/infrastructure/ui/GameManager.go. Adding a game here is the *only*
// place where a game is declared; the per-build tag init functions attach
// the matching factories.
var registry = []*Game{
	{Name: "blackjack", Category: CategoryCasino},
	{Name: "poker", Category: CategoryCasino},
	{Name: "oldmaid", Category: CategoryClassic},
	{Name: "daifugo", Category: CategoryClassic},
	{Name: "sevens", Category: CategoryClassic},
	{Name: "doubt", Category: CategoryClassic},
	{Name: "holdem", Category: CategoryCasino},
	{Name: "omaha", Category: CategoryCasino},
	{Name: "shortdeck", Category: CategoryCasino},
	{Name: "pineapple", Category: CategoryCasino},
	{Name: "hearts", Category: CategoryClassic},
	{Name: "memory", Category: CategorySolo},
	{Name: "klondike", Category: CategorySolo},
	{Name: "freecell", Category: CategorySolo},
	{Name: "baccarat", Category: CategoryCasino},
	{Name: "spades", Category: CategoryClassic},
	{Name: "crazyeights", Category: CategoryClassic},
	{Name: "ginrummy", Category: CategorySolo},
	{Name: "canasta", Category: CategorySolo},
	{Name: "spider", Category: CategorySolo},
	{Name: "napoleon", Category: CategoryClassic},
	{Name: "indianpoker", Category: CategoryCasino},
	{Name: "videopoker", Category: CategoryCasino},
	{Name: "deuceswild", Category: CategoryCasino},
	{Name: "jokerpoker", Category: CategoryCasino},
	{Name: "euchre", Category: CategoryClassic},
	{Name: "pyramid", Category: CategorySolo},
	{Name: "tripeaks", Category: CategorySolo},
	{Name: "cribbage", Category: CategorySolo},
	{Name: "threecard", Category: CategoryCasino},
	{Name: "ohhell", Category: CategoryClassic},
	{Name: "bridge", Category: CategoryClassic},
	{Name: "speed", Category: CategoryClassic},
	{Name: "gofish", Category: CategoryClassic},
	{Name: "pinochle", Category: CategoryClassic},
	{Name: "golf", Category: CategorySolo},
	{Name: "pigtail", Category: CategoryClassic},
	{Name: "sevencardstud", Category: CategoryCasino},
	{Name: "clocksolitaire", Category: CategorySolo},
	{Name: "durak", Category: CategoryClassic},
	{Name: "fortythieves", Category: CategorySolo},
	{Name: "paigow", Category: CategoryCasino},
	{Name: "twotenjack", Category: CategoryClassic},
	{Name: "caribbeanstud", Category: CategoryCasino},
	{Name: "war", Category: CategoryClassic},
	{Name: "canfield", Category: CategorySolo},
	{Name: "fiftyone", Category: CategoryClassic},
	{Name: "yukon", Category: CategorySolo},
	{Name: "whist", Category: CategoryClassic},
	{Name: "letitride", Category: CategoryCasino},
	{Name: "pokersquares", Category: CategorySolo},
	{Name: "pageone", Category: CategoryClassic},
	{Name: "reddog", Category: CategoryCasino},
	{Name: "razz", Category: CategoryCasino},
	{Name: "scorpion", Category: CategorySolo},
}

// All returns a fresh copy of the registry in canonical order.
func All() []*Game {
	out := make([]*Game, len(registry))
	copy(out, registry)
	return out
}

// ByCategory returns the games assigned to cat in canonical order.
func ByCategory(cat Category) []*Game {
	var out []*Game
	for _, g := range registry {
		if g.Category == cat {
			out = append(out, g)
		}
	}
	return out
}

// find locates a game by name; returns nil if not found.
func find(name string) *Game {
	for _, g := range registry {
		if g.Name == name {
			return g
		}
	}
	return nil
}

// BindWebController attaches the server-side factory for the named game.
// Intended for use from games_server.go's init. Panics if the name is
// unknown, which surfaces typos at package-load time.
func BindWebController(name string, f func() WebController) {
	g := find(name)
	if g == nil {
		panic(fmt.Sprintf("games: BindWebController: unknown game %q", name))
	}
	g.NewWebController = f
}

// BindWorker attaches the WASM worker registration for the named game.
// Called from the per-category sub-packages' init functions. Panics on an
// unknown name or when the target game's Category does not match (catches
// the common error of registering a game in the wrong worker sub-package).
func BindWorker(name string, cat Category, r func(mux *http.ServeMux) error) {
	g := find(name)
	if g == nil {
		panic(fmt.Sprintf("games: BindWorker: unknown game %q", name))
	}
	if g.Category != cat {
		panic(fmt.Sprintf("games: BindWorker: %q belongs to %s but was registered under %s",
			name, g.Category, cat))
	}
	g.RegisterWorker = r
}

// RegisterCategory registers every game in cat onto mux using the KV-backed
// session provider. Returns the first error encountered. Call from a
// worker's main after blank-importing the matching sub-package.
func RegisterCategory(mux *http.ServeMux, cat Category) error {
	for _, g := range ByCategory(cat) {
		if g.RegisterWorker == nil {
			return fmt.Errorf("games: %q has no RegisterWorker (missing from %s sub-package init)",
				g.Name, cat)
		}
		if err := g.RegisterWorker(mux); err != nil {
			return fmt.Errorf("games: register %q: %w", g.Name, err)
		}
	}
	return nil
}
