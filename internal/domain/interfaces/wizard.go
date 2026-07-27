package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// WizardGame ウィザードゲームインタフェース
type WizardGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerBid プレイヤーがビッドする
	PlayerBid(bid int) error
	// CpuBid CPUプレイヤーがビッドする
	CpuBid()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ラウンドの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.WizardConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.WizardConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.WizardPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 現在のビッド手番が人間かを返す
	IsHumanBidTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTotalRounds 総ラウンド数を取得する
	GetTotalRounds() int
	// GetHandSize 現在のラウンドの手札枚数を取得する
	GetHandSize() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetTrumpCard 切り札カードを取得する
	GetTrumpCard() *domain.Card
	// GetTrumpSuit 切り札スートを取得する (-1 = 切り札なし)
	GetTrumpSuit() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetBidPlayerIdx ビッドプレイヤーインデックスを取得する
	GetBidPlayerIdx() int
	// GetRestrictedBid ディーラーが選択できないビッド値を返す (Wizardは常に-1)
	GetRestrictedBid() int
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.WizardPlayer
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.WizardHint
}
