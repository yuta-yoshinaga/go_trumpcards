//go:build !js || !wasm || extra3

package interfaces

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"

// RamschGame Ramsch game interface used by the use-case layer.
//
// **Skat との差は「入札が無い」こと。** 落札・スカット取り・ゲーム宣言の
// メソッドは持たない ── 配ったらすぐトリックが始まり、切り札はジャック 4 枚で
// 固定される。得点は罰点で、最も多く取った人が失点する。
type RamschGame interface {
	BaseGame

	// Reset starts a new game session.
	Reset()
	// NextRound advances to the next round.
	NextRound()

	// PlayerPlay human plays a card.
	PlayerPlay(cardIndex int) error
	// CpuPlay CPU plays a card.
	CpuPlay()

	// ResolveTrick resolves the current trick.
	ResolveTrick()
	// NextTrick begins the next trick.
	NextTrick()
	// ScoreRound finalises the round score.
	ScoreRound()

	// GetConfig returns the current configuration.
	GetConfig() domain.RamschConfig
	// SetConfig sets the configuration.
	SetConfig(cfg domain.RamschConfig)

	// GetGameEndFlag reports whether the game session has ended.
	GetGameEndFlag() bool
	// GetPhase returns the current phase.
	GetPhase() domain.RamschPhase
	// IsHumanTurn reports whether the current trick player is human.
	IsHumanTurn() bool

	// GetRoundNumber returns the round number.
	GetRoundNumber() int
	// GetTrickNumber returns the trick number.
	GetTrickNumber() int
	// GetCurrentPlayerIdx returns the player whose turn it is to play a card.
	GetCurrentPlayerIdx() int
	// GetCurrentTrick returns the current trick (in play order).
	GetCurrentTrick() []*domain.TrickCard
	// GetForehandIdx returns the forehand index (leads the first trick).
	GetForehandIdx() int
	// GetMiddlehandIdx returns the middlehand index.
	GetMiddlehandIdx() int
	// GetRearhandIdx returns the rearhand index.
	GetRearhandIdx() int
	// GetDealerIdx returns the dealer index.
	GetDealerIdx() int

	// GetSkat は伏せてある 2 枚を返す。**最終トリックの獲得者が受け取る**ので、
	// 誰も触らない中立の札ではなく、最後まで勝敗に効く。
	GetSkat() []*domain.Card
	// GetCardPoints は playerIdx がこれまでに集めた点を返す。**多いほど不利。**
	GetCardPoints(playerIdx int) int
	// GetLoserIdx は最も点を集めてしまったプレイヤーを返す。
	// ラウンド終了前、同点、Durchmarsch のときは -1。
	GetLoserIdx() int
	// IsDurchmarsch は 1 人が全トリックを取ったか（逆転勝ち）を返す。
	IsDurchmarsch() bool
	// GetDurchmarschIdx は総取りしたプレイヤーを返す（無ければ -1）。
	GetDurchmarschIdx() int
	// GetLeadPlayerIdx returns the lead player's index.
	GetLeadPlayerIdx() int

	// GetPlayerCnt returns the player count.
	GetPlayerCnt() int
	// GetPlayer returns the i-th player.
	GetPlayer(i int) *domain.RamschPlayer
	// GetActionLog returns the round action log.
	GetActionLog() []*domain.ActionLogEntry

	// GetValidPlayIndices returns indices of legally playable cards.
	GetValidPlayIndices(playerIdx int) []int
	// GetHint returns a hint for the human player.
	GetHint() *domain.RamschHint
}
