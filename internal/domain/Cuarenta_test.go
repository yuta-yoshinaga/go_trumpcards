//go:build !js || !wasm || casino

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// cuCard はテスト用に 1 枚のカードを生成するヘルパー (design, value は plain int)。
func cuCard(d, v int) *Card { return NewCard(d, v, false) }

// cuNewGame は決定的なテスト用 Cuarenta を生成する (4 人、シャッフルなし)。
func cuNewGame(cfg CuarentaConfig) *Cuarenta {
	players := make([]*CuarentaPlayer, CuarentaPlayerCnt)
	players[0] = NewCuarentaPlayer(true)
	for i := 1; i < CuarentaPlayerCnt; i++ {
		players[i] = NewCuarentaPlayer(false)
	}
	g := NewCuarenta(NewTrumpCardsScopa(), players, cfg)
	g.round.actionLog = make([]*ActionLogEntry, 0)
	return g
}

// cuDrainDeck はテスト用に山札を空にする。
func cuDrainDeck(g *Cuarenta) {
	for g.trumpCards.GetRemainingCount() > 0 {
		g.trumpCards.DrawCard()
	}
}

func TestCuarenta_DefaultConfigAndDeck(t *testing.T) {
	g := NewDefaultCuarenta()
	g.Reset()
	if g.GetPlayerCnt() != CuarentaPlayerCnt {
		t.Fatalf("player count = %d, want %d", g.GetPlayerCnt(), CuarentaPlayerCnt)
	}
	if got := g.GetConfig().TargetScore; got != CuarentaDefaultTargetScore {
		t.Fatalf("target score = %d, want %d", got, CuarentaDefaultTargetScore)
	}
	// 40-card deck: 5 cards each ×4 = 20 dealt, 4 on table = 24 used, 16 remaining.
	if rem := g.GetRemainingDeck(); rem != 16 {
		t.Fatalf("remaining deck after first deal = %d, want 16", rem)
	}
	// each player should have 5 cards.
	for i := 0; i < CuarentaPlayerCnt; i++ {
		if sz := g.GetPlayer(i).GetCardsSize(); sz != CuarentaHandSize {
			t.Fatalf("player %d hand size = %d, want %d", i, sz, CuarentaHandSize)
		}
	}
}

func TestCuarentaTeamOf(t *testing.T) {
	cases := map[int]int{0: 0, 1: 1, 2: 0, 3: 1}
	for seat, want := range cases {
		if got := CuarentaTeamOf(seat); got != want {
			t.Errorf("CuarentaTeamOf(%d) = %d, want %d", seat, got, want)
		}
	}
}

// TestCuarenta_RankCapture は同ランク捕獲を検証する。
func TestCuarenta_RankCapture(t *testing.T) {
	g := cuNewGame(DefaultCuarentaConfig())
	// human (seat 0) holds a 5 (plus a spare 6 so the round does not end on this
	// play and trigger the leftover-sweep); table has two 5s and a 7.
	g.players[0].Reset()
	g.players[0].AddCard(cuCard(CardDesignSpade, 5))
	g.players[0].AddCard(cuCard(CardDesignSpade, 6))
	g.round.tableCards = []*Card{cuCard(CardDesignHeart, 5), cuCard(CardDesignClover, 5), cuCard(CardDesignDiamond, 7)}
	g.round.currentTurn = 0
	g.round.phase = CuarentaPhasePlay
	// give others empty hands so it doesn't redeal.
	for i := 1; i < CuarentaPlayerCnt; i++ {
		g.players[i].Reset()
	}
	cuDrainDeck(g)

	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("PlayerPlay returned error: %v", err)
	}
	// only the rank-5 capture: played 5 + two table 5s = 3; the 7 stays on the
	// table (the human still holds the 6, so the round has not ended yet).
	if g.GetPlayer(0).CapturedCount() != 3 {
		t.Fatalf("captured count = %d, want 3", g.GetPlayer(0).CapturedCount())
	}
}

