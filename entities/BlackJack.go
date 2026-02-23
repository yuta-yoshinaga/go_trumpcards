package entities

// GameResult ゲーム勝敗結果
type GameResult int

const (
	GameResultWin  GameResult = 1
	GameResultDraw GameResult = 0
	GameResultLose GameResult = -1
)

// ブラックジャックフェーズ定数
const (
	BJPhaseBet       = 1 // ベットフェーズ
	BJPhaseDeal      = 2 // ディールフェーズ
	BJPhaseInsurance = 3 // インシュランスフェーズ
	BJPhaseAction    = 4 // アクションフェーズ
	BJPhaseEnd       = 5 // 終了フェーズ
)

// ブラックジャックデフォルト値
const (
	BJDefaultChips = 1000 // デフォルトチップ
	BJMinBet       = 10   // 最低ベット額
)

// BlackJack ブラックジャッククラス
type BlackJack struct {
	trumpCards         *TrumpCards      // トランプカード
	player             *BlackJackPlayer // プレイヤー
	dealer             *BlackJackPlayer // ディーラー
	gameEndFlag        bool             // ゲーム終了フラグ
	phase              int              // 現在のフェーズ
	playerHands        []*BlackJackHand // プレイヤーハンド（スプリット対応）
	currentHandIdx     int              // 現在操作中のハンドインデックス
	insuranceBet       int              // インシュランスベット額
	insuranceAvailable bool             // インシュランス可能フラグ
}

// NewDefaultBlackJack デフォルト設定のブラックジャックを生成するファクトリ関数
func NewDefaultBlackJack() *BlackJack {
	bj := NewBlackJack(NewTrumpCards(0), NewBlackJackPlayer(), NewBlackJackPlayer())
	bj.player.SetChips(BJDefaultChips)
	bj.dealer.SetChips(BJDefaultChips)
	return bj
}

// NewBlackJack コンストラクタ
func NewBlackJack(trumpCards *TrumpCards, player *BlackJackPlayer, dealer *BlackJackPlayer) *BlackJack {
	return &BlackJack{
		trumpCards:  trumpCards,
		player:     player,
		dealer:     dealer,
		gameEndFlag: false,
		phase:      BJPhaseBet,
		playerHands: []*BlackJackHand{NewBlackJackHand()},
	}
}

// Reset ゲーム初期化
func (b *BlackJack) Reset() {
	b.gameEndFlag = false
	b.phase = BJPhaseBet
	b.currentHandIdx = 0
	b.insuranceBet = 0
	b.insuranceAvailable = false
	b.playerHands = []*BlackJackHand{NewBlackJackHand()}
	// チップが0以下ならデフォルト値にリセット
	if b.player.GetChips() <= 0 {
		b.player.SetChips(BJDefaultChips)
	}
	if b.dealer.GetChips() <= 0 {
		b.dealer.SetChips(BJDefaultChips)
	}
	// 山札シャッフル
	for i := 0; i < 10; i++ {
		b.trumpCards.Shuffle()
	}
	// プレイヤー・ディーラー初期化
	b.player.Reset()
	b.dealer.Reset()
}

// PlayerBet プレイヤーベット
func (b *BlackJack) PlayerBet(amount int) bool {
	if b.phase != BJPhaseBet {
		return false
	}
	if amount < BJMinBet {
		return false
	}
	if !b.player.SubtractChips(amount) {
		return false
	}
	b.playerHands[0].SetBet(amount)

	// カードを2枚ずつ配る
	for i := 0; i < 2; i++ {
		b.playerHands[0].AddCard(b.trumpCards.DrawCard())
		b.dealer.AddCard(b.trumpCards.DrawCard())
	}
	// プレイヤーのハンドをPlayerにも同期
	b.syncPlayerCards()

	b.phase = BJPhaseDeal

	// ディーラーの表向きカード(1枚目)がエースならインシュランス可能
	if b.dealer.GetCard(0) != nil && b.dealer.GetCard(0).GetValue() == 1 {
		b.insuranceAvailable = true
		b.phase = BJPhaseInsurance
	} else {
		b.phase = BJPhaseAction
		// ナチュラルBJチェック
		b.checkNaturalBlackJack()
	}
	return true
}

// PlayerInsurance プレイヤーインシュランス
func (b *BlackJack) PlayerInsurance() bool {
	if b.phase != BJPhaseInsurance {
		return false
	}
	cost := b.playerHands[0].GetBet() / 2
	if !b.player.SubtractChips(cost) {
		return false
	}
	b.insuranceBet = cost
	b.phase = BJPhaseAction
	b.checkNaturalBlackJack()
	return true
}

// PlayerDeclineInsurance プレイヤーインシュランス辞退
func (b *BlackJack) PlayerDeclineInsurance() {
	if b.phase != BJPhaseInsurance {
		return
	}
	b.phase = BJPhaseAction
	b.checkNaturalBlackJack()
}

