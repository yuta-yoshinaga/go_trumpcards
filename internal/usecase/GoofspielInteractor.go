//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GoofspielInteractorIF ゴフスピールインタラクターインタフェース
type GoofspielInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.GoofspielConfig) string
	// Bid 入札札を伏せる
	Bid(cardIndex int) string
	// NextRound 次の賞札をめくる
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.GoofspielConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// GoofspielInteractor ゴフスピールインタラクタークラス
type GoofspielInteractor struct {
	GameBase[interfaces.GoofspielGame]
	sp presenter.GoofspielPresenter
}

// NewGoofspielInteractor コンストラクタ
func NewGoofspielInteractor(s interfaces.GoofspielGame, sp presenter.GoofspielPresenter) *GoofspielInteractor {
	mustNotNil("GoofspielInteractor", map[string]any{"s": s, "sp": sp})
	return &GoofspielInteractor{GameBase: GameBase[interfaces.GoofspielGame]{Game: s}, sp: sp}
}

// Reset ゲーム初期化
func (gi *GoofspielInteractor) Reset() string {
	gi.Game.Reset()
	return gi.sp.Output(gi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (gi *GoofspielInteractor) ResetWithConfig(cfg domain.GoofspielConfig) string {
	return resetWithValidatedConfig(gi.Game, gi.sp, cfg, gi.Game.SetConfig, gi.Reset)
}

// Bid 入札札を伏せる
//
// **CPU を先に動かしません。** 入札は同時なので、人間が伏せた時点で CPU も伏せ、
// ドメイン側が一斉に公開します。先に CPU を動かすと、公開前に相手の札が
// 出力に載ってしまいます。
func (gi *GoofspielInteractor) Bid(cardIndex int) string {
	if out, blocked := guardGameEnd(gi.Game, gi.sp); blocked {
		return out
	}
	if err := gi.Game.PlayerBid(cardIndex); err != nil {
		return gi.sp.Output(gi.Game, err)
	}
	return gi.sp.Output(gi.Game, nil)
}

// NextRound 次の賞札をめくる
func (gi *GoofspielInteractor) NextRound() string {
	if out, blocked := guardGameEnd(gi.Game, gi.sp); blocked {
		return out
	}
	if err := gi.Game.NextRound(); err != nil {
		return gi.sp.Output(gi.Game, err)
	}
	return gi.sp.Output(gi.Game, nil)
}

// GiveUp 投了する
func (gi *GoofspielInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(gi.Game, gi.sp); blocked {
		return out
	}
	gi.Game.GiveUp()
	return gi.sp.Output(gi.Game, nil)
}

// GetConfig 現在の設定を取得
func (gi *GoofspielInteractor) GetConfig() domain.GoofspielConfig { return gi.Game.GetConfig() }

// Hint ヒント取得
func (gi *GoofspielInteractor) Hint() string { return gi.sp.HintOutput(gi.Game) }

// ActionLog 棋譜を出力する
func (gi *GoofspielInteractor) ActionLog() string { return gi.sp.ActionLogOutput(gi.Game) }

// RestoreGoofspielInteractor deserialises JSON into a GoofspielInteractor.
func RestoreGoofspielInteractor(data []byte, sp presenter.GoofspielPresenter) (*GoofspielInteractor, error) {
	return restoreAndBuild[domain.Goofspiel](data, func(g *domain.Goofspiel) *GoofspielInteractor {
		return NewGoofspielInteractor(g, sp)
	})
}
