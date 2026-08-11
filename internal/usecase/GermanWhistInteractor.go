//go:build !js || !wasm || classic

package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// GermanWhistInteractorIF ジャーマンホイストインタラクターインタフェース
type GermanWhistInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	// Reset ゲーム初期化
	Reset() string
	// Play カードをプレイ
	Play(cardIndex int) string
	// GiveUp 投了する
	GiveUp() string
	// Hint ヒント取得
	Hint() string
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// GermanWhistInteractor ジャーマンホイストインタラクタークラス
type GermanWhistInteractor struct {
	GameBase[interfaces.GermanWhistGame]
	gp presenter.GermanWhistPresenter
}

// NewGermanWhistInteractor コンストラクタ
func NewGermanWhistInteractor(g interfaces.GermanWhistGame, gp presenter.GermanWhistPresenter) *GermanWhistInteractor {
	mustNotNil("GermanWhistInteractor", map[string]any{"g": g, "gp": gp})
	return &GermanWhistInteractor{GameBase: GameBase[interfaces.GermanWhistGame]{Game: g}, gp: gp}
}

// Reset ゲーム初期化
func (gi *GermanWhistInteractor) Reset() string {
	gi.Game.Reset()
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// Play カードをプレイ
func (gi *GermanWhistInteractor) Play(cardIndex int) string {
	if out, blocked := guardNotPlayable(gi.Game, gi.gp); blocked {
		return out
	}
	if err := gi.Game.PlayerPlay(cardIndex); err != nil {
		return gi.gp.Output(gi.Game, err)
	}
	gi.runCpuTurns()
	return gi.gp.Output(gi.Game, nil)
}

// GiveUp 投了する
func (gi *GermanWhistInteractor) GiveUp() string {
	if out, blocked := guardGameEnd(gi.Game, gi.gp); blocked {
		return out
	}
	gi.Game.GiveUp()
	return gi.gp.Output(gi.Game, nil)
}

// Hint ヒント取得
func (gi *GermanWhistInteractor) Hint() string {
	return gi.gp.HintOutput(gi.Game)
}

// ActionLog 棋譜を出力する
func (gi *GermanWhistInteractor) ActionLog() string {
	return gi.gp.ActionLogOutput(gi.Game)
}

// runCpuTurns 人間の手番になるかゲームが終わるまで CPU を進める。
//
// 共通の runCpuTurnsLoop は使えない。あれは「プレイ用フェーズがひとつ」で
// トリック解決が別ステップになっているゲーム向けで、ジャーマンホイストは
// 前半 (Draw) と後半 (Scoring) の**どちらもプレイ用フェーズ**、かつ
// トリックは 2 枚揃った時点でドメインが自分で解決するため。
func (gi *GermanWhistInteractor) runCpuTurns() {
	for turns := 0; !gi.Game.GetGameEndFlag() && !gi.Game.IsHumanTurn(); turns++ {
		// 進まない CpuPlay でハングしないための上限 (#4607 と同じ理由)。
		if turns >= maxCpuTurnsPerCall {
			return
		}
		gi.Game.CpuPlay()
	}
}

// RestoreGermanWhistInteractor deserialises JSON into a GermanWhistInteractor.
func RestoreGermanWhistInteractor(data []byte, gp presenter.GermanWhistPresenter) (*GermanWhistInteractor, error) {
	return restoreAndBuild[domain.GermanWhist](data, func(g *domain.GermanWhist) *GermanWhistInteractor {
		return &GermanWhistInteractor{GameBase: GameBase[interfaces.GermanWhistGame]{Game: g}, gp: gp}
	})
}
