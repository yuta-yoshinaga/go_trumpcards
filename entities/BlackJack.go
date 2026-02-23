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
	BJMaxHands     = 4    // スプリットによる最大ハンド数
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
	lastError          string           // 直前の操作エラーメッセージ（プレゼンター読み取り後クリア）
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
	// 初期状態でもシャッフルしておく（Reset前にbetが呼ばれた場合の予測可能性を防ぐ）
	trumpCards.Shuffle()
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
	// チップが最低ベット額未満ならデフォルト値にリセット
	if b.player.GetChips() < BJMinBet {
		b.player.SetChips(BJDefaultChips)
	}
	if b.dealer.GetChips() < BJMinBet {
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

// drawCard 山札からカードを1枚引く（デッキ枯渇時はnilを返す）
func (b *BlackJack) drawCard() *Card {
	card := b.trumpCards.DrawCard()
	return card
}

// PlayerBet プレイヤーベット
func (b *BlackJack) PlayerBet(amount int) bool {
	b.lastError = ""
	if b.phase != BJPhaseBet {
		b.lastError = "Bet is only allowed during the bet phase."
		return false
	}
	if amount < BJMinBet || amount%BJMinBet != 0 {
		b.lastError = "Invalid bet amount."
		return false
	}
	if !b.player.SubtractChips(amount) {
		b.lastError = "Insufficient chips."
		return false
	}
	b.playerHands[0].SetBet(amount)

	// カードを2枚ずつ配る
	dealFailed := false
	for i := 0; i < 2; i++ {
		if card := b.drawCard(); card != nil {
			b.playerHands[0].AddCard(card)
		} else {
			dealFailed = true
		}
		if card := b.drawCard(); card != nil {
			b.dealer.AddCard(card)
		} else {
			dealFailed = true
		}
	}

	// 山札枯渇で必要な枚数を配れなかった場合、ベットを返却してリセット
	if dealFailed {
		b.player.AddChips(amount)
		b.playerHands[0].Reset()
		b.dealer.Reset()
		return false
	}
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
	b.lastError = ""
	if b.phase != BJPhaseInsurance {
		b.lastError = "Insurance is not available now."
		return false
	}
	// ベット額はBJMinBetの倍数（偶数）なので端数は発生しない
	cost := b.playerHands[0].GetBet() / 2
	if !b.player.SubtractChips(cost) {
		b.lastError = "Insufficient chips for insurance."
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
	card := b.drawCard()
	if card == nil {
		return
	}
	hand.AddCard(card)
	if hand.GetScore() >= 22 {
		// バースト
		hand.SetBusted(true)
		b.advanceHand()
	}
}

// PlayerStand プレイヤースタンド
func (b *BlackJack) PlayerStand() {
	if b.phase != BJPhaseAction {
		return
	}
	hand := b.playerHands[b.currentHandIdx]
	hand.SetStood(true)
	b.advanceHand()
}

// PlayerDoubleDown プレイヤーダブルダウン
func (b *BlackJack) PlayerDoubleDown() bool {
	b.lastError = ""
	if b.phase != BJPhaseAction {
		b.lastError = "Double down is not allowed now."
		return false
	}
	hand := b.playerHands[b.currentHandIdx]
	if hand.GetCardsSize() != 2 {
		b.lastError = "Double down is only allowed with 2 cards."
		return false
	}
	if hand.IsFinished() {
		b.lastError = "This hand is already finished."
		return false
	}
	bet := hand.GetBet()
	if !b.player.SubtractChips(bet) {
		b.lastError = "Insufficient chips for double down."
		return false
	}
	hand.SetBet(bet * 2)
	hand.SetDoubled(true)
	// ダブルダウンは1枚だけ引いてスタンド
	if card := b.drawCard(); card != nil {
		hand.AddCard(card)
	}
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
	b.lastError = ""
	if b.phase != BJPhaseAction {
		b.lastError = "Split is not allowed now."
		return false
	}
	if len(b.playerHands) >= BJMaxHands {
		b.lastError = "Maximum number of hands reached."
		return false
	}
	hand := b.playerHands[b.currentHandIdx]
	if !hand.CanSplit() {
		b.lastError = "Split is not allowed for this hand."
		return false
	}
	bet := hand.GetBet()
	if !b.player.SubtractChips(bet) {
		b.lastError = "Insufficient chips for split."
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

	// 各ハンドに1枚ずつ配る（山札枯渇時は自動スタンド）
	if card := b.drawCard(); card != nil {
		hand.AddCard(card)
	} else {
		hand.SetStood(true)
	}
	if card := b.drawCard(); card != nil {
		newHand.AddCard(card)
	} else {
		newHand.SetStood(true)
	}

	// 新しいハンドを挿入
	b.playerHands = append(b.playerHands[:b.currentHandIdx+1], append([]*BlackJackHand{newHand}, b.playerHands[b.currentHandIdx+1:]...)...)

	// エースのスプリットの場合、両ハンドを自動スタンド
	if firstCard.GetValue() == 1 {
		hand.SetStood(true)
		newHand.SetStood(true)
	}

	// 全ハンドが完了している場合（エーススプリットまたは山札枯渇）、次へ進む
	allFinished := true
	for _, h := range b.playerHands {
		if !h.IsFinished() {
			allFinished = false
			break
		}
	}
	if allFinished {
		b.advanceHand()
	}

	return true
}

// advanceHand 次の未完了ハンドに進む。全ハンド完了ならディーラープレイ→精算
func (b *BlackJack) advanceHand() {
	// 次の未完了ハンドを探す
	for i := 0; i < len(b.playerHands); i++ {
		if !b.playerHands[i].IsFinished() {
			b.currentHandIdx = i
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
			card := b.drawCard()
			if card == nil {
				b.DealerStand()
				break
			}
			b.dealer.AddCard(card)
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
				// 3:2配当: ベット額 + ベット額*3/2（ベット額はBJMinBetの倍数なので端数なし）
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
	playerScore := hand.GetScore()
	playerCardsSize := hand.GetCardsSize()
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

// GetLastError 直前の操作エラーメッセージを取得しクリアする
func (b *BlackJack) GetLastError() string {
	msg := b.lastError
	b.lastError = ""
	return msg
}

// SetPlayerHands プレイヤーハンド設定（テスト用）
func (b *BlackJack) SetPlayerHands(hands []*BlackJackHand) {
	b.playerHands = hands
}

// SetPhase フェーズ設定（テスト用）
func (b *BlackJack) SetPhase(phase int) {
	b.phase = phase
	b.gameEndFlag = phase == BJPhaseEnd
	if phase == BJPhaseInsurance {
		b.insuranceAvailable = true
	}
}