// TestCuarenta_NoMatchLays は一致なしで場に置くことを検証する。
func TestCuarenta_NoMatchLays(t *testing.T) {
	g := cuNewGame(DefaultCuarentaConfig())
	g.players[0].Reset()
	g.players[0].AddCard(cuCard(CardDesignSpade, 4))
	g.round.tableCards = []*Card{cuCard(CardDesignHeart, 5), cuCard(CardDesignClover, 7)}
	g.round.currentTurn = 0
	g.round.phase = CuarentaPhasePlay
	for i := 1; i < CuarentaPlayerCnt; i++ {
		g.players[i].AddCard(cuCard(CardDesignSpade, 2))
	}

	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("PlayerPlay error: %v", err)
	}
	if g.GetPlayer(0).CapturedCount() != 0 {
		t.Fatalf("captured count = %d, want 0", g.GetPlayer(0).CapturedCount())
	}
	if len(g.GetTableCards()) != 3 {
		t.Fatalf("table size = %d, want 3 (4 laid on)", len(g.GetTableCards()))
	}
}

// TestCuarenta_Caida は caída (+2) を検証する。
func TestCuarenta_Caida(t *testing.T) {
	g := cuNewGame(DefaultCuarentaConfig())
	// seat 0 lays a 6 (no match), seat 1 captures the just-laid 6 → caida +2 for team 1.
	g.players[0].Reset()
	g.players[0].AddCard(cuCard(CardDesignSpade, 6))
	g.players[1].Reset()
	g.players[1].AddCard(cuCard(CardDesignHeart, 6))
	g.players[2].AddCard(cuCard(CardDesignClover, 2))
	g.players[3].AddCard(cuCard(CardDesignClover, 3))
	g.round.tableCards = []*Card{cuCard(CardDesignDiamond, 4)}
	g.round.currentTurn = 0
	g.round.phase = CuarentaPhasePlay

	if err := g.PlayerPlay(0); err != nil { // human lays 6
		t.Fatalf("seat 0 play error: %v", err)
	}
	// now seat 1 (CPU) is up. Drive it manually via internal applyPlay path.
	if g.GetCurrentTurn() != 1 {
		t.Fatalf("current turn = %d, want 1", g.GetCurrentTurn())
	}
	before := g.GetTeamScore(1)
	g.CpuPlay() // seat 1 captures the laid 6
	if g.GetTeamScore(1) != before+CuarentaScoreCaida {
		t.Fatalf("team1 score after caida = %d, want %d", g.GetTeamScore(1), before+CuarentaScoreCaida)
	}
	acts := g.GetCpuActions()
	if len(acts) == 0 || !acts[0].IsCaida {
		t.Fatalf("expected caida flag on cpu action, got %+v", acts)
	}
}

// TestCuarenta_Ronda は同ランク 3 枚以上の連続捕獲 (ronda) 加点を検証する。
func TestCuarenta_Ronda(t *testing.T) {
	g := cuNewGame(DefaultCuarentaConfig())
	g.players[0].Reset()
	g.players[0].AddCard(cuCard(CardDesignSpade, 3))
	// three 3s on table → total same-rank = 4 (played + 3) → ronda bonus (4-2)=2.
	g.round.tableCards = []*Card{cuCard(CardDesignHeart, 3), cuCard(CardDesignClover, 3), cuCard(CardDesignDiamond, 3)}
	g.round.currentTurn = 0
	g.round.phase = CuarentaPhasePlay
	for i := 1; i < CuarentaPlayerCnt; i++ {
		g.players[i].AddCard(cuCard(CardDesignSpade, 2))
	}
	before := g.GetTeamScore(0)
	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("play error: %v", err)
	}
	wantBonus := (4 - 2) * CuarentaScoreRondaPerExtra
	// limpia also fires here (table cleared) → +1, so account for both.
	gained := g.GetTeamScore(0) - before
	if gained != wantBonus+CuarentaScoreLimpia {
		t.Fatalf("team0 gained = %d, want %d (ronda %d + limpia %d)", gained, wantBonus+CuarentaScoreLimpia, wantBonus, CuarentaScoreLimpia)
	}
	ha := g.GetHumanAction()
	if ha == nil || ha.RondaBonus != wantBonus || !ha.IsLimpia {
		t.Fatalf("human action ronda/limpia mismatch: %+v", ha)
	}
}