// checkNaturalBlackJack ナチュラルBJチェック（ディール直後）
func (b *BlackJack) checkNaturalBlackJack() {
	dealerBJ := b.dealer.GetCardsSize() == 2 && b.dealer.GetScore() == 21
	playerBJ := b.playerHands[0].IsBlackJack()

	if dealerBJ || playerBJ {
		// ナチュラルBJがあればすぐに終了
		b.endGame()
	}
}

// PlayerHit プレイヤーヒット
func (b *BlackJack) PlayerHit() {
	if b.phase != BJPhaseAction {
		return
	}
	hand := b.playerHands[b.currentHandIdx]
	if hand.IsFinished() {
		return
	}
	hand.AddCard(b.trumpCards.DrawCard())
	b.syncPlayerCards()
	if hand.GetScore() >= 22 {
		// バースト
		hand.SetBusted(true)
		b.advanceHand()
	}
}

// PlayerStand プレイヤースタンド
func (b *BlackJack) PlayerStand() {
	if b.phase != BJPhaseAction {
		// 後方互換: phaseがBJPhaseActionでなくてもDealerHitを呼べるようにする
		if !b.gameEndFlag {
			b.DealerHit()
		}
		return
	}
	hand := b.playerHands[b.currentHandIdx]
	hand.SetStood(true)
	b.advanceHand()
}

// PlayerDoubleDown プレイヤーダブルダウン
func (b *BlackJack) PlayerDoubleDown() bool {
	if b.phase != BJPhaseAction {
		return false
	}
	hand := b.playerHands[b.currentHandIdx]
	if hand.GetCardsSize() != 2 {
		return false
	}
	if hand.IsFinished() {
		return false
	}
	bet := hand.GetBet()
	if !b.player.SubtractChips(bet) {
		return false
	}
	hand.SetBet(bet * 2)
	hand.SetDoubled(true)
	// ダブルダウンは1枚だけ引いてスタンド
	hand.AddCard(b.trumpCards.DrawCard())
	b.syncPlayerCards()
	if hand.GetScore() >= 22 {
		hand.SetBusted(true)
	} else {
		hand.SetStood(true)
	}
	b.advanceHand()
	return true
}

// PlayerSplit プレイヤースプリット
func (b *BlackJack) PlayerSplit() bool {
	if b.phase != BJPhaseAction {
		return false
	}
	hand := b.playerHands[b.currentHandIdx]
	if !hand.CanSplit() {
		return false
	}
	bet := hand.GetBet()
	if !b.player.SubtractChips(bet) {
		return false
	}

	// 2枚目のカードを取り出して新しいハンドを作る
	secondCard := hand.GetCard(1)
	// 元のハンドを1枚目だけにする
	firstCard := hand.GetCard(0)
	hand.Reset()
	hand.SetBet(bet)
	hand.AddCard(firstCard)

	// 新しいハンドを作成
	newHand := NewBlackJackHand()
	newHand.SetBet(bet)
	newHand.AddCard(secondCard)

	// 各ハンドに1枚ずつ配る
	hand.AddCard(b.trumpCards.DrawCard())
	newHand.AddCard(b.trumpCards.DrawCard())

	// 新しいハンドを挿入
	b.playerHands = append(b.playerHands[:b.currentHandIdx+1], append([]*BlackJackHand{newHand}, b.playerHands[b.currentHandIdx+1:]...)...)

	// エースのスプリットの場合、両ハンドを自動スタンド
	if firstCard.GetValue() == 1 {
		hand.SetStood(true)
		newHand.SetStood(true)
		b.syncPlayerCards()
		b.advanceHand()
	} else {
		b.syncPlayerCards()
	}

	return true
}

// advanceHand 次の未完了ハンドに進む。全ハンド完了ならディーラープレイ→精算
func (b *BlackJack) advanceHand() {
	// 次の未完了ハンドを探す
	for i := 0; i < len(b.playerHands); i++ {
		if !b.playerHands[i].IsFinished() {
			b.currentHandIdx = i
			b.syncPlayerCards()
			return
		}
	}
	// 全ハンド完了 → ディーラープレイ
	b.dealerPlay()
}

// dealerPlay ディーラーのカードドロー＆精算
func (b *BlackJack) dealerPlay() {
	// 全ハンドがバーストしていればディーラーはカードを引かない
	allBusted := true
	for _, hand := range b.playerHands {
		if !hand.IsBusted() {
			allBusted = false
			break
		}
	}
	if !allBusted {
		b.DealerHit()
	} else {
		b.endGame()
	}
}

// endGame ゲーム終了処理
func (b *BlackJack) endGame() {
	b.resolvePayouts()
	b.gameEndFlag = true
	b.phase = BJPhaseEnd
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
	b.endGame()
}

