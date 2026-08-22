//go:build test && (!js || !wasm || extra)

package presenter

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// --- fixtures ---

// specSeatSpec describes one seat of a Speculation table.
type specSeatSpec struct {
	name   string
	chips  int
	hidden []*domain.Card
	best   *domain.Card
}

// specBoard is a whole board, built field by field so nothing depends on the
// shuffle. Every presenter test starts from specDefaultBoard() and changes only
// the fields it is about.
type specBoard struct {
	phase       domain.SpeculationPhase
	seats       []specSeatSpec
	trumpSuit   int
	trumpCard   *domain.Card
	pot         int
	turnSeat    int
	bestSeat    int
	offerFrom   int
	offerTo     int
	offerAmount int
	roundNo     int
	winnerSeat  int
	gameEnd     bool
	cfg         domain.SpeculationConfig
	log         []*domain.ActionLogEntry
}

func specCard(design, value int) *domain.Card { return domain.NewCard(design, value, true) }

// specDefaultBoard is a mid-round board: nobody has taken the lead, no auction
// is open, and every "no seat" field carries the -1 sentinel.
func specDefaultBoard() specBoard {
	cfg := domain.NewDefaultSpeculationConfig()
	return specBoard{
		phase: domain.SpeculationPhaseFlip,
		seats: []specSeatSpec{
			{name: "You", chips: 190, hidden: []*domain.Card{specCard(domain.CardDesignDiamond, 9), specCard(domain.CardDesignDiamond, 10)}},
			{name: "CPU1", chips: 175, hidden: []*domain.Card{specCard(domain.CardDesignDiamond, 4)}},
			{name: "CPU2", chips: 220, hidden: nil},
			{name: "CPU3", chips: 60, hidden: []*domain.Card{specCard(domain.CardDesignDiamond, 2)}},
		},
		trumpSuit:   domain.CardDesignSpade,
		trumpCard:   specCard(domain.CardDesignSpade, 3),
		pot:         40,
		turnSeat:    1,
		bestSeat:    -1,
		offerFrom:   -1,
		offerTo:     -1,
		offerAmount: 0,
		roundNo:     2,
		winnerSeat:  -1,
		gameEnd:     false,
		cfg:         cfg,
	}
}

// mock turns a board into a MockSpeculationGame. Every getter is registered
// once, so a test that changes a field changes exactly what the presenter reads.
func (b specBoard) mock() *interfaces.MockSpeculationGame {
	m := new(interfaces.MockSpeculationGame)
	players := make([]*domain.SpeculationPlayer, len(b.seats))
	for i, s := range b.seats {
		p := domain.NewSpeculationPlayer(s.name, s.chips)
		p.SetHidden(s.hidden)
		p.SetBest(s.best)
		players[i] = p
	}
	m.On("GetPhase").Return(b.phase).Maybe()
	m.On("GetPlayers").Return(players).Maybe()
	m.On("GetConfig").Return(b.cfg).Maybe()
	m.On("GetTrumpSuit").Return(b.trumpSuit).Maybe()
	m.On("GetTrumpCard").Return(b.trumpCard).Maybe()
	m.On("GetPot").Return(b.pot).Maybe()
	m.On("GetTurnSeat").Return(b.turnSeat).Maybe()
	m.On("GetBestSeat").Return(b.bestSeat).Maybe()
	m.On("GetOfferFrom").Return(b.offerFrom).Maybe()
	m.On("GetOfferTo").Return(b.offerTo).Maybe()
	m.On("GetOfferAmount").Return(b.offerAmount).Maybe()
	m.On("GetRoundNo").Return(b.roundNo).Maybe()
	m.On("GetWinnerSeat").Return(b.winnerSeat).Maybe()
	m.On("GetGameEndFlag").Return(b.gameEnd).Maybe()
	m.On("GetActionLog").Return(b.log).Maybe()
	return m
}

// specLineContaining returns the single output line that mentions needle.
func specLineContaining(t *testing.T, out, needle string) string {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, needle) {
			return line
		}
	}
	t.Fatalf("no line containing %q in:\n%s", needle, out)
	return ""
}

