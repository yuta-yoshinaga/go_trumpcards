// Package games is the single source of truth for the set of games exposed
// by the HTTP server and the Cloudflare Workers.
//
// The registry is split across build-tagged files so that Cloudflare Worker
// binaries (TinyGo / WASM) stay under the 1 MB gzipped free-tier limit:
//
//   - registry.go (this file, no tag)  — types and bare metadata (Name +
//     Category) for all 58 games. Cheap; no references to game code.
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

// String returns the lowercase worker name (casino/classic/solo). Panics on
// an unknown value — consistent with BindWebController/BindWorker, which also
// panic on API misuse. A silent fallback would mask bugs such as reading an
// uninitialised Category or forgetting a case after adding a new value.
func (c Category) String() string {
	switch c {
	case CategoryCasino:
		return "casino"
	case CategoryClassic:
		return "classic"
	case CategorySolo:
		return "solo"
	default:
		panic(fmt.Sprintf("games: unknown Category %d", int(c)))
	}
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

	// Description is the human-readable short title used by CLI listings and
	// help. Kept on Game itself (not in the CLI-only ui package) so that the
	// Name→Description mapping has a single source of truth (issue #1459).
	Description string

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
	{Name: "blackjack", Category: CategoryCasino, Description: "BlackJack (ブラックジャック)"},
	{Name: "poker", Category: CategoryCasino, Description: "5-card Draw Poker (ポーカー)"},
	{Name: "oldmaid", Category: CategoryClassic, Description: "Old Maid (ババ抜き)"},
	{Name: "daifugo", Category: CategoryClassic, Description: "Daifugo / Great Fool (大富豪)"},
	{Name: "sevens", Category: CategoryClassic, Description: "Sevens (7並べ)"},
	{Name: "doubt", Category: CategoryClassic, Description: "Doubt (ダウト)"},
	{Name: "holdem", Category: CategoryCasino, Description: "Texas Hold'em (テキサスホールデム)"},
	{Name: "omaha", Category: CategoryCasino, Description: "Omaha Hold'em (オマハホールデム)"},
	{Name: "shortdeck", Category: CategoryCasino, Description: "Short Deck (6+ Hold'em) (ショートデック)"},
	{Name: "pineapple", Category: CategoryCasino, Description: "Pineapple Poker (パイナップルポーカー)"},
	{Name: "hearts", Category: CategoryClassic, Description: "Hearts (ハーツ)"},
	{Name: "memory", Category: CategorySolo, Description: "Memory / Concentration (神経衰弱)"},
	{Name: "klondike", Category: CategorySolo, Description: "Klondike Solitaire (ソリティア)"},
	{Name: "freecell", Category: CategorySolo, Description: "FreeCell (フリーセル)"},
	{Name: "baccarat", Category: CategoryCasino, Description: "Baccarat (バカラ)"},
	{Name: "spades", Category: CategoryClassic, Description: "Spades (スペード)"},
	{Name: "crazyeights", Category: CategoryClassic, Description: "Crazy Eights (クレイジーエイト)"},
	{Name: "ginrummy", Category: CategorySolo, Description: "Gin Rummy (ジンラミー)"},
	{Name: "canasta", Category: CategorySolo, Description: "Canasta (カナスタ)"},
	{Name: "spider", Category: CategorySolo, Description: "Spider Solitaire (スパイダーソリティア)"},
	{Name: "napoleon", Category: CategoryClassic, Description: "Napoleon (ナポレオン)"},
	{Name: "indianpoker", Category: CategoryCasino, Description: "Indian Poker (インディアンポーカー)"},
	{Name: "videopoker", Category: CategoryCasino, Description: "Video Poker Jacks or Better (ビデオポーカー)"},
	{Name: "deuceswild", Category: CategoryCasino, Description: "Deuces Wild (デューシーズワイルド)"},
	{Name: "jokerpoker", Category: CategoryCasino, Description: "Joker Poker (ジョーカーポーカー)"},
	{Name: "euchre", Category: CategoryClassic, Description: "Euchre (ユーカー)"},
	{Name: "pyramid", Category: CategorySolo, Description: "Pyramid (ピラミッド)"},
	{Name: "tripeaks", Category: CategorySolo, Description: "TriPeaks (トリピークス)"},
	{Name: "cribbage", Category: CategorySolo, Description: "Cribbage (クリベッジ)"},
	{Name: "threecard", Category: CategoryCasino, Description: "Three Card Poker (スリーカードポーカー)"},
	{Name: "ohhell", Category: CategoryClassic, Description: "Oh Hell (オー・ヘル)"},
	{Name: "bridge", Category: CategoryClassic, Description: "Contract Bridge (コントラクトブリッジ)"},
	{Name: "speed", Category: CategoryClassic, Description: "Speed (スピード)"},
	{Name: "gofish", Category: CategoryClassic, Description: "Go Fish (ゴーフィッシュ)"},
	{Name: "pinochle", Category: CategoryClassic, Description: "Pinochle (ピノクル)"},
	{Name: "golf", Category: CategorySolo, Description: "Golf Solitaire (ゴルフ)"},
	{Name: "pigtail", Category: CategoryClassic, Description: "Pig's Tail (ブタのしっぽ)"},
	{Name: "sevencardstud", Category: CategoryCasino, Description: "Seven Card Stud (セブンカードスタッド)"},
	{Name: "clocksolitaire", Category: CategorySolo, Description: "Clock Solitaire (クロックソリティア)"},
	{Name: "durak", Category: CategoryClassic, Description: "Durak / Fool (ドゥラーク)"},
	{Name: "fortythieves", Category: CategorySolo, Description: "Forty Thieves (フォーティシーブス)"},
	{Name: "paigow", Category: CategoryCasino, Description: "Pai Gow Poker (パイガオポーカー)"},
	{Name: "twotenjack", Category: CategoryClassic, Description: "Two Ten Jack (ツーテンジャック)"},
	{Name: "caribbeanstud", Category: CategoryCasino, Description: "Caribbean Stud Poker (カリビアンスタッドポーカー)"},
	{Name: "war", Category: CategoryClassic, Description: "War (戦争)"},
	{Name: "canfield", Category: CategorySolo, Description: "Canfield Solitaire (キャンフィールド)"},
	{Name: "fiftyone", Category: CategoryClassic, Description: "Fifty-one (フィフティワン)"},
	{Name: "yukon", Category: CategorySolo, Description: "Yukon Solitaire (ユーコン)"},
	{Name: "whist", Category: CategoryClassic, Description: "Whist (ホイスト)"},
	{Name: "letitride", Category: CategoryCasino, Description: "Let It Ride (レット・イット・ライド)"},
	{Name: "pokersquares", Category: CategorySolo, Description: "Poker Squares (ポーカー・スクエアズ)"},
	{Name: "pageone", Category: CategoryClassic, Description: "Page One (ページワン)"},
	{Name: "reddog", Category: CategoryCasino, Description: "Red Dog (レッドドッグ)"},
	{Name: "badugi", Category: CategoryCasino, Description: "Badugi (バドゥーギ)"},
	{Name: "razz", Category: CategoryCasino, Description: "Razz (ラズ)"},
	{Name: "scorpion", Category: CategorySolo, Description: "Scorpion (スコーピオン)"},
	{Name: "accordion", Category: CategorySolo, Description: "Accordion (アコーディオン)"},
	{Name: "trash", Category: CategoryClassic, Description: "Trash (トラッシュ)"},
}

