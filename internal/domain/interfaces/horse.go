//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// HorseGame は H.O.R.S.E. のゲームインタフェース。
type HorseGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextHand ハンドを閉じて次へ進む
	NextHand() error
	// PlayerAction 人間の手をいまの種目へ渡す
	PlayerAction(action, amount, humanPlayMs int) error

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.HorseConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.HorseConfig)

	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.HorsePhase
	// GetGameEndFlag 終局したかを取得する
	GetGameEndFlag() bool
	// GetDiscipline いまの種目を取得する
	GetDiscipline() domain.HorseDiscipline
	// GetDisciplineLetter いまの種目の頭文字を取得する
	GetDisciplineLetter() string
	// GetHandInDiscipline いまの種目で何ハンド目かを取得する
	GetHandInDiscipline() int
	// GetHandNumber 通算ハンド数を取得する
	GetHandNumber() int
	// GetSeatChips 指定席のチップを取得する
	GetSeatChips(i int) int
	// GetSeatName 指定席の表示名を取得する
	GetSeatName(i int) string
	// GetSeatIsHuman 指定席が人間かを取得する
	GetSeatIsHuman(i int) bool
	// GetSeatCount 席数を取得する
	GetSeatCount() int
	// GetHumanSeat 人間の席を取得する
	GetHumanSeat() int
	// GetCurrentTurn 手番の席を取得する (正本の番号)
	GetCurrentTurn() int
	// IsHumanTurn 人間の手番かを取得する
	IsHumanTurn() bool
	// GetPot いまの種目のポットを取得する
	GetPot() int
	// GetTablePhase いまの種目のフェーズ番号を取得する
	GetTablePhase() int
	// WinnerSeat チップが最も多い席を取得する
	WinnerSeat() int
	// GetSeatCards 指定席から見えている札を取得する (CPU は表向きのみ)
	GetSeatCards(seat int) []*domain.Card
	// GetCommunityCards いまの種目の共有札を取得する (スタッド系は空)
	GetCommunityCards() []*domain.Card
	// GetSeatLiveChips 指定席のいまの残高を取得する (ハンド中は卓の値)
	GetSeatLiveChips(seat int) int
	// GetToCall 人間の席がコールに要する額を取得する (0 ならチェック可)
	GetToCall() int
	// GetMinRaise いまの種目の最小レイズ幅を取得する
	GetMinRaise() int
}
