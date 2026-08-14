//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PolignacGame ポリニャック (Polignac) ゲームインタフェース
type PolignacGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// DeclareCapot capot（全トリック獲得）を宣言する
	DeclareCapot() error
	// PassDeclaration 宣言せずにプレイへ進む
	PassDeclaration() error
	// PlayerPlay プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// CpuPlay CPUプレイヤーが1枚出す
	CpuPlay()
	// NextRound 次のラウンドを開始する
	NextRound()
	// GiveUp 投了する
	GiveUp()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.PolignacConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.PolignacConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.PolignacPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// IsDeclarePhase capot 宣言の受付中かを返す
	IsDeclarePhase() bool
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
	// GetTrickNumber 現在のトリック番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetCapotIdx capot 宣言者を取得する (-1: 宣言なし)
	GetCapotIdx() int
	// GetCapotTricks capot 宣言者が取ったトリック数を取得する
	GetCapotTricks() int
	// GetCurrentTrick 現在のトリックを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.PolignacPlayer
	// GetWinnerIdx 勝者プレイヤーインデックスを取得する (-1: 未確定/同点)
	GetWinnerIdx() int
	// GetHint ヒントを取得する
	GetHint() *domain.PolignacHint
}
