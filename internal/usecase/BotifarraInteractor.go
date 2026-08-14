//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// BotifarraInteractorIF ボティファラインタラクターインタフェース
type BotifarraInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.BotifarraConfig) string
	// Declare 切り札を宣言する
	Declare(suit int) string
	// Delegate 宣言を相方に委ねる
	Delegate() string
	// Double 倍付けを宣言する
	Double() string
	// PassDouble 倍付けを見送る
	PassDouble() string
	// PlayCard 札を出す
	PlayCard(cardIndex int) string
	// NextRound 次のラウンドを配る
	NextRound() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.BotifarraConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// BotifarraInteractor ボティファラインタラクタークラス
type BotifarraInteractor struct {
	GameBase[interfaces.BotifarraGame]
	bp presenter.BotifarraPresenter
}

// NewBotifarraInteractor コンストラクタ
func NewBotifarraInteractor(b interfaces.BotifarraGame, bp presenter.BotifarraPresenter) *BotifarraInteractor {
	mustNotNil("BotifarraInteractor", map[string]any{"b": b, "bp": bp})
	return &BotifarraInteractor{GameBase: GameBase[interfaces.BotifarraGame]{Game: b}, bp: bp}
}

// Reset ゲーム初期化
func (bi *BotifarraInteractor) Reset() string {
	bi.Game.Reset()
	return bi.bp.Output(bi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (bi *BotifarraInteractor) ResetWithConfig(cfg domain.BotifarraConfig) string {
	return resetWithValidatedConfig(bi.Game, bi.bp, cfg, bi.Game.SetConfig, bi.Reset)
}

// Declare 切り札を宣言する
func (bi *BotifarraInteractor) Declare(suit int) string {
	return bi.runGuarded(func() error { return bi.Game.Declare(suit) })
}

// Delegate 宣言を相方に委ねる
func (bi *BotifarraInteractor) Delegate() string {
	return bi.runGuarded(bi.Game.Delegate)
}

// Double 倍付けを宣言する
func (bi *BotifarraInteractor) Double() string {
	return bi.runGuarded(bi.Game.Double)
}

// PassDouble 倍付けを見送る
func (bi *BotifarraInteractor) PassDouble() string {
	return bi.runGuarded(bi.Game.PassDouble)
}

// PlayCard 札を出す
func (bi *BotifarraInteractor) PlayCard(cardIndex int) string {
	return bi.runGuarded(func() error { return bi.Game.PlayCard(cardIndex) })
}

// NextRound 次のラウンドを配る
func (bi *BotifarraInteractor) NextRound() string {
	return bi.runGuarded(bi.Game.NextRound)
}

// runGuarded は終局後の操作を弾いてから action を実行し、結果を出力する。
//
// **ドメインの各メソッドが自分でフェーズを見ている**ので、ここでは終局だけを
// 見ます。二重にフェーズ判定を書くと、片方だけ直したときにずれます。
func (bi *BotifarraInteractor) runGuarded(action func() error) string {
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	if err := action(); err != nil {
		return bi.bp.Output(bi.Game, err)
	}
	return bi.bp.Output(bi.Game, nil)
}

// GiveUp 投了する
func (bi *BotifarraInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(bi.Game, bi.bp); blocked {
		return out
	}
	bi.Game.GiveUp()
	return bi.bp.Output(bi.Game, nil)
}

// GetConfig 現在の設定を取得
func (bi *BotifarraInteractor) GetConfig() domain.BotifarraConfig { return bi.Game.GetConfig() }

// Hint ヒント取得
func (bi *BotifarraInteractor) Hint() string { return bi.bp.HintOutput(bi.Game) }

// ActionLog 棋譜を出力する
func (bi *BotifarraInteractor) ActionLog() string { return bi.bp.ActionLogOutput(bi.Game) }

// RestoreBotifarraInteractor deserialises JSON into a BotifarraInteractor.
func RestoreBotifarraInteractor(data []byte, bp presenter.BotifarraPresenter) (*BotifarraInteractor, error) {
	return restoreAndBuild[domain.Botifarra](data, func(g *domain.Botifarra) *BotifarraInteractor {
		return NewBotifarraInteractor(g, bp)
	})
}
