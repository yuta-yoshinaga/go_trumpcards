//go:build !js || !wasm || classic

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// ColourWhistGame カラーホイストゲームインタフェース
type ColourWhistGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// Bid 契約を宣言する (ColourWhistContractNone でパス)
	Bid(contract int) error
	// Call 切り札を決め、Samen なら相方の札を指名する
	Call(trumpSuit int) error
	// PlayCard 札を出す
	PlayCard(cardIndex int) error
	// NextRound 次のラウンドを配る
	NextRound() error
	// GiveUp 投了する
	GiveUp()
	// CpuPlay CPUの手番を進める
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.ColourWhistConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.ColourWhistConfig)

	// GetPhase 現在のフェーズ
	GetPhase() int
	// GetGameEndFlag ゲーム終了フラグ
	GetGameEndFlag() bool
	// IsHumanTurn 人間の入力を待っているか
	IsHumanTurn() bool
	// GetValidPlayIndices 出せる手札の位置
	GetValidPlayIndices(playerIdx int) []int
	// IsDeclarerSide 席が契約側かを返す
	IsDeclarerSide(playerIdx int) bool
	// GetDealerIdx 親の席
	GetDealerIdx() int
	// GetContract 成立した契約
	GetContract() int
	// GetDeclarerIdx 契約者の席 (-1: 未定)
	GetDeclarerIdx() int
	// GetPartnerIdx 相方の席 (-1: 相方なし、または未公開)
	GetPartnerIdx() int
	// GetCalledCard 指名した札 (nil: 指名なし)
	GetCalledCard() *domain.Card
	// GetTrumpSuit 切り札
	GetTrumpSuit() int
	// IsTroelForced 配りで Troel が強制成立したか
	IsTroelForced() bool
	// HasPassed 席が降りたか
	HasPassed(i int) bool
	// GetCurrentTurn いまの手番の席
	GetCurrentTurn() int
	// GetTrick 進行中のトリック
	GetTrick() []*domain.TrickCard
	// GetLastTrick 直前に完成したトリック
	GetLastTrick() []*domain.TrickCard
	// GetLastTrickWinner 直前のトリックを取った席
	GetLastTrickWinner() int
	// GetTrickCount このラウンドで完成したトリック数
	GetTrickCount() int
	// GetDeclarerTricks 契約側が取ったトリック数
	GetDeclarerTricks() int
	// GetRoundNumber ラウンド数
	GetRoundNumber() int
	// GetPlayerCnt 人数
	GetPlayerCnt() int
	// GetPlayer 席 i のプレイヤー
	GetPlayer(i int) *domain.ColourWhistPlayer
	// GetWinnerIdx 勝者の席 (-1: 未確定)
	GetWinnerIdx() int
	// GetHint 助言
	GetHint() *domain.ColourWhistHint
}
