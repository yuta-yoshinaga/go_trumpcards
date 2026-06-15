//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// FortyFivesInteractorIF オークション・フォーティファイブズのインタラクターインタフェース
type FortyFivesInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.FortyFivesConfig) string
	// Bid 入札する (0=Pass,15,20,25)
	Bid(bid int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextTrick 次のトリックへ進む
	NextTrick() string
	// NextRound ラウンドをスコアリングして次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.FortyFivesConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// FortyFivesInteractor オークション・フォーティファイブズのインタラクタークラス
type FortyFivesInteractor struct {
	GameBase[interfaces.FortyFivesGame]
	sp presenter.FortyFivesPresenter
}

// NewFortyFivesInteractor コンストラクタ
func NewFortyFivesInteractor(g interfaces.FortyFivesGame, sp presenter.FortyFivesPresenter) *FortyFivesInteractor {
	mustNotNil("FortyFivesInteractor", map[string]any{"g": g, "sp": sp})
	return &FortyFivesInteractor{GameBase: GameBase[interfaces.FortyFivesGame]{Game: g}, sp: sp}
}

// Reset ゲーム初期化
func (fi *FortyFivesInteractor) Reset() string {
	fi.Game.Reset()
	fi.advanceCpu()
	return fi.sp.Output(fi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (fi *FortyFivesInteractor) ResetWithConfig(cfg domain.FortyFivesConfig) string {
	return resetWithValidatedConfig(fi.Game, fi.sp, cfg, fi.Game.SetConfig, fi.Reset)
}

// Bid 入札する (0=Pass,15,20,25)
func (fi *FortyFivesInteractor) Bid(bid int) string {
	if out, blocked := guardGameEnd(fi.Game, fi.sp); blocked {
		return out
	}
	if err := fi.Game.PlayerBid(domain.FortyFivesBid(bid)); err != nil {
		return fi.sp.Output(fi.Game, err)
	}
	fi.advanceCpu()
	return fi.sp.Output(fi.Game, nil)
}

// Play カードをプレイ
func (fi *FortyFivesInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(fi.Game, fi.sp); blocked {
		return out
	}
	if err := fi.Game.PlayerPlay(cardIndex); err != nil {
		return fi.sp.Output(fi.Game, err)
	}
	// 人間が最後のカードを出してトリック完了した場合、即座に解決する。
	if fi.Game.GetPhase() == domain.FortyFivesPhaseTrickEnd {
		fi.Game.ResolveTrick()
	}
	fi.advanceCpu()
	return fi.sp.Output(fi.Game, nil)
}

// NextTrick 次のトリックへ進む
func (fi *FortyFivesInteractor) NextTrick() string {
	fi.Game.NextTrick()
	fi.advanceCpu()
	return fi.sp.Output(fi.Game, nil)
}

// NextRound ラウンドをスコアリングして次のラウンドへ進む
func (fi *FortyFivesInteractor) NextRound() string {
	fi.Game.ScoreRound()
	if out, blocked := guardGameEnd(fi.Game, fi.sp); blocked {
		return out
	}
	fi.Game.NextRound()
	fi.advanceCpu()
	return fi.sp.Output(fi.Game, nil)
}

// GetConfig 現在の設定を取得
func (fi *FortyFivesInteractor) GetConfig() domain.FortyFivesConfig {
	return fi.Game.GetConfig()
}

// Hint ヒント取得
func (fi *FortyFivesInteractor) Hint() string {
	return fi.sp.HintOutput(fi.Game)
}

// ActionLog 棋譜を出力する
func (fi *FortyFivesInteractor) ActionLog() string {
	return fi.sp.ActionLogOutput(fi.Game)
}

// advanceCpu 入札 → プレイ の順に CPU を自動進行させる。
// 入札フェーズでは人間の手番になるまで CPU に入札させ、入札が締まって
// プレイフェーズに入ったら人間の手番までトリックを進める。
func (fi *FortyFivesInteractor) advanceCpu() {
	runCpuBidsLoop(fi.Game, domain.FortyFivesPhaseBid)
	fi.runCpuTurns()
}

// runCpuTurns プレイフェーズで CPU ターンを自動実行する。
func (fi *FortyFivesInteractor) runCpuTurns() {
	runCpuTurnsLoop(fi.Game, trickPhases[domain.FortyFivesPhase]{
		play:     domain.FortyFivesPhasePlay,
		trickEnd: domain.FortyFivesPhaseTrickEnd,
		roundEnd: domain.FortyFivesPhaseRoundEnd,
		gameEnd:  domain.FortyFivesPhaseGameEnd,
	})
}

// RestoreFortyFivesInteractor deserialises JSON into a FortyFivesInteractor.
func RestoreFortyFivesInteractor(data []byte, sp presenter.FortyFivesPresenter) (*FortyFivesInteractor, error) {
	return restoreAndBuild[domain.FortyFives](data, func(g *domain.FortyFives) *FortyFivesInteractor {
		return &FortyFivesInteractor{GameBase: GameBase[interfaces.FortyFivesGame]{Game: g}, sp: sp}
	})
}
