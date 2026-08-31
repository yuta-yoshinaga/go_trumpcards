//go:build test

package domain

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ckCard creates a Card with the given design and value.
func ckCard(d, v int) *Card { return NewCard(d, v, false) }

// ckNewGame builds a Cuckoo game with the given config and a fixed rng seed.
func ckNewGame(cfg CuckooConfig) *Cuckoo {
	players := []*CuckooPlayer{
		NewCuckooPlayer(true),
		NewCuckooPlayer(false),
		NewCuckooPlayer(false),
		NewCuckooPlayer(false),
	}
	g := NewCuckoo(NewTrumpCards(0), players, cfg)
	g.SetRand(rand.New(rand.NewSource(1)))
	return g
}

// ckSetHand assigns each player one card and resets lives, then puts the game
// into the turn phase with player 0 to act (dealer is the last seat, 3).
func ckSetHand(g *Cuckoo, vals [CuckooPlayerCnt]int, lives int) {
	for i, p := range g.players {
		p.SetCard(ckCard(CardDesignSpade, vals[i]))
		p.SetLives(lives)
	}
	g.dealerIdx = CuckooPlayerCnt - 1
	g.actedCount = 0
	g.roundLosers = make([]int, 0)
	for i := range g.revealedKings {
		g.revealedKings[i] = false
	}
	g.clearPending()
	g.currentPlayerIdx = 0
	g.phase = CuckooPhaseTurn
	g.gameEndFlag = false
}

func TestCuckoo_ResetDealsOneCardEach(t *testing.T) {
	g := ckNewGame(DefaultCuckooConfig())
	g.Reset()
	for i := 0; i < g.GetPlayerCnt(); i++ {
		p := g.GetPlayer(i)
		if p.GetCardsSize() != 1 {
			t.Errorf("player %d has %d cards, want 1", i, p.GetCardsSize())
		}
		if p.GetLives() != CuckooStartLives {
			t.Errorf("player %d lives = %d, want %d", i, p.GetLives(), CuckooStartLives)
		}
	}
	if g.GetPhase() != CuckooPhaseTurn {
		t.Errorf("phase = %d, want Turn", g.GetPhase())
	}
}

func TestCuckoo_Keep(t *testing.T) {
	g := ckNewGame(DefaultCuckooConfig())
	ckSetHand(g, [CuckooPlayerCnt]int{5, 6, 7, 8}, 3)
	if err := g.PlayerKeep(); err != nil {
		t.Fatalf("PlayerKeep: %v", err)
	}
	if g.GetCurrentPlayerIdx() != 1 {
		t.Errorf("current = %d, want 1", g.GetCurrentPlayerIdx())
	}
	if g.GetPlayer(0).CardValue() != 5 {
		t.Errorf("player 0 card changed after keep")
	}
}

func TestCuckoo_SwapExchangesCards(t *testing.T) {
	g := ckNewGame(DefaultCuckooConfig())
	ckSetHand(g, [CuckooPlayerCnt]int{2, 9, 7, 8}, 3) // player 1 has no king
	if err := g.PlayerSwap(); err != nil {
		t.Fatalf("PlayerSwap: %v", err)
	}
	if g.GetPlayer(0).CardValue() != 9 || g.GetPlayer(1).CardValue() != 2 {
		t.Errorf("swap failed: p0=%d p1=%d, want 9/2", g.GetPlayer(0).CardValue(), g.GetPlayer(1).CardValue())
	}
}

func TestCuckoo_SwapBlockedByKingRefuse_HumanNeighbour(t *testing.T) {
	g := ckNewGame(DefaultCuckooConfig())
	// Make seat 1 human so the Refuse phase pauses for input.
	g.players[1] = NewCuckooPlayer(true)
	g.players[0] = NewCuckooPlayer(false)
	ckSetHand(g, [CuckooPlayerCnt]int{4, CuckooKingValue, 7, 8}, 3)
	g.currentPlayerIdx = 0
	// Player 0 (CPU) swaps toward player 1 (human, King) -> Refuse phase.
	g.attemptSwap(0)
	if g.GetPhase() != CuckooPhaseRefuse {
		t.Fatalf("phase = %d, want Refuse", g.GetPhase())
	}
	if g.GetPendingSwapTo() != 1 {
		t.Fatalf("pendingSwapTo = %d, want 1", g.GetPendingSwapTo())
	}
	if err := g.PlayerRefuse(); err != nil {
		t.Fatalf("PlayerRefuse: %v", err)
	}
	if g.GetPlayer(0).CardValue() != 4 {
		t.Errorf("requester card changed after refuse: %d", g.GetPlayer(0).CardValue())
	}
	if !g.IsKingRevealed(1) {
		t.Errorf("king should be revealed after refuse")
	}
}