// resolvePayouts 全ハンドの精算
func (b *BlackJack) resolvePayouts() {
	dealerScore := b.dealer.GetScore()
	dealerBJ := b.dealer.GetCardsSize() == 2 && dealerScore == 21

	// インシュランスの精算
	if b.insuranceBet > 0 {
		if dealerBJ {
			// ディーラーがBJなのでインシュランス勝ち（2:1配当 = 元本+2倍）
			b.player.AddChips(b.insuranceBet * 3)
		}
		// ディーラーがBJでなければインシュランスは没収（何もしない）
	}

	for _, hand := range b.playerHands {
		bet := hand.GetBet()
		result := b.judgeHand(hand)
		switch result {
		case GameResultWin:
			// ナチュラルBJ（2枚で21、スプリットからでない場合）は3:2配当
			if hand.IsBlackJack() && len(b.playerHands) == 1 {
				// 3:2配当: ベット額 + ベット額*3/2
				b.player.AddChips(bet + bet*3/2)
			} else {
				// 通常勝利: ベット額 + ベット額
				b.player.AddChips(bet * 2)
			}
		case GameResultDraw:
			// 引き分け: ベット額返却
			b.player.AddChips(bet)
		case GameResultLose:
			// 負け: 没収（何もしない、既に減算済み）
		}
	}
}

// judgeHand 個別ハンドの勝敗判定
func (b *BlackJack) judgeHand(hand *BlackJackHand) GameResult {
	// ハンドにカードがない場合はプレイヤーのカードを使う（後方互換）
	playerScore := hand.GetScore()
	playerCardsSize := hand.GetCardsSize()
	if playerCardsSize == 0 {
		playerScore = b.player.GetScore()
		playerCardsSize = b.player.GetCardsSize()
	}
	dealerScore := b.dealer.GetScore()

	// プレイヤーがバーストしているなら負け
	if playerScore > 21 {
		return GameResultLose
	}
	// ディーラーバーストしているので勝ち
	if dealerScore > 21 {
		return GameResultWin
	}
	// プレイヤーの方が21に近いので勝ち
	if playerScore > dealerScore {
		return GameResultWin
	}
	// ディーラーの方が21に近いので負け
	if dealerScore > playerScore {
		return GameResultLose
	}

	// スコアが同じ場合、ナチュラルブラックジャックを確認
	dealerBJ := b.dealer.GetCardsSize() == 2 && dealerScore == 21
	playerBJ := playerCardsSize == 2 && playerScore == 21 && len(b.playerHands) == 1

	if playerBJ && !dealerBJ {
		return GameResultWin
	}
	if dealerBJ && !playerBJ {
		return GameResultLose
	}

	return GameResultDraw
}

// GameJudgment ゲーム勝敗判定（後方互換: ハンド0の結果を返す）
func (b *BlackJack) GameJudgment() GameResult {
	return b.GameJudgmentForHand(0)
}

// GameJudgmentForHand 指定ハンドの勝敗判定
func (b *BlackJack) GameJudgmentForHand(handIdx int) GameResult {
	if handIdx < 0 || handIdx >= len(b.playerHands) {
		return GameResultLose
	}
	return b.judgeHand(b.playerHands[handIdx])
}

// syncPlayerCards プレイヤーのcardsを現在のハンド(currentHandIdx)と同期
func (b *BlackJack) syncPlayerCards() {
	if b.currentHandIdx < len(b.playerHands) {
		hand := b.playerHands[b.currentHandIdx]
		b.player.Reset()
		for i := 0; i < hand.GetCardsSize(); i++ {
			b.player.AddCard(hand.GetCard(i))
		}
	}
}

// GetGameEndFlag ゲーム終了フラグ
func (b *BlackJack) GetGameEndFlag() bool {
	return b.gameEndFlag
}

// GetPlayer プレイヤー
func (b *BlackJack) GetPlayer() *BlackJackPlayer {
	return b.player
}

// GetDealer ディーラー
func (b *BlackJack) GetDealer() *BlackJackPlayer {
	return b.dealer
}

// GetPhase 現在のフェーズ
func (b *BlackJack) GetPhase() int {
	return b.phase
}

// GetPlayerHands プレイヤーハンド一覧
func (b *BlackJack) GetPlayerHands() []*BlackJackHand {
	return b.playerHands
}

// GetCurrentHandIdx 現在操作中のハンドインデックス
func (b *BlackJack) GetCurrentHandIdx() int {
	return b.currentHandIdx
}

// GetInsuranceBet インシュランスベット額
func (b *BlackJack) GetInsuranceBet() int {
	return b.insuranceBet
}

// IsInsuranceAvailable インシュランス可能か
func (b *BlackJack) IsInsuranceAvailable() bool {
	return b.insuranceAvailable
}

// GetTrumpCards トランプカード取得（テスト用）
func (b *BlackJack) GetTrumpCards() *TrumpCards {
	return b.trumpCards
}

// SetPhase フェーズ設定（テスト用）
func (b *BlackJack) SetPhase(phase int) {
	b.phase = phase
	b.gameEndFlag = phase == BJPhaseEnd
	if phase == BJPhaseInsurance {
		b.insuranceAvailable = true
	}
}
