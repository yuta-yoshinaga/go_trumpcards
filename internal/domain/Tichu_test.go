//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newAllCpuTichu() *Tichu {
	players := []*TichuPlayer{
		NewTichuPlayer(false),
		NewTichuPlayer(false),
		NewTichuPlayer(false),
		NewTichuPlayer(false),
	}
	return NewTichu(NewTrumpCards(TichuJokerCount), players, DefaultTichuConfig())
}

// drive a deal to completion with all-CPU players.
func driveToEnd(t *testing.T, g *Tichu) {
	t.Helper()
	guard := 0
	for g.GetPhase() == TichuPhaseDeclare {
		g.CpuPlay()
		guard++
		if guard > 100 {
			t.Fatal("declare phase did not terminate")
		}
	}
	guard = 0
	for !g.GetGameEndFlag() {
		g.CpuPlay()
		guard++
		if guard > 5000 {
			t.Fatal("play phase did not terminate")
		}
	}
}

func TestTichuFullGameAllCpu(t *testing.T) {
	oneTwos := 0
	for iter := 0; iter < 400; iter++ {
		g := newAllCpuTichu()
		g.Reset()
		driveToEnd(t, g)

		scores := g.GetScores()
		if g.GetIsOneTwo() {
			oneTwos++
		} else {
			declared := false
			for i := 0; i < TichuPlayerCnt; i++ {
				if g.GetPlayer(i).GetDeclType() != TichuDeclNone {
					declared = true
				}
			}
			if !declared && scores[0]+scores[1] != 100 {
				t.Fatalf("card points sum=%d want 100 (scores=%v)", scores[0]+scores[1], scores)
			}
		}

		// JSON round-trip preserves the result.
		data, err := json.Marshal(g)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var g2 Tichu
		if err := json.Unmarshal(data, &g2); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if g2.GetScores() != scores {
			t.Fatalf("round-trip score mismatch: %v vs %v", g2.GetScores(), scores)
		}
	}
	if oneTwos == 0 {
		t.Error("expected at least one one-two across many games")
	}
}

func TestTichuResetState(t *testing.T) {
	g := NewDefaultTichu()
	g.Reset()
	if g.GetPhase() != TichuPhaseDeclare {
		t.Errorf("phase after reset = %d, want declare", g.GetPhase())
	}
	total := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		total += g.GetPlayer(i).GetCardsSize()
	}
	if total != TichuHandSize*TichuPlayerCnt {
		t.Errorf("total dealt = %d, want %d", total, TichuHandSize*TichuPlayerCnt)
	}
	if g.GetPlayer(99) != nil {
		t.Error("out of range player should be nil")
	}
	// start leader holds the mahjong
	leader := g.GetStartLeader()
	found := false
	p := g.GetPlayer(leader)
	for i := 0; i < p.GetCardsSize(); i++ {
		if tichuSpecialKind(p.GetCard(i)) == TichuMahjong {
			found = true
		}
	}
	if !found {
		t.Error("start leader should hold the mahjong")
	}
}

func TestTichuPhaseGuards(t *testing.T) {
	g := NewDefaultTichu()
	g.Reset()
	// PlayerPlay before play phase -> error
	if err := g.PlayerPlay([]int{0}); err == nil {
		t.Error("play during declare phase should error")
	}
	// finish the game then declare/play should error
	cpuG := newAllCpuTichu()
	cpuG.Reset()
	driveToEnd(t, cpuG)
	if err := cpuG.PlayerDeclare(TichuDeclTichu); err == nil {
		t.Error("declare after end should error")
	}
	if err := cpuG.PlayerPlay([]int{0}); err == nil {
		t.Error("play after end should error")
	}
}

func TestTichuHumanDeclarationFlow(t *testing.T) {
	g := NewDefaultTichu()
	g.Reset()
	for g.GetPhase() == TichuPhaseDeclare && !g.IsHumanTurn() {
		g.CpuPlay()
	}
	if g.GetPhase() == TichuPhaseDeclare && g.IsHumanTurn() {
		if err := g.PlayerDeclare(99); err == nil {
			t.Error("invalid declaration should error")
		}
		if err := g.PlayerDeclare(TichuDeclTichu); err != nil {
			t.Errorf("valid declaration failed: %v", err)
		}
		if g.GetPlayer(0).GetDeclType() != TichuDeclTichu {
			t.Error("declaration not recorded")
		}
	}
}

