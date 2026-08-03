//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// ブラックジャックフェーズ定数
const (
	BJPhaseBet            = 1 // ベットフェーズ
	BJPhaseDeal           = 2 // ディールフェーズ
	BJPhaseInsurance      = 3 // インシュランスフェーズ
	BJPhaseAction         = 4 // アクションフェーズ
	BJPhaseEnd            = 5 // 終了フェーズ
	BJPhaseEarlySurrender = 6 // アーリーサレンダーフェーズ
)

// ブラックジャックデフォルト値
const (
	BJDefaultChips      = 1000  // デフォルトチップ
	BJMinBet            = 10    // 最低ベット額
	BJMaxBet            = 10000 // 最大ベット額（オーバーフロー防止）
	BJMaxHands          = 4     // スプリットによる最大ハンド数（単一初期ハンド時）
	BJMaxInitialHands   = 3     // マルチハンド最大初期ハンド数
	BJMaxMultiHandTotal = 8     // マルチハンド時の総ハンド数上限
	BJDefaultDecks      = 1     // デフォルトデッキ数
)

// BJValidDeckCounts 有効なデッキ数
var BJValidDeckCounts = []int{1, 2, 4, 6, 8}

// BJMaxCpuPlayers CPUプレイヤー最大数
const BJMaxCpuPlayers = 3

// BJCpuBetAmount CPU自動ベット額
const BJCpuBetAmount = 50

