//go:build !js || !wasm || extra4

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// PutGame プットゲームインタフェース
type PutGame interface {
	BaseGame
	// Reset マッチを初期化する
	Reset()
	// PlayerPlay 人間プレイヤーがカードを出す
	PlayerPlay(cardIndex int) error
	// DeclarePut 人間プレイヤーが Put を宣言 (または再引き上げ) する
	DeclarePut() error
	// RespondPut 人間プレイヤーが Put 宣言に応答する (true=受諾 / false=拒否)
	RespondPut(accept bool) error
	// CpuStep CPU が1アクション実行する
	CpuStep()
	// Next バサ終了 / マノ終了から次の状態へ進める
	Next()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.PutConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.PutConfig)

	// GetGameEndFlag マッチ終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.PutPhase
	// IsHumanTurn 現在の手番 (プレイまたは応答) が人間かを返す
	IsHumanTurn() bool
	// CanDeclarePut 現在の手番の人間が Put 宣言可能かを返す
	CanDeclarePut() bool
	// GetHandNumber 現在のマノ番号を取得する
	GetHandNumber() int
	// GetTrickNumber 現在のバサ番号を取得する
	GetTrickNumber() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックスを取得する
	GetCurrentPlayerIdx() int
	// GetResponderIdx 応答すべきプレイヤーインデックスを取得する (-1: 応答待ちでない)
	GetResponderIdx() int
	// GetCurrentTrick 現在のバサを取得する
	GetCurrentTrick() []*domain.TrickCard
	// GetTrickResults 当該マノで完了したバサの勝者リストを取得する
	GetTrickResults() []int
	// GetLeadPlayerIdx リードプレイヤーインデックスを取得する
	GetLeadPlayerIdx() int
	// GetManoIdx 親 (elder hand) インデックスを取得する
	GetManoIdx() int
	// GetDealerIdx ディーラーインデックスを取得する
	GetDealerIdx() int
	// GetHandStake 現在の確定賭け点を取得する
	GetHandStake() int
	// GetAcceptedLevel 受諾済みベッティングレベルを取得する
	GetAcceptedLevel() int
	// GetPendingLevel 応答待ち提示レベルを取得する (0: 応答待ちでない)
	GetPendingLevel() int
	// GetPutCallerIdx 応答待ちの宣言者インデックスを取得する (-1: なし)
	GetPutCallerIdx() int
	// GetMatchTarget マッチ目標点を取得する
	GetMatchTarget() int
	// GetPlayerMatchPoints プレイヤーのマッチ累積点を取得する
	GetPlayerMatchPoints(i int) int
	// GetHandWinnerIdx 直近マノの勝者を取得する (-1: 未確定)
	GetHandWinnerIdx() int
	// GetWinnerIdx マッチ勝者インデックスを取得する (-1: 未確定)
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.PutPlayer
	// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
	GetValidPlayIndices(playerIdx int) []int
	// GetHint ヒントを取得する
	GetHint() *domain.PutHint
}