// --- CUI: board ---

func TestSpeculationCuiPresenter_Output_HeaderLines(t *testing.T) {
	out := new(SpeculationCuiPresenter).Output(specDefaultBoard().mock(), nil)

	// roundNo is 0-based internally and 1-based on screen.
	assert.Contains(t, out, "ラウンド: 3 / 5")
	assert.Contains(t, out, "ポット: 40")
	assert.Contains(t, out, "切り札: ♠")
	assert.Contains(t, out, "フェーズ: めくり")
}

// TestSpeculationCuiPresenter_Output_RoundNumberOnTheResultScreen pins that the
// result screen names the round that just finished. roundNo advances at the
// moment the round resolves, so a plain +1 would head the round-1 result with
// "round 2" -- a screen about a round nobody has played yet.
func TestSpeculationCuiPresenter_Output_RoundNumberOnTheResultScreen(t *testing.T) {
	cp := new(SpeculationCuiPresenter)

	mid := specDefaultBoard() // roundNo 2, still being played
	assert.Contains(t, cp.Output(mid.mock(), nil), "ラウンド: 3 / 5")

	// The same roundNo, but the round has now resolved: roundNo already counts
	// it, so the header must not advance again.
	done := specDefaultBoard()
	done.roundNo = 3
	done.phase = domain.SpeculationPhaseResult
	done.winnerSeat = 0
	assert.Contains(t, cp.Output(done.mock(), nil), "ラウンド: 3 / 5")

	end := specDefaultBoard()
	end.roundNo = 5
	end.phase = domain.SpeculationPhaseGameEnd
	end.gameEnd = true
	end.winnerSeat = 1
	out := cp.Output(end.mock(), nil)
	assert.Contains(t, out, "ラウンド: 5 / 5")
	// **Never past the total.** "Round 6 / 5" on the final screen reads as a
	// counting bug to the player.
	assert.NotContains(t, out, "ラウンド: 6 / 5")
}

func TestSpeculationCuiPresenter_Output_TrumpSuitSymbols(t *testing.T) {
	cases := []struct {
		suit int
		want string
	}{
		{domain.CardDesignSpade, "♠"},
		{domain.CardDesignClover, "♣"},
		{domain.CardDesignHeart, "♥"},
		{domain.CardDesignDiamond, "♦"},
		// **Undecided is not spade.** 0 is the joker design, and -1 is what
		// Reset leaves before the trump card is turned; both must read as "-".
		{-1, "-"},
		{domain.CardDesignJoker, "-"},
	}
	for _, tc := range cases {
		b := specDefaultBoard()
		b.trumpSuit = tc.suit
		out := new(SpeculationCuiPresenter).Output(b.mock(), nil)
		assert.Contains(t, specLineContaining(t, out, "切り札:"), "切り札: "+tc.want)
	}
}

func TestSpeculationCuiPresenter_Output_OneLinePerSeat(t *testing.T) {
	out := new(SpeculationCuiPresenter).Output(specDefaultBoard().mock(), nil)

	assert.Contains(t, out, "You: チップ 190 / 伏せ札 2枚")
	assert.Contains(t, out, "CPU1: チップ 175 / 伏せ札 1枚")
	assert.Contains(t, out, "CPU2: チップ 220 / 伏せ札 0枚")
	assert.Contains(t, out, "CPU3: チップ 60 / 伏せ札 1枚")
}

