//go:build !js || !wasm || extra5

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// MarriageGame マリッジゲームインタフェース
type MarriageGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// NextRound 次のラウンドを開始する
	NextRound()
	// PlayerDrawFromStock プレイヤーが山札からカードを引く
	PlayerDrawFromStock() error
	// PlayerDrawFromDiscard プレイヤーが捨て札トップからカードを引く
	PlayerDrawFromDiscard() error
	// PlayerDiscard プレイヤーがカードを捨てる
	PlayerDiscard(cardIndex int) error
	// PlayerDeclare プレイヤーが宣言する（22 枚目をフィニッシュに捨て、残り 21 枚を判定）
	PlayerDeclare(cardIndex int) error
	// CpuPlay CPU プレイヤーが 1 ターン実行する
	CpuPlay()

	// GetConfig ゲーム設定を取得する
	GetConfig() domain.MarriageConfig
	// SetConfig ゲーム設定をセットする
	SetConfig(cfg domain.MarriageConfig)

	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.MarriagePhase
	// IsHumanTurn 現在の手番が人間か
	IsHumanTurn() bool
	// GetRoundNumber 現在のラウンド番号
	GetRoundNumber() int
	// GetTargetRounds ゲーム終了までのラウンド数
	GetTargetRounds() int
	// GetCurrentPlayerIdx 現在のプレイヤーインデックス
	GetCurrentPlayerIdx() int
	// GetDealerIdx ディーラーインデックス
	GetDealerIdx() int
	// GetDiscardTop 捨て札の一番上のカード
	GetDiscardTop() *domain.Card
	// GetDrawPileCount 山札の残り枚数
	GetDrawPileCount() int
	// GetWildJoker ワイルドジョーカーカード（nil の場合あり）
	GetWildJoker() *domain.Card
	// GetWildRank ワイルドランク（0 = 指定なし）
	GetWildRank() int
	// GetWinnerIdx 勝者インデックス（-1 未確定）
	GetWinnerIdx() int
	// GetPlayerCnt プレイヤー数
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.MarriagePlayer
	// GetDeclarerIdx 宣言したプレイヤーインデックス（-1 = 宣言なし）
	GetDeclarerIdx() int
	// GetDeclarationValid 直近の宣言が有効だったか
	GetDeclarationValid() bool
	// PlayerDeadwoodValue プレイヤー i のデッドウッド採点値
	PlayerDeadwoodValue(i int) int
	// PlayerMaalValue プレイヤー i のマール点
	PlayerMaalValue(i int) int
	// PlayerHasPureSequence プレイヤー i がピュアシーケンスを持つか
	PlayerHasPureSequence(i int) bool
}
