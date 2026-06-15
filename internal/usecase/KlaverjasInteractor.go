//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// KlaverjasInteractorIF クラヴァヤスのインタラクターインタフェース
type KlaverjasInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.KlaverjasConfig) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.KlaverjasConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// KlaverjasInteractor クラヴァヤスのインタラクタークラス
type KlaverjasInteractor struct {
	GameBase[interfaces.KlaverjasGame]
	kp presenter.KlaverjasPresenter
}

// NewKlaverjasInteractor コンストラクタ
func NewKlaverjasInteractor(g interfaces.KlaverjasGame, kp presenter.KlaverjasPresenter) *KlaverjasInteractor {
	mustNotNil("KlaverjasInteractor", map[string]any{"g": g, "kp": kp})
	return &KlaverjasInteractor{GameBase: GameBase[interfaces.KlaverjasGame]{Game: g}, kp: kp}
}

// Reset ゲーム初期化
func (ki *KlaverjasInteractor) Reset() string {
	ki.Game.Reset()
	ki.runCpuTurns()
	return ki.kp.Output(ki.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ki *KlaverjasInteractor) ResetWithConfig(cfg domain.KlaverjasConfig) string {
	return resetWithValidatedConfig(ki.Game, ki.kp, cfg, ki.Game.SetConfig, ki.Reset)
}

// Play カードをプレイ
func (ki *KlaverjasInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ki.Game, ki.kp); blocked {
		return out
	}
	err := ki.Game.PlayerPlay(cardIndex)
	if err != nil {
		return ki.kp.Output(ki.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ki.Game.GetPhase() == domain.KlaverjasPhaseTrickEnd {
		ki.Game.ResolveTrick()
	}
	ki.runCpuTurns()
	return ki.kp.Output(ki.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ki *KlaverjasInteractor) NextTrick() string {
	ki.Game.NextTrick()
	ki.runCpuTurns()
	return ki.kp.Output(ki.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ki *KlaverjasInteractor) NextRound() string {
	ki.Game.ScoreRound()
	if out, blocked := guardGameEnd(ki.Game, ki.kp); blocked {
		return out
	}
	ki.Game.NextRound()
	ki.runCpuTurns()
	return ki.kp.Output(ki.Game, nil)
}

// GetConfig 現在の設定を取得
func (ki *KlaverjasInteractor) GetConfig() domain.KlaverjasConfig {
	return ki.Game.GetConfig()
}

// Hint ヒント取得
func (ki *KlaverjasInteractor) Hint() string {
	return ki.kp.HintOutput(ki.Game)
}

// ActionLog 棋譜を出力する
func (ki *KlaverjasInteractor) ActionLog() string {
	return ki.kp.ActionLogOutput(ki.Game)
}

// runCpuTurns ゲーム終了・人間の手番・トリック/ラウンド終了になるまで CPU ターンを実行する。
func (ki *KlaverjasInteractor) runCpuTurns() {
	runCpuTurnsLoop(ki.Game, trickPhases[domain.KlaverjasPhase]{
		play:     domain.KlaverjasPhasePlay,
		trickEnd: domain.KlaverjasPhaseTrickEnd,
		roundEnd: domain.KlaverjasPhaseRoundEnd,
		gameEnd:  domain.KlaverjasPhaseGameEnd,
	})
}

// RestoreKlaverjasInteractor deserialises JSON into a KlaverjasInteractor.
func RestoreKlaverjasInteractor(data []byte, kp presenter.KlaverjasPresenter) (*KlaverjasInteractor, error) {
	return restoreAndBuild[domain.Klaverjas](data, func(g *domain.Klaverjas) *KlaverjasInteractor {
		return &KlaverjasInteractor{GameBase: GameBase[interfaces.KlaverjasGame]{Game: g}, kp: kp}
	})
}
