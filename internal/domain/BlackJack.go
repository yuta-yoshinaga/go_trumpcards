package domain

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
	BJDefaultDecks = 1    // デフォルトデッキ数
)

// BJValidDeckCounts 有効なデッキ数
var BJValidDeckCounts = []int{1, 2, 4, 6, 8}

// BJMaxCpuPlayers CPUプレイヤー最大数
const BJMaxCpuPlayers = 3

// BJCpuBetAmount CPU自動ベット額
const BJCpuBetAmount = 50

// BlackJack ブラックジャッククラス
type BlackJack struct {
	trumpCards         *TrumpCards         // トランプカード
	player             *BlackJackPlayer    // プレイヤー
	dealer             *BlackJackPlayer    // ディーラー
	gameEndFlag        bool                // ゲーム終了フラグ
	phase              int                 // 現在のフェーズ
	playerHands        []*BlackJackHand    // プレイヤーハンド（スプリット対応）
	currentHandIdx     int                 // 現在操作中のハンドインデックス
	insuranceBet       int                 // インシュランスベット額
	insuranceAvailable bool                // インシュランス可能フラグ
	deckCount          int                 // デッキ数
	hintEnabled        bool                // ヒント有効フラグ
	config             BlackJackConfig     // ゲーム設定
	runningCount       int                 // ランニングカウント (Hi-Lo)
	holeCardCounted    bool                // ホールカードをカウント済みか
	deckCountChanged   bool                // デッキ数変更フラグ（シュー再構築判定用）
	cpuPlayers         []*BlackJackCpuSeat // CPUプレイヤー
}

// NewDefaultBlackJack デフォルト設定のブラックジャックを生成するファクトリ関数
func NewDefaultBlackJack() *BlackJack {
	bj := NewBlackJack(NewTrumpCards(0), NewBlackJackPlayer(), NewBlackJackPlayer())
	bj.player.SetChips(BJDefaultChips)
	bj.dealer.SetChips(BJDefaultChips)
	bj.deckCount = BJDefaultDecks
	bj.config = DefaultBlackJackConfig()
	return bj
}

// NewBlackJack コンストラクタ
func NewBlackJack(trumpCards *TrumpCards, player *BlackJackPlayer, dealer *BlackJackPlayer) *BlackJack {
	// 初期状態でもシャッフルしておく（Reset前にbetが呼ばれた場合の予測可能性を防ぐ）
	trumpCards.Shuffle()
	return &BlackJack{
		trumpCards:  trumpCards,
		player:      player,
		dealer:      dealer,
		gameEndFlag: false,
		phase:       BJPhaseBet,
		playerHands: []*BlackJackHand{NewBlackJackHand()},
		config:      DefaultBlackJackConfig(),
	}
}

// Reset ゲーム初期化（hintEnabled/deckCount/configは保持）
func (b *BlackJack) Reset() {
	b.gameEndFlag = false
	b.phase = BJPhaseBet
	b.currentHandIdx = 0
	b.insuranceBet = 0
	b.insuranceAvailable = false
	b.playerHands = []*BlackJackHand{NewBlackJackHand()}
	b.holeCardCounted = false
	// チップが最低ベット額未満ならデフォルト値にリセット
	if b.player.GetChips() < BJMinBet {
		b.player.SetChips(BJDefaultChips)
	}
	if b.dealer.GetChips() < BJMinBet {
		b.dealer.SetChips(BJDefaultChips)
	}
	// デッキ数を反映してシュー再構築判定
	if b.deckCount <= 0 {
		b.deckCount = BJDefaultDecks
	}
	needReshuffle := b.trumpCards == nil ||
		b.deckCountChanged ||
		b.trumpCards.GetRemainingCount()*4 < b.trumpCards.GetTotalCount() // 残り25%未満
	if needReshuffle {
		b.trumpCards = NewTrumpCardsWithDecks(b.deckCount, 0)
		// 山札シャッフル
		for i := 0; i < 10; i++ {
			b.trumpCards.Shuffle()
		}
		b.runningCount = 0
		b.deckCountChanged = false
	}
	// プレイヤー・ディーラー初期化
	b.player.Reset()
	b.dealer.Reset()
	// CPUプレイヤー初期化
	b.initCpuPlayers()
}