func TestCuckoo_AcceptSwap(t *testing.T) {
	g := ckNewGame(DefaultCuckooConfig())
	g.players[1] = NewCuckooPlayer(true)
	g.players[0] = NewCuckooPlayer(false)
	ckSetHand(g, [CuckooPlayerCnt]int{4, CuckooKingValue, 7, 8}, 3)
	g.currentPlayerIdx = 0
	g.attemptSwap(0)
	if err := g.PlayerAcceptSwap(); err != nil {
		t.Fatalf("PlayerAcceptSwap: %v", err)
	}
	if g.GetPlayer(0).CardValue() != CuckooKingValue || g.GetPlayer(1).CardValue() != 4 {
		t.Errorf("accept swap failed: p0=%d p1=%d", g.GetPlayer(0).CardValue(), g.GetPlayer(1).CardValue())
	}
}

func TestCuckoo_CpuAutoRefusesWithKing(t *testing.T) {
	g := ckNewGame(DefaultCuckooConfig())
	ckSetHand(g, [CuckooPlayerCnt]int{4, CuckooKingValue, 7, 8}, 3)
	// Player 0 (human seat) swaps toward CPU player 1 holding a King.
	g.attemptSwap(0)
	// CPU auto-refuses, so no Refuse phase remains and the turn has advanced.
	if g.GetPhase() == CuckooPhaseRefuse {
		t.Errorf("CPU should auto-refuse, not pause in Refuse phase")
	}
	if g.GetPlayer(0).CardValue() != 4 {
		t.Errorf("requester card changed after CPU refuse: %d", g.GetPlayer(0).CardValue())
	}
	if !g.IsKingRevealed(1) {
		t.Errorf("CPU king should be revealed")
	}
}

func TestCuckoo_DealerSwapsWithStock(t *testing.T) {
	g := ckNewGame(DefaultCuckooConfig())
	ckSetHand(g, [CuckooPlayerCnt]int{5, 6, 7, 2}, 3)
	g.dealerIdx = 3
	g.currentPlayerIdx = 3
	g.actedCount = 3 // dealer is last to act
	g.stock = []*Card{ckCard(CardDesignHeart, 10)}
	g.attemptSwap(3)
	if g.GetPlayer(3).CardValue() != 10 {
		t.Errorf("dealer card = %d, want 10 from stock", g.GetPlayer(3).CardValue())
	}
}

func TestCuckoo_RoundEndLowestLosesLife(t *testing.T) {
	g := ckNewGame(DefaultCuckooConfig())
	ckSetHand(g, [CuckooPlayerCnt]int{3, 9, 10, 11}, 3)
	g.endRound()
	if g.GetPlayer(0).GetLives() != 2 {
		t.Errorf("lowest player lives = %d, want 2", g.GetPlayer(0).GetLives())
	}
	if g.GetPlayer(1).GetLives() != 3 {
		t.Errorf("non-lowest player lost a life")
	}
	if g.GetRoundLowest() != 3 {
		t.Errorf("roundLowest = %d, want 3", g.GetRoundLowest())
	}
	if len(g.GetRoundLosers()) != 1 || g.GetRoundLosers()[0] != 0 {
		t.Errorf("round losers = %v, want [0]", g.GetRoundLosers())
	}
}

func TestCuckoo_RoundEndTieBothLoseLife(t *testing.T) {
	g := ckNewGame(DefaultCuckooConfig())
	ckSetHand(g, [CuckooPlayerCnt]int{4, 4, 10, 11}, 3)
	g.endRound()
	if g.GetPlayer(0).GetLives() != 2 || g.GetPlayer(1).GetLives() != 2 {
		t.Errorf("tied players lives = %d/%d, want 2/2", g.GetPlayer(0).GetLives(), g.GetPlayer(1).GetLives())
	}
	if len(g.GetRoundLosers()) != 2 {
		t.Errorf("round losers = %v, want 2 losers", g.GetRoundLosers())
	}
}

