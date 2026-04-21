package domain

// tournamentBase groups the rebuy/addon/handCount state shared by the
// tournament-mode poker variants (Holdem, Omaha, ShortDeck, Pineapple,
// SevenCardStud). Extracted per issue #1463 so rule changes to tournament
// mechanics (e.g. a new maximum rebuy rule) can be made in one place.
//
// Usage: embed in a poker domain struct so the three fields promote onto
// the outer type. Call initTournamentState(playerCount) from the
// constructor and from any Resize that changes the player count — both
// rebuy/addon slices need to match len(players) or index accesses panic.
//
// Migration is deliberately incremental: each game can adopt the
// embedding independently because promoted-field access leaves existing
// h.handCount / h.rebuyCounts / h.addonUsed call sites unchanged. The
// JSON snapshot shape is preserved as long as the owning struct declares
// its own exported fields for HandCount/RebuyCounts/AddonUsed (they all
// currently do), so no Marshal/Unmarshal rework is required.
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
