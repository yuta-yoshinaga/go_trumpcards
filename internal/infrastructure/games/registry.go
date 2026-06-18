// Package games is the single source of truth for the set of games exposed
// by the HTTP server and the Cloudflare Workers.
//
// The registry is split across build-tagged files so that Cloudflare Worker
// binaries (TinyGo / WASM) stay under the 1 MB gzipped free-tier limit:
//
//   - registry.go (this file, no tag)  — types and bare metadata (Name +
//     Category) for all 120 games. Cheap; no references to game code.
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
	{Name: "bigtwo", Category: CategoryClassic},
	{Name: "sevens", Category: CategoryClassic},
	{Name: "doubt", Category: CategoryClassic},
	{Name: "holdem", Category: CategoryCasino},
	{Name: "omaha", Category: CategoryCasino},
	{Name: "omahahilo", Category: CategoryCasino},
	{Name: "bigo", Category: CategoryCasino},
	{Name: "bigohilo", Category: CategoryCasino},
	{Name: "shortdeck", Category: CategoryCasino},
	{Name: "pineapple", Category: CategoryCasino},
	{Name: "crazypineapple", Category: CategoryCasino},
	{Name: "irishpoker", Category: CategoryCasino},
	{Name: "hearts", Category: CategoryClassic},
	{Name: "memory", Category: CategorySolo},
	{Name: "klondike", Category: CategorySolo},
	{Name: "freecell", Category: CategorySolo},
	{Name: "seahaventowers", Category: CategorySolo},
	{Name: "cruel", Category: CategorySolo},
	{Name: "baccarat", Category: CategoryCasino},
	{Name: "spades", Category: CategoryClassic},
	{Name: "crazyeights", Category: CategoryClassic},
	{Name: "ginrummy", Category: CategorySolo},
	{Name: "canasta", Category: CategorySolo},
	{Name: "spider", Category: CategorySolo},
	// Napoleon is a trick-taking game bucketed into the casino worker purely for
	// binary-size balancing (#2126): it is one of the heaviest games. Category is
	// only a size bucket.
	{Name: "napoleon", Category: CategoryCasino},
	{Name: "indianpoker", Category: CategoryCasino},
	{Name: "videopoker", Category: CategoryCasino},
	{Name: "deuceswild", Category: CategoryCasino},
	{Name: "jokerpoker", Category: CategoryCasino},
	// Euchre is a trick-taking game bucketed into the SOLO worker purely for
	// binary-size balancing (#2126): the classic worker is the constrained one,
	// and solo has more headroom than casino now. Category is only a size bucket.
	{Name: "euchre", Category: CategorySolo},
	{Name: "pyramid", Category: CategorySolo},
	{Name: "tripeaks", Category: CategorySolo},
	{Name: "cribbage", Category: CategorySolo},
	{Name: "threecard", Category: CategoryCasino},
	{Name: "ohhell", Category: CategoryClassic},
	// Bridge is a trick-taking game bucketed into the casino worker purely for
	// binary-size balancing (#2126): it is one of the heaviest games and the
	// classic worker is at the 1 MB gzip limit. Category is only a size bucket.
	{Name: "bridge", Category: CategoryCasino},
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
	{Name: "texasholdembonus", Category: CategoryCasino},
	{Name: "war", Category: CategoryClassic},
	{Name: "canfield", Category: CategorySolo},
	{Name: "fiftyone", Category: CategoryClassic},
	{Name: "yukon", Category: CategorySolo},
	{Name: "russiansolitaire", Category: CategorySolo},
	{Name: "whist", Category: CategoryClassic},
	{Name: "letitride", Category: CategoryCasino},
	{Name: "pokersquares", Category: CategorySolo},
	{Name: "pageone", Category: CategoryClassic},
	{Name: "reddog", Category: CategoryCasino},
	{Name: "badugi", Category: CategoryCasino},
	{Name: "deucetoseven", Category: CategoryCasino},
	{Name: "razz", Category: CategoryCasino},
	{Name: "scorpion", Category: CategorySolo},
	{Name: "wasp", Category: CategorySolo},
	{Name: "accordion", Category: CategorySolo},
	{Name: "trash", Category: CategoryClassic},
	{Name: "sevenbridge", Category: CategorySolo},
	{Name: "president", Category: CategoryClassic},
	{Name: "cassino", Category: CategoryClassic},
	{Name: "spanish21", Category: CategoryCasino},
	{Name: "calculation", Category: CategorySolo},
	{Name: "spiteandmalice", Category: CategoryClassic},
	// Skat is a trick-taking game bucketed into the casino worker purely for
	// binary-size balancing (#2126). Category is only a size bucket.
	{Name: "skat", Category: CategoryCasino},
	{Name: "shithead", Category: CategoryClassic},
	{Name: "nertz", Category: CategoryClassic},
	{Name: "slapjack", Category: CategoryClassic},
	{Name: "egyptianratscrew", Category: CategoryClassic},
	{Name: "bakersdozen", Category: CategorySolo},
	{Name: "tonk", Category: CategoryClassic},
	{Name: "casinowar", Category: CategoryCasino},
	{Name: "pitch", Category: CategoryClassic},
	{Name: "dragontiger", Category: CategoryCasino},
	{Name: "blackjackswitch", Category: CategoryCasino},
	{Name: "montecarlo", Category: CategorySolo},
	{Name: "contractrummy", Category: CategorySolo},
	{Name: "ultimatetexasholdem", Category: CategoryCasino},
	{Name: "crescent", Category: CategorySolo},
	{Name: "mississippistud", Category: CategoryCasino},
	// Belote is bucketed into the casino worker purely for binary-size balancing
	// (#2126). Category is only a size bucket.
	{Name: "belote", Category: CategoryCasino},
	{Name: "spiderette", Category: CategorySolo},
	// Mighty is a trick-taking game, but it is bucketed into the casino worker
	// purely for binary-size balancing (#2126): it is one of the heaviest games
	// and the classic worker is at the 1 MB gzip limit. Category is only a
	// per-worker size bucket with no user-facing meaning.
	{Name: "mighty", Category: CategoryCasino},
	{Name: "oasispoker", Category: CategoryCasino},
	{Name: "beleagueredcastle", Category: CategorySolo},
	// Piquet is a trick-taking game bucketed into the SOLO worker purely for
	// binary-size balancing (#2126). Category is only a size bucket.
	{Name: "piquet", Category: CategorySolo},
	{Name: "casinoholdem", Category: CategoryCasino},
	{Name: "callbreak", Category: CategoryClassic},
	// Tarneeb is a trick-taking game bucketed into the casino worker purely for
	// binary-size balancing (#2126). Category is only a size bucket.
	{Name: "tarneeb", Category: CategoryCasino},
	{Name: "highcardflush", Category: CategoryCasino},
	{Name: "briscola", Category: CategoryClassic},
	{Name: "gaps", Category: CategorySolo},
	{Name: "fourcardpoker", Category: CategoryCasino},
	{Name: "rummy500", Category: CategorySolo},
	{Name: "eightoff", Category: CategorySolo},
	{Name: "russianpoker", Category: CategoryCasino},
	{Name: "penguin", Category: CategorySolo},
	{Name: "chinesepoker", Category: CategoryCasino},
	{Name: "sixcardgolf", Category: CategoryClassic},
	// Dou Dizhu is bucketed into the casino worker purely for binary-size
	// balancing (#2126). Category is only a size bucket.
	{Name: "doudizhu", Category: CategoryClassic},
	{Name: "truco", Category: CategoryClassic},
	{Name: "scopa", Category: CategoryClassic},
	{Name: "acesup", Category: CategorySolo},
	// Barbu is a classic compendium trick-taking game, but it is bucketed into
	// the solo worker because the classic worker is at the 1 MB gzip free-tier
	// limit. Category here is purely a binary-size bucket (see package doc).
	{Name: "barbu", Category: CategorySolo},
	// Macau is a classic Crazy Eights variant, but it is bucketed into the solo
	// worker because the classic worker is at the 1 MB gzip free-tier limit.
	// Category here is purely a binary-size bucket (see package doc).
	{Name: "macau", Category: CategorySolo},
	// Thirty-One (Scat) is a classic draw-and-discard pub game, but it is
	// bucketed into the solo worker because the classic worker is at the 1 MB
	// gzip free-tier limit. Category here is purely a binary-size bucket.
	{Name: "thirtyone", Category: CategorySolo},
	// Tien Len (Vietnamese Big Two) is a classic shedding game, but it is
	// bucketed into the solo worker because the classic worker is at the 1 MB
	// gzip free-tier limit. Category here is purely a binary-size bucket.
	{Name: "tienlen", Category: CategorySolo},
	// Osmosis (浸透) is a foundation-only solitaire bucketed into the solo
	// worker. Category here is purely a binary-size bucket.
	{Name: "osmosis", Category: CategorySolo},
	// 500 (Five Hundred) is a trick-taking game (auction + kitty exchange +
	// bowers/joker). It is bucketed into the solo worker only because the
	// classic worker is at the 1 MB gzip free-tier limit. Category here is
	// purely a binary-size bucket.
	{Name: "fivehundred", Category: CategorySolo},
	// Schnapsen / Sixty-Six is a 2-player trick-taking game (marriages + draw
	// from stock). It is bucketed into the solo worker only because the classic
	// worker is at the 1 MB gzip free-tier limit. Category here is purely a
	// binary-size bucket.
	{Name: "schnapsen", Category: CategorySolo},
	// Burraco is a Canasta-derived rummy game. Bucketed into the solo worker
	// (the classic worker is at the 1 MB gzip free-tier limit). Category here is
	// purely a binary-size bucket.
	{Name: "burraco", Category: CategorySolo},
	// Yaniv (ヤニブ) is a draw-and-discard hand-reduction game. The issue
	// proposed the classic worker, but that worker is at the 1 MB gzip free-tier
	// limit, so Yaniv is bucketed into the casino worker. Category here is purely
	// a binary-size bucket (see package doc).
	{Name: "yaniv", Category: CategoryCasino},
	// Gong Zhu (拱猪 / Chinese Hearts) is a trick-taking game with positive and
	// negative point cards, a doubling card, and an exposure phase. The issue
	// proposed the classic worker, but that worker is at the 1 MB gzip free-tier
	// limit, so Gong Zhu is bucketed into the solo worker (most headroom).
	// Category here is purely a binary-size bucket (see package doc).
	{Name: "gongzhu", Category: CategorySolo},
	// Bristol is a tableau/reserve solitaire (build-down tableau, 3 fans, stock).
	// Bucketed into the solo worker (classic worker is at the 1 MB gzip limit).
	{Name: "bristol", Category: CategorySolo},
	// Bid Whist is a 4-player partnership trick-taking game with jokers, a 6-card
	// kitty and Uptown/Downtown/No-Trump bidding. Conceptually a "classic"
	// trick-taker, but the classic worker is at the 1 MB gzip limit, so it is
	// bucketed into the solo worker. Category here is purely a binary-size bucket.
	{Name: "bidwhist", Category: CategorySolo},
	// Tressette (トレセッテ) is an Italian no-trump must-follow trick-taking
	// team game on the 40-card Briscola deck. The issue proposed the classic
	// worker, but that worker is at the 1 MB gzip free-tier limit, so Tressette
	// is bucketed into the casino worker (most headroom). Category here is purely
	// a binary-size bucket (see package doc).
	{Name: "tressette", Category: CategoryCasino},
	// Easthaven (イーストヘイブン) is a Klondike/Spider hybrid solitaire:
	// alternating-color descending tableau with A-K foundations (Klondike) but
	// a Spider-style stock that deals one card to every column. Solo worker.
	{Name: "easthaven", Category: CategorySolo},
	// Tichu (ティチュー) is a 4-player partnership shedding game (Daifugo-like
	// combinations + special cards Dragon/Phoenix/Dog/Mahjong). The issue proposed
	// the classic worker, but that worker is at the 1 MB gzip free-tier limit, so
	// Tichu is bucketed into the casino worker. Category is purely a binary-size
	// bucket here (see package doc).
	{Name: "tichu", Category: CategoryClassic},
	// Baker's Game is FreeCell's same-suit ancestor; it reuses the FreeCell
	// engine (domain.NewDefaultBakersGame) and ships in the solo worker.
	{Name: "bakersgame", Category: CategorySolo},
	// Bourré fuses poker-style ante/draw betting with must-follow trick-taking.
	// Category is purely a binary-size bucket here (see package doc).
	{Name: "bourre", Category: CategoryCasino},
	// Sheepshead (シープスヘッド) is a German-American 5-player trick-taking game
	// with a fixed-trump system (all Queens + all Jacks + all Diamonds) and a
	// secret picker/partner formed via a called Ace. Casino worker (binary-size
	// bucket only; see package doc).
	{Name: "sheepshead", Category: CategoryCasino},
	// Doppelkopf (ドッペルコップ) is a German 4-player partnership trick-taking
	// game on a doubled 48-card deck with a fixed trump (♥10 Dulle + all Q + all
	// J + all ♦) and secret Re/Kontra teams formed by the two Q♣ holders. Casino
	// worker (binary-size bucket only; see package doc).
	{Name: "doppelkopf", Category: CategoryCasino},
	// Mus (ムス) is a Basque 4-player 2-team vying (betting) game on a 40-card
	// Latin deck: four wager rounds (Grande/Chica/Pares/Juego) with paso/envido/
	// ordago and a mus card-exchange phase. Casino worker (binary-size bucket).
	{Name: "mus", Category: CategoryCasino},
	// Tute (トゥーテ) is a Spanish 40-card trump trick-taking game for 4 players
	// (2v2) with K+Q marriage declarations (cante) and a 4-King/4-Queen instant
	// win. Casino worker (binary-size bucket).
	{Name: "tute", Category: CategoryCasino},
	// Sueca (スエカ) is a Portuguese/Brazilian 40-card trump trick-taking game for
	// 4 players (2v2) with A=11/7=10 scoring. Casino worker (binary-size bucket).
	{Name: "sueca", Category: CategoryCasino},
	{Name: "fortyfives", Category: CategoryCasino},
	{Name: "twentynine", Category: CategoryCasino},
	// Klaverjas (クラヴァヤス) is a Dutch Jass-family trump trick-taking game for
	// 4 players (2v2) with the J(20)>9(14) trump rank and Roem melds. Casino
	// worker (binary-size bucket).
	{Name: "klaverjas", Category: CategoryClassic},
	{Name: "manille", Category: CategoryClassic},
	{Name: "marias", Category: CategoryClassic},
	{Name: "sedma", Category: CategoryClassic},
	{Name: "solowhist", Category: CategoryClassic},
	{Name: "knockoutwhist", Category: CategoryClassic},
	{Name: "nap", Category: CategoryClassic},
	{Name: "preference", Category: CategoryClassic},
	{Name: "spoilfive", Category: CategoryClassic},
	// Court Piece (コートピース / Rang / Hokm) is a Pakistani/Iranian 4-player
	// (2v2) trick-taking game where the caller declares trump after peeking at
	// the first 5 cards; 7+ tricks wins the round (Sar), consecutive wins score
	// a Court bonus. Casino worker (binary-size bucket).
	{Name: "courtpiece", Category: CategoryCasino},
	// Bezique (ベジーク) is a French 2-player declaration trick game (the ancestor
	// of Pinochle) using a 64-card deck. Trick winners declare melds (marriages,
	// Bezique = ♠Q+♦J, four-of-a-kind); after the stock empties play becomes
	// strict must-follow. Classic worker (binary-size bucket — casino was full).
	{Name: "bezique", Category: CategoryClassic},
	// Écarté (エカルテ) is a French 2-player trick game (32-card deck) with an
	// exchange-negotiation phase (propose/accept/refuse/discard), King-of-trump
	// and Vole bonuses, then 5 strict-follow tricks. Casino worker bucket.
	{Name: "ecarte", Category: CategoryCasino},
	// Three Card Brag (スリーカード・ブラグ) is a British 3-card vying/betting game
	// (an ancestor of poker) for 4 players with Blind/Seen betting, ante/pot, and
	// the ranking Prial > Running Flush > Run > Flush > Pair > High Card. Casino
	// worker bucket.
	{Name: "threecardbrag", Category: CategoryCasino},
	// Teen Patti (ティーンパッティ) is the South-Asian version of Three Card Brag:
	// 4-player Blind/Seen betting on a 52-card deck, sharing Brag's 3-card hand
	// ranking, plus a Side Show (request a private hand comparison with the
	// previous Seen player). Casino worker bucket.
	{Name: "teenpatti", Category: CategoryCasino},
	// Scopone (スコポーネ) is the 4-player, 2-team "scientific" version of Scopa:
	// all 40 cards are dealt at once and captured by summing to the played card's
	// value, scoring carte/denari/sevens/settebello/scopa per team. Classic worker
	// bucket — it reuses Scopa's (classic) capture & scoring helpers.
	{Name: "scopone", Category: CategoryClassic},
	// Escoba (エスコバ) is a Spanish 4-player free-for-all capture game in the
	// Scopa family: capture table cards summing to exactly 15 (figures J/Q/K =
	// 8/9/10), with Escoba sweeps and Espada/Oro/seven scoring. Classic worker
	// bucket — it reuses Scopa's (classic) subset-sum capture helper.
	{Name: "escoba", Category: CategoryClassic},
	// Hand and Foot: Canasta-family two-stage game (each player holds a "hand"
	// and a "foot"), 4 players / 2 teams, 216-card deck (4 decks + 8 jokers).
	// Solo worker bucket — it reuses Canasta's (solo) meld/canasta/red-3 helpers.
	{Name: "handandfoot", Category: CategorySolo},
	// Conquian: the Mexican 2-player ancestor of rummy. 40-card Latin deck
	// (standard 52 minus 8/9/10), table melds (sets + runs with 7–J adjacency),
	// forced use of a taken discard, win by melding out the whole hand. Solo
	// worker bucket — it reuses Gin Rummy's (untagged) meld helpers.
	{Name: "conquian", Category: CategorySolo},
	// Chinchón: Spanish/Argentine 7-card rummy in the Gin Rummy family. 40-card
	// Latin deck (no 8/9/10), draw/knock/layoff with deadwood scoring, plus the
	// "Chinchón" instant win (7 consecutive cards of one suit). Solo worker
	// bucket — it reuses Gin Rummy's (untagged) meld/deadwood helpers.
	{Name: "chinchon", Category: CategorySolo},
	// Kalooki: Jamaican/British joker-wild rummy, 2–4 players, two 52-card decks
	// plus 2 jokers (106 cards). First melds must total ≥51 points (opening
	// requirement); jokers are wild and a meld containing one scores 1.5×. Solo
	// worker bucket — it reuses the Contract Rummy meld/layoff structure.
	{Name: "kalooki", Category: CategorySolo},
	// Three Thirteen: American progressive rummy, 2–4 players, two 52-card decks
	// (104 cards). Eleven rounds deal 3..13 cards; the rank equal to the deal
	// count is wild that round. Lowest cumulative deadwood after round 11 wins.
	// Solo worker bucket — it reuses the Gin Rummy meld/deadwood approach.
	{Name: "threethirteen", Category: CategorySolo},
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

// AllCategories returns every Category value in canonical display order
// (casino, classic, solo). The returned slice is fresh per call so callers
// cannot mutate package state. Adding a new Category value to the iota above
// requires extending this slice — that intentional coupling is the SSoT
// guarantee that consumers (e.g. the CLI --help summary) cannot drift out
// of sync with the registry.
func AllCategories() []Category {
	return []Category{CategoryCasino, CategoryClassic, CategorySolo}
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