// drawCard 山札からカードを1枚引く（デッキ枯渇時はnilを返す）
func (b *BlackJack) drawCard() *Card {
	card := b.trumpCards.DrawCard()
	return card
}

// PlayerBet プレイヤーベット
func (b *BlackJack) PlayerBet(amount int) error {
	if b.phase != BJPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if amount < BJMinBet || amount%BJMinBet != 0 {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	if !b.player.SubtractChips(amount) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
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
		return ErrDeckExhausted
	}
	b.phase = BJPhaseDeal

	// CPUプレイヤーの自動ベットとカード配布
	b.cpuBetAndDeal()

	// カウンティング: プレイヤーの2枚 + ディーラーのアップカード + CPU表向きカードをカウント
	if b.config.CountingEnabled {
		for i := 0; i < b.playerHands[0].GetCardsSize(); i++ {
			b.updateRunningCount(b.playerHands[0].GetCard(i))
		}
		if b.dealer.GetCardsSize() > 0 {
			b.updateRunningCount(b.dealer.GetCard(0)) // アップカードのみ
		}
		for _, cpu := range b.cpuPlayers {
			for _, hand := range cpu.GetHands() {
				for j := 0; j < hand.GetCardsSize(); j++ {
					b.updateRunningCount(hand.GetCard(j))
				}
			}
		}
	}

	// ディーラーの表向きカード(1枚目)がエースならインシュランス可能
	if b.dealer.GetCard(0) != nil && b.dealer.GetCard(0).GetValue() == 1 {
		b.insuranceAvailable = true
		b.phase = BJPhaseInsurance
	} else {
		b.phase = BJPhaseAction
		// ナチュラルBJチェック
		b.checkNaturalBlackJack()
	}
	return nil
}

// PlayerInsurance プレイヤーインシュランス
func (b *BlackJack) PlayerInsurance() error {
	if b.phase != BJPhaseInsurance {
		return NewDomainError(ErrWrongPhase, "Insurance is not available now.")
	}
	// ベット額はBJMinBetの倍数（偶数）なので端数は発生しない
	cost := b.playerHands[0].GetBet() / 2
	if !b.player.SubtractChips(cost) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for insurance.")
	}
	b.insuranceBet = cost
	b.phase = BJPhaseAction
	b.checkNaturalBlackJack()
	return nil
}

// PlayerDeclineInsurance プレイヤーインシュランス辞退
func (b *BlackJack) PlayerDeclineInsurance() error {
	if b.phase != BJPhaseInsurance {
		return NewDomainError(ErrWrongPhase, "Insurance decline is not available now.")
	}
	b.phase = BJPhaseAction
	b.checkNaturalBlackJack()
	return nil
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
func (b *BlackJack) PlayerHit() error {
	if b.phase != BJPhaseAction {
		return NewDomainError(ErrWrongPhase, "Hit is not allowed now.")
	}
	hand := b.playerHands[b.currentHandIdx]
	if hand.IsFinished() {
		return NewDomainError(ErrHandFinished, "This hand is already finished.")
	}
	card := b.drawCard()
	if card == nil {
		return ErrDeckExhausted
	}
	hand.AddCard(card)
	b.updateRunningCount(card)
	if hand.GetScore() >= 22 {
		// バースト
		hand.SetBusted(true)
		b.advanceHand()
	}
	return nil
}

// PlayerStand プレイヤースタンド
func (b *BlackJack) PlayerStand() error {
	if b.phase != BJPhaseAction {
		return NewDomainError(ErrWrongPhase, "Stand is not allowed now.")
	}
	hand := b.playerHands[b.currentHandIdx]
	if hand.IsFinished() {
		return NewDomainError(ErrHandFinished, "This hand is already finished.")
	}
	hand.SetStood(true)
	b.advanceHand()
	return nil
}

// PlayerDoubleDown プレイヤーダブルダウン
func (b *BlackJack) PlayerDoubleDown() error {
	if b.phase != BJPhaseAction {
		return NewDomainError(ErrWrongPhase, "Double down is not allowed now.")
	}
	hand := b.playerHands[b.currentHandIdx]
	if hand.GetCardsSize() != 2 {
		return NewDomainError(ErrInvalidPlay, "Double down is only allowed with 2 cards.")
	}
	if hand.IsFinished() {
		return NewDomainError(ErrHandFinished, "This hand is already finished.")
	}
	bet := hand.GetBet()
	if !b.player.SubtractChips(bet) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for double down.")
	}
	hand.SetBet(bet * 2)
	hand.SetDoubled(true)
	// ダブルダウンは1枚だけ引いてスタンド
	card := b.drawCard()
	if card == nil {
		// デッキ枯渇: ベットと状態を元に戻す
		b.player.AddChips(bet)
		hand.SetBet(bet)
		hand.SetDoubled(false)
		return ErrDeckExhausted
	}
	hand.AddCard(card)
	b.updateRunningCount(card)
	if hand.GetScore() >= 22 {
		hand.SetBusted(true)
	} else {
		hand.SetStood(true)
	}
	b.advanceHand()
	return nil
}