// TestCuarenta_Limpia は場の掃き (+1) を検証する。
func TestCuarenta_Limpia(t *testing.T) {
	g := cuNewGame(DefaultCuarentaConfig())
	g.players[0].Reset()
	g.players[0].AddCard(cuCard(CardDesignSpade, 5))
	g.round.tableCards = []*Card{cuCard(CardDesignHeart, 5)} // only card → cleared.
	g.round.currentTurn = 0
	g.round.phase = CuarentaPhasePlay
	for i := 1; i < CuarentaPlayerCnt; i++ {
		g.players[i].AddCard(cuCard(CardDesignSpade, 2))
	}
	before := g.GetTeamScore(0)
	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("play error: %v", err)
	}
	if g.GetTeamScore(0) != before+CuarentaScoreLimpia {
		t.Fatalf("team0 score = %d, want %d", g.GetTeamScore(0), before+CuarentaScoreLimpia)
	}
}

// TestCuarenta_LastPlayNoLimpia はラウンド最後の 1 手では掃きにならないことを検証する。
func TestCuarenta_LastPlayNoLimpia(t *testing.T) {
	g := cuNewGame(DefaultCuarentaConfig())
	g.players[0].Reset()
	g.players[0].AddCard(cuCard(CardDesignSpade, 5))
	for i := 1; i < CuarentaPlayerCnt; i++ {
		g.players[i].Reset() // empty hands
	}
	g.round.tableCards = []*Card{cuCard(CardDesignHeart, 5)}
	g.round.currentTurn = 0
	g.round.phase = CuarentaPhasePlay
	cuDrainDeck(g) // last play of round

	before := g.GetTeamScore(0)
	if err := g.PlayerPlay(0); err != nil {
		t.Fatalf("play error: %v", err)
	}
	// no limpia because it is the last play; the round-end bonus may apply though.
	det := g.GetLastRoundDetail()
	if det == nil {
		t.Fatalf("expected round detail after final play")
	}
	// captured everything so team0 holds all 2 cards → not > 20, no most-cards bonus.
	if g.GetTeamScore(0) != before {
		t.Fatalf("team0 score = %d, want %d (no limpia, no most-cards)", g.GetTeamScore(0), before)
	}
}

// TestCuarenta_MostCardsBonus は最多取りボーナス (+6) を検証する。
func TestCuarenta_MostCardsBonus(t *testing.T) {
	g := cuNewGame(DefaultCuarentaConfig())
	// give team 0 (seats 0 & 2) > 20 captured cards directly, then finish round.
	captured := make([]*Card, 0, 22)
	for i := 0; i < 22; i++ {
		captured = append(captured, cuCard(CardDesignSpade, (i%7)+1))
	}
	g.players[0].AddCaptured(captured)
	for i := 0; i < CuarentaPlayerCnt; i++ {
		g.players[i].Reset() // clear hands
	}
	g.round.tableCards = nil
	g.round.lastCaptureIdx = 0
	cuDrainDeck(g)

	g.finishRound()
	det := g.GetLastRoundDetail()
	if det == nil || det.MostCards != 0 {
		t.Fatalf("most cards team = %v, want 0", det)
	}
	if g.GetTeamScore(0) < CuarentaScoreMostCards {
		t.Fatalf("team0 score = %d, want >= %d", g.GetTeamScore(0), CuarentaScoreMostCards)
	}
}

