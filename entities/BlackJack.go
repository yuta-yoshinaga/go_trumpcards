package entities

// BlackJack ブラックジャッククラス
type BlackJack struct {
	trumpCards  *TrumpCards      // トランプカード
	player      *BlackJackPlayer // プレイヤー
	dealer      *BlackJackPlayer // ディーラー
	gameEndFlag bool             // ゲーム終了フラグ
}

// NewBlackJack コンストラクタ
func NewBlackJack(trumpCards *TrumpCards, player *BlackJackPlayer, dealer *BlackJackPlayer) *BlackJack {
	return &BlackJack{
		trumpCards:  trumpCards,
		player:      player,
		dealer:      dealer,
		gameEndFlag: false,
	}
}

// Reset ゲーム初期化
func (b *BlackJack) Reset() {
	b.gameEndFlag = false
	// 山札シャッフル
	for i := 0; i < 10; i++ {
		b.trumpCards.Shuffle()
	}
	// プレイヤー・ディーラー初期化
	b.player.Reset()
	b.dealer.Reset()
	// プレイヤー・ディーラー手札を2枚づつ配る
	for i := 0; i < 2; i++ {
		b.player.AddCard(b.trumpCards.DrawCard())
		b.dealer.AddCard(b.trumpCards.DrawCard())
	}
}

// PlayerHit プレイヤーヒット
func (b *BlackJack) PlayerHit() {
	if !b.gameEndFlag {
		b.player.AddCard(b.trumpCards.DrawCard())
		if 22 <= b.player.GetScore() {
			// バーストしたので強制終了
			b.PlayerStand()
		}
	}
}

// PlayerStand プレイヤースタンド
func (b *BlackJack) PlayerStand() {
	b.DealerHit()
}

// DealerHit ディーラーヒット
func (b *BlackJack) DealerHit() {
	for {
		if b.dealer.GetScore() < 17 {
			// ディーラーは自分の手持ちのカードの合計が「17」以上になるまでヒットし続ける（カードを引き続ける）
			b.dealer.AddCard(b.trumpCards.DrawCard())
		} else {
			// ディーラーは自分の手持ちカードの合計が「17」以上になったらステイする（カードを引かない）。
			b.DealerStand()
			break
		}
	}
}

// DealerStand ディーラースタンド
func (b *BlackJack) DealerStand() {
	b.gameEndFlag = true
}

// GameJudgment ゲーム勝敗判定
func (b *BlackJack) GameJudgment() int {
	res := 0
	score1 := b.player.GetScore()
	score2 := b.dealer.GetScore()
	diff1 := 21 - score1
	diff2 := 21 - score2
	if 22 <= score1 && 22 <= score2 {
		// プレイヤー・ディーラー共にバーストしているので負け
		res = -1
	} else if 22 <= score1 && score2 <= 21 {
		// プレイヤーバーストしているので負け
		res = -1
	} else if score1 <= 21 && 22 <= score2 {
		// ディーラーバーストしているので勝ち
		res = 1
	} else {
		if diff1 == diff2 {
			// 同スコアなら引き分け
			res = 0
			if score1 == 21 && b.player.GetCardsSize() == 2 && b.dealer.GetCardsSize() != 2 {
				// プレイヤーのみがピュアブラックジャックならプレイヤーの勝ち
				res = 1
			}
		} else if diff1 < diff2 {
			// プレイヤーの方が21に近いので勝ち
			res = 1
		} else {
			// ディーラーの方が21に近いので負け
			res = -1
		}
	}
	return res
}

// GetGameEndFlag ゲーム終了フラグ
func (b *BlackJack) GetGameEndFlag() bool {
	return b.gameEndFlag
}

//GetPlayer プレイヤー
func (b *BlackJack) GetPlayer() *BlackJackPlayer {
	return b.player
}

//GetDealer ディーラー
func (b *BlackJack) GetDealer() *BlackJackPlayer {
	return b.dealer
}