// PlayerSplit プレイヤースプリット
func (b *BlackJack) PlayerSplit() error {
	if b.phase != BJPhaseAction {
		return NewDomainError(ErrWrongPhase, "Split is not allowed now.")
	}
	if len(b.playerHands) >= BJMaxHands {
		return NewDomainError(ErrInvalidPlay, "Maximum number of hands reached.")
	}
	hand := b.playerHands[b.currentHandIdx]
	if hand.IsFinished() {
		return NewDomainError(ErrHandFinished, "This hand is already finished.")
	}
	if !hand.CanSplit() {
		return NewDomainError(ErrInvalidPlay, "Split is not allowed for this hand.")
	}
	bet := hand.GetBet()
	if !b.player.SubtractChips(bet) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for split.")
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
	card1 := b.drawCard()
	if card1 == nil {
		// デッキ枯渇: 元のハンドを復元してベットを返却
		hand.Reset()
		hand.SetBet(bet)
		hand.AddCard(firstCard)
		hand.AddCard(secondCard)
		b.player.AddChips(bet)
		return ErrDeckExhausted
	}
	hand.AddCard(card1)
	b.updateRunningCount(card1)

	card2 := b.drawCard()
	if card2 == nil {
		// デッキ枯渇: 部分的なドローを元に戻し、元のハンドを復元してベットを返却
		hand.Reset()
		hand.SetBet(bet)
		hand.AddCard(firstCard)
		hand.AddCard(secondCard)
		b.player.AddChips(bet)
		// card1のランニングカウント更新を元に戻す
		b.runningCount -= hiLoValue(card1)
		return ErrDeckExhausted
	}
	newHand.AddCard(card2)
	b.updateRunningCount(card2)

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

	return nil
}

// advanceHand 次の未完了ハンドに進む。全ハンド完了ならCPUプレイ→ディーラープレイ→精算
func (b *BlackJack) advanceHand() {
	// 次の未完了ハンドを探す
	for i := 0; i < len(b.playerHands); i++ {
		if !b.playerHands[i].IsFinished() {
			b.currentHandIdx = i
			return
		}
	}
	// 全人間ハンド完了 → CPUプレイ → ディーラープレイ
	b.cpuPlay()
	b.dealerPlay()
}

// allPlayerHandsDone 全プレイヤーハンド（人間+CPU）がバーストまたはサレンダーしているか
func (b *BlackJack) allPlayerHandsDone() bool {
	for _, hand := range b.playerHands {
		if !hand.IsBusted() && !hand.IsSurrendered() {
			return false
		}
	}
	for _, cpu := range b.cpuPlayers {
		for _, hand := range cpu.GetHands() {
			if !hand.IsBusted() && !hand.IsSurrendered() {
				return false
			}
		}
	}
	return true
}

// dealerPlay ディーラーのカードドロー＆精算
func (b *BlackJack) dealerPlay() {
	if !b.allPlayerHandsDone() {
		b.DealerHit()
	} else {
		// 全ハンドバースト/サレンダーでもホールカードをカウント
		if b.config.CountingEnabled && !b.holeCardCounted && b.dealer.GetCardsSize() >= 2 {
			b.updateRunningCount(b.dealer.GetCard(1))
			b.holeCardCounted = true
		}
		b.endGame()
	}
}

