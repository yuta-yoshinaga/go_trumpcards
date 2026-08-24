//go:build !js || !wasm || casino

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// HorseInteractorIF は H.O.R.S.E. のインタラクターインタフェース。
type HorseInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.HorseConfig) string
	// Action 人間の手をいまの種目へ渡す
	Action(action, amount, humanPlayMs int) string
	// Exchange 引き直しをいまの種目へ渡す (ドロー系の種目のみ)
	Exchange(indices []int) string
	// NextHand 次のハンドへ進む
	NextHand() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.HorseConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// HorseInteractor は H.O.R.S.E. のインタラクター。
type HorseInteractor struct {
	GameBase[interfaces.HorseGame]
	hp presenter.HorsePresenter
}

// NewHorseInteractor コンストラクタ。
func NewHorseInteractor(g interfaces.HorseGame, hp presenter.HorsePresenter) *HorseInteractor {
	mustNotNil("HorseInteractor", map[string]any{"g": g, "hp": hp})
	return &HorseInteractor{GameBase: GameBase[interfaces.HorseGame]{Game: g}, hp: hp}
}

// Reset ゲーム初期化。
func (hi *HorseInteractor) Reset() string {
	hi.Game.Reset()
	return hi.hp.Output(hi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化。
//
// **バリアントは要求では変えられない。** どのローテーションを回すかは
// 「いま遊んでいるゲーム」そのもので、設定の 1 項目ではない ── 要求から
// 受け取ると、Eight-Game Mix の卓に H.O.R.S.E. の設定を送るだけで 8 種目が
// 5 種目に化ける。
//
// **席数はここで丸めない。** バリアントに無い席数は `Validate` が弾き、卓は
// いまの設定のまま残る ── 黙って別の席数に直すと、要求した卓とは違う卓が
// 「成功」として返る。
func (hi *HorseInteractor) ResetWithConfig(cfg domain.HorseConfig) string {
	cfg.Variant = hi.Game.GetConfig().Variant
	return resetWithValidatedConfig(hi.Game, hi.hp, cfg, hi.Game.SetConfig, hi.Reset)
}

// Action 人間の手をいまの種目へ渡す。
//
// **CPU を回す呼び出しは要らない。** 種目の `PlayerAction` が内部で CPU の手番を
// 回し切り、人間の手番かハンドの決着で戻ってくる。
func (hi *HorseInteractor) Action(action, amount, humanPlayMs int) string {
	if out, blocked := guardGameEnd(hi.Game, hi.hp); blocked {
		return out
	}
	if err := hi.Game.PlayerAction(action, amount, humanPlayMs); err != nil {
		return hi.hp.Output(hi.Game, err)
	}
	return hi.hp.Output(hi.Game, nil)
}

// Exchange 引き直しをいまの種目へ渡す。
//
// **空のスライスはスタンドパット。** 「引かない」も 1 つの手なので、
// 何も選ばない呼び出しを弾いてはいけない。
func (hi *HorseInteractor) Exchange(indices []int) string {
	if out, blocked := guardGameEnd(hi.Game, hi.hp); blocked {
		return out
	}
	if err := hi.Game.PlayerExchange(indices); err != nil {
		return hi.hp.Output(hi.Game, err)
	}
	return hi.hp.Output(hi.Game, nil)
}

// NextHand 次のハンドへ進む。
func (hi *HorseInteractor) NextHand() string {
	if out, blocked := guardGameEnd(hi.Game, hi.hp); blocked {
		return out
	}
	if err := hi.Game.NextHand(); err != nil {
		return hi.hp.Output(hi.Game, err)
	}
	return hi.hp.Output(hi.Game, nil)
}

// GetConfig 現在の設定を取得。
func (hi *HorseInteractor) GetConfig() domain.HorseConfig { return hi.Game.GetConfig() }

// Hint ヒント取得。
func (hi *HorseInteractor) Hint() string { return hi.hp.HintOutput(hi.Game) }

// ActionLog 棋譜を出力する。
func (hi *HorseInteractor) ActionLog() string { return hi.hp.ActionLogOutput(hi.Game) }

// RestoreHorseInteractor deserialises JSON into a HorseInteractor.
func RestoreHorseInteractor(data []byte, hp presenter.HorsePresenter) (*HorseInteractor, error) {
	return restoreAndBuild[domain.Horse](data, func(g *domain.Horse) *HorseInteractor {
		return &HorseInteractor{GameBase: GameBase[interfaces.HorseGame]{Game: g}, hp: hp}
	})
}