// TestSpeculationCuiPresenter_Output_HidesFaceDownCards is the game's core
// secret: the face-down cards must never be printed. Seeing them would tell the
// player exactly what to bid, and the auction would stop being a gamble.
//
// Every seat's face-down cards are diamonds here and nothing else on the board
// is, so a leak shows up as a diamond in the text.
func TestSpeculationCuiPresenter_Output_HidesFaceDownCards(t *testing.T) {
	b := specDefaultBoard()
	b.seats[2].best = specCard(domain.CardDesignClover, 12)
	b.bestSeat = 2
	out := new(SpeculationCuiPresenter).Output(b.mock(), nil)

	for _, s := range b.seats {
		for _, c := range s.hidden {
			assert.NotContains(t, out, cuiCardStr(c),
				"a face-down card leaked into the board")
		}
	}
	// Named explicitly too, so the test still says what it protects even if
	// cuiCardStr changes.
	assert.NotContains(t, out, "DIAMOND")
	assert.NotContains(t, out, "♦9")
	// The counts, by contrast, must be there -- they are the public part.
	assert.Contains(t, out, "伏せ札 2枚")
	assert.Contains(t, out, "伏せ札 1枚")
}

func TestSpeculationCuiPresenter_Output_HoldsMarkerOnlyOnTheLeader(t *testing.T) {
	b := specDefaultBoard()
	b.seats[2].best = specCard(domain.CardDesignClover, 12)
	b.bestSeat = 2
	out := new(SpeculationCuiPresenter).Output(b.mock(), nil)

	assert.Contains(t, specLineContaining(t, out, "CPU2:"), "【最高札 CLOVER 12】")
	assert.NotContains(t, specLineContaining(t, out, "You:"), "【最高札")
	assert.NotContains(t, specLineContaining(t, out, "CPU1:"), "【最高札")
	assert.NotContains(t, specLineContaining(t, out, "CPU3:"), "【最高札")
	assert.Equal(t, 1, strings.Count(out, "【最高札"))
}

func TestSpeculationCuiPresenter_Output_NoHoldsMarkerBeforeAnyoneLeads(t *testing.T) {
	out := new(SpeculationCuiPresenter).Output(specDefaultBoard().mock(), nil)
	assert.NotContains(t, out, "【最高札")
}

// --- CUI: the offer line ---

// specAuctionBoard is a board with a live auction. `humanSells` picks the
// direction: seat 0 owns the best trump and a CPU is buying, or the other way
// round.
func specAuctionBoard(humanSells bool) specBoard {
	b := specDefaultBoard()
	b.phase = domain.SpeculationPhaseAuction
	b.offerAmount = 34
	if humanSells {
		b.seats[0].best = specCard(domain.CardDesignSpade, 11)
		b.bestSeat = 0
		b.offerFrom, b.offerTo = 3, 0
		return b
	}
	b.seats[2].best = specCard(domain.CardDesignSpade, 11)
	b.bestSeat = 2
	b.offerFrom, b.offerTo = 0, 2
	return b
}

func TestSpeculationCuiPresenter_Output_OfferLineOnlyDuringTheAuction(t *testing.T) {
	auction := new(SpeculationCuiPresenter).Output(specAuctionBoard(true).mock(), nil)
	assert.Contains(t, auction, "を提示しています")

	// The same offer fields, but the phase has moved on: no offer line.
	b := specAuctionBoard(true)
	b.phase = domain.SpeculationPhaseFlip
	flip := new(SpeculationCuiPresenter).Output(b.mock(), nil)
	assert.NotContains(t, flip, "を提示しています")
	assert.NotContains(t, flip, "譲るそうです")
}

// TestSpeculationCuiPresenter_Output_OfferLineWordedByDirection pins that the
// two sides of the deal read differently. **Selling and buying are opposite
// moves**; one wording for both would tell the player to press `a` without
// saying whether that spends chips or earns them.
func TestSpeculationCuiPresenter_Output_OfferLineWordedByDirection(t *testing.T) {
	selling := new(SpeculationCuiPresenter).Output(specAuctionBoard(true).mock(), nil)
	buying := new(SpeculationCuiPresenter).Output(specAuctionBoard(false).mock(), nil)

	assert.Contains(t, selling, "CPU3 があなたの札に 34 を提示しています。")
	assert.Contains(t, selling, "a で売る")
	assert.NotContains(t, selling, "譲るそうです")

	assert.Contains(t, buying, "CPU2 は最高札を 34 で譲るそうです。")
	assert.Contains(t, buying, "a で買う")
	assert.Contains(t, buying, "bid <額>")
	assert.NotContains(t, buying, "を提示しています")

	assert.NotEqual(t, specLineContaining(t, selling, "34"), specLineContaining(t, buying, "34"))
}

