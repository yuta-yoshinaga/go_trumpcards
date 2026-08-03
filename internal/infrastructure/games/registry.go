// Package games is the single source of truth for the set of games exposed
// by the HTTP server and the Cloudflare Workers.
//
// The registry is split across build-tagged files so that Cloudflare Worker
// binaries (TinyGo / WASM) stay under the 1 MB gzipped free-tier limit:
//
//   - registry.go (this file, no tag)  — types and bare metadata (Name +
//     Category) for all 219 games. Cheap; no references to game code.
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
	// CategoryExtra is the fourth size bucket, added when the other three
	// approached the 1 MB gzip per-worker limit. Like the others it is purely
	// a binary-size bucket, not a user-facing taxonomy: it holds an overflow
	// mix of rummy, shedding/matching and light banking games moved off the
	// other workers to keep every binary under the free-tier limit.
	CategoryExtra
	// CategoryExtra2 and CategoryExtra3 are the fifth and sixth size buckets
	// (ADR-0036), added when all four earlier workers reached 93-99% of the
	// 1 MB gzip limit at 219 games. The deliberately colourless names restate
	// what the doc comments have said since ADR-0027: a Category is a
	// binary-size bucket, never a taxonomy. A name like "puzzle" would invite
	// the next person to reason about where a game "belongs".
	//
	// New games go to whichever worker has the most headroom. Do not think
	// about genre.
	CategoryExtra2
	// CategoryExtra3 is the sixth size bucket. See CategoryExtra2.
	CategoryExtra3
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
	case CategoryExtra:
		return "extra"
	case CategoryExtra2:
		return "extra2"
	case CategoryExtra3:
		return "extra3"
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
	{Name: "bigtwo", Category: CategoryExtra2},
	{Name: "sevens", Category: CategoryClassic},
	{Name: "doubt", Category: CategoryExtra2},
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
	{Name: "ginrummy", Category: CategoryExtra},
	// Indian Rummy (13-card) is a draw-and-discard rummy.
	{Name: "indianrummy", Category: CategoryExtra},
	{Name: "canasta", Category: CategoryExtra},
	{Name: "spider", Category: CategorySolo},
	// Napoleon is a trick-taking game.
	{Name: "napoleon", Category: CategoryCasino},
	{Name: "indianpoker", Category: CategoryCasino},
	{Name: "videopoker", Category: CategoryCasino},
	{Name: "deuceswild", Category: CategoryCasino},
	{Name: "jokerpoker", Category: CategoryCasino},
	// Euchre is a trick-taking game.
	{Name: "euchre", Category: CategorySolo},
	{Name: "pyramid", Category: CategorySolo},
	{Name: "tripeaks", Category: CategorySolo},
	{Name: "cribbage", Category: CategoryExtra3},
	{Name: "threecard", Category: CategoryCasino},
	{Name: "ohhell", Category: CategoryClassic},
	// Ninety-Nine (David Parlett) is a trick-taking game; it shares its trick-play code
	// with ohhell.
	{Name: "ninetynine", Category: CategoryClassic},
	// Bridge is a trick-taking game and one of the heaviest in the registry.
	{Name: "bridge", Category: CategoryExtra3},
	{Name: "speed", Category: CategoryExtra2},
	{Name: "gofish", Category: CategoryExtra2},
	{Name: "pinochle", Category: CategoryExtra2},
	{Name: "golf", Category: CategorySolo},
	{Name: "pigtail", Category: CategoryExtra2},
	{Name: "sevencardstud", Category: CategoryCasino},
	{Name: "clocksolitaire", Category: CategorySolo},
	{Name: "durak", Category: CategoryClassic},
	{Name: "fortythieves", Category: CategorySolo},
	{Name: "paigow", Category: CategoryCasino},
	{Name: "twotenjack", Category: CategoryClassic},
	{Name: "caribbeanstud", Category: CategoryCasino},
	{Name: "texasholdembonus", Category: CategoryCasino},
	{Name: "war", Category: CategoryExtra2},
	{Name: "canfield", Category: CategorySolo},
	{Name: "fiftyone", Category: CategoryExtra2},
	{Name: "yukon", Category: CategorySolo},
	{Name: "russiansolitaire", Category: CategorySolo},
	{Name: "whist", Category: CategoryClassic},
	{Name: "catchten", Category: CategoryClassic},
	{Name: "letitride", Category: CategoryCasino},
	{Name: "pokersquares", Category: CategorySolo},
	{Name: "pageone", Category: CategoryClassic},
	{Name: "reddog", Category: CategoryCasino},
	{Name: "badugi", Category: CategoryCasino},
	{Name: "deucetoseven", Category: CategoryCasino},
	{Name: "razz", Category: CategoryCasino},

	// Seven Card Stud Hi-Lo splits the pot with an eight-or-better low. It shares
	// the whole stud stack -- only the showdown differs -- so it must sit in the
	// same bucket as the poker evaluator it reuses.
	{Name: "sevencardstudhilo", Category: CategoryCasino},
	{Name: "scorpion", Category: CategorySolo},
	{Name: "wasp", Category: CategorySolo},
	{Name: "accordion", Category: CategorySolo},
	{Name: "trash", Category: CategoryExtra2},
	{Name: "sevenbridge", Category: CategoryExtra3},
	{Name: "president", Category: CategoryClassic},
	{Name: "cassino", Category: CategoryClassic},
	{Name: "spanish21", Category: CategoryCasino},
	{Name: "calculation", Category: CategorySolo},
	// Sir Tommy is one of the oldest recorded patiences: deal the stock one card at
	// a time, open foundations with Aces, and build them up A-K ignoring suit.
	{Name: "sirtommy", Category: CategoryExtra2},
	// Bisley deals every card face-up and runs two foundation sets at once:
	// four ascending by suit from the dealt Aces, four descending that open
	// when their King is played.
	{Name: "bisley", Category: CategoryExtra2},
	// Napoleon's Square is a two-deck square tableau whose same-suit runs move as
	// a unit; eight foundations sit in the middle, two per suit.
	{Name: "napoleonssquare", Category: CategoryExtra2},
	// Grandfather's Clock builds twelve suit-ordered faces up to the rank of the
	// hour they sit at, wrapping King to Ace; the tableau is a plain down-by-one.
	{Name: "grandfathersclock", Category: CategoryExtra2},
	// Miss Milligan deals eight at a time onto every column and lets a stuck
	// player hold one run aside ("waiving") once the stock is gone.
	{Name: "missmilligan", Category: CategoryExtra2},
	// Duchess (Glenwood) lets the player pick the foundation's starting rank from
	// the top of a reserve fan; empty columns may only be filled from the reserve
	// while any of it remains.
	{Name: "duchess", Category: CategoryExtra2},
	// Windmill builds a cross: one centre foundation running A-K four times
	// through, and four corner foundations running King down to Ace.
	{Name: "windmill", Category: CategoryExtra2},
	// American Toad is the two-deck Canfield: a 20-card reserve, eight columns,
	// and eight foundations that wrap from a player-independent base rank.
	{Name: "americantoad", Category: CategoryExtra2},
	{Name: "spiteandmalice", Category: CategoryExtra2},
	// Skat is a trick-taking game.
	{Name: "skat", Category: CategoryExtra3},
	// Congress deals one card to each of eight piles and keeps the other 96 as
	// the stock; empty piles can only be refilled from the stock or the waste.
	{Name: "congress", Category: CategoryExtra3},
	// Terrace builds its foundations up in ALTERNATING COLOUR from a
	// player-chosen base rank; the 11-card terrace feeds them and nothing else.
	{Name: "terrace", Category: CategoryExtra3},
	// Braid's four braid fields refill from the braid's tail and are the only
	// thing that consumes it; the eight helper fields refill from the waste only.
	{Name: "braid", Category: CategoryExtra2},
	// Pontoon deals every hand face down, including the banker's, and ranks a
	// two-card 21 above a five-card trick above any total.
	{Name: "pontoon", Category: CategoryExtra2},
	// Sette e Mezzo plays to 7.5 on a 40-card deck where face cards are worth
	// half a point, and the king of coins is wild for 0.5 or 1-7.
	{Name: "settemezzo", Category: CategoryExtra2},
	// Niu Niu finds three of five cards summing to a multiple of ten; the
	// remaining pair's last digit is the rank, and the multiplier follows it.
	{Name: "niuniu", Category: CategoryExtra3},

	// Bura leads up to three cards of one suit at a time and is won by
	// claiming 31 card points, not by reaching them.
	{Name: "bura", Category: CategoryExtra3},

	// Mushi is hanafuda on 40 cards -- June and July are out -- and is a
	// hana-awase game: no koi-koi stop, the round runs to the last card.
	{Name: "mushi", Category: CategoryExtra2},

	// Toepen's ranking is the standard order inverted -- 10 high, jack low --
	// and only the winner of the final trick escapes the penalty.
	{Name: "toepen", Category: CategoryExtra3},

	// Chinese Ten captures by summing to ten (A-9) or by rank (10-K), and only
	// the RED cards score.
	{Name: "chineseten", Category: CategoryExtra2},

	// Skitgubbe's two phases are different games: a two-player duel that
	// collects cards, then a durak-style beat-or-pick-up shed.
	{Name: "trex", Category: CategoryExtra3},

	{Name: "skitgubbe", Category: CategoryExtra3},

	// Laugh and Lie Down is a 17th-century fishing game: one card takes one or
	// three of a rank, and a player who cannot capture feeds their whole hand
	// to the table.
	// Sjavs has six permanent trumps (both black queens and all four jacks), so
	// a "trump" is not the same thing as "a card of the trump suit".
	{Name: "loba", Category: CategoryExtra2},

	{Name: "sjavs", Category: CategoryExtra2},

	{Name: "laughandliedown", Category: CategoryExtra2},

	{Name: "shithead", Category: CategoryClassic},
	{Name: "nertz", Category: CategoryExtra2},
	{Name: "slapjack", Category: CategoryClassic},
	{Name: "egyptianratscrew", Category: CategoryClassic},
	{Name: "bakersdozen", Category: CategorySolo},
	{Name: "tonk", Category: CategoryClassic},
	{Name: "casinowar", Category: CategoryCasino},
	{Name: "pitch", Category: CategoryClassic},
	{Name: "dragontiger", Category: CategoryCasino},
	{Name: "blackjackswitch", Category: CategoryCasino},
	{Name: "montecarlo", Category: CategorySolo},
	{Name: "contractrummy", Category: CategoryExtra},
	{Name: "ultimatetexasholdem", Category: CategoryCasino},
	{Name: "crescent", Category: CategorySolo},
	{Name: "mississippistud", Category: CategoryCasino},
	// Belote is a 4-player partnership trick-taking game on the 32-card deck.
	{Name: "belote", Category: CategoryExtra3},
	{Name: "spiderette", Category: CategorySolo},
	// Mighty is a trick-taking game.
	{Name: "mighty", Category: CategoryExtra2},
	{Name: "oasispoker", Category: CategoryCasino},
	{Name: "beleagueredcastle", Category: CategorySolo},
	// Streets and Alleys is a Beleaguered Castle variant.
	{Name: "streetsandalleys", Category: CategoryExtra},
	// King Albert is an English open patience (FreeCell family).
	{Name: "kingalbert", Category: CategoryExtra},
	{Name: "flowergarden", Category: CategoryExtra},
	{Name: "fortyandeight", Category: CategoryExtra3},
	// Agnes Sorel is a Klondike+Canfield hybrid patience.
	{Name: "agnes", Category: CategoryExtra},
	// Sultan of Turkey is a two-deck King-foundation patience.
	{Name: "sultan", Category: CategoryExtra},
	// Piquet is a trick-taking game.
	{Name: "piquet", Category: CategoryExtra3},
	{Name: "casinoholdem", Category: CategoryCasino},
	{Name: "callbreak", Category: CategoryClassic},
	// Tarneeb is a trick-taking game.
	{Name: "tarneeb", Category: CategoryCasino},
	{Name: "highcardflush", Category: CategoryCasino},
	{Name: "briscola", Category: CategoryClassic},
	{Name: "gaps", Category: CategorySolo},
	{Name: "fourcardpoker", Category: CategoryCasino},
	{Name: "rummy500", Category: CategoryExtra},
	{Name: "eightoff", Category: CategorySolo},
	{Name: "russianpoker", Category: CategoryCasino},
	{Name: "penguin", Category: CategorySolo},
	{Name: "chinesepoker", Category: CategoryCasino},
	{Name: "sixcardgolf", Category: CategoryExtra2},
	// Dou Dizhu (fight the landlord) is a 3-player climbing/shedding game.
	{Name: "doudizhu", Category: CategoryClassic},
	{Name: "truco", Category: CategoryClassic},
	{Name: "scopa", Category: CategoryClassic},
	{Name: "acesup", Category: CategorySolo},
	// Barbu is a compendium trick-taking game.
	{Name: "barbu", Category: CategorySolo},
	// Macau is a Crazy Eights variant.
	{Name: "macau", Category: CategorySolo},
	// Thirty-One (Scat) is a draw-and-discard pub game.
	{Name: "thirtyone", Category: CategorySolo},
	// Tien Len (Vietnamese Big Two) is a shedding game.
	{Name: "tienlen", Category: CategorySolo},
	// Osmosis (浸透) is a foundation-only solitaire.
	{Name: "osmosis", Category: CategorySolo},
	// 500 (Five Hundred) is a trick-taking game (auction + kitty exchange + bowers/joker).
	{Name: "fivehundred", Category: CategorySolo},
	// Schnapsen / Sixty-Six is a 2-player trick-taking game (marriages + draw from stock).
	{Name: "schnapsen", Category: CategorySolo},
	// Burraco is a Canasta-derived rummy game.
	{Name: "burraco", Category: CategoryExtra},
	// Yaniv (ヤニブ) is a draw-and-discard hand-reduction game.
	{Name: "yaniv", Category: CategorySolo},
	// Gong Zhu (拱猪 / Chinese Hearts) is a trick-taking game with positive and negative
	// point cards, a doubling card, and an exposure phase.
	{Name: "gongzhu", Category: CategorySolo},
	// Bristol is a tableau/reserve solitaire (build-down tableau, 3 fans, stock).
	{Name: "bristol", Category: CategorySolo},
	// Bid Whist is a 4-player partnership trick-taking game with jokers, a 6-card kitty
	// and Uptown/Downtown/No-Trump bidding.
	{Name: "bidwhist", Category: CategorySolo},
	// Tressette (トレセッテ) is an Italian no-trump must-follow trick-taking team game on the
	// 40-card Briscola deck.
	{Name: "tressette", Category: CategoryCasino},
	// Easthaven (イーストヘイブン) is a Klondike/Spider hybrid solitaire: alternating-color
	// descending tableau with A-K foundations (Klondike) but a Spider-style stock that
	// deals one card to every column.
	{Name: "easthaven", Category: CategorySolo},
	// Tichu (ティチュー) is a 4-player partnership shedding game (Daifugo-like combinations +
	// special cards Dragon/Phoenix/Dog/Mahjong).
	{Name: "tichu", Category: CategoryExtra2},
	// Baker's Game is FreeCell's same-suit ancestor; it reuses the FreeCell engine
	// (domain.NewDefaultBakersGame).
	{Name: "bakersgame", Category: CategorySolo},
	// Bourré fuses poker-style ante/draw betting with must-follow trick-taking.
	{Name: "bourre", Category: CategoryCasino},
	// Sheepshead (シープスヘッド) is a German-American 5-player trick-taking game with a fixed-
	// trump system (all Queens + all Jacks + all Diamonds) and a secret picker/partner
	// formed via a called Ace.
	{Name: "sheepshead", Category: CategoryExtra3},
	// Doppelkopf (ドッペルコップ) is a German 4-player partnership trick-taking game on a doubled
	// 48-card deck with a fixed trump (♥10 Dulle + all Q + all J + all ♦) and secret
	// Re/Kontra teams formed by the two Q♣ holders.
	{Name: "doppelkopf", Category: CategoryCasino},
	// Mus (ムス) is a Basque 4-player 2-team vying (betting) game on a 40-card Latin deck:
	// four wager rounds (Grande/Chica/Pares/Juego) with paso/envido/ ordago and a mus
	// card-exchange phase.
	{Name: "mus", Category: CategoryCasino},
	// Tute (トゥーテ) is a Spanish 40-card trump trick-taking game for 4 players (2v2) with
	// K+Q marriage declarations (cante) and a 4-King/4-Queen instant win.
	{Name: "tute", Category: CategoryCasino},
	// Sueca (スエカ) is a Portuguese/Brazilian 40-card trump trick-taking game for 4 players
	// (2v2) with A=11/7=10 scoring.
	{Name: "sueca", Category: CategoryCasino},
	{Name: "fortyfives", Category: CategoryCasino},
	{Name: "twentynine", Category: CategoryCasino},
	// Klaverjas (クラヴァヤス) is a Dutch Jass-family trump trick-taking game for 4 players
	// (2v2) with the J(20)>9(14) trump rank and Roem melds.
	{Name: "klaverjas", Category: CategoryClassic},
	{Name: "manille", Category: CategoryClassic},
	{Name: "marias", Category: CategoryClassic},
	{Name: "sedma", Category: CategoryClassic},
	{Name: "solowhist", Category: CategoryClassic},
	{Name: "knockoutwhist", Category: CategoryClassic},
	{Name: "nap", Category: CategoryClassic},
	{Name: "preference", Category: CategoryClassic},
	{Name: "ganjifa", Category: CategoryExtra},
	{Name: "vira", Category: CategoryExtra},
	{Name: "spoilfive", Category: CategoryClassic},
	// Court Piece (コートピース / Rang / Hokm) is a Pakistani/Iranian 4-player (2v2) trick-
	// taking game where the caller declares trump after peeking at the first 5 cards; 7+
	// tricks wins the round (Sar), consecutive wins score a Court bonus.
	{Name: "courtpiece", Category: CategoryCasino},
	// Bezique (ベジーク) is a French 2-player declaration trick game (the ancestor of
	// Pinochle) using a 64-card deck. Trick winners declare melds (marriages, Bezique =
	// ♠Q+♦J, four-of-a-kind); after the stock empties play becomes strict must-follow.
	{Name: "bezique", Category: CategoryClassic},
	// Écarté (エカルテ) is a French 2-player trick game (32-card deck) with an exchange-
	// negotiation phase (propose/accept/refuse/discard), King-of-trump and Vole bonuses,
	// then 5 strict-follow tricks.
	{Name: "ecarte", Category: CategoryCasino},
	// Three Card Brag (スリーカード・ブラグ) is a British 3-card vying/betting game (an ancestor of
	// poker) for 4 players with Blind/Seen betting, ante/pot, and the ranking Prial >
	// Running Flush > Run > Flush > Pair > High Card.
	{Name: "threecardbrag", Category: CategoryCasino},
	// Teen Patti (ティーンパッティ) is the South-Asian version of Three Card Brag: 4-player
	// Blind/Seen betting on a 52-card deck, sharing Brag's 3-card hand ranking, plus a
	// Side Show (request a private hand comparison with the previous Seen player).
	{Name: "teenpatti", Category: CategoryCasino},
	// Scopone (スコポーネ) is the 4-player, 2-team "scientific" version of Scopa: all 40 cards
	// are dealt at once and captured by summing to the played card's value, scoring
	// carte/denari/sevens/settebello/scopa per team.
	{Name: "scopone", Category: CategoryClassic},
	// Escoba (エスコバ) is a Spanish 4-player free-for-all capture game in the Scopa family:
	// capture table cards summing to exactly 15 (figures J/Q/K = 8/9/10), with Escoba
	// sweeps and Espada/Oro/seven scoring.
	{Name: "escoba", Category: CategoryClassic},
	// Hand and Foot: Canasta-family two-stage game (each player holds a "hand" and a
	// "foot"), 4 players / 2 teams, 216-card deck (4 decks + 8 jokers).
	{Name: "handandfoot", Category: CategoryExtra},
	// Conquian: the Mexican 2-player ancestor of rummy. 40-card Latin deck (standard 52
	// minus 8/9/10), table melds (sets + runs with 7–J adjacency), forced use of a taken
	// discard, win by melding out the whole hand.
	{Name: "conquian", Category: CategoryExtra},
	// Chinchón: Spanish/Argentine 7-card rummy in the Gin Rummy family. 40-card Latin deck
	// (no 8/9/10), draw/knock/layoff with deadwood scoring, plus the "Chinchón" instant
	// win (7 consecutive cards of one suit).
	{Name: "chinchon", Category: CategoryExtra},
	// Kalooki: Jamaican/British joker-wild rummy, 2–4 players, two 52-card decks plus 2
	// jokers (106 cards). First melds must total ≥51 points (opening requirement); jokers
	// are wild and a meld containing one scores 1.5×.
	{Name: "kalooki", Category: CategoryExtra},
	// Three Thirteen: American progressive rummy, 2–4 players, two 52-card decks (104
	// cards). Eleven rounds deal 3..13 cards; the rank equal to the deal count is wild
	// that round. Lowest cumulative deadwood after round 11 wins.
	{Name: "threethirteen", Category: CategoryExtra},
	// Mao: a Crazy Eights / Macau–style shedding game with a secret "hidden rule" the
	// human must infer (penalties for non-compliance, a half-hint after three correct
	// follows). 4 players, 52-card deck, magic cards (8 wild, A skip, 2 draw-two).
	{Name: "mao", Category: CategoryExtra3},
	// Spoons: American party speed game. 4 players, 52-card deck; pass cards around until
	// someone collects four of a kind, then everyone races to grab one of the N-1 spoons.
	// Missing out earns a letter (S-P-O-O-N-S); six letters eliminates you, last player
	// standing wins.
	{Name: "spoons", Category: CategoryExtra2},
	// Kemps: 4-player, 2-team matching game. Swap cards through a shared field until you
	// collect four of a kind, then your partner signals secretly and your team declares
	// "Kemps!" for a point — or the opponents call "Counter-Kemps!" to steal it. First
	// team to 5 wins.
	{Name: "kemps", Category: CategoryExtra2},
	// Cuckoo (Chase the Ace / Ranter-Go-Round): a European life-survival game. 4 players
	// each hold one card and 3 lives; on your turn keep or swap with your neighbour (a
	// King holder may refuse). The lowest card each round loses a life; last player
	// standing wins.
	{Name: "cuckoo", Category: CategoryExtra2},
	// Pişti: a popular Turkish fishing/capture game. 2–4 players; play a card matching the
	// pile top (or any Jack) to capture the whole pile. Matching a lone card scores a
	// Pişti (+10; +20 for Jack-on-Jack).
	{Name: "pishti", Category: CategoryExtra2},
	// Cuarenta: the national card game of Ecuador. 4 players in 2 teams, 40-card deck (no
	// 8/9/10); capture by rank with caída/ronda/limpia bonuses, first team to 40 points
	// wins.
	{Name: "cuarenta", Category: CategoryExtra2},
	// Five Card Stud: one of the oldest stud poker variants. 2–6 players, one face-down
	// hole card plus four face-up cards dealt over four betting streets (bring-in on the
	// lowest up-card), standard poker showdown.
	{Name: "fivecardstud", Category: CategoryCasino},
	// Faro: a 19th-century American banking game. The player places chips on a 13-rank
	// layout (a copper bets the rank to lose); the bank deals cards in turns of two
	// (losing card then winning card), with a half-collect on splits and a final 3-card
	// call.
	{Name: "faro", Category: CategoryExtra2},
	// Open Face Chinese Poker (OFC): a modern Chinese-poker variant. Players receive cards
	// and place them one at a time into three rows (top/middle/ bottom) face-up with no
	// rearranging; rows must rank bottom >= middle >= top or the hand fouls. Royalties
	// reward strong rows and QQ+ on top earns Fantasyland.
	{Name: "openfacechinese", Category: CategoryCasino},
	// Russian Bank (Crapette): 2-player competitive solitaire on two decks. Each player
	// races to empty their 13-card reserve onto 8 shared foundations (A-up by suit) and 4
	// shared tableau columns (alternating colour, descending), discarding from hand to end
	// a turn. Catch the CPU leaving a forced foundation move with "stop".
	{Name: "russianbank", Category: CategorySolo},
	// La Belle Lucie: classic French fan solitaire. 52 cards are dealt into 17 fans of 3
	// plus a single; only each fan's top card moves, building it down in suit onto another
	// fan or up from the Ace on 4 foundations. When stuck, gather and reshuffle (up to 3
	// redeals).
	{Name: "labellelucie", Category: CategoryClassic},
	// Simple Simon: an easier Spider-family solitaire. All 52 cards are dealt face-up into
	// 10 columns with no stock; move single cards or same-suit descending runs, and a
	// complete K-down-to-A same-suit run is removed. Clear all four suits to win.
	{Name: "simplesimon", Category: CategoryClassic},
	// Double Klondike (Gargantua): a two-deck Klondike. 104 cards over 9 tableau columns
	// and 8 foundations (two A-K piles per suit); deal/draw/waste play as in Klondike.
	// Clear all eight foundations to win.
	{Name: "doubleklondike", Category: CategoryExtra2},
	// Black Hole: a one-deck patience by David Parlett. 51 cards dealt into 17 fans of
	// three around a central foundation (the black hole) seeded with the ♠A; play a fan
	// top whose rank is ±1 (any suit, no K-A wrap) onto the pile. Absorb all 52 cards to
	// win.
	{Name: "blackhole", Category: CategorySolo},
	// Beggar-My-Neighbour: classic English 2-player capture game. 52 cards split evenly;
	// players alternate turning top cards onto a central pile. Penalty cards (J=1, Q=2,
	// K=3, A=4) force the opponent to pay that many cards; a new penalty card during
	// payment flips the obligation. The player who collects all 52 cards wins.
	{Name: "beggarmyneighbour", Category: CategoryExtra2},
	// All Fours (Seven Up / Old Sledge): classic English 2-player trick-taking game with a
	// beg/stand negotiation and a turn-up trump. Each deal scores High/Low/Jack/Game;
	// first to 7 points wins.
	{Name: "allfours", Category: CategoryClassic},
	// Prší (チェコ版クレイジーエイト / Mau Mau): a Czech shedding game on a 32-card pack (7..A). Match
	// the discard top by suit or rank; 7 forces the next player to draw 2 (7s stack), Ace
	// and Under (Jack) skip the next player. First to empty their hand wins.
	{Name: "prsi", Category: CategoryClassic},
	// Jass (Schieber): a 36-card (6..A) 4-player/2-team Swiss trump trick-taker with
	// Schieber bidding, Weis melds, and the Stöck (trump K+Q) bonus. The trump Jack
	// (Bauer) and 9 (Nell) outrank the Ace. Modelled on belote.
	{Name: "jass", Category: CategoryExtra3},
	// Gaigel: a 48-card (A,10,K,Q,J,7 doubled) 4-player/2-team Schwabian point-trick game
	// in the Schnapsen/66 family. A stock/talon refills hands in phase 1 (optional
	// follow); phase 2 enforces must-follow. Marriage (trump K+Q = 40, else 20) scores to
	// the team. First team to 101 wins. Composes jass (4p/2-team structure) + schnapsen
	// (marriage/points/stock).
	{Name: "gaigel", Category: CategoryExtra},
	// Thousand (Tysiąc): a Polish/East-European 3-player bidding trick-taker on a 24-card
	// pack (9,J,Q,K,10,A). Players bid from 100 in +10 steps; the last bidder becomes
	// declarer, takes the 3-card talon (widow) and passes one card to each opponent.
	// Declaring a marriage (K+Q of a suit) on lead sets that suit as trump and scores
	// 40/60/80/100 (♠/♣/♦/♥); trump changes dynamically. Declarer scores ±contract; others
	// round to 10. First to 1000 wins. Modelled on mariáš.
	{Name: "tysiac", Category: CategoryExtra},
	// Calabresella (Terziglio): an Italian (Calabrian) 3-player no-trump trick-taker in
	// the Tressette family. One soloist plays against a 2-player coalition on a 40-card
	// deck (A,2..7,J,Q,K). Each player gets 12 cards; 4 form the monte (widow). Bidding is
	// pass/chiamo (stake 1)/solo (stake 2); the soloist takes the monte and discards down
	// to 12. Tressette rank (3>2>A>K>Q>J>7>6>5>4) and points (11/deal via thirds +
	// ultima); the soloist must take more than half to win.
	{Name: "calabresella", Category: CategoryExtra},
	// Ombre (Hombre): a 17th-century Spanish 3-player soloist-vs-coalition trick-taker,
	// ancestor of all solo games. 40-card deck (A,2..7,J,Q,K); 9 cards each, 13 unused.
	// Bidding is pass/entrar/solo; the winner (Ombre) picks trump and plays alone against
	// the other two. The trump group is Spadille (♠A) > Manille (7 of trump) > Basto (♣A)
	// > Punto (A of a red trump) > K>Q>J>6..2. Must-follow; more tricks than each opponent
	// = Sacar (win), tied = Puesta, beaten = Codille.
	{Name: "ombre", Category: CategoryExtra3},
	// Ulti (Ulti / Ultimó): a Hungarian 3-player contract trick-taker. One declarer (the
	// human) vs a 2-CPU coalition. 32-card deck (A,10,K,Q,J,9,8,7); trick rank
	// A>10>K>Q>J>9>8>7. 10 cards each + a 2-card talon. Reduced ruleset: the declarer non-
	// competitively declares one of three contracts — Party (name trump, take >half the
	// 126 card points), Betli (no trump, lose every trick), or Durchmarsch (no trump, win
	// every trick) — takes the talon, discards 2, then leads 10 tricks. Coin settlement
	// ±2/±5/±6 per defender.
	{Name: "ulti", Category: CategoryExtra3},
	// King (Greek/Brazilian compendium): a 4-player 52-card trick-avoidance game. Each
	// deal the dealer picks one of 7 not-yet-played contracts (No Tricks / No Hearts / No
	// Queens / No King♥ / No Last Two / No Men / King-Trump); the negatives penalise
	// capturing, King-Trump rewards tricks with a chosen trump. Play all 7 contracts once;
	// highest total (least penalty) wins.
	{Name: "king", Category: CategoryExtra},
	// Cinch (Double Pedro / High Five): a 4-player All-Fours/Pitch-family auction trick-
	// taker on a 52-card deck. Deal 9 each; players bid 1-14 or pass; the high bidder
	// names trump and leads. Capture point cards (14/deal): High(A)=1, King=1,
	// Ten("Game")=1, Jack=1, Right Pedro (5 of trump)=5, Left Pedro (5 of same colour as
	// trump)=5. The Left Pedro is treated as a trump ranking just below the trump 5. The
	// bidder's side must make its bid or is set back; first to the target score wins.
	{Name: "cinch", Category: CategoryExtra},
	// Loo (Lanterloo): a classic English pot/gambling trick-taking game. 4 players ante to
	// a carried-over pot, a turn-up sets trump, and each player decides to play or pass.
	// Players who play compete over 5 tricks (must-follow-and-head); each trick wins 1/5
	// of the pot, and a player who plays but takes no trick is "looed" and pays a penalty
	// into the next pot. Chips accumulate over repeated deals (no target-score race).
	{Name: "loo", Category: CategoryExtra3},
	// Basra (Bastra): an Egyptian/Levantine fishing (capture) game on a 52-card deck. 4
	// players (you + 3 CPU, individual scoring); each is dealt 4 cards with 4 face-up on
	// the table. A played number card captures same-rank cards and any table subset
	// summing to its value; a Jack sweeps the whole table (except other Jacks). Clearing
	// the table with a single non-Jack card scores a "Basra" bonus. Deal fresh hands until
	// the stock is exhausted, then score most cards, 7♦, 10♦, each Ace, and each Basra.
	{Name: "basra", Category: CategoryExtra3},
	// Tablanet (Tablić): a Balkan fishing (capture) game on a 52-card deck, closely
	// related to Basra. 4 players (you + 3 CPU, individual scoring); each is dealt 4 cards
	// with 4 face-up on the table. A played number card captures same-rank cards and any
	// table subset summing to its value; a Jack sweeps the whole table (except other
	// Jacks). Clearing the table with a single non-Jack card scores a "Tabla" bonus. Deal
	// fresh hands until the stock is exhausted, then score the traditional Tablanet
	// points: most cards, each Ace, each Jack, 10♦, 2♣, and each Tabla.
	{Name: "tablanet", Category: CategoryExtra3},
	// Trente et Quarante (Rouge et Noir): a French casino banking game — the simplest
	// possible, with no player card decisions. On a 6-deck (312-card) shoe the dealer
	// deals a Noir (black) row then a Rouge (red) row, each until its running total (A=1,
	// 2–10 pip, J/Q/K=10) reaches 31–40; the lower total wins. The player bets, before the
	// deal, on Noir, Rouge, Couleur (first card's color matches the winning row's color)
	// or Inverse (differs). Even-money payout; a tie is a push, except a tie at 31
	// ("Refait") takes half the stake for the house. Chips persist across rounds.
	{Name: "trenteetquarante", Category: CategoryExtra},
	// Guts: a simple American poker-vying pot game on a 52-card deck. 2–7 players ante to
	// a pot and get 2 cards each, then simultaneously declare "in" (stay) or "out" (fold).
	// Among the players who stayed, the best 2-card hand (a pair beats two non-paired
	// cards; else high card then kicker, Ace high) takes the whole pot; every other "in"
	// player must MATCH the pot into the next round's pot — the escalation/penalty. Chips
	// accumulate; the game ends after a fixed number of rounds or when fewer than two
	// players can ante, and the richest player wins.
	{Name: "guts", Category: CategoryExtra},
	// Bouillotte: an 18th-century French poker ancestor, a vying/betting pot game on a
	// 20-card deck (A, K, Q, 9, 8 × 4 suits). 3–4 players ante to a pot, are dealt 3 cards
	// each, and a shared "retourne" card is turned face up. Players bet in turn (call /
	// raise "vie" by the ante, capped; or fold). At showdown the best hand wins the whole
	// pot: a brelan (three of a kind) beats everything — a "favori" (a pair completed by
	// the retourne) beats a same-rank "simple" — otherwise high card wins (ties to the
	// earliest seat). Chips accumulate; the game ends after a fixed number of rounds, and
	// the richest player wins.
	{Name: "bouillotte", Category: CategoryExtra3},
	// Primero: a 16th-century Renaissance vying/betting pot game, an ancestor of poker, on
	// a 40-card deck (A,2,3,4,5,6,7,J,Q,K × 4 suits). 2–6 players ante to a pot and are
	// dealt 4 cards each (no shared card). Players bet in turn (call / raise "vie" by the
	// ante, capped; or fold). At showdown the best hand wins the whole pot, ranked by
	// bespoke prime-point values: a Fluxus (flush) beats a Supremus (four suits, points >=
	// 50), which beats a Primero (four suits, points < 50), which beats a Numerus (best
	// single-suit point sum); ties go to the earliest seat. Chips accumulate; the game
	// ends after a fixed number of rounds, and the richest player wins.
	{Name: "primero", Category: CategoryExtra3},
	// Michigan (a.k.a. Newmarket / Boodle / Chicago): a "stops" family gambling party game
	// on a standard 52-card deck. 3–8 players each spread an ante across four fixed center
	// "boodle" cards (A♥, K♣, Q♦, J♠), then all 52 cards are dealt round-robin to the
	// players plus one face-down "dead hand" (widow). The player left of the dealer leads
	// the lowest card of a suit; the sequence climbs in that suit (♥3→♥4→♥5) passing to
	// whoever holds the next card, until a STOP (the next card is in the dead hand or past
	// the King), when the last player starts a new sequence. Playing a card matching a
	// boodle collects that boodle's chips. The round ends the instant a player empties
	// their hand; unclaimed boodle chips carry over. Chips accumulate; the game ends after
	// a fixed number of rounds, and the richest player wins.
	{Name: "michigan", Category: CategoryExtra3},
	// Watten: a Bavarian/Austrian 4-player/2-team trick-taker on a 32-card pack (7..A)
	// with a bluff-raise stake mechanic. The dealer declares a Schlag rank and a critical
	// (trump) suit; ranking is fixed Max(♥K) > Belli(♦K) > Spitz(♦7) > Schlag cards >
	// critical-suit cards > plain. Teams may raise the deal's stake ("gehen"); the
	// opposing team holds or folds. First team to 15 wins. Modelled on jass (4p/2-team) +
	// truco (raise/respond).
	{Name: "watten", Category: CategoryExtra},
	// Carioca is a South-American contract rummy (7 progressive rounds of set/run
	// contracts) played with 108 cards (two 52-card decks + 4 wild jokers), 3-6 players.
	// Modelled on contractrummy (same 7-round contract table + draw/discard/meld/go-out
	// engine) with a double-deck+jokers deck and configurable player count.
	{Name: "carioca", Category: CategoryExtra},
	// Samba is a Canasta variant that adds sequence melds ("sambas") and a third deck (3
	// decks + 6 jokers = 162 cards). It is a 4-player partnership rummy game (seats 0 & 2
	// vs 1 & 3). Modelled on canasta (same wild-aware set melds, canasta/red-3/take-the-
	// pile/go-out engine) extended with same-suit sequence melds and team scoring.
	{Name: "samba", Category: CategoryExtra},
	// Anaconda ("Pass the Trash") is an American home-poker variant on a 52-card deck, 3-7
	// players. Everyone antes and is dealt 7 cards, then passes cards to the left in three
	// sub-rounds (3, then 2, then 1), keeps the best 5, and reveals them one at a time
	// with a betting round (check/call, raise, fold) before each reveal. The best 5-card
	// poker hand at showdown wins the pot; folding to a single player wins immediately.
	// Chips accumulate; the game ends after a fixed number of rounds and the richest
	// player wins.
	{Name: "anaconda", Category: CategoryExtra},
	// Machiavelli (マキャヴェッリ) is an Italian rummy — Rummikub with cards — where all melds
	// live on a single SHARED TABLE that a player may freely rebuild on their turn (moving
	// cards between melds) as long as every meld stays valid and at least one hand card is
	// added. Two 52-card decks (104 cards, no jokers), 2–5 players, sets (same rank,
	// distinct suits) and runs (same-suit consecutive).
	{Name: "machiavelli", Category: CategoryExtra},
	// Panguingue (Pan) is a multi-deck draw-and-discard rummy.
	{Name: "pan", Category: CategoryExtra},
	// Wizard is a 60-card (52 + 4 wizards + 4 jesters) exact-bid trick-taker; its
	// wizard/jester cards are the first to use the non-52 procedural render path
	// (ADR-0033).
	{Name: "wizard", Category: CategoryExtra3},
	// Oicho-Kabu is a kabufuda (40-card, values 1-10) baccarat-style banking game; its
	// cards use the non-52 procedural render path (ADR-0033).
	{Name: "oichokabu", Category: CategoryExtra},
	// Rook is a 57-card (4 colors 1-14 + Rook bird) 2-team point-trick game; its special-
	// deck cards use the non-52 procedural render path (ADR-0033).
	{Name: "rook", Category: CategoryExtra3},
	// Koi-Koi is a 48-card hanafuda capture game with yaku scoring; the hanafuda cards use
	// the non-52 procedural render path (ADR-0033).
	{Name: "koikoi", Category: CategoryExtra3},
	// Go-Stop (Godori) is a Korean hanafuda capture game (same 48-card Hwatu deck
	// as Koi-Koi) with Gwang/Godori scoring + Go/Stop; procedural render (ADR-0033).
	{Name: "gostop", Category: CategoryExtra},
	// Hachi-Hachi is the classic 3-player Japanese hanafuda game (88-point
	// settlement); reuses the hanafuda deck + procedural render path (ADR-0033).
	{Name: "hachihachi", Category: CategoryExtra},
	// French Tarot is a 78-card tarot trick-taker (4 suits×14 + 21 atouts + Excuse);
	// the first tarot-deck game on the non-52 procedural render path (ADR-0033).
	{Name: "frenchtarot", Category: CategoryExtra},
	// Königrufen is an Austrian tarock trick-taker (54-card tarock deck) with the
	// call-a-king hidden-partnership mechanic; procedural render path (ADR-0033).
	{Name: "koenigrufen", Category: CategoryExtra},
	// Scarto is the simplest Italian (Piedmontese) tarocchi trick-taker on the 78-card
	// tarot deck; procedural render path (ADR-0033).
	// Tarocchini (タロッキーニ / Ottocento) is a Bolognese 62-card tarot game for 4 in
	// fixed 2v2 teams. The four papi rank equal and the LATER-played one wins the trick,
	// which no other game here does; the dealer buries 2 surplus cards (scarto).
	{Name: "tarocchini", Category: CategorySolo},
	{Name: "scarto", Category: CategoryExtra3},
	// Cego is a German (Baden) tarock trick-taker on the 54-card tarock deck with the
	// signature Cego-blind swap; procedural render path (ADR-0033).
	{Name: "cego", Category: CategoryExtra3},
	// Zheng Shangyou is a Chinese climbing/shedding game (ancestor of Big Two / Daifugo)
	// on a 54-card deck (52 + 2 jokers); suits are irrelevant to rank strength.
	{Name: "zheng", Category: CategorySolo},
	// Desmoche is a Nicaraguan rummy: nine dealt, and the pot goes to whoever melds
	// exactly ten cards. Poker hand rankings play no part despite the family
	// resemblance, and "desmoche" itself is the move of reusing a card from one of
	// your own face-up melds in another.
	{Name: "desmoche", Category: CategoryExtra3},
	// Zwicker is a north-German fishing game on 55 cards (52 + three jokers worth a
	// fixed 15/20/25). Unlike Cassino, aces and court cards each carry TWO matching
	// values chosen by the player, and they join sums -- which is why it cannot reuse
	// the Cassino engine. "Zwick" is the bonus for clearing the table, not the name
	// for a multi-group capture.
	{Name: "zwicker", Category: CategoryExtra2},
	// Poch is a 15th-century German three-stage game on 32 cards with a nine-pool
	// board. Pools that go unclaimed carry over, which is what the game runs on.
	// The middle stage compares same-rank sets (4 > 3 > 2) -- there is no bluff and
	// no declaration -- and "Pocher" is one of the nine pools, not a move.
	{Name: "poch", Category: CategoryExtra3},
	// Pope Joan is the ancestor of the stops family: 51 cards (the 8D is removed so a
	// run always dies at the 7D), a board of eight named compartments, and a dead hand
	// whose last card turns for trump. Compartments pay only on the trump suit, and
	// whoever holds the Pope (9D) is excused the per-card payment at the end.
	{Name: "popejoan", Category: CategoryExtra3},
	// Le Nain Jaune is the French stops game on a board of five boxes (D10, CJ, SQ,
	// HK and D7 -- the yellow dwarf itself). Unlike Pope Joan the run IGNORES SUIT and
	// simply climbs by rank, ending on a king; and the loser pays the winner in card
	// POINTS, not in cards.
	{Name: "nainjaune", Category: CategoryExtra3},
	// Kille is the Swedish Cuckoo game played with its own 42-card pack (21
	// denominations twice over, a single suit). Five of the picture cards break the
	// exchange: the Cuckoo ends the round on the spot, the Hussar cuts down the
	// challenger, the Pig unwinds the swap and bites its own holder, and the
	// Cavalier and Inn pass the challenge along to the next seat.
	{Name: "kille", Category: CategoryExtra3},
	// Klaberjass is the two-player ancestor of the Jass family, on a 32-card pack
	// of which only 18 cards are dealt. The trump jack (20) and nine (14) outrank
	// the ace, sequences are contested so that only the better holder scores, and
	// a maker who fails to score MORE than the opponent goes bete and hands over
	// the whole hand.
	{Name: "klaberjass", Category: CategoryExtra3},
	// Kaiser is the Saskatchewan partnership bidding game on a 34-card pack: the
	// usual A-K-Q-J-10-9-8-7 in each suit PLUS the five of hearts (+5) and the
	// three of spades (-3), which is why 4x8 cards leave a two-card kitty. The
	// declarer takes the kitty and discards two, but may never discard either
	// scoring card.
	{Name: "kaiser", Category: CategoryExtra3},
	// Boston is the 18th-century Whist derivative whose auction ladder INTERLEAVES
	// the misere bids with the trick bids -- Little Misere ranks below seven
	// tricks, Grand Misere below nine -- and adds Piccolissimo, which wants
	// EXACTLY one trick. Trick bids may call a partner (two against two); the
	// miseres, Piccolissimo and the slams are played alone against three.
	{Name: "boston", Category: CategoryExtra3},
	// Vint is the Russian ancestor of contract bridge, played WITHOUT a dummy.
	// Both sides score below the line for every trick they take, whether or not
	// the contract was made; the trick value depends on the level as well as the
	// denomination; and the bidding order is spades < clubs < diamonds < hearts <
	// no trump -- the reverse of bridge, with spades LOWEST.
	{Name: "vint", Category: CategoryExtra3},
	// bideuchre -- Bid Euchre. A 24-card euchre variant played in partnerships
	// where the whole pack is dealt out (24 / 4 = 6 each), so unlike classic
	// euchre there is NO kitty and no turn-up. Bidding starts at three tricks and
	// each bid must beat the last -- except the DEALER, who may take the contract
	// by equalling it. The declarer then names a trump suit or one of two
	// no-trump forms; at no trump LOW the ranking reverses and the nine is
	// highest. A made contract scores each side its own tricks; a set costs the
	// declaring side its BID (not the tricks it took), while the defenders still
	// score theirs. First side to 32 wins.
	{Name: "bideuchre", Category: CategoryExtra2},
	// sixbidsolo -- Six-Bid Solo. A 3-player American descendant of skat on a
	// 36-card pack where the ten outranks the king. ELEVEN cards each plus a
	// THREE-CARD WIDOW (11 x 3 + 3 = 36), and the widow is credited to the
	// declarer at the end -- except at either misere. Six bids ascend and swap
	// the target and the payment together: a plain bid must EXCEED 60 (so 61)
	// and settles on the difference from 60, while a misere asks for zero card
	// POINTS rather than zero tricks. A call solo also names a card whose
	// holder must exchange it.
	{Name: "sixbidsolo", Category: CategoryExtra2},
	// karnoffel -- Karnoffel, the oldest card game known by name (1426).
	// Four players in two partnerships on a 48-card pack with the ACES
	// removed. FIVE cards each, dealt with the first card face up in front
	// of each player; THE LOWEST of those four face-up cards picks the
	// chosen suit. Inside it the JACK is the Karnoffel and beats everything,
	// the SEVEN is the devil and beats all but the Karnoffel ONLY WHEN LED
	// (losing to every card otherwise, and barred from the opening lead),
	// then 6 (Pope) and 2 (Kaiser); the 3, 4 and 5 are partial trumps that
	// still lose to kings, to kings and queens, and to every face card.
	// Following suit is not required. Three of five tricks takes the hand.
	{Name: "karnoffel", Category: CategoryClassic},
	// literature -- Literature, a deduction fishing game for six players in
	// two teams of three, SEATED ALTERNATELY, on a 48-card pack with the
	// EIGHTS removed. Eight half-suits of six: low (2-7) and high (9-A) per
	// suit. You may ask AN OPPONENT ONLY, for a half-suit you already hold,
	// and only for a card you do NOT hold. Claiming has THREE outcomes: all
	// six with your team placed right wins it; all six with your team but
	// MISPLACED CANCELS it, so it goes to nobody; an opponent holding one
	// gives it to them. Winning takes FIVE half-suits -- a majority of eight
	// -- and because cancellations belong to neither side the totals need not
	// add to eight.
	{Name: "literature", Category: CategorySolo},
	// guandan -- Guandan, a two-pack climbing game for four players in two
	// partnerships sitting OPPOSITE, 27 cards each from 108 (52x2 + 4 jokers).
	// Each hand is played at a LEVEL: cards of that rank sit ABOVE THE ACE and
	// below the black joker, and the two HEARTS among them are WILD. Going out
	// first and second climbs FOUR levels, first and third two, first and
	// fourth one -- there is no climb of three. Between hands the losers pay
	// TRIBUTE (highest card, wilds excluded) and receive one back, unless a
	// payer holds both red jokers, which cancels tribute outright. Climbing
	// past the ace wins the game.
	{Name: "guandan", Category: CategoryExtra2},
	// shengji -- Sheng Ji (Tractor), a two-pack point-trick game for four
	// players in two partnerships sitting OPPOSITE. 25 cards each from 108,
	// leaving an EIGHT-CARD KITTY -- 108 divides by four, but dealing 27 each
	// would leave no kitty, and the kitty is what the defenders capture (with
	// a multiplier) by taking the last trick. Trumps are NOT just the trump
	// suit: every card of the hand's LEVEL rank, in all four suits, plus all
	// four jokers. Trump is declared by SHOWING a level card, not by bidding,
	// and only a stronger showing overrides. The DEFENDERS collect the 5s,
	// 10s and kings (200 in the pack); the declarers win by holding them
	// under 80. Climbing stops at the ace, which must then be held to win.
	{Name: "shengji", Category: CategoryClassic},
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
// (casino, classic, solo, extra, extra2, extra3). The returned slice is fresh per
// call so callers
// cannot mutate package state. Adding a new Category value to the iota above
// requires extending this slice — that intentional coupling is the SSoT
// guarantee that consumers (e.g. the CLI --help summary) cannot drift out
// of sync with the registry.
func AllCategories() []Category {
	return []Category{
		CategoryCasino, CategoryClassic, CategorySolo,
		CategoryExtra, CategoryExtra2, CategoryExtra3,
	}
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