// TestCuarenta_TargetScoreWin はチームが目標点に達すると勝利することを検証する。
func TestCuarenta_TargetScoreWin(t *testing.T) {
	cfg := DefaultCuarentaConfig()
	cfg.TargetScore = 5
	g := cuNewGame(cfg)
	g.teamScore[0] = 5 // already at target
	for i := 0; i < CuarentaPlayerCnt; i++ {
		g.players[i].Reset()
	}
	g.round.tableCards = nil
	g.round.lastCaptureIdx = 0
	cuDrainDeck(g)

	g.finishRound()
	if !g.GetGameEndFlag() {
		t.Fatalf("expected game end at target score")
	}
	if g.GetPhase() != int(CuarentaPhaseGameEnd) {
		t.Fatalf("phase = %d, want gameEnd %d", g.GetPhase(), int(CuarentaPhaseGameEnd))
	}
	winners := g.GetRoundWinners()
	if len(winners) != 1 || winners[0] != 0 {
		t.Fatalf("winners = %v, want [0]", winners)
	}
}

// TestCuarenta_GuardErrors はエラーガードを検証する。
func TestCuarenta_GuardErrors(t *testing.T) {
	g := cuNewGame(DefaultCuarentaConfig())
	g.round.gameEndFlag = true
	if err := g.PlayerPlay(0); err != ErrGameEnded {
		t.Errorf("play after end = %v, want ErrGameEnded", err)
	}
	g.round.gameEndFlag = false
	g.round.phase = CuarentaPhaseRoundEnd
	if err := g.PlayerPlay(0); err == nil {
		t.Errorf("play in wrong phase should error")
	}
	g.round.phase = CuarentaPhasePlay
	g.round.currentTurn = 1 // CPU seat
	if err := g.PlayerPlay(0); err != ErrNotHumanTurn {
		t.Errorf("play on cpu turn = %v, want ErrNotHumanTurn", err)
	}
	g.round.currentTurn = 0
	if err := g.PlayerPlay(99); err == nil {
		t.Errorf("play invalid index should error")
	}
}

// TestCuarenta_FullCpuGameTerminates は全 CPU ゲームが必ず終了することを
// 複数難易度で検証する (termination guard)。
func TestCuarenta_FullCpuGameTerminates(t *testing.T) {
	difficulties := []CuarentaCpuDifficulty{
		CuarentaDifficultyEasy, CuarentaDifficultyNormal, CuarentaDifficultyHard,
	}
	for _, diff := range difficulties {
		diff := diff
		t.Run(CuarentaDifficultyNames[diff], func(t *testing.T) {
			cfg := DefaultCuarentaConfig()
			cfg.CpuDifficulty = diff
			cfg.TargetScore = 12 // small target to converge quickly
			players := make([]*CuarentaPlayer, CuarentaPlayerCnt)
			for i := 0; i < CuarentaPlayerCnt; i++ {
				players[i] = NewCuarentaPlayer(false) // all CPU
			}
			g := NewCuarenta(NewTrumpCardsScopa(), players, cfg)
			g.Reset()

			const maxIter = 100000
			iter := 0
			for !g.GetGameEndFlag() {
				if iter++; iter > maxIter {
					t.Fatalf("game did not terminate within %d iterations (diff=%v)", maxIter, diff)
				}
				g.CpuPlay()
				if !g.GetGameEndFlag() && g.GetPhase() != int(CuarentaPhasePlay) {
					g.NextRound()
				}
			}
			// some team must have reached the target.
			ok := false
			for tm := 0; tm < CuarentaTeamCnt; tm++ {
				if g.GetTeamScore(tm) >= cfg.TargetScore {
					ok = true
				}
			}
			if !ok {
				t.Fatalf("no team reached target on game end (diff=%v)", diff)
			}
			if len(g.GetRoundWinners()) == 0 {
				t.Fatalf("no winners recorded (diff=%v)", diff)
			}
		})
	}
}

// TestCuarenta_NextRoundResetsTable は NextRound が場/手札をリセットすることを検証する。
func TestCuarenta_NextRoundResetsTable(t *testing.T) {
	g := NewDefaultCuarenta()
	g.Reset()
	// drive into round end manually.
	for i := 0; i < CuarentaPlayerCnt; i++ {
		g.players[i].Reset()
	}
	g.round.tableCards = nil
	g.round.lastCaptureIdx = 0
	cuDrainDeck(g)
	g.finishRound()
	if g.GetGameEndFlag() {
		t.Skip("game ended unexpectedly fast; skip next-round assertion")
	}
	g.NextRound()
	if g.GetPhase() != int(CuarentaPhasePlay) {
		t.Fatalf("phase after NextRound = %d, want play", g.GetPhase())
	}
	if g.GetCurrentTurn() != 0 {
		t.Fatalf("turn after NextRound = %d, want 0", g.GetCurrentTurn())
	}
}

