//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// BridgeGame ブリッジゲームインタフェース
type BridgeGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerBid 人間プレイヤーがビッドする
	PlayerBid(bidType int, level int, suit int) error
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
	GetConfig() domain.BridgeConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.BridgeConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.BridgePhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 現在のビッド手番が人間かを返す
	IsHumanBidTurn() bool
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
	// GetBidPlayerIdx ビッドプレイヤーインデックスを取得する
	GetBidPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetTrumpSuit 切り札スートを取得する (-1 = NoTrump)
	GetTrumpSuit() int
	// GetContractLevel コントラクトレベルを取得する
	GetContractLevel() int
	// GetContractSuit コントラクトスートを取得する
	GetContractSuit() int
	// GetDoubled ダブル状態を取得する (0=なし, 1=ダブル, 2=リダブル)
	GetDoubled() int
	// BridgeMinLegalBid 現在のコントラクトを上回る最も低いビッドを返す
	BridgeMinLegalBid() (level, suit int, ok bool)
	// BridgeCanDouble 今ダブルできるかを返す
	BridgeCanDouble(playerIdx int) bool
	// BridgeCanRedouble 今リダブルできるかを返す
	BridgeCanRedouble(playerIdx int) bool
	// GetDeclarerIdx デクレアラーインデックスを取得する
	GetDeclarerIdx() int
	// GetDummyIdx ダミーインデックスを取得する
	GetDummyIdx() int
	// GetBidHistory ビッド履歴を取得する
	GetBidHistory() []*domain.BridgeBidEntry
	// GetVulnerability バルネラビリティを取得する
	GetVulnerability(team int) bool
	// GetTeamScore チームスコアを取得する
	GetTeamScore(team int) int
	// GetGamesWon 勝利ゲーム数を取得する
	GetGamesWon(team int) int
	// GetBelowLine ライン以下スコアを取得する
	GetBelowLine(team int) int
	// GetWinnerTeam 勝利チームを取得する
	GetWinnerTeam() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.BridgePlayer
	// IsOpeningLeadDone オープニングリード完了かを取得する
	IsOpeningLeadDone() bool
	// GetDummyHand ダミーの手札を取得する
	GetDummyHand() []*domain.Card
	// GetHint ヒントを取得する
	GetHint() *domain.BridgeHint
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
}
