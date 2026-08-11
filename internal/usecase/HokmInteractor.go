//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// HokmInteractorIF ホクムインタラクターインタフェース
type HokmInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化
	ResetWithConfig(cfg domain.HokmConfig) string
	// DeclareTrump 切り札スートを宣言する
	DeclareTrump(suit int) string
	// Play カードをプレイ
	Play(cardIndex int) string
	// NextHand 次のハンドへ進む
	NextHand() string
	// GiveUp 投了する
	GiveUp() string
	// GetConfig 現在の設定を取得
	GetConfig() domain.HokmConfig
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// HokmInteractor ホクムインタラクタークラス
type HokmInteractor struct {
	GameBase[interfaces.HokmGame]
	hp presenter.HokmPresenter
}

// NewHokmInteractor コンストラクタ
func NewHokmInteractor(h interfaces.HokmGame, hp presenter.HokmPresenter) *HokmInteractor {
	mustNotNil("HokmInteractor", map[string]any{"h": h, "hp": hp})
	return &HokmInteractor{GameBase: GameBase[interfaces.HokmGame]{Game: h}, hp: hp}
}

// Reset ゲーム初期化。配り終えたら人間の番まで進める。
//
// **宣言だけでなくプレイも進める。** 親が CPU なら切り札はその場で決まるので、
// ここで打たせないとリード（親）のまま止まった盤面を返してしまう。
func (hi *HokmInteractor) Reset() string {
	hi.Game.Reset()
	hi.advanceToHuman()
	return hi.hp.Output(hi.Game, nil)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (hi *HokmInteractor) ResetWithConfig(cfg domain.HokmConfig) string {
	return resetWithValidatedConfig(hi.Game, hi.hp, cfg, hi.Game.SetConfig, hi.Reset)
}

// DeclareTrump 切り札スートを宣言する
func (hi *HokmInteractor) DeclareTrump(suit int) string {
	if out, blocked := guardGameEnd(hi.Game, hi.hp); blocked {
		return out
	}
	if err := hi.Game.PlayerDeclareTrump(suit); err != nil {
		return hi.hp.Output(hi.Game, err)
	}
	hi.runCpuTurns()
	return hi.hp.Output(hi.Game, nil)
}

// Play カードをプレイ
func (hi *HokmInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(hi.Game, hi.hp); blocked {
		return out
	}
	if err := hi.Game.PlayerPlay(cardIndex); err != nil {
		return hi.hp.Output(hi.Game, err)
	}
	hi.runCpuTurns()
	return hi.hp.Output(hi.Game, nil)
}

// NextHand 次のハンドへ進む
func (hi *HokmInteractor) NextHand() string {
	if out, blocked := guardGameEnd(hi.Game, hi.hp); blocked {
		return out
	}
	hi.Game.NextHand()
	// 次のハンドも切り札の宣言から始まるので、人間の番まで進める。
	hi.advanceToHuman()
	return hi.hp.Output(hi.Game, nil)
}

// GiveUp 投了する
func (hi *HokmInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(hi.Game, hi.hp); blocked {
		return out
	}
	hi.Game.GiveUp()
	return hi.hp.Output(hi.Game, nil)
}

// GetConfig 現在の設定を取得
func (hi *HokmInteractor) GetConfig() domain.HokmConfig { return hi.Game.GetConfig() }

// Hint ヒント取得
func (hi *HokmInteractor) Hint() string { return hi.hp.HintOutput(hi.Game) }

// ActionLog 棋譜を出力する
func (hi *HokmInteractor) ActionLog() string { return hi.hp.ActionLogOutput(hi.Game) }

// advanceToHuman 宣言 → プレイ の順に、人間の番まで CPU を進める
func (hi *HokmInteractor) advanceToHuman() {
	hi.runCpuTrump()
	hi.runCpuTurns()
}

// runCpuTrump CPU の親なら切り札を宣言させる
func (hi *HokmInteractor) runCpuTrump() {
	if hi.Game.GetGameEndFlag() {
		return
	}
	if hi.Game.GetPhase() != domain.HokmPhaseTrump || hi.Game.IsHumanTrumpTurn() {
		return
	}
	hi.Game.CpuDeclareTrump()
}

// runCpuTurns 人間の手番になるかハンド／ゲームが終わるまで CPU を進める
func (hi *HokmInteractor) runCpuTurns() {
	for turns := 0; !hi.Game.GetGameEndFlag() && !hi.Game.IsHumanTurn(); turns++ {
		if turns >= maxCpuTurnsPerCall {
			return
		}
		if hi.Game.GetPhase() != domain.HokmPhasePlay {
			return
		}
		hi.Game.CpuPlay()
	}
}

// RestoreHokmInteractor deserialises JSON into a HokmInteractor.
func RestoreHokmInteractor(data []byte, hp presenter.HokmPresenter) (*HokmInteractor, error) {
	return restoreAndBuild[domain.Hokm](data, func(g *domain.Hokm) *HokmInteractor {
		return &HokmInteractor{GameBase: GameBase[interfaces.HokmGame]{Game: g}, hp: hp}
	})
}
