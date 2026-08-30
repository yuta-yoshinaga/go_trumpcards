//go:build test

package presenter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func setupNertzCuiMockDefaults(g *interfaces.MockNertzGame) {
	cfg := domain.DefaultNertzConfig()
	g.On("GetConfig").Return(cfg).Maybe()
	g.On("GetRoundNo").Return(1).Maybe()
	g.On("GetMoveCount").Return(2).Maybe()
	g.On("GetWinnerIdx").Return(-1).Maybe()
	g.On("GetMatchWinner").Return(-1).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("GetPhase").Return(domain.NertzPhasePlaying).Maybe()
	human := newNertzPlayerWithCards("You", false, 0)
	cpu := newNertzPlayerWithCards("CPU1", true, 1)
	g.On("GetPlayers").Return([]*domain.NertzPlayer{human, cpu}).Maybe()
	g.On("GetFoundations").Return([]*domain.NertzFoundation{
		newNertzFoundationWithAce(),
		domain.NewNertzFoundation(),
	}).Maybe()
}

// 生スコアだけでは何点で決着するのか分からない。Web は同じ値をスコアバーの
// aria-valuemax にしている。目標そのものと、各席の残りの両方を見る。
func TestNertzCuiPresenter_ShowsTheTargetAndEachSeatsDistanceToIt(t *testing.T) {
	g := new(interfaces.MockNertzGame)
	cfg := domain.DefaultNertzConfig()
	// 既定 (NertzTargetScoreDefault = 100) と**違う**値にする。既定と同じにすると、
	// 定数を直に読む実装と設定を読む実装が見分けられない。
	cfg.TargetScore = 250
	g.On("GetConfig").Return(cfg).Maybe()
	g.On("GetRoundNo").Return(1).Maybe()
	g.On("GetMoveCount").Return(2).Maybe()
	g.On("GetWinnerIdx").Return(-1).Maybe()
	g.On("GetMatchWinner").Return(-1).Maybe()
	g.On("CanUndo").Return(false).Maybe()
	g.On("GetPhase").Return(domain.NertzPhasePlaying).Maybe()
	human := newNertzPlayerWithCards("You", false, 0)
	human.SetScore(70)
	cpu := newNertzPlayerWithCards("CPU1", true, 1)
	cpu.SetScore(300) // 目標を越えている席。残りは負でなく 0 と出す。
	g.On("GetPlayers").Return([]*domain.NertzPlayer{human, cpu}).Maybe()
	g.On("GetFoundations").Return([]*domain.NertzFoundation{
		newNertzFoundationWithAce(),
		domain.NewNertzFoundation(),
	}).Maybe()

	out := new(NertzCuiPresenter).Output(g, nil)
	assert.Contains(t, out, i18n.Tf("nertz.targetLine", "target", "250"))
	assert.Contains(t, out, i18n.Tf("nertz.playerLine",
		"idx", "0", "label", i18n.T("nertz.labelHuman"), "name", "You", "score", "70", "remaining", "180"))
	assert.Contains(t, out, i18n.Tf("nertz.playerLine",
		"idx", "1", "label", i18n.T("nertz.labelCpu"), "name", "CPU1", "score", "300", "remaining", "0"))
	assert.NotContains(t, out, "-50")
}

