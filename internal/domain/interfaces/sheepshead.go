//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// SheepsheadGame シープスヘッドのゲームインタフェース
type SheepsheadGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerPick ピック(true)/パス(false)を選択する
	PlayerPick(pick bool) error
	// PlayerBury ピッカーが2枚を埋める
	PlayerBury(indices []int) error
	// PlayerCall ピッカーが呼びスートを指定する
	PlayerCall(suit int) error
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
	GetConfig() domain.SheepsheadConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.SheepsheadConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.SheepsheadPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetBlind ブラインドを取得する
	GetBlind() []*domain.Card
	// GetBuried 埋め札を取得する
	GetBuried() []*domain.Card
	// GetPickerIdx ピッカーのインデックスを取得する
	GetPickerIdx() int
	// GetPartnerIdx 相棒のインデックスを取得する
	GetPartnerIdx() int
	// GetCalledSuit 呼びスートを取得する
	GetCalledSuit() int
	// IsPartnerRevealed 相棒が判明済みかを取得する
	IsPartnerRevealed() bool
	// GetPassCount 現ピックフェーズのパス人数を取得する
	GetPassCount() int
	// GetRoundPickerPoints 直近ラウンドのピッカー組得点を取得する
	GetRoundPickerPoints() int
	// GetRoundMultiplier 直近ラウンドの倍率を取得する
	GetRoundMultiplier() int
	// GetRoundPickerWon 直近ラウンドでピッカー組が勝ったかを取得する
	GetRoundPickerWon() bool
	// GetWinnerIdx 勝者インデックスを取得する
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.SheepsheadPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetCallableSuits 呼べるフェイルスート一覧を取得する
	GetCallableSuits() []int
	// GetHint ヒントを取得する
	GetHint() *domain.SheepsheadHint
}