// BlackJack ブラックジャッククラス
type BlackJack struct {
	trumpCards         *TrumpCards             // トランプカード
	player             *BlackJackPlayer        // プレイヤー
	dealer             *BlackJackPlayer        // ディーラー
	gameEndFlag        bool                    // ゲーム終了フラグ
	phase              int                     // 現在のフェーズ
	playerHands        []*BlackJackHand        // プレイヤーハンド（スプリット対応）
	currentHandIdx     int                     // 現在操作中のハンドインデックス
	insuranceBet       int                     // インシュランスベット額
	insuranceAvailable bool                    // インシュランス可能フラグ
	deckCount          int                     // デッキ数
	hintEnabled        bool                    // ヒント有効フラグ
	config             BlackJackConfig         // ゲーム設定
	variant            *BlackJackVariantConfig // バリアント設定 (nil = 標準BJ)
	runningCount       int                     // ランニングカウント
	holeCardCounted    bool                    // ホールカードをカウント済みか
	deckCountChanged   bool                    // デッキ数変更フラグ（シュー再構築判定用）
	cpuPlayers         []*BlackJackCpuSeat     // CPUプレイヤー
	perfectPairsBet    int                     // Perfect Pairsサイドベット額
	twentyOnePlus3Bet  int                     // 21+3サイドベット額
	sideBetResults     []*BJSideBetResult      // サイドベット結果
	multiHandCount     int                     // マルチハンド数（0=デフォルト1）
	actionLog          []*ActionLogEntry       // 棋譜
	bonusKeys          []string                // 当ラウンドで成立したバリアントボーナスのi18nキー (Spanish 21 等)
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

// NewSpanish21BlackJack スパニッシュ21バリアントのブラックジャックを生成するファクトリ関数
func NewSpanish21BlackJack() *BlackJack {
	variant := Spanish21Variant()
	bj := NewBlackJack(variant.DeckBuilder(BJDefaultDecks), NewBlackJackPlayer(), NewBlackJackPlayer())
	bj.player.SetChips(BJDefaultChips)
	bj.dealer.SetChips(BJDefaultChips)
	bj.deckCount = BJDefaultDecks
	bj.config = DefaultBlackJackConfig()
	bj.config.Variant = variant.Name
	bj.variant = variant
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
	b.perfectPairsBet = 0
	b.twentyOnePlus3Bet = 0
	b.sideBetResults = nil
	b.bonusKeys = nil
	b.multiHandCount = 0
	b.actionLog = nil
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
	pen := b.GetDeckPenetration()
	needReshuffle := b.trumpCards == nil ||
		b.deckCountChanged ||
		b.trumpCards.GetRemainingCount()*100 < b.trumpCards.GetTotalCount()*(100-pen) // ペネトレーション率に基づく
	if needReshuffle {
		if b.variant != nil && b.variant.DeckBuilder != nil {
			b.trumpCards = b.variant.DeckBuilder(b.deckCount)
		} else {
			b.trumpCards = NewTrumpCardsWithDecks(b.deckCount, 0)
		}
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
func (b *BlackJack) PlayerBet(amount, ppBet, t3Bet, handCount int) error {
	if b.phase != BJPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if amount < BJMinBet || amount%BJMinBet != 0 || amount > BJMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	// ハンド数バリデーション
	if handCount == 0 {
		handCount = 1
	}
	if handCount < 1 || handCount > BJMaxInitialHands {
		return NewDomainError(ErrInvalidAmount, "Invalid hand count. Must be 1-3.")
	}
	// サイドベットのバリデーション: 0 または (BJMinBet以上かつBJMinBetの倍数かつBJMaxBet以下)
	if ppBet != 0 && (ppBet < BJMinBet || ppBet%BJMinBet != 0 || ppBet > BJMaxBet) {
		return NewDomainError(ErrInvalidAmount, "Invalid Perfect Pairs bet amount.")
	}
	if t3Bet != 0 && (t3Bet < BJMinBet || t3Bet%BJMinBet != 0 || t3Bet > BJMaxBet) {
		return NewDomainError(ErrInvalidAmount, "Invalid poker-hand side bet amount.")
	}
	totalCost := amount*handCount + ppBet + t3Bet
	if !b.player.SubtractChips(totalCost) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}

	// マルチハンド用にハンドを作成
	b.playerHands = make([]*BlackJackHand, handCount)
	for i := 0; i < handCount; i++ {
		b.playerHands[i] = NewBlackJackHand()
		b.playerHands[i].SetBet(amount)
	}
	b.multiHandCount = handCount
	b.perfectPairsBet = ppBet
	b.twentyOnePlus3Bet = t3Bet

	// カードをインターリーブで配る: hand0-card1, hand1-card1, ..., dealer-card1, hand0-card2, ...
	dealFailed := false
	for round := 0; round < 2; round++ {
		for i := 0; i < handCount; i++ {
			if card := b.drawCard(); card != nil {
				b.playerHands[i].AddCard(card)
			} else {
				dealFailed = true
			}
		}
		if card := b.drawCard(); card != nil {
			b.dealer.AddCard(card)
		} else {
			dealFailed = true
		}
	}

	// 山札枯渇で必要な枚数を配れなかった場合、ベットを返却してリセット
	if dealFailed {
		b.player.AddChips(totalCost)
		for i := range b.playerHands {
			b.playerHands[i].Reset()
		}
		b.playerHands = []*BlackJackHand{NewBlackJackHand()}
		b.dealer.Reset()
		b.perfectPairsBet = 0
		b.twentyOnePlus3Bet = 0
		b.multiHandCount = 0
		return ErrDeckExhausted
	}
	b.phase = BJPhaseDeal
	b.appendLog(0, "bet", fmt.Sprintf("bet %d chips", amount), nil)

	// サイドベット判定（ハンド0のみ）
	b.evaluateSideBets()

	// CPUプレイヤーの自動ベットとカード配布
	b.cpuBetAndDeal()

	// カウンティング: 全プレイヤーハンドの2枚 + ディーラーのアップカード + CPU表向きカードをカウント
	if b.config.CountingEnabled {
		for _, hand := range b.playerHands {
			for i := 0; i < hand.GetCardsSize(); i++ {
				b.updateRunningCount(hand.GetCard(i))
			}
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
		b.cpuInsurance()
	} else if b.config.SurrenderRule == BJSurrenderEarly {
		b.phase = BJPhaseEarlySurrender
		b.cpuEarlySurrender()
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
	b.appendLog(0, "insurance", "insurance", nil)
	b.afterInsurance()
	return nil
}

// PlayerDeclineInsurance プレイヤーインシュランス辞退
func (b *BlackJack) PlayerDeclineInsurance() error {
	if b.phase != BJPhaseInsurance {
		return NewDomainError(ErrWrongPhase, "Insurance decline is not available now.")
	}
	b.appendLog(0, "insurance", "decline insurance", nil)
	b.afterInsurance()
	return nil
}

// afterInsurance インシュランス後の分岐処理
func (b *BlackJack) afterInsurance() {
	if b.config.SurrenderRule == BJSurrenderEarly {
		b.phase = BJPhaseEarlySurrender
		b.cpuEarlySurrender()
	} else {
		b.phase = BJPhaseAction
		b.checkNaturalBlackJack()
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
	b.appendLog(0, "hit", "hit", []*Card{card})
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
	b.appendLog(0, "stand", "stand", nil)
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
	if hand.IsFromSplit() && !b.config.DoubleAfterSplit {
		return NewDomainError(ErrInvalidPlay, "Double after split is not allowed.")
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
	b.appendLog(0, "doubledown", "double down", []*Card{card})
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
	maxHands := BJMaxHands
	if b.multiHandCount > 1 {
		maxHands = BJMaxMultiHandTotal
	}
	if len(b.playerHands) >= maxHands {
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
	hand.SetFromSplit(true)
	hand.AddCard(firstCard)

	// 新しいハンドを作成
	newHand := NewBlackJackHand()
	newHand.SetBet(bet)
	newHand.SetFromSplit(true)
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
		b.runningCount -= countingValue(card1, b.config.CountingSystem)
		return ErrDeckExhausted
	}
	newHand.AddCard(card2)
	b.updateRunningCount(card2)

	// 新しいハンドを挿入
	b.playerHands = append(b.playerHands[:b.currentHandIdx+1], append([]*BlackJackHand{newHand}, b.playerHands[b.currentHandIdx+1:]...)...)
	b.appendLog(0, "split", "split", nil)

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

// PlayerSurrender プレイヤーサレンダー（ベット半額返却して降りる）
func (b *BlackJack) PlayerSurrender() error {
	if b.phase != BJPhaseAction {
		return NewDomainError(ErrWrongPhase, "Surrender is not allowed now.")
	}
	if b.config.SurrenderRule == BJSurrenderNone {
		return NewDomainError(ErrInvalidPlay, "Surrender is disabled.")
	}
	hand := b.playerHands[b.currentHandIdx]
	if !hand.CanSurrender() {
		return NewDomainError(ErrInvalidPlay, "Surrender is not allowed for this hand.")
	}
	// 半額返却
	halfBet := hand.GetBet() / 2
	b.player.AddChips(halfBet)
	hand.SetSurrendered(true)
	b.appendLog(0, "surrender", "surrender", nil)
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

// GetDeckPenetration デッキペネトレーション率取得
func (b *BlackJack) GetDeckPenetration() int {
	if b.config.DeckPenetration <= 0 {
		return BJDefaultPenetration
	}
	return b.config.DeckPenetration
}

// ToggleHint ヒント表示のON/OFF切り替え
func (b *BlackJack) ToggleHint() {
	b.hintEnabled = !b.hintEnabled
}

// IsHintEnabled ヒント有効か
func (b *BlackJack) IsHintEnabled() bool {
	return b.hintEnabled
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
	if err := config.Validate(); err != nil {
		return err
	}
	// カウンティングシステム変更時はランニングカウントをリセット
	if config.CountingSystem != b.config.CountingSystem {
		b.runningCount = 0
	}
	// バリアント変更時はデッキ再構築フラグを立てる
	if config.Variant != b.config.Variant {
		b.deckCountChanged = true
		b.variant = ResolveBlackJackVariant(config.Variant)
	}
	b.config = config
	return nil
}

// GetVariant バリアント設定取得 (nil = 標準BJ)
func (b *BlackJack) GetVariant() *BlackJackVariantConfig {
	return b.variant
}

// GetCpuPlayers CPUプレイヤー一覧
func (b *BlackJack) GetCpuPlayers() []*BlackJackCpuSeat {
	return b.cpuPlayers
}

// GetMultiHandCount マルチハンド数取得
func (b *BlackJack) GetMultiHandCount() int {
	if b.multiHandCount <= 0 {
		return 1
	}
	return b.multiHandCount
}

// GetBonusKeys 当ラウンドで成立したバリアントボーナスのi18nキー一覧を返す
// (Spanish 21 の 7-7-7 / 6-7-8 / 5枚以上21 等)。ボーナスなしの場合は空。
func (b *BlackJack) GetBonusKeys() []string {
	return b.bonusKeys
}

// SetMultiHandCount マルチハンド数設定（テスト用）
func (b *BlackJack) SetMultiHandCount(count int) {
	b.multiHandCount = count
}

// SetBonusKeys 成立したバリアントボーナスのi18nキー一覧を設定（テスト用）
func (b *BlackJack) SetBonusKeys(keys []string) {
	b.bonusKeys = keys
}

// GetPerfectPairsBet Perfect Pairsベット額取得
func (b *BlackJack) GetPerfectPairsBet() int {
	return b.perfectPairsBet
}

// Get21Plus3Bet 21+3ベット額取得
func (b *BlackJack) Get21Plus3Bet() int {
	return b.twentyOnePlus3Bet
}

// CanSurrenderHand プレイヤーハンドのサレンダー可否判定
func (b *BlackJack) CanSurrenderHand(handIdx int) bool {
	if b.config.SurrenderRule == BJSurrenderNone {
		return false
	}
	if handIdx < 0 || handIdx >= len(b.playerHands) {
		return false
	}
	return b.playerHands[handIdx].CanSurrender()
}

// CanSurrenderCpuHand CPUハンドのサレンダー可否判定
func (b *BlackJack) CanSurrenderCpuHand(cpuIdx, handIdx int) bool {
	if b.config.SurrenderRule == BJSurrenderNone {
		return false
	}
	if cpuIdx < 0 || cpuIdx >= len(b.cpuPlayers) {
		return false
	}
	hands := b.cpuPlayers[cpuIdx].GetHands()
	if handIdx < 0 || handIdx >= len(hands) {
		return false
	}
	return hands[handIdx].CanSurrender()
}

// PlayerEarlySurrender プレイヤーアーリーサレンダー
func (b *BlackJack) PlayerEarlySurrender() error {
	if b.phase != BJPhaseEarlySurrender {
		return NewDomainError(ErrWrongPhase, "Early surrender is not available now.")
	}
	hand := b.playerHands[b.currentHandIdx]
	if !hand.CanSurrender() {
		return NewDomainError(ErrInvalidPlay, "Early surrender is not allowed for this hand.")
	}
	// 半額返却
	halfBet := hand.GetBet() / 2
	b.player.AddChips(halfBet)
	hand.SetSurrendered(true)
	b.appendLog(0, "surrender", "early surrender", nil)
	b.advanceEarlySurrender()
	return nil
}

// PlayerDeclineEarlySurrender プレイヤーアーリーサレンダー辞退
func (b *BlackJack) PlayerDeclineEarlySurrender() error {
	if b.phase != BJPhaseEarlySurrender {
		return NewDomainError(ErrWrongPhase, "Decline early surrender is not available now.")
	}
	b.advanceEarlySurrender()
	return nil
}

// advanceEarlySurrender アーリーサレンダーフェーズの次ハンドへ進む
func (b *BlackJack) advanceEarlySurrender() {
	// 次の未完了ハンドを探す
	for i := b.currentHandIdx + 1; i < len(b.playerHands); i++ {
		if !b.playerHands[i].IsFinished() {
			b.currentHandIdx = i
			return
		}
	}
	// 全ハンド処理完了 → アクションフェーズへ
	b.currentHandIdx = 0
	b.phase = BJPhaseAction
	b.checkNaturalBlackJack()
}

// GetActionLog 棋譜を取得する
func (b *BlackJack) GetActionLog() []*ActionLogEntry { return b.actionLog }

// appendLog 棋譜にエントリを追加する
func (b *BlackJack) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	b.actionLog = append(b.actionLog, &ActionLogEntry{
		TurnNumber: len(b.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// blackJackJSON is the JSON wire format for BlackJack.
type blackJackJSON struct {
	TrumpCards         *TrumpCards         `json:"tc"`
	Player             *BlackJackPlayer    `json:"pl"`
	Dealer             *BlackJackPlayer    `json:"dl"`
	GameEndFlag        bool                `json:"ge"`
	Phase              int                 `json:"ps"`
	PlayerHands        []*BlackJackHand    `json:"ph"`
	CurrentHandIdx     int                 `json:"ci"`
	InsuranceBet       int                 `json:"ib"`
	InsuranceAvailable bool                `json:"ia"`
	DeckCount          int                 `json:"dc"`
	HintEnabled        bool                `json:"he"`
	Config             BlackJackConfig     `json:"cf"`
	RunningCount       int                 `json:"rc"`
	HoleCardCounted    bool                `json:"hc"`
	DeckCountChanged   bool                `json:"dd"`
	CpuPlayers         []*BlackJackCpuSeat `json:"cp"`
	PerfectPairsBet    int                 `json:"pp"`
	TwentyOnePlus3Bet  int                 `json:"t3"`
	SideBetResults     []*BJSideBetResult  `json:"sb"`
	MultiHandCount     int                 `json:"mh"`
	ActionLog          []*ActionLogEntry   `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (b *BlackJack) MarshalJSON() ([]byte, error) {
	return json.Marshal(blackJackJSON{
		TrumpCards:         b.trumpCards,
		Player:             b.player,
		Dealer:             b.dealer,
		GameEndFlag:        b.gameEndFlag,
		Phase:              b.phase,
		PlayerHands:        b.playerHands,
		CurrentHandIdx:     b.currentHandIdx,
		InsuranceBet:       b.insuranceBet,
		InsuranceAvailable: b.insuranceAvailable,
		DeckCount:          b.deckCount,
		HintEnabled:        b.hintEnabled,
		Config:             b.config,
		RunningCount:       b.runningCount,
		HoleCardCounted:    b.holeCardCounted,
		DeckCountChanged:   b.deckCountChanged,
		CpuPlayers:         b.cpuPlayers,
		PerfectPairsBet:    b.perfectPairsBet,
		TwentyOnePlus3Bet:  b.twentyOnePlus3Bet,
		SideBetResults:     b.sideBetResults,
		MultiHandCount:     b.multiHandCount,
		ActionLog:          b.actionLog,
	})
}

// blackJackMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const blackJackMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (b *BlackJack) UnmarshalJSON(data []byte) error {
	var j blackJackJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.PlayerHands) > blackJackMaxSliceLen || len(j.CpuPlayers) > blackJackMaxSliceLen ||
		len(j.SideBetResults) > blackJackMaxSliceLen || len(j.ActionLog) > blackJackMaxSliceLen {
		return fmt.Errorf("blackjack: input array exceeds maximum allowed size")
	}
	b.trumpCards = j.TrumpCards
	if b.trumpCards == nil {
		b.trumpCards = NewTrumpCards(0)
	}
	b.player = j.Player
	if b.player == nil {
		b.player = NewBlackJackPlayer()
	}
	b.dealer = j.Dealer
	if b.dealer == nil {
		b.dealer = NewBlackJackPlayer()
	}
	b.gameEndFlag = j.GameEndFlag
	b.phase = j.Phase
	b.playerHands = j.PlayerHands
	if b.playerHands == nil {
		b.playerHands = []*BlackJackHand{NewBlackJackHand()}
	}
	b.currentHandIdx = j.CurrentHandIdx
	b.insuranceBet = j.InsuranceBet
	b.insuranceAvailable = j.InsuranceAvailable
	b.deckCount = j.DeckCount
	b.hintEnabled = j.HintEnabled
	b.config = j.Config
	b.variant = ResolveBlackJackVariant(b.config.Variant)
	b.runningCount = j.RunningCount
	b.holeCardCounted = j.HoleCardCounted
	b.deckCountChanged = j.DeckCountChanged
	b.cpuPlayers = j.CpuPlayers
	if b.cpuPlayers == nil {
		b.cpuPlayers = make([]*BlackJackCpuSeat, 0)
	}
	b.perfectPairsBet = j.PerfectPairsBet
	b.twentyOnePlus3Bet = j.TwentyOnePlus3Bet
	b.sideBetResults = j.SideBetResults
	if b.sideBetResults == nil {
		b.sideBetResults = make([]*BJSideBetResult, 0)
	}
	b.multiHandCount = j.MultiHandCount
	b.actionLog = j.ActionLog
	if b.actionLog == nil {
		b.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
