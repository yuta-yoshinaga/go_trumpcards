//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// SevenBridgeInteractorIF セブンブリッジインタラクターインタフェース
type SevenBridgeInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.SevenBridgeConfig) string
	// DrawFromStock 山札からカードを引く
	DrawFromStock() string
	// ClaimPon ポンで捨て札を取得する
	ClaimPon(cardIndices []int) string
	// ClaimChi チーで捨て札を取得する
	ClaimChi(cardIndices []int) string
	// Meld メルドを場に出す
	Meld(cardIndices []int) string
	// Layoff 既存メルドにカードを 1 枚追加する
	Layoff(targetPlayerIdx, meldIdx, cardIndex int) string
	// Discard カードを捨てる
	Discard(cardIndex int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.SevenBridgeConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// SevenBridgeInteractor セブンブリッジインタラクター
type SevenBridgeInteractor struct {
	GameBase[interfaces.SevenBridgeGame]
	gp presenter.SevenBridgePresenter
}

// NewSevenBridgeInteractor コンストラクタ
func NewSevenBridgeInteractor(g interfaces.SevenBridgeGame, gp presenter.SevenBridgePresenter) *SevenBridgeInteractor {
	mustNotNil("SevenBridgeInteractor", map[string]any{"g": g, "gp": gp})
	return &SevenBridgeInteractor{GameBase: GameBase[interfaces.SevenBridgeGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (ci *SevenBridgeInteractor) Reset() string {
	ci.Game.Reset()
	ci.runCpuTurns()
	return ci.gp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *SevenBridgeInteractor) ResetWithConfig(cfg domain.SevenBridgeConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.gp, cfg, ci.Game.SetConfig, ci.Reset)
}

// DrawFromStock 山札からカードを引く
func (ci *SevenBridgeInteractor) DrawFromStock() string {
	return humanAction(ci.Game, ci.gp, func() error { return ci.Game.PlayerDrawFromStock() }, ci.runCpuTurns)
}

// ClaimPon ポンで捨て札を取得する
func (ci *SevenBridgeInteractor) ClaimPon(cardIndices []int) string {
	return humanAction(ci.Game, ci.gp, func() error { return ci.Game.PlayerClaimPon(cardIndices) }, ci.runCpuTurns)
}

// ClaimChi チーで捨て札を取得する
func (ci *SevenBridgeInteractor) ClaimChi(cardIndices []int) string {
	return humanAction(ci.Game, ci.gp, func() error { return ci.Game.PlayerClaimChi(cardIndices) }, ci.runCpuTurns)
}

// Meld メルドを場に出す
func (ci *SevenBridgeInteractor) Meld(cardIndices []int) string {
	return humanAction(ci.Game, ci.gp, func() error { return ci.Game.PlayerMeld(cardIndices) }, ci.runCpuTurns)
}

// Layoff 既存メルドにカードを 1 枚追加する
func (ci *SevenBridgeInteractor) Layoff(targetPlayerIdx, meldIdx, cardIndex int) string {
	return humanAction(ci.Game, ci.gp, func() error { return ci.Game.PlayerLayoff(targetPlayerIdx, meldIdx, cardIndex) }, ci.runCpuTurns)
}

// Discard カードを捨てる
func (ci *SevenBridgeInteractor) Discard(cardIndex int) string {
	return humanAction(ci.Game, ci.gp, func() error { return ci.Game.PlayerDiscard(cardIndex) }, ci.runCpuTurns)
}

// NextRound 次のラウンドへ進む
func (ci *SevenBridgeInteractor) NextRound() string {
	return advanceRound(ci.Game, ci.gp, ci.runCpuTurns)
}

// GetConfig 現在の設定を取得
func (ci *SevenBridgeInteractor) GetConfig() domain.SevenBridgeConfig {
	return ci.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ci *SevenBridgeInteractor) ActionLog() string {
	return ci.gp.ActionLogOutput(ci.Game)
}

// Hint ヒント取得
func (ci *SevenBridgeInteractor) Hint() string {
	return ci.gp.HintOutput(ci.Game)
}

// runCpuTurns CPU ターンを連続で処理する
func (ci *SevenBridgeInteractor) runCpuTurns() {
	runCpuTurnsUntil(ci.Game, func() bool {
		phase := ci.Game.GetPhase()
		return phase == domain.SevenBridgePhaseRoundEnd || phase == domain.SevenBridgePhaseGameEnd || ci.Game.IsHumanTurn()
	}, ci.Game.CpuPlay)
}

// RestoreSevenBridgeInteractor deserialises JSON into a SevenBridgeInteractor.
func RestoreSevenBridgeInteractor(data []byte, gp presenter.SevenBridgePresenter) (*SevenBridgeInteractor, error) {
	return restoreAndBuild[domain.SevenBridge](data, func(g *domain.SevenBridge) *SevenBridgeInteractor {
		return &SevenBridgeInteractor{GameBase: GameBase[interfaces.SevenBridgeGame]{Game: g}, gp: gp}
	})
}
