//go:build !js || !wasm || extra4

package usecase

import (
	"errors"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ChemindeFerInteractorIF シュマン・ド・フェールインタラクターインタフェース
type ChemindeFerInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.ChemindeFerConfig) string
	// SetStake 親がバンク額を張る
	SetStake(amount int) string
	// PlaceBet 子が賭ける
	PlaceBet(seatIdx, amount int) string
	// PunterDraw 子側が 3 枚目を引く
	PunterDraw() string
	// PunterStand 子側が立つ
	PunterStand() string
	// BankerDraw 親が 3 枚目を引く
	BankerDraw() string
	// BankerStand 親が立つ
	BankerStand() string
	// DrawOrStand 手番の側へ引き/立ちを送る
	DrawOrStand(draw bool) string
	// PassBank 親が自分から親を降りる
	PassBank() string
	// NextRound 次のラウンドを始める
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.ChemindeFerConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ChemindeFerInteractor シュマン・ド・フェールインタラクタークラス
type ChemindeFerInteractor struct {
	GameBase[interfaces.ChemindeFerGame]
	cp presenter.ChemindeFerPresenter
}

// NewChemindeFerInteractor コンストラクタ
func NewChemindeFerInteractor(c interfaces.ChemindeFerGame, cp presenter.ChemindeFerPresenter) *ChemindeFerInteractor {
	mustNotNil("ChemindeFerInteractor", map[string]any{"c": c, "cp": cp})
	return &ChemindeFerInteractor{GameBase: GameBase[interfaces.ChemindeFerGame]{Game: c}, cp: cp}
}

// Reset ゲーム初期化
func (ci *ChemindeFerInteractor) Reset() string {
	ci.Game.Reset()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *ChemindeFerInteractor) ResetWithConfig(cfg domain.ChemindeFerConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, cfg, ci.Game.SetConfig, ci.Reset)
}

// SetStake 親がバンク額を張る
func (ci *ChemindeFerInteractor) SetStake(amount int) string {
	return ci.runGuarded(func() error { return ci.Game.SetStake(amount) })
}

// PlaceBet 子が賭ける
func (ci *ChemindeFerInteractor) PlaceBet(seatIdx, amount int) string {
	return ci.runGuarded(func() error { return ci.Game.PlaceBet(seatIdx, amount) })
}

// PunterDraw 子側が 3 枚目を引く
func (ci *ChemindeFerInteractor) PunterDraw() string {
	return ci.runGuarded(ci.Game.PunterDraw)
}

// PunterStand 子側が立つ
func (ci *ChemindeFerInteractor) PunterStand() string {
	return ci.runGuarded(ci.Game.PunterStand)
}

// BankerDraw 親が 3 枚目を引く
func (ci *ChemindeFerInteractor) BankerDraw() string {
	return ci.runGuarded(ci.Game.BankerDraw)
}

// BankerStand 親が立つ
func (ci *ChemindeFerInteractor) BankerStand() string {
	return ci.runGuarded(ci.Game.BankerStand)
}

// ErrChemindeFerNotDrawPhase は引くか立つかを決める場面ではないときのエラー。
var ErrChemindeFerNotDrawPhase = errors.New("chemindefer: nobody is deciding on a third card")

// DrawOrStand は手番の側へ引き/立ちを送る。
//
// **どちらの側かはドメインのフェーズが決める。** CUI に側を書かせると、画面側が
// フェーズをもう 1 つ持つことになってドメインとずれる。ここは規則を作り直しているの
// ではなく、ドメインが公開している現在のフェーズをそのまま読んで振り分けているだけ。
func (ci *ChemindeFerInteractor) DrawOrStand(draw bool) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	var action func() error
	switch ci.Game.GetPhase() {
	case domain.ChemindeFerPhasePunterDraw:
		action = ci.Game.PunterStand
		if draw {
			action = ci.Game.PunterDraw
		}
	case domain.ChemindeFerPhaseBankerDraw:
		action = ci.Game.BankerStand
		if draw {
			action = ci.Game.BankerDraw
		}
	default:
		return ci.cp.Output(ci.Game, ErrChemindeFerNotDrawPhase)
	}
	if err := action(); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	return ci.cp.Output(ci.Game, nil)
}

// PassBank 親が自分から親を降りる
func (ci *ChemindeFerInteractor) PassBank() string {
	return ci.runGuarded(ci.Game.PassBank)
}

// NextRound 次のラウンドを始める
func (ci *ChemindeFerInteractor) NextRound() string {
	return ci.runGuarded(ci.Game.NextRound)
}

// runGuarded は終局後の操作を弾いてから action を実行し、結果を出力する。
//
// **フェーズ判定はドメインに任せます。** 二重に書くとずれます。張り・賭け・引きの
// どれが今できるかはドメインだけが知っていて、ここが知っているのは「終局したか」だけ。
func (ci *ChemindeFerInteractor) runGuarded(action func() error) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	if err := action(); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	return ci.cp.Output(ci.Game, nil)
}

// GiveUp 投了する
func (ci *ChemindeFerInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	ci.Game.GiveUp()
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *ChemindeFerInteractor) GetConfig() domain.ChemindeFerConfig { return ci.Game.GetConfig() }

// Hint ヒント取得
func (ci *ChemindeFerInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する
func (ci *ChemindeFerInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// RestoreChemindeFerInteractor deserialises JSON into a ChemindeFerInteractor.
func RestoreChemindeFerInteractor(data []byte, cp presenter.ChemindeFerPresenter) (*ChemindeFerInteractor, error) {
	return restoreAndBuild[domain.ChemindeFer](data, func(g *domain.ChemindeFer) *ChemindeFerInteractor {
		return NewChemindeFerInteractor(g, cp)
	})
}