func TestSpeculationCuiPresenter_Output_OfferLineSkippedWhenSeatsAreOutOfRange(t *testing.T) {
	b := specAuctionBoard(true)
	b.offerFrom, b.offerTo = 9, 0
	out := new(SpeculationCuiPresenter).Output(b.mock(), nil)
	assert.NotContains(t, out, "を提示しています")

	b = specAuctionBoard(true)
	b.offerFrom, b.offerTo = -1, -1
	out = new(SpeculationCuiPresenter).Output(b.mock(), nil)
	assert.NotContains(t, out, "を提示しています")
}

// --- CUI: the result line ---

func TestSpeculationCuiPresenter_Output_ResultYouWin(t *testing.T) {
	b := specDefaultBoard()
	b.phase = domain.SpeculationPhaseResult
	b.winnerSeat = 0
	out := new(SpeculationCuiPresenter).Output(b.mock(), nil)

	assert.Contains(t, out, "あなたがポットを取りました！")
	assert.NotContains(t, out, "がポットを取りました。")
	assert.NotContains(t, out, "切り札が1枚も出ませんでした")
}

func TestSpeculationCuiPresenter_Output_ResultCpuWin(t *testing.T) {
	b := specDefaultBoard()
	b.phase = domain.SpeculationPhaseResult
	b.winnerSeat = 2
	out := new(SpeculationCuiPresenter).Output(b.mock(), nil)

	assert.Contains(t, out, "CPU2 がポットを取りました。")
	assert.NotContains(t, out, "あなたがポットを取りました")
}

// TestSpeculationCuiPresenter_Output_ResultVoidRound pins the third outcome.
// **A void round is not a loss.** If no trump was turned up, nobody takes the
// pot and the stakes come back, so it must not be announced as a CPU win.
func TestSpeculationCuiPresenter_Output_ResultVoidRound(t *testing.T) {
	b := specDefaultBoard()
	b.phase = domain.SpeculationPhaseResult
	b.winnerSeat = -1
	out := new(SpeculationCuiPresenter).Output(b.mock(), nil)

	assert.Contains(t, out, "切り札が1枚も出ませんでした。参加料は返却されます。")
	assert.NotContains(t, out, "ポットを取りました")
}

func TestSpeculationCuiPresenter_Output_NoResultLineMidRound(t *testing.T) {
	out := new(SpeculationCuiPresenter).Output(specDefaultBoard().mock(), nil)
	assert.NotContains(t, out, "ポットを取りました")
	assert.NotContains(t, out, "切り札が1枚も出ませんでした")
	assert.NotContains(t, out, "最終チップ")
}

func TestSpeculationCuiPresenter_Output_FinalChipsOnGameEnd(t *testing.T) {
	b := specDefaultBoard()
	b.phase = domain.SpeculationPhaseGameEnd
	b.gameEnd = true
	b.winnerSeat = 0
	out := new(SpeculationCuiPresenter).Output(b.mock(), nil)

	assert.Contains(t, out, "最終チップ: 190")
	assert.Contains(t, out, "あなたがポットを取りました！")
}

func TestSpeculationCuiPresenter_Output_Error(t *testing.T) {
	out := new(SpeculationCuiPresenter).Output(specDefaultBoard().mock(), errors.New("boom"))
	assert.Contains(t, out, "boom")
	assert.Contains(t, out, i18n.ErrorLinePrefix)
}

// --- CUI: phaseStr ---

