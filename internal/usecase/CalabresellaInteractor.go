//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// CalabresellaInteractorIF カラブレセッラ (Calabresella) のインタラクターインタフェース
type CalabresellaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.CalabresellaConfig) string
	// Bid ビッドする (pass/chiamo/solo)
	Bid(bid domain.CalabresellaBid) string
	// Discard monte 交換で1枚を捨てる
	Discard(cardIndex int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.CalabresellaConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// CalabresellaInteractor カラブレセッラのインタラクタークラス
type CalabresellaInteractor struct {
	GameBase[interfaces.CalabresellaGame]
	tp presenter.CalabresellaPresenter
}

// NewCalabresellaInteractor コンストラクタ
func NewCalabresellaInteractor(g interfaces.CalabresellaGame, tp presenter.CalabresellaPresenter) *CalabresellaInteractor {
	mustNotNil("CalabresellaInteractor", map[string]any{"g": g, "tp": tp})
	return &CalabresellaInteractor{GameBase: GameBase[interfaces.CalabresellaGame]{Game: g}, tp: tp}
}

// calabresellaTrickPhases Calabresella のトリックフェーズ定数
func calabresellaTrickPhases() trickPhases[domain.CalabresellaPhase] {
	return trickPhases[domain.CalabresellaPhase]{
		play:     domain.CalabresellaPhasePlay,
		trickEnd: domain.CalabresellaPhaseTrickEnd,
		roundEnd: domain.CalabresellaPhaseRoundEnd,
		gameEnd:  domain.CalabresellaPhaseGameEnd,
	}
}

// Reset ゲーム初期化
func (ci *CalabresellaInteractor) Reset() string {
	ci.Game.Reset()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *CalabresellaInteractor) ResetWithConfig(cfg domain.CalabresellaConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.tp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Bid ビッドする
func (ci *CalabresellaInteractor) Bid(bid domain.CalabresellaBid) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerBid(bid); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Discard monte 交換で1枚を捨てる
func (ci *CalabresellaInteractor) Discard(cardIndex int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerDiscard(cardIndex); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// Play カードをプレイ
func (ci *CalabresellaInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(ci.Game, ci.tp); blocked {
		return out
	}
	if err := ci.Game.PlayerPlay(cardIndex); err != nil {
		return ci.tp.Output(ci.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if ci.Game.GetPhase() == domain.CalabresellaPhaseTrickEnd {
		ci.Game.ResolveTrick()
	}
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextTrick 次のトリックへ進む
func (ci *CalabresellaInteractor) NextTrick() string {
	ci.Game.NextTrick()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (ci *CalabresellaInteractor) NextRound() string {
	ci.Game.ScoreRound()
	if out, blocked := guardGameEnd(ci.Game, ci.tp); blocked {
		return out
	}
	ci.Game.NextRound()
	ci.advance()
	return ci.tp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *CalabresellaInteractor) GetConfig() domain.CalabresellaConfig {
	return ci.Game.GetConfig()
}

// Hint ヒント取得
func (ci *CalabresellaInteractor) Hint() string {
	return ci.tp.HintOutput(ci.Game)
}

// ActionLog 棋譜を出力する
func (ci *CalabresellaInteractor) ActionLog() string {
	return ci.tp.ActionLogOutput(ci.Game)
}

// advance CPU のビッド・プレイを人間の手番もしくはトリック/ラウンド終了になるまで自動実行する。
// まずビッドフェーズの CPU を消化し (CPU ソリストの monte 交換は domain 側で自動)、続いて
// プレイフェーズの CPU ターンを共通ループで処理する。
func (ci *CalabresellaInteractor) advance() {
	runCpuBidsLoop[domain.CalabresellaPhase](ci.Game, domain.CalabresellaPhaseBid)
	runCpuTurnsLoop(ci.Game, calabresellaTrickPhases())
}

// RestoreCalabresellaInteractor deserialises JSON into a CalabresellaInteractor.
func RestoreCalabresellaInteractor(data []byte, tp presenter.CalabresellaPresenter) (*CalabresellaInteractor, error) {
	return restoreAndBuild[domain.Calabresella](data, func(g *domain.Calabresella) *CalabresellaInteractor {
		return &CalabresellaInteractor{GameBase: GameBase[interfaces.CalabresellaGame]{Game: g}, tp: tp}
	})
}
