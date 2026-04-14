package domain

// communityCardBettingBase はコミュニティカード系ポーカー（Holdem, Omaha, ShortDeck, Pineapple）
// で共通のベッティング状態フィールドとヘルパーメソッドをまとめた構造体。
// 各バリアントに埋め込んで使用する。
type communityCardBettingBase struct {
	pot         int
	lastBet     int
	minRaise    int
	raiseCount  int
	actedFlags  []bool
	gameEndFlag bool
}

// bettingState は ExecuteBettingAction に渡す BettingState を構築する。
// ActedFlags はスライス参照を共有するため、ExecuteBettingAction 内の変更が直接反映される。
func (b *communityCardBettingBase) bettingState() *BettingState {
	return &BettingState{
		Pot:        b.pot,
		LastBet:    b.lastBet,
		MinRaise:   b.minRaise,
		RaiseCount: b.raiseCount,
		ActedFlags: b.actedFlags,
	}
}

// syncBettingState は ExecuteBettingAction 後の BettingState を書き戻す。
func (b *communityCardBettingBase) syncBettingState(state *BettingState) {
	b.pot = state.Pot
	b.lastBet = state.LastBet
	b.minRaise = state.MinRaise
	b.raiseCount = state.RaiseCount
}

// isBettingRoundComplete は全アクティブプレイヤーがアクション済みか判定する。
func (b *communityCardBettingBase) isBettingRoundComplete(players []BettingPlayer) bool {
	for i, p := range players {
		if p.GetFolded() || p.GetAllIn() {
			continue
		}
		if !b.actedFlags[i] {
			return false
		}
	}
	return true
}

// findNextActiveTurn は現在のターンから次のアクティブプレイヤーのインデックスを返す。
// 見つからない場合は -1 を返す。
func (b *communityCardBettingBase) findNextActiveTurn(currentTurn int, players []BettingPlayer) int {
	for i := 1; i <= len(players); i++ {
		next := (currentTurn + i) % len(players)
		if !players[next].GetFolded() && !players[next].GetAllIn() && !b.actedFlags[next] {
			return next
		}
	}
	return -1
}

// toBettingPlayers は BettingPlayer インターフェースのスライスに変換する汎用ヘルパー。
func toBettingPlayers[T BettingPlayer](players []T) []BettingPlayer {
	bp := make([]BettingPlayer, len(players))
	for i, p := range players {
		bp[i] = p
	}
	return bp
}
