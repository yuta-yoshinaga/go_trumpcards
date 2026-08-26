//go:build !js || !wasm || casino

package usecase

import (
	"errors"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// DramahaInteractorIF ドラマハホールデムインタラクターインタフェース
type DramahaInteractorIF interface {
	// Snapshot serialises game state for KV persistence.
	Snapshot() ([]byte, error)
	TournamentInteractorIF
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化 (profileData: JSONプロファイル、nilなら無視)
	ResetWithConfig(cfg domain.DramahaConfig, profileData []byte) string
	// Action プレイヤーアクション実行
	Action(action int, amount int, humanPlayMs int) string
	// Draw ドローラウンドでホールカードを引き直す (0 始まりの位置)
	Draw(indices []int) string
	// GetConfig 現在の設定を取得
	GetConfig() domain.DramahaConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// DramahaInteractor ドラマハホールデムインタラクタークラス
type DramahaInteractor struct {
	GameBase[interfaces.DramahaGame]
	op presenter.DramahaPresenter
	tournamentActions[interfaces.DramahaGame]
}

// NewDramahaInteractor コンストラクタ
func NewDramahaInteractor(o interfaces.DramahaGame, op presenter.DramahaPresenter) *DramahaInteractor {
	mustNotNil("DramahaInteractor", map[string]any{"o": o, "op": op})
	return &DramahaInteractor{
		GameBase:          GameBase[interfaces.DramahaGame]{Game: o},
		op:                op,
		tournamentActions: newTournamentActions[interfaces.DramahaGame](o, op),
	}
}

// Reset ゲーム初期化
func (oi *DramahaInteractor) Reset() string {
	return execAndPresent(oi.Game, oi.op, oi.Game.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (oi *DramahaInteractor) ResetWithConfig(cfg domain.DramahaConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return oi.op.Output(oi.Game, err)
	}
	// **収まらない席数は黙って丸めず断る。** NewDramahaPlayersForTable は
	// 安全のため 4 に丸めるが、丸めるだけだと「6 人卓にした」と頼んだ側には
	// 効いたのか無視されたのか分からない。CUI も Web API もここを通るので、
	// 判断はこの 1 箇所に置く。
	if cfg.TableSize > 0 && cfg.TableSize != domain.HoldemTableSize4 {
		return oi.op.Output(oi.Game, errDramahaTableSize(cfg.TableSize))
	}
	if cfg.TableSize > 0 && cfg.TableSize != oi.Game.GetPlayerCnt() {
		oi.Game.Resize(domain.NewDramahaPlayersForTable(cfg.TableSize))
	}
	oi.Game.SetConfig(cfg)
	err := oi.Game.Reset()
	if len(profileData) > 0 {
		_ = oi.Game.ImportProfile(profileData)
	}
	return oi.op.Output(oi.Game, err)
}

// Action プレイヤーアクション実行
func (oi *DramahaInteractor) Action(action int, amount int, humanPlayMs int) string {
	return execAndPresent(oi.Game, oi.op, func() error { return oi.Game.PlayerAction(action, amount, humanPlayMs) })
}

// Draw はドローラウンドで人間席のホールカードを引き直す。
//
// **席は人間 (0) 固定。** どの席を引かせるかを外から渡せると、Web からの
// リクエストで CPU の手札を作り替えられてしまう。
func (oi *DramahaInteractor) Draw(indices []int) string {
	return execAndPresent(oi.Game, oi.op, func() error { return oi.Game.Draw(0, indices) })
}

// GetConfig 現在の設定を取得
func (oi *DramahaInteractor) GetConfig() domain.DramahaConfig {
	return oi.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (oi *DramahaInteractor) ActionLog() string {
	return oi.op.ActionLogOutput(oi.Game)
}

// RestoreDramahaInteractor deserialises JSON into an DramahaInteractor.
func RestoreDramahaInteractor(data []byte, op presenter.DramahaPresenter) (*DramahaInteractor, error) {
	o, err := restoreGame[domain.Dramaha](data)
	if err != nil {
		return nil, err
	}
	return &DramahaInteractor{GameBase: GameBase[interfaces.DramahaGame]{Game: o}, op: op, tournamentActions: newTournamentActions[interfaces.DramahaGame](o, op)}, nil
}

// errDramahaTableSize は山に収まらない席数を断るエラー。
//
// ドラマハは 1 席が最悪 10 枚 (ホール 5 + 交換 5) 使い、ボードに 5 枚要るので
// 必要枚数は 10N+5。52 枚に収まるのは 4-max だけ (6-max=65, 9-max=95)。
func errDramahaTableSize(size int) error {
	return errors.New(i18n.Tf("dramaha.tableSizeFixed", "val", strconv.Itoa(size)))
}