func TestTichuHumanPlayFlow(t *testing.T) {
	g := NewDefaultTichu()
	g.Reset()
	for g.GetPhase() == TichuPhaseDeclare {
		if g.IsHumanTurn() {
			_ = g.PlayerDeclare(TichuDeclNone)
		} else {
			g.CpuPlay()
		}
	}
	for !g.GetGameEndFlag() && !g.IsHumanTurn() {
		g.CpuPlay()
	}
	if g.GetGameEndFlag() {
		return
	}
	if g.GetTableCombo() == nil {
		// leading: a pass is illegal, an out-of-range index errors, a single succeeds
		if err := g.PlayerPlay([]int{}); err == nil {
			t.Error("pass while leading should error")
		}
		if err := g.PlayerPlay([]int{999}); err == nil {
			t.Error("out of range index should error")
		}
		if err := g.PlayerPlay([]int{0}); err != nil {
			t.Errorf("leading lowest single should succeed: %v", err)
		}
	} else {
		if err := g.PlayerPlay([]int{}); err != nil {
			t.Errorf("pass while following should succeed: %v", err)
		}
	}
}

func TestTichuGettersAndConfig(t *testing.T) {
	g := NewDefaultTichu()
	g.Reset()
	if g.IsHumanTurn() != (g.GetCurrentTurn() == 0) {
		// player 0 is the only human
		if g.GetPlayer(g.GetCurrentTurn()).GetIsHuman() != g.IsHumanTurn() {
			t.Error("IsHumanTurn inconsistent")
		}
	}
	if g.HasPendingAction() {
		t.Error("Tichu never has pending action")
	}
	cfg := g.GetConfig()
	cfg.CpuDifficulty = TichuDifficultyHard
	g.SetConfig(cfg)
	if g.GetConfig().CpuDifficulty != TichuDifficultyHard {
		t.Error("SetConfig failed")
	}
	if g.GetBombCount() < 0 {
		t.Error("bomb count negative")
	}
	// drive one declaration so the action log is exercised
	if g.IsHumanTurn() {
		_ = g.PlayerDeclare(TichuDeclNone)
	} else {
		g.CpuPlay()
	}
	if len(g.GetActionLog()) == 0 {
		t.Error("action log should record the declaration")
	}
	_ = g.GetFinishOrder()
	_ = g.GetLastPlayIdx()
	_ = g.GetCpuActions()
	_ = g.GetHumanAction()
}

func TestTichuTeamOf(t *testing.T) {
	if TichuTeamOf(0) != TichuTeamOf(2) {
		t.Error("players 0 and 2 should be same team")
	}
	if TichuTeamOf(1) != TichuTeamOf(3) {
		t.Error("players 1 and 3 should be same team")
	}
	if TichuTeamOf(0) == TichuTeamOf(1) {
		t.Error("players 0 and 1 should be opposing teams")
	}
}

func TestTichuConfigValidate(t *testing.T) {
	if err := DefaultTichuConfig().Validate(); err != nil {
		t.Errorf("default config invalid: %v", err)
	}
	bad := TichuConfig{CpuDifficulty: 99}
	if err := bad.Validate(); err == nil {
		t.Error("out of range difficulty should be invalid")
	}
	data, err := json.Marshal(DefaultTichuConfig())
	if err != nil {
		t.Fatalf("config marshal: %v", err)
	}
	var c TichuConfig
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatalf("config unmarshal: %v", err)
	}
}

// テスト用セッターが効くこと。presenter からしか呼ばないとドメインの
// カバレッジに乗らない。
func TestTichu_TestSetters(t *testing.T) {
	g := newAllCpuTichu()
	g.Reset()
	g.SetBombCountForTest(3)
	assert.Equal(t, 3, g.GetBombCount())
	g.SetIsOneTwoForTest(true)
	assert.True(t, g.GetIsOneTwo())
}
