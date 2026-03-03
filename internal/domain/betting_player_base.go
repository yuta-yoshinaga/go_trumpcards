package domain

// bettingPlayerBase ベッティング系ゲームの共通プレイヤー状態
// PokerPlayer と HoldemPlayer で共有する。
type bettingPlayerBase struct {
	handRank   int  // ハンドランク
	folded     bool // フォールド済
	allIn      bool // オールイン済
	currentBet int  // 現ラウンドベット額
}

// GetHandRank ハンドランク取得
func (b *bettingPlayerBase) GetHandRank() int { return b.handRank }

// SetHandRank ハンドランク設定
func (b *bettingPlayerBase) SetHandRank(rank int) { b.handRank = rank }

// GetFolded フォールド状態取得
func (b *bettingPlayerBase) GetFolded() bool { return b.folded }

// SetFolded フォールド状態設定
func (b *bettingPlayerBase) SetFolded(folded bool) { b.folded = folded }

// GetAllIn オールイン状態取得
func (b *bettingPlayerBase) GetAllIn() bool { return b.allIn }

// SetAllIn オールイン状態設定
func (b *bettingPlayerBase) SetAllIn(allIn bool) { b.allIn = allIn }

// GetCurrentBet 現ラウンドベット取得
func (b *bettingPlayerBase) GetCurrentBet() int { return b.currentBet }

// SetCurrentBet 現ラウンドベット設定
func (b *bettingPlayerBase) SetCurrentBet(bet int) { b.currentBet = bet }