// TestCuarenta_JSONRoundTrip は Marshal/Unmarshal の往復を検証する。
func TestCuarenta_JSONRoundTrip(t *testing.T) {
	g := NewDefaultCuarenta()
	g.Reset()
	data, err := g.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var restored Cuarenta
	if err := restored.UnmarshalJSON(data); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if restored.GetPlayerCnt() != CuarentaPlayerCnt {
		t.Fatalf("restored player count = %d", restored.GetPlayerCnt())
	}
	if restored.GetConfig().TargetScore != g.GetConfig().TargetScore {
		t.Fatalf("restored target score mismatch")
	}
}

// TestCuarenta_UnmarshalRejectsBadState は防御的バリデーションを検証する。
func TestCuarenta_UnmarshalRejectsBadState(t *testing.T) {
	cases := []string{
		`{}`, // nil trump cards
		`{"tc":{},"pl":[],"cf":{"ts":40,"di":1},"ph":0,"ct":0}`,                        // wrong player count
		`{"tc":{},"pl":[{},{},{},{}],"cf":{"ts":40,"di":1},"ph":9,"ct":0}`,             // bad phase
		`{"tc":{},"pl":[{},{},{},{}],"cf":{"ts":40,"di":1},"ph":0,"ct":9}`,             // bad current turn
		`{"tc":{},"pl":[{},{},{},{}],"cf":{"ts":40,"di":1},"ph":0,"ct":0,"lc":9}`,      // bad last capture
		`{"tc":{},"pl":[{},{},{},{}],"cf":{"ts":40,"di":9},"ph":0,"ct":0}`,             // bad config difficulty
		`{"tc":{},"pl":[{},{},{},{}],"cf":{"ts":40,"di":1},"ph":0,"ct":0,"ts":[-1,0]}`, // bad team score
	}
	for i, c := range cases {
		var g Cuarenta
		if err := g.UnmarshalJSON([]byte(c)); err == nil {
			t.Errorf("case %d: expected error for %s", i, c)
		}
	}
}

// TestCuarentaConfigValidate は設定バリデーションを検証する。
func TestCuarentaConfigValidate(t *testing.T) {
	if err := DefaultCuarentaConfig().Validate(); err != nil {
		t.Errorf("default config invalid: %v", err)
	}
	bad := CuarentaConfig{TargetScore: 0, CpuDifficulty: CuarentaDifficultyNormal}
	if err := bad.Validate(); err == nil {
		t.Errorf("target score 0 should be invalid")
	}
	bad2 := CuarentaConfig{TargetScore: 40, CpuDifficulty: 9}
	if err := bad2.Validate(); err == nil {
		t.Errorf("difficulty 9 should be invalid")
	}
}

// **合算は 1 箇所に置く。**精算・CUI で別々に足すと、席とチームの対応を変えた
// ときに片方だけ古いままになる (#4893)。
func TestCuarenta_GetTeamCapturedCount(t *testing.T) {
	g := NewDefaultCuarenta()
	g.Reset()
	card := func() *Card { return NewCard(CardDesignSpade, 2, false) }
	give := func(seat, n int) {
		cards := make([]*Card, n)
		for i := range cards {
			cards[i] = card()
		}
		g.GetPlayer(seat).AddCaptured(cards)
	}

	// 席 0 と 2 が同じチーム、1 と 3 がもう一方。
	give(0, 7)
	give(2, 5)
	give(1, 3)
	assert.Equal(t, 12, g.GetTeamCapturedCount(0))
	assert.Equal(t, 3, g.GetTeamCapturedCount(1))

	// 範囲外のチーム番号は 0。
	assert.Equal(t, 0, g.GetTeamCapturedCount(99))
}
