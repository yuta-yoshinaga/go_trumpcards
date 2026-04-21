package domain

// tournamentBase holds the rebuy/addon/handCount state shared by tournament-mode poker variants.
type tournamentBase struct {
	rebuyCounts []int  // プレイヤーごとのリバイ回数
	addonUsed   []bool // プレイヤーごとのアドオン使用フラグ
	handCount   int    // ハンド数 (トーナメントモード用)
}

// initTournamentState reinitialises the tournament slices for playerCount.
// Safe to call on a zero-value tournamentBase; resets handCount to 0.
func (t *tournamentBase) initTournamentState(playerCount int) {
	t.rebuyCounts = make([]int, playerCount)
	t.addonUsed = make([]bool, playerCount)
	t.handCount = 0
}
