//go:build !js || !wasm || extra3

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// UltiInteractorIF ウルティ (Ulti) のインタラクターインタフェース
type UltiInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.UltiConfig) string
	// Bid コントラクトを宣言する (party は切り札スートも)
	Bid(contract domain.UltiContract, trumpSuit int) string
	// Discard タロン受け取り後に 2 枚を捨てる
	Discard(cardIndices []int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ディールをスコアリングして次のディールへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.UltiConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// UltiInteractor ウルティのインタラクタークラス
type UltiInteractor struct {
	GameBase[interfaces.UltiGame]
	tp presenter.UltiPresenter
}

// NewUltiInteractor コンストラクタ
func NewUltiInteractor(g interfaces.UltiGame, tp presenter.UltiPresenter) *UltiInteractor {
	mustNotNil("UltiInteractor", map[string]any{"g": g, "tp": tp})
	return &UltiInteractor{GameBase: GameBase[interfaces.UltiGame]{Game: g}, tp: tp}
}

// ultiTrickPhases Ulti のトリックフェーズ定数
func ultiTrickPhases() trickPhases[domain.UltiPhase] {
	return trickPhases[domain.UltiPhase]{
		play:     domain.UltiPhasePlay,
		trickEnd: domain.UltiPhaseTrickEnd,
		roundEnd: domain.UltiPhaseRoundEnd,
		gameEnd:  domain.UltiPhaseGameEnd,
	}
}

// Reset ゲーム初期化
func (ci *UltiInteractor) Reset() string {
	ci.Game.Reset()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *UltiInteractor) ResetWithConfig(cfg domain.UltiConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.tp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Bid コントラクトを宣言する
func (ci *UltiInteractor) Bid(contract domain.UltiContract, trumpSuit int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerBid(contract, trumpSuit); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Discard タロン受け取り後に 2 枚を捨てる
func (ci *UltiInteractor) Discard(cardIndices []int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerDiscard(cardIndices); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Play カードをプレイ
func (ci *UltiInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(cardIndex); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ci.Game.GetPhase() == domain.UltiPhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ci *UltiInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextRound ディールをスコアリングして次のディールへ進む
func (ci *UltiInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *UltiInteractor) GetConfig() domain.UltiConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *UltiInteractor) Hint() string {
	return ci.tp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *UltiInteractor) ActionLog() string {
	return ci.tp.ActionLogOutput(ci.Game)
}

// advance CPU のプレイを人間の手番もしくはトリック/ラウンド終了になるまで自動実行する。
// 入札は非競争のため CPU の宣言ループは実質即座に抜ける。
func (ci *UltiInteractor) advance() {
	runCpuBidsLoop[domain.UltiPhase](ci.Game, domain.UltiPhaseBid)
	runCpuTurnsLoop(ci.Game, ultiTrickPhases())
}

// RestoreUltiInteractor deserialises JSON into an UltiInteractor.
func RestoreUltiInteractor(data []byte, tp presenter.UltiPresenter) (*UltiInteractor, error) {
	return restoreAndBuild[domain.Ulti](data, func(g *domain.Ulti) *UltiInteractor {
		return &UltiInteractor{GameBase: GameBase[interfaces.UltiGame]{Game: g}, tp: tp}
	})
}
