//go:build !js || !wasm || extra2

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// KempsGame はケムプス (Kemps) のゲームインタフェース。
type KempsGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// ResetWithConfig 設定を更新してゲームを初期化する
	ResetWithConfig(cfg domain.KempsConfig)
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerSwap 人間が手札の 1 枚をフィールドの 1 枚と交換する
	PlayerSwap(handIndex, fieldIndex int) error
	// PlayerPass 人間が交換せずにパスする
	PlayerPass() error
	// PlayerSetSignal 人間が秘密のシグナル種別を設定する
	PlayerSetSignal(signalType int)
	// PlayerDeclareKemps 人間が Kemps を宣言する
	PlayerDeclareKemps() error
	// PlayerDeclareCounterKemps 人間が相手 targetSeat に Counter-Kemps を宣言する
	PlayerDeclareCounterKemps(targetSeat int) error
	// CpuPlay CPU の手番を 1 ステップ進める
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.KempsConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.KempsConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.KempsPhase
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.KempsPlayer
	// GetWinnerTeam 勝利チームを取得する (-1=未確定)
	GetWinnerTeam() int
	// GetTeamScore チームの得点を取得する
	GetTeamScore(team int) int
	// GetFieldSize フィールドのカード枚数を取得する
	GetFieldSize() int
	// GetFieldCard フィールドの指定インデックスのカードを取得する
	GetFieldCard(i int) *domain.Card
	// GetCurrentPlayerIdx 現在の手番プレイヤーを取得する
	GetCurrentPlayerIdx() int
	// GetSignalType 人間が設定したシグナル種別を取得する
	GetSignalType() domain.SignalType
	// GetFourHolderIdx 現在フォーオブアカインドを保持するプレイヤーを取得する (-1=なし)
	GetFourHolderIdx() int
	// IsPartnerSignaling 人間チームのシグナル状態を取得する (人間にのみ公開)
	IsPartnerSignaling() bool
	// IsOpponentSignaling 相手チームのフォーオブアカインドの気配を取得する
	IsOpponentSignaling() bool
	// GetRoundResult 直近ラウンドの結果コードを取得する
	GetRoundResult() int
	// GetRoundWinnerTeam 直近ラウンドで得点したチームを取得する (-1=なし)
	GetRoundWinnerTeam() int
	// GetRoundNumber 現在のラウンド番号を取得する
	GetRoundNumber() int
}
