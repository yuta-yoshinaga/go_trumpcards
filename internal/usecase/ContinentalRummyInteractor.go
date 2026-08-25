//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// ContinentalRummyInteractorIF はコンチネンタル・ラミーのインタラクターインタフェース。
type ContinentalRummyInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.ContinentalRummyConfig) string
	// DrawStock 山札から 1 枚引く
	DrawStock() string
	// DrawDiscard 捨て札の一番上を取る
	DrawDiscard() string
	// Discard 手札の i 番を捨てる
	Discard(i int) string
	// GoOut 15 枚を並べて上がる
	GoOut(i int) string
	// NextRound 次のラウンドへ進む
	NextRound() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.ContinentalRummyConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// ContinentalRummyInteractor はコンチネンタル・ラミーのインタラクター。
type ContinentalRummyInteractor struct {
	GameBase[interfaces.ContinentalRummyGame]
	cp presenter.ContinentalRummyPresenter
}

// NewContinentalRummyInteractor コンストラクタ。
func NewContinentalRummyInteractor(g interfaces.ContinentalRummyGame, cp presenter.ContinentalRummyPresenter) *ContinentalRummyInteractor {
	mustNotNil("ContinentalRummyInteractor", map[string]any{"g": g, "cp": cp})
	return &ContinentalRummyInteractor{
		GameBase: GameBase[interfaces.ContinentalRummyGame]{Game: g}, cp: cp,
	}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (ci *ContinentalRummyInteractor) Reset() string {
	ci.Game.Reset()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (ci *ContinentalRummyInteractor) ResetWithConfig(config domain.ContinentalRummyConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, config, ci.Game.SetConfig, ci.Reset)
}

// DrawStock 山札から 1 枚引く。
//
// **山と捨て札は別の入口にする。** 引数ひとつで分けると、既定値のまま届いた
// 要求がどちらかに黙って倒れる。
func (ci *ContinentalRummyInteractor) DrawStock() string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	return ci.cp.Output(ci.Game, ci.Game.DrawStock())
}

// DrawDiscard 捨て札の一番上を取る。
func (ci *ContinentalRummyInteractor) DrawDiscard() string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	return ci.cp.Output(ci.Game, ci.Game.DrawDiscard())
}

// Discard 手札の i 番を捨てる。
func (ci *ContinentalRummyInteractor) Discard(i int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	return ci.cp.Output(ci.Game, ci.Game.Discard(i))
}

// GoOut 15 枚を並べて上がる。
func (ci *ContinentalRummyInteractor) GoOut(i int) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	return ci.cp.Output(ci.Game, ci.Game.GoOut(i))
}

// NextRound 次のラウンドへ進む。
//
// **ドメインが終局とフェーズの両方を見ている**ので、ここで同じ検査は重ねない。
func (ci *ContinentalRummyInteractor) NextRound() string {
	ci.Game.NextRound()
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を返す。
func (ci *ContinentalRummyInteractor) GetConfig() domain.ContinentalRummyConfig {
	return ci.Game.GetConfig()
}

// Hint ヒントを出力する。
func (ci *ContinentalRummyInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する。
func (ci *ContinentalRummyInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// RestoreContinentalRummyInteractor deserialises JSON into an interactor.
func RestoreContinentalRummyInteractor(data []byte, cp presenter.ContinentalRummyPresenter) (*ContinentalRummyInteractor, error) {
	return restoreAndBuild[domain.ContinentalRummy](data, func(g *domain.ContinentalRummy) *ContinentalRummyInteractor {
		return &ContinentalRummyInteractor{
			GameBase: GameBase[interfaces.ContinentalRummyGame]{Game: g}, cp: cp,
		}
	})
}
