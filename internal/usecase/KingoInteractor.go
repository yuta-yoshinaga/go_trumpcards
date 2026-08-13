//go:build !js || !wasm || extra

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// KingoInteractorIF キンゴインタラクターインタフェース
type KingoInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.KingoConfig) string
	// Bet 子として張る
	Bet(amount int) string
	// Deal 親として配る
	Deal() string
	// NextRound 次のラウンドを始める
	NextRound() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.KingoConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// KingoInteractor キンゴインタラクタークラス
type KingoInteractor struct {
	GameBase[interfaces.KingoGame]
	cp presenter.KingoPresenter
}

// NewKingoInteractor コンストラクタ
func NewKingoInteractor(c interfaces.KingoGame, cp presenter.KingoPresenter) *KingoInteractor {
	mustNotNil("KingoInteractor", map[string]any{"c": c, "cp": cp})
	return &KingoInteractor{GameBase: GameBase[interfaces.KingoGame]{Game: c}, cp: cp}
}

// Reset ゲーム初期化
func (ci *KingoInteractor) Reset() string {
	ci.Game.Reset()
	return ci.cp.Output(ci.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ci *KingoInteractor) ResetWithConfig(cfg domain.KingoConfig) string {
	return resetWithValidatedConfig(ci.Game, ci.cp, cfg, ci.Game.SetConfig, ci.Reset)
}

// Bet は子としての張りを処理する。
//
// **額はそのまま通す。** 手札は配る前なので、額を「良かれと思って」丸めると
// 見えていない情報を根拠にしたことになる ── 範囲外はドメインが弾く。
func (ci *KingoInteractor) Bet(amount int) string {
	return ci.runGuarded(func() error { return ci.Game.PlaceBet(amount) })
}

// Deal は親として配る。
func (ci *KingoInteractor) Deal() string {
	return ci.runGuarded(ci.Game.Deal)
}

// NextRound 次のラウンドを始める
func (ci *KingoInteractor) NextRound() string { return ci.runGuarded(ci.Game.NextRound) }

// runGuarded は終局後の操作を弾いてから action を実行して出力する。
//
// **CPU を進める呼び出しは要らない。** 子の張りは配る前に出そろい、決着は
// その場で終わるので、人間の 1 手のあとに動く CPU がいない。
func (ci *KingoInteractor) runGuarded(action func() error) string {
	if out, blocked := guardGameEnd(ci.Game, ci.cp); blocked {
		return out
	}
	if err := action(); err != nil {
		return ci.cp.Output(ci.Game, err)
	}
	return ci.cp.Output(ci.Game, nil)
}

// GetConfig 現在の設定を取得
func (ci *KingoInteractor) GetConfig() domain.KingoConfig { return ci.Game.GetConfig() }

// Hint ヒント取得
func (ci *KingoInteractor) Hint() string { return ci.cp.HintOutput(ci.Game) }

// ActionLog 棋譜を出力する
func (ci *KingoInteractor) ActionLog() string { return ci.cp.ActionLogOutput(ci.Game) }

// RestoreKingoInteractor deserialises JSON into an interactor.
func RestoreKingoInteractor(data []byte, cp presenter.KingoPresenter) (*KingoInteractor, error) {
	return restoreAndBuild[domain.Kingo](data,
		func(g *domain.Kingo) *KingoInteractor { return NewKingoInteractor(g, cp) })
}