// All returns a value-level copy of the registry in canonical order.
// Returning values (not pointers) ensures callers cannot mutate the global
// registry's fields — the two function-pointer fields are cheap to copy.
func All() []Game {
	out := make([]Game, len(registry))
	for i, g := range registry {
		out[i] = *g
	}
	return out
}

// ByCategory returns the games assigned to cat in canonical order. Returned
// values are copies of the registry entries (see All).
func ByCategory(cat Category) []Game {
	var out []Game
	for _, g := range registry {
		if g.Category == cat {
			out = append(out, *g)
		}
	}
	return out
}

// descriptionCache holds Name→Description for every registered game. Built
// once at package load from the static registry; safe to return by reference
// because the registry is never mutated after init.
var descriptionCache = func() map[string]string {
	m := make(map[string]string, len(registry))
	for _, g := range registry {
		m[g.Name] = g.Description
	}
	return m
}()

// Descriptions returns Name→Description for every registered game (issue
// #1459 SSoT). The returned map is the package's cached copy — callers must
// not mutate it. Previously this allocated a fresh map on every call, which
// turned entry.Description() lookups in CLI loops into O(N²) work.
func Descriptions() map[string]string {
	return descriptionCache
}

// Description returns the description for a single game by name, or "" if
// name is unknown. O(1) via the cached descriptions map; safe to call in
// tight loops.
func Description(name string) string {
	return descriptionCache[name]
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
