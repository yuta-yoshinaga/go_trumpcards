package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// DurakGame ドゥラークゲームインタフェース
type DurakGame interface {
	BaseGame
	// Reset ゲームを初期化する
	Reset()
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// PlayerAttack 人間プレイヤーが攻撃カードを出す
	PlayerAttack(cardIdx int) error
	// PlayerDefend 人間プレイヤーが防御カードを出す
	PlayerDefend(attackIdx, handIdx int) error
	// PlayerPass 人間プレイヤーがパス (攻撃停止)
	PlayerPass() error
	// PlayerTakeCards 人間プレイヤーがカードを引き取る (防御放棄)
	PlayerTakeCards() error
	// CpuPlay CPUプレイヤーが1ターン実行する
	CpuPlay()
	// HasPendingAction ペンディングアクションがあるかを返す
	HasPendingAction() bool
	// SetConfig ゲーム設定をセットする
	SetConfig(config domain.DurakConfig)
	// SortHumanHand 人間の手札をソートする
	SortHumanHand(mode domain.DurakSortMode) error

	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.DurakPlayer
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetPhase 現在のフェーズを取得する
	GetPhase() domain.DurakPhase
	// GetAttackerIdx 攻撃者インデックスを取得する
	GetAttackerIdx() int
	// GetDefenderIdx 防御者インデックスを取得する
	GetDefenderIdx() int
	// GetTablePairs テーブル上のカードペアを取得する
	GetTablePairs() []*domain.DurakTablePair
	// GetTrumpSuit 切り札スートを取得する
	GetTrumpSuit() int
	// GetTrumpCard 切り札カード (山札底) を取得する
	GetTrumpCard() *domain.Card
	// GetStockCount 山札残数を取得する
	GetStockCount() int
	// GetLoserIdx 敗者インデックスを取得する
	GetLoserIdx() int
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.DurakConfig
	// GetSortMode 現在のソートモードを取得する
	GetSortMode() domain.DurakSortMode
	// GetCpuActions CPU行動記録を取得する
	GetCpuActions() []*domain.DurakCpuAction
	// GetHumanAction 人間の最後の行動記録を取得する
	GetHumanAction() *domain.DurakCpuAction
	// GetBoutNumber バウト番号を取得する
	GetBoutNumber() int
	// GetHint サーバー計算の推奨手を返す (手番でなければ nil)
	GetHint() *domain.DurakHint
}