func TestSpeculationCuiPresenter_PhaseStr(t *testing.T) {
	cp := new(SpeculationCuiPresenter)
	cases := []struct {
		phase domain.SpeculationPhase
		want  string
	}{
		{domain.SpeculationPhaseFlip, "めくり"},
		{domain.SpeculationPhaseAuction, "競り"},
		{domain.SpeculationPhaseResult, "ラウンド結果"},
		{domain.SpeculationPhaseGameEnd, "終了"},
		{domain.SpeculationPhase(99), "不明"},
		{domain.SpeculationPhase(-1), "不明"},
	}
	seen := make(map[string]bool, len(cases))
	for _, tc := range cases {
		got := cp.phaseStr(tc.phase)
		assert.Equal(t, tc.want, got)
		seen[got] = true
	}
	// Five distinct labels: a switch that collapsed two phases onto one string
	// would still satisfy every Equal above only if the table were wrong.
	assert.Len(t, seen, 5)
}

// --- CUI: hints ---

func TestSpeculationCuiPresenter_HintOutput_Flip(t *testing.T) {
	out := new(SpeculationCuiPresenter).HintOutput(specDefaultBoard().mock())
	assert.Contains(t, out, "めくって様子を見ましょう")
}

// TestSpeculationCuiPresenter_HintOutput_DiffersByPhase pins that the advice is
// actually about the situation: the flip hint and the auction hint must not be
// the same sentence.
func TestSpeculationCuiPresenter_HintOutput_DiffersByPhase(t *testing.T) {
	cp := new(SpeculationCuiPresenter)
	flip := cp.HintOutput(specDefaultBoard().mock())
	auction := cp.HintOutput(specAuctionBoard(true).mock())
	other := func() string {
		b := specDefaultBoard()
		b.phase = domain.SpeculationPhaseResult
		return cp.HintOutput(b.mock())
	}()

	assert.NotEqual(t, flip, auction)
	assert.NotEqual(t, flip, other)
	assert.Contains(t, other, "いま助言できることはありません")
}

// TestSpeculationCuiPresenter_HintOutput_SellingVsBuying pins the four auction
// hints. The rule is about how many cards are still face down, not about the
// card's rank: with plenty left the lead will likely be beaten (sell / pass),
// with few left it will probably hold (keep / buy).
func TestSpeculationCuiPresenter_HintOutput_SellingVsBuying(t *testing.T) {
	cp := new(SpeculationCuiPresenter)

	// 4 seats. "Plenty left" means more than 4 face-down cards in total.
	many := []*domain.Card{specCard(domain.CardDesignDiamond, 2), specCard(domain.CardDesignDiamond, 3)}
	few := []*domain.Card(nil)

	withHidden := func(b specBoard, per []*domain.Card) specBoard {
		for i := range b.seats {
			b.seats[i].hidden = per
		}
		return b
	}

	sell := cp.HintOutput(withHidden(specAuctionBoard(true), many).mock())
	hold := cp.HintOutput(withHidden(specAuctionBoard(true), few).mock())
	pass := cp.HintOutput(withHidden(specAuctionBoard(false), many).mock())
	buy := cp.HintOutput(withHidden(specAuctionBoard(false), few).mock())

	assert.Contains(t, sell, "売り時です")
	assert.Contains(t, hold, "持ち続けましょう")
	assert.Contains(t, pass, "見送りましょう")
	assert.Contains(t, buy, "買い時です")

	// The seller's advice must never tell a seller to buy, and vice versa.
	assert.NotContains(t, sell, "買い時")
	assert.NotContains(t, buy, "売り時")
	assert.NotEqual(t, sell, pass)
	assert.NotEqual(t, hold, buy)
}

// TestSpeculationI18nKeysResolve is the negative control for every assertion
// above: i18n.T returns the key itself when a translation is missing, so a
// Contains(out, i18n.T(key)) check would pass in exactly the broken case.
func TestSpeculationI18nKeysResolve(t *testing.T) {
	keys := []string{
		"speculation.roundLine", "speculation.potLine", "speculation.trumpLine",
		"speculation.phaseLine", "speculation.seatLine", "speculation.holdsLine",
		"speculation.offerToYou", "speculation.offerFromYou",
		"speculation.phaseFlip", "speculation.phaseAuction", "speculation.phaseResult",
		"speculation.phaseGameEnd", "speculation.phaseUnknown",
		"speculation.youWin", "speculation.seatWins", "speculation.voidRound",
		"speculation.finalChips",
		"speculation.hintFlip", "speculation.hintSell", "speculation.hintHold",
		"speculation.hintBuy", "speculation.hintPass", "speculation.hintNone",
	}
	for _, k := range keys {
		assert.NotEqual(t, k, i18n.T(k), "missing translation for %s", k)
	}
}