func TestCuckoo_EliminationAndWin(t *testing.T) {
	g := ckNewGame(DefaultCuckooConfig())
	ckSetHand(g, [CuckooPlayerCnt]int{2, 9, 10, 11}, 1) // player 0 at 1 life, lowest
	g.endRound()
	if !g.GetPlayer(0).IsEliminated() {
		t.Errorf("player 0 should be eliminated")
	}
	if !g.GetGameEndFlag() {
		t.Errorf("game should end when human is eliminated")
	}
	if g.GetPhase() != CuckooPhaseGameEnd {
		t.Errorf("phase = %d, want GameEnd", g.GetPhase())
	}
}

func TestCuckoo_WrongPhaseAndTurnGuards(t *testing.T) {
	g := ckNewGame(DefaultCuckooConfig())
	ckSetHand(g, [CuckooPlayerCnt]int{5, 6, 7, 8}, 3)
	g.gameEndFlag = true
	if err := g.PlayerKeep(); err != ErrGameEnded {
		t.Errorf("PlayerKeep after end = %v, want ErrGameEnded", err)
	}
	g.gameEndFlag = false
	g.phase = CuckooPhaseRoundEnd
	if err := g.PlayerSwap(); err != ErrWrongPhase {
		t.Errorf("PlayerSwap wrong phase = %v, want ErrWrongPhase", err)
	}
	g.phase = CuckooPhaseTurn
	g.currentPlayerIdx = 1 // CPU
	if err := g.PlayerKeep(); err != ErrNotHumanTurn {
		t.Errorf("PlayerKeep not human turn = %v, want ErrNotHumanTurn", err)
	}
}

func TestCuckoo_JSONRoundTrip(t *testing.T) {
	g := NewDefaultCuckoo()
	g.SetRand(rand.New(rand.NewSource(2)))
	g.Reset()
	data, err := g.MarshalJSON()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var g2 Cuckoo
	if err := g2.UnmarshalJSON(data); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if g2.GetPlayerCnt() != CuckooPlayerCnt {
		t.Errorf("player count after round trip = %d", g2.GetPlayerCnt())
	}
}

func TestCuckoo_UnmarshalRejectsBadInput(t *testing.T) {
	cases := []string{
		`{"pl":[]}`,                                 // wrong player count
		`{"pl":[null,null,null,null]}`,              // nil players
		`{"cf":{"cd":9,"il":3},"pl":[{},{},{},{}]}`, // invalid config
	}
	for _, c := range cases {
		var g Cuckoo
		if err := g.UnmarshalJSON([]byte(c)); err == nil {
			t.Errorf("UnmarshalJSON(%s) = nil err, want error", c)
		}
	}
}

// runCuckooToEnd drives a full CPU-only game to completion and returns the
// number of rounds played. It fails if the game does not terminate.
func runCuckooToEnd(t *testing.T, cfg CuckooConfig, seed int64) {
	t.Helper()
	players := []*CuckooPlayer{
		NewCuckooPlayer(false),
		NewCuckooPlayer(false),
		NewCuckooPlayer(false),
		NewCuckooPlayer(false),
	}
	g := NewCuckoo(NewTrumpCards(0), players, cfg)
	g.SetRand(rand.New(rand.NewSource(seed)))
	g.Reset()

	const maxSteps = 100000
	for step := 0; step < maxSteps; step++ {
		if g.GetGameEndFlag() {
			break
		}
		switch g.GetPhase() {
		case CuckooPhaseTurn, CuckooPhaseRefuse:
			g.CpuPlay()
		case CuckooPhaseRoundEnd:
			g.NextRound()
		case CuckooPhaseGameEnd:
		}
	}
	if !g.GetGameEndFlag() {
		t.Fatalf("full-CPU game did not terminate (seed %d)", seed)
	}
	if g.GetWinnerIdx() < 0 || g.GetWinnerIdx() >= CuckooPlayerCnt {
		t.Errorf("winner idx out of range: %d", g.GetWinnerIdx())
	}
	// Exactly one player should remain with lives, others eliminated.
	active := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if !g.GetPlayer(i).IsEliminated() {
			active++
		}
	}
	if active != 1 {
		t.Errorf("active players at end = %d, want 1 (seed %d)", active, seed)
	}
}

