//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// GoofspielGame ゴフスピールゲームインタフェース
type GoofspielGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerBid 人間が入札札を伏せる
	PlayerBid(cardIndex int) error
	// NextRound 公開状態から次の賞札をめくる
	NextRound() error
	// CpuPlay CPUの入札を進める
	CpuPlay()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.GoofspielConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.GoofspielConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.GoofspielPhase
	// IsHumanTurn 人間の入力を待っているかを返す
	IsHumanTurn() bool
	// GetValidBidIndices 入札できる手札の位置を返す
	GetValidBidIndices(playerIdx int) []int
	// GetCurrentPrize いま公開されている賞札を返す (nil: 公開中でない)
	GetCurrentPrize() *domain.Card
	// GetCarriedPrizes 持ち越された賞札を返す
	GetCarriedPrizes() []*domain.Card
	// PrizeValue いま懸かっている得点を返す (持ち越しを含む)
	PrizeValue() int
	// GetPrizeRemaining まだめくっていない賞札の枚数を返す
	GetPrizeRemaining() int
	// HasBid 席iが伏せ終えたかを返す
	HasBid(i int) bool
	// GetRevealedBids 直前に公開された入札を返す
	GetRevealedBids() []*domain.Card
	// GetLastWinnerIdx 直前のラウンドの勝者を返す (-1: 同点)
	GetLastWinnerIdx() int
	// GetLastGained 直前のラウンドで動いた得点を返す
	GetLastGained() int
	// GetRoundNumber ラウンド数を返す
	GetRoundNumber() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.GoofspielPlayer
	// GetWinnerIdx 勝った席を取得する (-1: 未確定)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.GoofspielHint
}