// endGame ゲーム終了処理
func (b *BlackJack) endGame() {
	b.resolvePayouts()
	b.resolvePayoutsCpu()
	b.gameEndFlag = true
	b.phase = BJPhaseEnd
}

// judgeHandCore 共通ハンド勝敗判定ロジック
// handCount はスプリットBJ抑制用: handCount==1 の場合のみプレイヤーBJとして判定
func (b *BlackJack) judgeHandCore(hand *BlackJackHand, handCount int /* 1=single hand, >1=split */) GameResult {
	playerScore := hand.GetScore()
	dealerScore := b.dealer.GetScore()

	if playerScore > 21 {
		return GameResultLose
	}
	if dealerScore > 21 {
		return GameResultWin
	}
	if playerScore > dealerScore {
		return GameResultWin
	}
	if dealerScore > playerScore {
		return GameResultLose
	}

	// スコアが同じ場合、ナチュラルブラックジャックを確認
	dealerBJ := b.dealer.GetCardsSize() == 2 && dealerScore == 21
	playerBJ := hand.GetCardsSize() == 2 && playerScore == 21 && handCount == 1

	if playerBJ && !dealerBJ {
		return GameResultWin
	}
	if dealerBJ && !playerBJ {
		return GameResultLose
	}

	return GameResultDraw
}

// payoutHand 共通ハンド精算ロジック
// handCount はBJ 3:2配当判定用: handCount==1 かつ BJ の場合のみ 3:2 配当
func payoutHand(player *BlackJackPlayer, hand *BlackJackHand, handCount int /* 1=single hand, >1=split */, result GameResult) {
	bet := hand.GetBet()
	switch result {
	case GameResultWin:
		if hand.IsBlackJack() && handCount == 1 {
			player.AddChips(bet + bet*3/2)
		} else {
			player.AddChips(bet * 2)
		}
	case GameResultDraw:
		player.AddChips(bet)
	case GameResultLose:
		// 没収（何もしない、既に減算済み）
	}
}

// dealerShouldHit ディーラーがヒットすべきか判定（ソフト17ルール対応）
func (b *BlackJack) dealerShouldHit() bool {
	score := b.dealer.GetScore()
	if score < 17 {
		return true
	}
	if score == 17 && b.config.DealerHitsSoft17 && b.dealer.IsSoft() {
		return true
	}
	return false
}

