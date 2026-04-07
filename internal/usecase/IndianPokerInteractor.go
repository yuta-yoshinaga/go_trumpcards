package usecase

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase/presenter"
)

// IndianPokerInteractorIF インディアンポーカーインタラクターインタフェース
type IndianPokerInteractorIF interface {
	// Reset ゲーム初期化
	Reset() string
	// ResetWithConfig 設定を変更してゲーム初期化 (profileData: JSONプロファイル、nilなら無視)
	ResetWithConfig(cfg domain.IndianPokerConfig, profileData []byte) string
	// Action プレイヤーアクション実行
	Action(action int, amount int, humanPlayMs int) string
	// GetConfig 現在の設定を取得
	GetConfig() domain.IndianPokerConfig
	// ActionLog 棋譜を出力する
	ActionLog() string
}

// IndianPokerInteractor インディアンポーカーインタラクタークラス
type IndianPokerInteractor struct {
	GameBase[interfaces.IndianPokerGame]
	ipp presenter.IndianPokerPresenter
}

// NewIndianPokerInteractor コンストラクタ
func NewIndianPokerInteractor(ip interfaces.IndianPokerGame, ipp presenter.IndianPokerPresenter) *IndianPokerInteractor {
	mustNotNil("IndianPokerInteractor", map[string]any{"ip": ip, "ipp": ipp})
	return &IndianPokerInteractor{GameBase: GameBase[interfaces.IndianPokerGame]{Game: ip}, ipp: ipp}
}

// Reset ゲーム初期化
func (ipi *IndianPokerInteractor) Reset() string {
	return execAndPresent(ipi.Game, ipi.ipp, ipi.Game.Reset)
}

// ResetWithConfig 設定を変更してゲーム初期化
func (ipi *IndianPokerInteractor) ResetWithConfig(cfg domain.IndianPokerConfig, profileData []byte) string {
	if err := cfg.Validate(); err != nil {
		return ipi.ipp.Output(ipi.Game, err)
	}
	ipi.Game.SetConfig(cfg)
	err := ipi.Game.Reset()
	if len(profileData) > 0 {
		_ = ipi.Game.ImportProfile(profileData)
	}
	return ipi.ipp.Output(ipi.Game, err)
}

// Action プレイヤーアクション実行
func (ipi *IndianPokerInteractor) Action(action int, amount int, humanPlayMs int) string {
	return execAndPresent(ipi.Game, ipi.ipp, func() error { return ipi.Game.PlayerAction(action, amount, humanPlayMs) })
}

// GetConfig 現在の設定を取得
func (ipi *IndianPokerInteractor) GetConfig() domain.IndianPokerConfig {
	return ipi.Game.GetConfig()
}

// ActionLog 棋譜を出力する
func (ipi *IndianPokerInteractor) ActionLog() string {
	return ipi.ipp.ActionLogOutput(ipi.Game)
}

// RestoreIndianPokerInteractor deserialises JSON into an IndianPokerInteractor.
func RestoreIndianPokerInteractor(data []byte, ipp presenter.IndianPokerPresenter) (*IndianPokerInteractor, error) {
	ip, err := restoreGame[domain.IndianPoker](data)
	if err != nil {
		return nil, err
	}
	return &IndianPokerInteractor{GameBase: GameBase[interfaces.IndianPokerGame]{Game: ip}, ipp: ipp}, nil
}