func TestSpeculationCuiPresenter_ActionLogOutput(t *testing.T) {
	b := specDefaultBoard()
	b.gameEnd = true
	b.log = []*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "flip", Detail: "1-5"},
	}
	out := new(SpeculationCuiPresenter).ActionLogOutput(b.mock())
	assert.Contains(t, out, "flip")
}

// --- Web ---

func specWebOutput(t *testing.T, b specBoard, lastErr error) (string, *controller.SpeculationWebOutput) {
	t.Helper()
	raw := new(SpeculationWebPresenter).Output(b.mock(), lastErr)
	var out controller.SpeculationWebOutput
	require.NoError(t, json.Unmarshal([]byte(raw), &out), "web output is not valid JSON: %s", raw)
	return raw, &out
}

func TestSpeculationWebPresenter_Output_EveryFieldIsPopulated(t *testing.T) {
	b := specDefaultBoard()
	b.phase = domain.SpeculationPhaseAuction
	b.seats[2].best = specCard(domain.CardDesignClover, 12)
	b.bestSeat = 2
	b.offerFrom, b.offerTo, b.offerAmount = 0, 2, 34
	b.winnerSeat = 3
	b.gameEnd = true

	raw, out := specWebOutput(t, b, nil)

	assert.Equal(t, int(domain.SpeculationPhaseAuction), out.Phase)
	assert.Equal(t, domain.CardDesignSpade, out.TrumpSuit)
	require.NotNil(t, out.TrumpCard)
	assert.Equal(t, "SPADE", out.TrumpCard.Design)
	assert.Equal(t, 3, out.TrumpCard.Value)
	assert.Equal(t, 40, out.Pot)
	assert.Equal(t, 1, out.TurnSeat)
	assert.Equal(t, 2, out.BestSeat)
	assert.Equal(t, 0, out.OfferFrom)
	assert.Equal(t, 2, out.OfferTo)
	assert.Equal(t, 34, out.OfferAmount)
	assert.Equal(t, 2, out.RoundNo)
	assert.Equal(t, 3, out.WinnerSeat)
	assert.True(t, out.GameEndFlag)

	require.NotNil(t, out.Config)
	assert.Equal(t, domain.SpeculationDefaultPlayers, out.Config.Players)
	assert.Equal(t, domain.SpeculationDefaultChips, out.Config.InitialChips)
	assert.Equal(t, domain.SpeculationDefaultStake, out.Config.Stake)
	assert.Equal(t, domain.SpeculationDefaultRounds, out.Config.Rounds)

	require.Len(t, out.Seats, 4)
	assert.Equal(t, []string{"You", "CPU1", "CPU2", "CPU3"},
		[]string{out.Seats[0].Name, out.Seats[1].Name, out.Seats[2].Name, out.Seats[3].Name})
	assert.Equal(t, []int{190, 175, 220, 60},
		[]int{out.Seats[0].Chips, out.Seats[1].Chips, out.Seats[2].Chips, out.Seats[3].Chips})
	assert.Equal(t, []int{2, 1, 0, 1},
		[]int{out.Seats[0].HiddenCount, out.Seats[1].HiddenCount, out.Seats[2].HiddenCount, out.Seats[3].HiddenCount})

	// The best trump is public, and only on the seat that holds it.
	require.NotNil(t, out.Seats[2].Best)
	assert.Equal(t, "CLOVER", out.Seats[2].Best.Design)
	assert.Equal(t, 12, out.Seats[2].Best.Value)
	assert.Nil(t, out.Seats[0].Best)
	assert.Nil(t, out.Seats[1].Best)
	assert.Nil(t, out.Seats[3].Best)

	assert.Empty(t, out.Message)
	assert.Contains(t, raw, `"hiddenCount":2`)
}