// DealerHit ディーラーヒット
func (b *BlackJack) DealerHit() {
	// ホールカード（裏向きの2枚目）をカウント
	if b.config.CountingEnabled && !b.holeCardCounted && b.dealer.GetCardsSize() >= 2 {
		b.updateRunningCount(b.dealer.GetCard(1))
		b.holeCardCounted = true
	}
	for {
		if b.dealerShouldHit() {
			card := b.drawCard()
			if card == nil {
				b.DealerStand()
				break
			}
			b.dealer.AddCard(card)
			b.updateRunningCount(card)
		} else {
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
		if hand.IsSurrendered() {
			// サレンダー: 半額返却済み（PlayerSurrender内で処理）
			continue
		}
		result := b.judgeHand(hand)
		payoutHand(b.player, hand, len(b.playerHands), result)
	}
}

// judgeHand 個別ハンドの勝敗判定（人間プレイヤー用: スプリット時はBJ抑制）
func (b *BlackJack) judgeHand(hand *BlackJackHand) GameResult {
	return b.judgeHandCore(hand, len(b.playerHands))
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

// PlayerSurrender プレイヤーサレンダー（ベット半額返却して降りる）
func (b *BlackJack) PlayerSurrender() error {
	if b.phase != BJPhaseAction {
		return NewDomainError(ErrWrongPhase, "Surrender is not allowed now.")
	}
	hand := b.playerHands[b.currentHandIdx]
	if !hand.CanSurrender() {
		return NewDomainError(ErrInvalidPlay, "Surrender is not allowed for this hand.")
	}
	// 半額返却
	halfBet := hand.GetBet() / 2
	b.player.AddChips(halfBet)
	hand.SetSurrendered(true)
	b.advanceHand()
	return nil
}

// SetDeckCount デッキ数設定（BETフェーズのみ）
func (b *BlackJack) SetDeckCount(count int) error {
	if b.phase != BJPhaseBet {
		return NewDomainError(ErrWrongPhase, "Deck count can only be changed during the bet phase.")
	}
	valid := false
	for _, v := range BJValidDeckCounts {
		if v == count {
			valid = true
			break
		}
	}
	if !valid {
		return NewDomainError(ErrInvalidAmount, "Invalid deck count. Use 1, 2, 4, 6, or 8.")
	}
	if count != b.deckCount {
		b.deckCountChanged = true
	}
	b.deckCount = count
	return nil
}

// GetDeckCount デッキ数取得
func (b *BlackJack) GetDeckCount() int {
	if b.deckCount <= 0 {
		return BJDefaultDecks
	}
	return b.deckCount
}

// ToggleHint ヒント表示のON/OFF切り替え
func (b *BlackJack) ToggleHint() {
	b.hintEnabled = !b.hintEnabled
}

// IsHintEnabled ヒント有効か
func (b *BlackJack) IsHintEnabled() bool {
	return b.hintEnabled
}

// GetBasicStrategySuggestion ベーシックストラテジーによる推奨アクションを返す
func (b *BlackJack) GetBasicStrategySuggestion() BJSuggestedAction {
	if !b.hintEnabled {
		return BJSuggestNone
	}
	if b.phase != BJPhaseAction && b.phase != BJPhaseInsurance {
		return BJSuggestNone
	}
	if b.phase == BJPhaseInsurance {
		return BJSuggestDeclineInsurance
	}
	hand := b.playerHands[b.currentHandIdx]
	if hand.IsFinished() {
		return BJSuggestNone
	}
	dealerUpcard := b.dealer.GetCard(0)
	if dealerUpcard == nil {
		return BJSuggestNone
	}
	return GetBasicStrategyAction(hand, dealerUpcard, b.config.DealerHitsSoft17)
}

// GetConfig ゲーム設定取得
func (b *BlackJack) GetConfig() BlackJackConfig {
	return b.config
}

// SetConfig ゲーム設定（BETフェーズのみ）
func (b *BlackJack) SetConfig(config BlackJackConfig) error {
	if b.phase != BJPhaseBet {
		return NewDomainError(ErrWrongPhase, "Config can only be changed during the bet phase.")
	}
	if config.CpuPlayerCount < 0 || config.CpuPlayerCount > BJMaxCpuPlayers {
		return NewDomainError(ErrInvalidAmount, "CPU player count must be 0-3.")
	}
	b.config = config
	return nil
}

// hiLoValue Hi-Loカウンティングのカード値 (2-6: +1, 7-9: 0, 10/J/Q/K/A: -1)
func hiLoValue(card *Card) int {
	if card == nil {
		return 0
	}
	v := card.GetValue()
	switch {
	case v >= 2 && v <= 6:
		return 1
	case v >= 7 && v <= 9:
		return 0
	default: // 1(A), 10, 11(J), 12(Q), 13(K)
		return -1
	}
}

// updateRunningCount ランニングカウントを更新（countingEnabled時のみ）
func (b *BlackJack) updateRunningCount(card *Card) {
	if !b.config.CountingEnabled {
		return
	}
	b.runningCount += hiLoValue(card)
}

// GetRunningCount ランニングカウント取得
func (b *BlackJack) GetRunningCount() int {
	return b.runningCount
}

// GetTrueCount トゥルーカウント取得
func (b *BlackJack) GetTrueCount() float64 {
	if b.trumpCards == nil {
		return 0
	}
	remaining := b.trumpCards.GetRemainingCount()
	decksRemaining := float64(remaining) / 52.0
	if decksRemaining < 1.0 {
		decksRemaining = 1.0
	}
	return float64(b.runningCount) / decksRemaining
}

// IsCountingEnabled カウンティング有効か
func (b *BlackJack) IsCountingEnabled() bool {
	return b.config.CountingEnabled
}

// GetCpuPlayers CPUプレイヤー一覧
func (b *BlackJack) GetCpuPlayers() []*BlackJackCpuSeat {
	return b.cpuPlayers
}

// initCpuPlayers CPUプレイヤー初期化
func (b *BlackJack) initCpuPlayers() {
	count := b.config.CpuPlayerCount
	if count <= 0 {
		b.cpuPlayers = nil
		return
	}
	// 既存CPUプレイヤーを再利用（チップ引き継ぎ）
	if len(b.cpuPlayers) == count {
		for _, cpu := range b.cpuPlayers {
			cpu.Reset()
			if cpu.GetPlayer().GetChips() < BJMinBet {
				cpu.GetPlayer().SetChips(BJDefaultChips)
			}
		}
		return
	}
	b.cpuPlayers = make([]*BlackJackCpuSeat, count)
	for i := 0; i < count; i++ {
		b.cpuPlayers[i] = NewBlackJackCpuSeat()
	}
}

// cpuBetAndDeal CPUプレイヤーの自動ベットとカード配布
func (b *BlackJack) cpuBetAndDeal() {
	for _, cpu := range b.cpuPlayers {
		betAmount := BJCpuBetAmount
		if cpu.GetPlayer().GetChips() < betAmount {
			betAmount = cpu.GetPlayer().GetChips()
		}
		if betAmount < BJMinBet {
			continue
		}
		// ベット額をBJMinBetの倍数に丸める
		betAmount = (betAmount / BJMinBet) * BJMinBet
		cpu.GetPlayer().SubtractChips(betAmount)
		hand := cpu.GetHands()[0]
		hand.SetBet(betAmount)
		// カードを2枚配る
		dealFailed := false
		for i := 0; i < 2; i++ {
			card := b.drawCard()
			if card != nil {
				hand.AddCard(card)
			} else {
				dealFailed = true
			}
		}
		// 山札枯渇で必要な枚数を配れなかった場合、ベットを返却してリセット
		if dealFailed {
			cpu.GetPlayer().AddChips(betAmount)
			hand.Reset()
		}
	}
}

// cpuPlay CPUプレイヤーのベーシックストラテジープレイ
func (b *BlackJack) cpuPlay() {
	dealerUpcard := b.dealer.GetCard(0)
	if dealerUpcard == nil {
		return
	}
	for _, cpu := range b.cpuPlayers {
		b.cpuPlaySeat(cpu, dealerUpcard)
	}
}

// cpuPlaySeat 個別CPUプレイヤーのプレイ
func (b *BlackJack) cpuPlaySeat(cpu *BlackJackCpuSeat, dealerUpcard *Card) {
	handIdx := 0
	for handIdx < len(cpu.GetHands()) {
		hand := cpu.GetHands()[handIdx]
		if hand.IsFinished() || hand.GetCardsSize() == 0 {
			handIdx++
			continue
		}
		for !hand.IsFinished() {
			action := GetBasicStrategyAction(hand, dealerUpcard, b.config.DealerHitsSoft17)
			switch action {
			case BJSuggestHit:
				b.cpuHit(hand)
			case BJSuggestStand:
				hand.SetStood(true)
			case BJSuggestDouble:
				if hand.GetCardsSize() == 2 && cpu.GetPlayer().GetChips() >= hand.GetBet() {
					b.cpuDoubleDown(cpu, hand)
				} else {
					b.cpuHit(hand)
				}
			case BJSuggestDoubleStand:
				if hand.GetCardsSize() == 2 && cpu.GetPlayer().GetChips() >= hand.GetBet() {
					b.cpuDoubleDown(cpu, hand)
				} else {
					hand.SetStood(true)
				}
			case BJSuggestSplit:
				if hand.CanSplit() && len(cpu.GetHands()) < BJMaxHands && cpu.GetPlayer().GetChips() >= hand.GetBet() {
					b.cpuSplit(cpu, hand, handIdx, dealerUpcard)
					continue // cpuSplit may add hands, re-check current index
				}
				b.cpuHit(hand)
			case BJSuggestSurrender:
				if hand.CanSurrender() {
					halfBet := hand.GetBet() / 2
					cpu.GetPlayer().AddChips(halfBet)
					hand.SetSurrendered(true)
				} else {
					b.cpuHit(hand)
				}
			}
		}
		handIdx++
	}
}

// cpuHit CPUヒット
func (b *BlackJack) cpuHit(hand *BlackJackHand) {
	card := b.drawCard()
	if card == nil {
		hand.SetStood(true)
		return
	}
	hand.AddCard(card)
	b.updateRunningCount(card)
	if hand.GetScore() >= 22 {
		hand.SetBusted(true)
	}
}

// cpuDoubleDown CPUダブルダウン
func (b *BlackJack) cpuDoubleDown(cpu *BlackJackCpuSeat, hand *BlackJackHand) {
	bet := hand.GetBet()
	cpu.GetPlayer().SubtractChips(bet)
	hand.SetBet(bet * 2)
	hand.SetDoubled(true)
	card := b.drawCard()
	if card == nil {
		// デッキ枯渇: 元に戻す
		cpu.GetPlayer().AddChips(bet)
		hand.SetBet(bet)
		hand.SetDoubled(false)
		hand.SetStood(true)
		return
	}
	hand.AddCard(card)
	b.updateRunningCount(card)
	if hand.GetScore() >= 22 {
		hand.SetBusted(true)
	} else {
		hand.SetStood(true)
	}
}

// cpuSplit CPUスプリット
func (b *BlackJack) cpuSplit(cpu *BlackJackCpuSeat, hand *BlackJackHand, handIdx int, dealerUpcard *Card) {
	bet := hand.GetBet()
	cpu.GetPlayer().SubtractChips(bet)

	firstCard := hand.GetCard(0)
	secondCard := hand.GetCard(1)
	hand.Reset()
	hand.SetBet(bet)
	hand.AddCard(firstCard)

	newHand := NewBlackJackHand()
	newHand.SetBet(bet)
	newHand.AddCard(secondCard)

	// 各ハンドに1枚ずつ配る
	card1 := b.drawCard()
	if card1 == nil {
		// デッキ枯渇: 元のハンドを復元してベットを返却
		hand.Reset()
		hand.SetBet(bet)
		hand.AddCard(firstCard)
		hand.AddCard(secondCard)
		cpu.GetPlayer().AddChips(bet)
		hand.SetStood(true)
		return
	}
	hand.AddCard(card1)
	b.updateRunningCount(card1)

	card2 := b.drawCard()
	if card2 == nil {
		// デッキ枯渇: 部分的なドローを元に戻し、元のハンドを復元してベットを返却
		hand.Reset()
		hand.SetBet(bet)
		hand.AddCard(firstCard)
		hand.AddCard(secondCard)
		cpu.GetPlayer().AddChips(bet)
		// card1のランニングカウント更新を元に戻す
		b.runningCount -= hiLoValue(card1)
		hand.SetStood(true)
		return
	}
	newHand.AddCard(card2)
	b.updateRunningCount(card2)

	// 新しいハンドを挿入
	hands := cpu.GetHands()
	newHands := make([]*BlackJackHand, 0, len(hands)+1)
	newHands = append(newHands, hands[:handIdx+1]...)
	newHands = append(newHands, newHand)
	newHands = append(newHands, hands[handIdx+1:]...)
	cpu.SetHands(newHands)

	// エースのスプリットの場合、両ハンドを自動スタンド
	if firstCard.GetValue() == 1 {
		hand.SetStood(true)
		newHand.SetStood(true)
	}
}

// resolvePayoutsCpu CPUプレイヤーの精算
func (b *BlackJack) resolvePayoutsCpu() {
	for _, cpu := range b.cpuPlayers {
		for _, hand := range cpu.GetHands() {
			if hand.GetCardsSize() == 0 {
				continue
			}
			if hand.IsSurrendered() {
				continue
			}
			result := b.judgeCpuHand(hand)
			payoutHand(cpu.GetPlayer(), hand, len(cpu.GetHands()), result)
		}
	}
}

// judgeCpuHand CPU個別ハンドの勝敗判定（CPUはスプリット有無に関わらずBJ判定）
func (b *BlackJack) judgeCpuHand(hand *BlackJackHand) GameResult {
	return b.judgeHandCore(hand, 1)
}
