//go:build !js || !wasm || extra

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// FrenchTarotGame フレンチタロット (French Tarot) のゲームインタフェース
type FrenchTarotGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のディールを開始する
	NextRound()
	// PlayerBid 人間プレイヤーが入札する
	PlayerBid(bid domain.FrenchTarotBid) error
	// PlayerPass 人間プレイヤーがパスする
	PlayerPass() error
	// CpuBid CPU プレイヤーが 1 回入札する
	CpuBid()
	// PlayerDiscard 人間デクレアラーがシアン交換で 6 枚を捨てる
	PlayerDiscard(cardIndices []int) error
	// CpuDiscard CPU デクレアラーがシアン交換で 6 枚を捨てる
	CpuDiscard()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPU プレイヤーが 1 ターン実行する
	CpuPlay()
	// ResolveTrick トリックを解決する
	ResolveTrick()
	// NextTrick 次のトリックを開始する
	NextTrick()
	// ScoreRound ディールの得点を計算する
	ScoreRound()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.FrenchTarotConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.FrenchTarotConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.FrenchTarotPhase
	// IsHumanTurn 現在のプレイ手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 現在の入札手番が人間かを返す
	IsHumanBidTurn() bool
	// IsHumanDiscardTurn 現在のシアン交換手番が人間かを返す
	IsHumanDiscardTurn() bool
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
	// GetBidPlayerIdx 入札手番インデックスを取得する
	GetBidPlayerIdx() int
	// GetHighestBid 現在の最高入札を取得する
	GetHighestBid() domain.FrenchTarotBid
	// GetHighestBidder 最高入札者を取得する (-1=なし)
	GetHighestBidder() int
	// GetDeclarerIdx デクレアラーインデックスを取得する (-1=未確定)
	GetDeclarerIdx() int
	// GetContract コントラクト (確定入札) を取得する
	GetContract() domain.FrenchTarotBid
	// GetChienCount シアンの枚数を取得する
	GetChienCount() int
	// GetChien シアンを取得する
	GetChien() []*domain.Card
	// GetChienRevealed シアンが公開済みかを取得する
	GetChienRevealed() bool
	// GetStashOwner stash の所有側を取得する (0=デクレアラー, 1=防御側)
	GetStashOwner() int
	// GetPlayerScores プレイヤー別累積得点を取得する
	GetPlayerScores() [domain.FrenchTarotPlayerCnt]int
	// GetCardPoints プレイヤー i の獲得ハーフポイントを取得する
	GetCardPoints(i int) int
	// GetOutcome 直近ディールの結果を取得する
	GetOutcome() domain.FrenchTarotOutcome
	// GetResult 人間視点のマッチ結果を取得する
	GetResult() domain.FrenchTarotResult
	// GetWinnerPlayer 勝利プレイヤーを取得する (-1=未確定)
	GetWinnerPlayer() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.FrenchTarotPlayer
	// GetPlayableIndices プレイ可能なカードのインデックスを取得する
	GetPlayableIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.FrenchTarotHint
}
