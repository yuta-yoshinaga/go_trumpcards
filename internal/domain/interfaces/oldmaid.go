package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// OldMaidGame ババ抜きゲームインタフェース
type OldMaidGame interface {
	// Reset ゲームを初期化する
	Reset()
	// SetConfig ゲーム設定をセットする
	SetConfig(config domain.OldMaidConfig)
	// ArrangeTargetForHumanDraw 人間が引く前にCPU心理戦の配置を行う
	ArrangeTargetForHumanDraw()
	// GetGameEndFlag ゲーム終了フラグを取得する
	GetGameEndFlag() bool
	// IsHumanTurn 現在の手番が人間かを返す
	IsHumanTurn() bool
	// PlayerDraw 人間プレイヤーがカードを引く
	PlayerDraw(cardIdx int) error
	// CpuDraw CPUプレイヤーが1ターン実行する
	CpuDraw() error
	// ShuffleHumanHand 人間の手札をシャッフルする
	ShuffleHumanHand() error
	// ReorderHumanHand 人間の手札を指定順に並び替える
	ReorderHumanHand(indices []int) error

	// GetHumanProfile メタAIプロファイルを取得する
	GetHumanProfile() *domain.OldMaidHumanProfile
	// ResetProfile メタAIプロファイルをリセットする
	ResetProfile()
	// ExportProfile メタAIプロファイルをエクスポートする
	ExportProfile() interface{}
	// ImportProfile JSONバイトからメタAIプロファイルをインポートする
	ImportProfile(data []byte) error

	// GetPlayerCnt プレイヤー数を取得する
	GetPlayerCnt() int
	// GetPlayer 指定インデックスのプレイヤーを取得する
	GetPlayer(i int) *domain.OldMaidPlayer
	// GetHasDrawn 引きが発生したかを返す
	GetHasDrawn() bool
	// GetLastDrawPlayerIdx 最後に引いたプレイヤーのインデックスを取得する
	GetLastDrawPlayerIdx() int
	// GetLastDrawFromIdx 最後に引いた相手のインデックスを取得する
	GetLastDrawFromIdx() int
	// GetLastDrawCard 最後に引いたカードを取得する
	GetLastDrawCard() *domain.Card
	// GetLastDiscardedPairs 最後に捨てたペア数を取得する
	GetLastDiscardedPairs() int
	// GetLastDiscardedCards 最後に捨てたカードを取得する
	GetLastDiscardedCards() []*domain.Card
	// GetCpuActions CPUターンの行動履歴を取得する
	GetCpuActions() []*domain.OldMaidCpuAction
	// GetHumanAction 人間の最後の行動記録を取得する
	GetHumanAction() *domain.OldMaidCpuAction
	// GetDrawHistory ゲーム全体の引き履歴を取得する
	GetDrawHistory() []*domain.OldMaidDrawHistoryEntry
	// GetLoserIdx 負けプレイヤーインデックスを取得する
	GetLoserIdx() int
	// GetCurrentTurn 現在の手番プレイヤーインデックスを取得する
	GetCurrentTurn() int
	// GetNextDrawTargetIdx 次の引き相手のインデックスを取得する
	GetNextDrawTargetIdx() int
	// GetConfig ゲーム設定を取得する
	GetConfig() domain.OldMaidConfig
	// GetRemovedCard ジジ抜きで除外されたカードを取得する
	GetRemovedCard() *domain.Card
	// GetCpuHighlightedCardIdx CPU心理戦で強調された位置を取得する
	GetCpuHighlightedCardIdx() int
	// GetActionLog 棋譜を取得する
	GetActionLog() []*domain.ActionLogEntry
}