// TestSpeculationWebPresenter_Output_HidesFaceDownCards checks the raw JSON,
// not the decoded struct: a leak would arrive in a field the struct does not
// even declare, and decoding would throw it away before the assertion ran.
//
// Every face-down card on this board is a diamond and nothing else is, so a
// single "DIAMOND" anywhere in the payload is the leak.
func TestSpeculationWebPresenter_Output_HidesFaceDownCards(t *testing.T) {
	b := specDefaultBoard()
	b.seats[2].best = specCard(domain.CardDesignClover, 12)
	b.bestSeat = 2
	raw, out := specWebOutput(t, b, nil)

	assert.NotContains(t, raw, "DIAMOND", "a face-down card's suit reached the client")
	for _, v := range []string{`"value":9`, `"value":10`, `"value":4`, `"value":2`} {
		assert.NotContains(t, raw, v, "a face-down card's value reached the client")
	}
	assert.NotContains(t, raw, `"hidden":`)

	// The counts are the public half and must still be there.
	assert.Contains(t, raw, `"hiddenCount"`)
	assert.Equal(t, []int{2, 1, 0, 1},
		[]int{out.Seats[0].HiddenCount, out.Seats[1].HiddenCount, out.Seats[2].HiddenCount, out.Seats[3].HiddenCount})
}

// TestSpeculationWebPresenter_Output_NoSeatSentinels pins that "nobody" is -1.
// **0 is a valid seat** -- it is the human -- so a 0 here would draw the player
// as holding the best trump on a board where nobody does.
func TestSpeculationWebPresenter_Output_NoSeatSentinels(t *testing.T) {
	raw, out := specWebOutput(t, specDefaultBoard(), nil)

	assert.Equal(t, -1, out.BestSeat)
	assert.Equal(t, -1, out.OfferFrom)
	assert.Equal(t, -1, out.OfferTo)
	assert.Equal(t, -1, out.WinnerSeat)
	assert.NotEqual(t, 0, out.BestSeat)
	assert.NotEqual(t, 0, out.WinnerSeat)
	assert.Contains(t, raw, `"bestSeat":-1`)
	assert.Contains(t, raw, `"winnerSeat":-1`)
}

func TestSpeculationWebPresenter_Output_NoTrumpCardYet(t *testing.T) {
	b := specDefaultBoard()
	b.trumpCard = nil
	b.trumpSuit = -1
	_, out := specWebOutput(t, b, nil)

	assert.Nil(t, out.TrumpCard)
	assert.Equal(t, -1, out.TrumpSuit)
}

func TestSpeculationWebPresenter_Output_Error(t *testing.T) {
	_, out := specWebOutput(t, specDefaultBoard(), errors.New("not allowed in this phase"))
	assert.Equal(t, "not allowed in this phase", out.Message)
}

// TestSpeculationWebPresenter_HintOutput_ReturnsTheBoard pins the documented
// contract: the web GUI builds its own advice, so the hint endpoint answers
// with the board rather than a sentence.
func TestSpeculationWebPresenter_HintOutput_ReturnsTheBoard(t *testing.T) {
	b := specDefaultBoard()
	cp := new(SpeculationWebPresenter)
	assert.JSONEq(t, cp.Output(b.mock(), nil), cp.HintOutput(b.mock()))
}

func TestSpeculationWebPresenter_ActionLogOutput(t *testing.T) {
	b := specDefaultBoard()
	b.gameEnd = true
	b.log = []*domain.ActionLogEntry{
		{TurnNumber: 1, PlayerIdx: 0, ActionType: "flip", Detail: "1-5"},
	}
	out := new(SpeculationWebPresenter).ActionLogOutput(b.mock())
	assert.Contains(t, out, "flip")
	var decoded any
	assert.NoError(t, json.Unmarshal([]byte(out), &decoded))
}