func TestCuckoo_FullCpuGameTerminatesAllConfigs(t *testing.T) {
	difficulties := []CuckooCpuDifficulty{
		CuckooCpuDifficultyEasy,
		CuckooCpuDifficultyNormal,
		CuckooCpuDifficultyHard,
	}
	for _, d := range difficulties {
		for _, lives := range []int{1, 3, 5} {
			cfg := CuckooConfig{CpuDifficulty: d, InitialLives: lives}
			for seed := int64(1); seed <= 5; seed++ {
				runCuckooToEnd(t, cfg, seed*int64(lives+1))
			}
		}
	}
}

// **席順の隣は答えにならない。**脱落者は手番から飛ばされるので、交換の相手は
// 卓を回って次に**まだ残っている**席になる (#6467)。
func TestCuckoo_GetSwapTargetIdxSkipsEliminatedSeats(t *testing.T) {
	tests := []struct {
		name  string
		lives [CuckooPlayerCnt]int
		from  int
		want  int
	}{
		{"nobody is out yet", [CuckooPlayerCnt]int{3, 3, 3, 3}, 0, 1},
		{"the seat beside you is out", [CuckooPlayerCnt]int{3, 0, 3, 3}, 0, 2},
		{"two in a row are out", [CuckooPlayerCnt]int{3, 0, 0, 3}, 0, 3},
		{"it wraps around the table", [CuckooPlayerCnt]int{3, 3, 3, 3}, 3, 0},
		{"it wraps past an eliminated seat", [CuckooPlayerCnt]int{0, 3, 3, 3}, 3, 1},
		// 他に誰も残っていない = attemptSwap が「保持」として扱う場合。
		{"nobody else is left", [CuckooPlayerCnt]int{3, 0, 0, 0}, 0, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ckNewGame(DefaultCuckooConfig())
			for i, p := range g.players {
				p.SetLives(tt.lives[i])
			}
			assert.Equal(t, tt.want, g.GetSwapTargetIdx(tt.from))
		})
	}

	// 範囲外は -1。描画側が存在しない席を引きに行かないため。
	g := ckNewGame(DefaultCuckooConfig())
	assert.Equal(t, -1, g.GetSwapTargetIdx(-1))
	assert.Equal(t, -1, g.GetSwapTargetIdx(CuckooPlayerCnt))
}

// **アクセサと実際の交換先が食い違ってはいけない。**プロンプトが名前を出す以上、
// その席が本当に交換されることを、規則の再実装ではなく**札の移動**から確かめる。
// `GetPendingSwapTo` は King で拒否されない限り即座に解決して -1 に戻るので、
// そちらを見ても何も分からない。
func TestCuckoo_SwapTargetMatchesWhereTheSwapActuallyGoes(t *testing.T) {
	g := ckNewGame(DefaultCuckooConfig())
	ckSetHand(g, [CuckooPlayerCnt]int{5, 6, 7, 8}, 3)
	g.players[1].SetLives(0) // 席順の隣は脱落済み → 相手は 2 のはず

	target := g.GetSwapTargetIdx(0)
	require.Equal(t, 2, target)

	before := [CuckooPlayerCnt]int{}
	for i, p := range g.players {
		before[i] = p.CardValue()
	}
	g.attemptSwap(0)

	assert.Equal(t, before[target], g.players[0].CardValue(), "seat 0 did not receive the named seat's card")
	assert.Equal(t, before[0], g.players[target].CardValue(), "the named seat did not receive seat 0's card")
	// 飛ばされた席と無関係な席は動かない ── ここを見ないと「全員が回った」形と
	// 区別が付かない。
	assert.Equal(t, before[1], g.players[1].CardValue(), "an eliminated seat was traded with")
	assert.Equal(t, before[3], g.players[3].CardValue())
}
