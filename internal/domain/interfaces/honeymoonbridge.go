//go:build !js || !wasm || solo

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// HoneymoonBridgeGame ハネムーンブリッジゲームインタフェース
type HoneymoonBridgeGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// PlayerBid 人間が契約を宣言する
	PlayerBid(level, suit int) error
	// PlayerPass 人間が降りる
	PlayerPass() error
	// CpuBid CPUが宣言する
	CpuBid()
	// NextRound 次のディールを開始する
	NextRound()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.HoneymoonBridgeConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.HoneymoonBridgeConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.HoneymoonBridgePhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsHumanBidTurn 人間が宣言する番かを返す
	IsHumanBidTurn() bool
	// GetRoundNumber 現在のディール番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetStockSize 山札の残り枚数を取得する
	GetStockSize() int
	// GetTrumpSuit 切り札のスートを取得する (0: ノートランプ)
	GetTrumpSuit() int
	// GetDeclarerIdx 落札者を取得する (-1: 未定)
	GetDeclarerIdx() int
	// GetContractLevel 契約レベルを取得する (0: 未定)
	GetContractLevel() int
	// RequiredTricks 契約に必要なトリック数を取得する
	RequiredTricks() int
	// NextBid 次に宣言できる最小の契約を取得する
	NextBid() (int, int)
	// GetLastMade 直前のディールで契約が成立したかを取得する
	GetLastMade() bool
	// GetLastTricks 直前のディールで落札者が取ったトリック数を取得する
	GetLastTricks() int
	// GetLastPoints 直前のディールで動いた点数を取得する
	GetLastPoints() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.HoneymoonBridgePlayer
	// GetWinnerIdx 勝者を取得する (-1: 未確定)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.HoneymoonBridgeHint
}
