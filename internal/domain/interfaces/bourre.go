//go:build !js || !wasm || casino

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BourreGame ブーレゲームインタフェース
type BourreGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// PlayerDecide 人間プレイヤーが参加(true)/フォールド(false)を決める
	PlayerDecide(play bool) error
	// PlayerDraw 人間プレイヤーが手札を交換する
	PlayerDraw(indices []int) error
	// PlayerPlay 人間プレイヤーがカードをプレイする
	PlayerPlay(cardIndex int) error
	// NextHand 次のハンドを開始する
	NextHand()
	// CpuPlay CPUが1アクション実行する
	CpuPlay()
	// SetConfig ゲーム設定をセットする
	SetConfig(config domain.BourreConfig)

	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BourrePhase
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.BourrePlayer
	// GetCurrentPlayerIdx 現在の手番プレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetPot 現在のポットを取得する
	GetPot() int
	// GetCarryPot 繰り越しポットを取得する
	GetCarryPot() int
	// GetTrumpSuit 切り札スートを取得する
	GetTrumpSuit() int
	// GetTrumpCard 切り札カードを取得する
	GetTrumpCard() *domain.Card
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentTrick 進行中のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetLastTrick 直前のトリックを取得する
	GetLastTrick() []*domain.TrickCard
	// GetLastTrickWinner 直前トリックの勝者を取得する
	GetLastTrickWinner() int
	// GetLeadPlayerIdx リードプレイヤーを取得する
	GetLeadPlayerIdx() int
	// GetHandNumber 現在のハンド番号を取得する
	GetHandNumber() int
	// GetWinnerIdx ゲーム勝者を取得する
	GetWinnerIdx() int
	// GetLastResults 直前ハンドの結果を取得する
	GetLastResults() []*domain.BourreHandResult
	// GetValidPlayIndices 合法手のインデックス一覧を取得する
	GetValidPlayIndices(idx int) []int
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.BourreConfig
}
