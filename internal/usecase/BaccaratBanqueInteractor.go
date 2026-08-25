//go:build !js || !wasm || extra2

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BaccaratBanqueInteractorIF はバカラ・バンクのインタラクターインタフェース。
type BaccaratBanqueInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化 (新規ゲーム)
	Reset() string
	// ResetWithConfig 設定を変更してゲームを初期化
	ResetWithConfig(config domain.BaccaratBanqueConfig) string
	// Draw バンカーが 3 枚目を引く
	Draw() string
	// Stand バンカーが 3 枚目を引かない
	Stand() string
	// NextCoup 次のクーへ進む
	NextCoup() string
	// Retire バンクを降りる
	Retire() string
	// GetConfig 現在の設定を返す
	GetConfig() domain.BaccaratBanqueConfig
	// Hint ヒントを出力する
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BaccaratBanqueInteractor はバカラ・バンクのインタラクター。
type BaccaratBanqueInteractor struct {
	GameBase[interfaces.BaccaratBanqueGame]
	bp presenter.BaccaratBanquePresenter
}

// NewBaccaratBanqueInteractor コンストラクタ。
func NewBaccaratBanqueInteractor(g interfaces.BaccaratBanqueGame, bp presenter.BaccaratBanquePresenter) *BaccaratBanqueInteractor {
	mustNotNil("BaccaratBanqueInteractor", map[string]any{"g": g, "bp": bp})
	return &BaccaratBanqueInteractor{
		GameBase: GameBase[interfaces.BaccaratBanqueGame]{Game: g}, bp: bp,
	}
}

// Reset ゲーム初期化 (新規ゲーム)。
func (bi *BaccaratBanqueInteractor) Reset() string {
	bi.Game.Reset()
	return bi.bp.Output(bi.Game, nil)
}

// ResetWithConfig 設定を変更してゲームを初期化。
func (bi *BaccaratBanqueInteractor) ResetWithConfig(config domain.BaccaratBanqueConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.bp, config, bi.Game.SetConfig, bi.Reset)
}

// Draw バンカーが 3 枚目を引く。
func (bi *BaccaratBanqueInteractor) Draw() string { return bi.bankerDraw(true) }

// Stand バンカーが 3 枚目を引かない。
func (bi *BaccaratBanqueInteractor) Stand() string { return bi.bankerDraw(false) }

// bankerDraw は引く / 引かないを反映する。
//
// **引くと止まるは別の入口にする。** 引数ひとつで分けると、既定値のまま
// 届いた要求がどちらかに黙って倒れる。
func (bi *BaccaratBanqueInteractor) bankerDraw(draw bool) string {
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	return bi.bp.Output(bi.Game, bi.Game.BankerDraw(draw))
}

// NextCoup 次のクーへ進む。
//
// **終局と決着以外での呼び出しはドメインが弾く。** BaccaratBanque.NextCoup が
// gameEndFlag とフェーズの両方を見ているので、ここで同じ検査は重ねない。
func (bi *BaccaratBanqueInteractor) NextCoup() string {
	bi.Game.NextCoup()
	return bi.bp.Output(bi.Game, nil)
}

// Retire バンクを降りる。
func (bi *BaccaratBanqueInteractor) Retire() string {
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	return bi.bp.Output(bi.Game, bi.Game.Retire())
}

// GetConfig 現在の設定を返す。
func (bi *BaccaratBanqueInteractor) GetConfig() domain.BaccaratBanqueConfig {
	return bi.Game.GetConfig()
}

// Hint ヒントを出力する。
func (bi *BaccaratBanqueInteractor) Hint() string { return bi.bp.HintOutput(bi.Game) }

// ActionLog 棋譜を出力する。
func (bi *BaccaratBanqueInteractor) ActionLog() string { return bi.bp.ActionLogOutput(bi.Game) }

// RestoreBaccaratBanqueInteractor deserialises JSON into an interactor.
func RestoreBaccaratBanqueInteractor(data []byte, bp presenter.BaccaratBanquePresenter) (*BaccaratBanqueInteractor, error) {
	return restoreAndBuild[domain.BaccaratBanque](data, func(g *domain.BaccaratBanque) *BaccaratBanqueInteractor {
		return &BaccaratBanqueInteractor{
			GameBase: GameBase[interfaces.BaccaratBanqueGame]{Game: g}, bp: bp,
		}
	})
}