func TestNertzCuiPresenter_Output(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		setupNertzCuiMockDefaults(g)
		out := new(NertzCuiPresenter).Output(g, nil)
		assert.Contains(t, out, "Nertz / Pounce")
		assert.Contains(t, out, "Foundations")
		assert.Contains(t, out, "プレイ中")
	})
	t.Run("empty foundation and tableau", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		g.On("GetConfig").Return(domain.DefaultNertzConfig()).Maybe()
		g.On("GetRoundNo").Return(1).Maybe()
		g.On("GetMoveCount").Return(0).Maybe()
		g.On("GetWinnerIdx").Return(-1).Maybe()
		g.On("GetMatchWinner").Return(-1).Maybe()
		g.On("CanUndo").Return(false).Maybe()
		g.On("GetPhase").Return(domain.NertzPhasePlaying).Maybe()
		empty := domain.NewNertzPlayer("Empty", false, 0)
		g.On("GetPlayers").Return([]*domain.NertzPlayer{empty, nil}).Maybe()
		g.On("GetFoundations").Return([]*domain.NertzFoundation{nil, domain.NewNertzFoundation()}).Maybe()
		out := new(NertzCuiPresenter).Output(g, nil)
		assert.Contains(t, out, "(empty)")
	})
	t.Run("with error", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		setupNertzCuiMockDefaults(g)
		out := new(NertzCuiPresenter).Output(g, errors.New("boom"))
		assert.Contains(t, out, "boom")
	})
	t.Run("round end", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		g.On("GetConfig").Return(domain.DefaultNertzConfig()).Maybe()
		g.On("GetRoundNo").Return(1).Maybe()
		g.On("GetMoveCount").Return(15).Maybe()
		g.On("GetWinnerIdx").Return(0).Maybe()
		g.On("GetMatchWinner").Return(-1).Maybe()
		g.On("CanUndo").Return(false).Maybe()
		g.On("GetPhase").Return(domain.NertzPhaseRoundEnd).Maybe()
		g.On("GetPlayers").Return([]*domain.NertzPlayer{}).Maybe()
		g.On("GetFoundations").Return([]*domain.NertzFoundation{}).Maybe()
		out := new(NertzCuiPresenter).Output(g, nil)
		assert.Contains(t, out, "ラウンド終了")
	})
	t.Run("game end human", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		g.On("GetConfig").Return(domain.DefaultNertzConfig()).Maybe()
		g.On("GetRoundNo").Return(3).Maybe()
		g.On("GetMoveCount").Return(50).Maybe()
		g.On("GetWinnerIdx").Return(0).Maybe()
		g.On("GetMatchWinner").Return(0).Maybe()
		g.On("CanUndo").Return(false).Maybe()
		g.On("GetPhase").Return(domain.NertzPhaseGameEnd).Maybe()
		g.On("GetPlayers").Return([]*domain.NertzPlayer{}).Maybe()
		g.On("GetFoundations").Return([]*domain.NertzFoundation{}).Maybe()
		out := new(NertzCuiPresenter).Output(g, nil)
		assert.Contains(t, out, "あなたの勝ち")
	})
	t.Run("game end cpu", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		g.On("GetConfig").Return(domain.DefaultNertzConfig()).Maybe()
		g.On("GetRoundNo").Return(3).Maybe()
		g.On("GetMoveCount").Return(50).Maybe()
		g.On("GetWinnerIdx").Return(2).Maybe()
		g.On("GetMatchWinner").Return(2).Maybe()
		g.On("CanUndo").Return(false).Maybe()
		g.On("GetPhase").Return(domain.NertzPhaseGameEnd).Maybe()
		g.On("GetPlayers").Return([]*domain.NertzPlayer{}).Maybe()
		g.On("GetFoundations").Return([]*domain.NertzFoundation{}).Maybe()
		out := new(NertzCuiPresenter).Output(g, nil)
		assert.Contains(t, out, "プレイヤー2の勝ち")
	})
}

func TestNertzCuiPresenter_HintOutput(t *testing.T) {
	tests := []struct {
		name string
		hint *domain.NertzHint
		want string
	}{
		{"none", nil, "ヒントはありません"},
		{"nertz->foundation", &domain.NertzHint{FromZone: "nertz", FromCol: -1, CardIndex: -1, ToZone: "foundation", ToCol: 0}, "ナッツ → ファウンデーション0"},
		{"waste->tableau", &domain.NertzHint{FromZone: "waste", FromCol: -1, CardIndex: -1, ToZone: "tableau", ToCol: 1}, "ウェイスト → タブロー1"},
		{"tableau->foundation", &domain.NertzHint{FromZone: "tableau", FromCol: 0, CardIndex: 2, ToZone: "foundation", ToCol: 1}, "タブロー0(idx=2) → ファウンデーション1"},
		{"tableau->tableau no idx", &domain.NertzHint{FromZone: "tableau", FromCol: 0, CardIndex: -1, ToZone: "tableau", ToCol: 2}, "タブロー0 → タブロー2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := new(interfaces.MockNertzGame)
			g.On("GetHint").Return(tt.hint)
			out := new(NertzCuiPresenter).HintOutput(g)
			assert.Contains(t, out, tt.want)
		})
	}
}

func TestNertzCuiPresenter_HintZoneLabelUnknown(t *testing.T) {
	assert.Equal(t, "weird", nertzHintZoneLabel("weird", 0, 0))
}

func TestNertzCuiPresenter_ActionLogOutput(t *testing.T) {
	t.Run("playing", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		g.On("GetPhase").Return(domain.NertzPhasePlaying)
		assert.NotEmpty(t, new(NertzCuiPresenter).ActionLogOutput(g))
	})
	t.Run("after round end", func(t *testing.T) {
		g := new(interfaces.MockNertzGame)
		g.On("GetPhase").Return(domain.NertzPhaseRoundEnd)
		g.On("GetActionLog").Return([]*domain.ActionLogEntry{{TurnNumber: 1, ActionType: "moveNF"}})
		assert.NotEmpty(t, new(NertzCuiPresenter).ActionLogOutput(g))
	})
}
