package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// WizardPlayerCnt ウィザードプレイヤー数
const WizardPlayerCnt = 4

// Wizard ローカルデッキ定数。
// 標準52枚デッキ (design 1..4) に加え、ウィザード札4枚とジェスター札4枚を
// 使う60枚デッキ。Card.go の共有定数は変更せず、ここでローカルに定義する。
const (
	// WizardDesignWizard ウィザード札 (最強)
	WizardDesignWizard = 5
	// WizardDesignJester ジェスター札 (最弱)
	WizardDesignJester = 6
	// WizardDeckSize 60枚 (52 + 4 ウィザード + 4 ジェスター)
	WizardDeckSize = 60
	// WizardTotalRounds 総ラウンド数 (60 / WizardPlayerCnt = 15)
	WizardTotalRounds = WizardDeckSize / WizardPlayerCnt
)

// WizardPhase ゲームフェーズ
type WizardPhase int

// Wizardのフェーズ定数
const (
	// WizardPhaseBid ビッドフェーズ
	WizardPhaseBid WizardPhase = 0
	// WizardPhasePlay トリックプレイフェーズ
	WizardPhasePlay WizardPhase = 1
	// WizardPhaseTrickEnd トリック終了フェーズ
	WizardPhaseTrickEnd WizardPhase = 2
	// WizardPhaseRoundEnd ラウンド終了フェーズ
	WizardPhaseRoundEnd WizardPhase = 3
	// WizardPhaseGameEnd ゲーム終了フェーズ
	WizardPhaseGameEnd WizardPhase = 4
)

// WizardHint ヒント情報
type WizardHint struct {
	CardIndex *int   // 推奨カードインデックス (ビッド時nil)
	Bid       *int   // 推奨ビッド値 (プレイ時nil)
	Reason    string // ヒント理由キー
}

// Wizard ウィザードゲームクラス
type Wizard struct {
	deck             []*Card // 山札 (60枚から配り、末尾からpopする)
	players          []*WizardPlayer
	config           WizardConfig
	phase            WizardPhase
	roundNumber      int
	totalRounds      int
	handSize         int // 現在のラウンドの手札枚数 (= roundNumber)
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	bidPlayerIdx     int
	dealerIdx        int
	trumpCard        *Card // めくった切り札カード (nil = 山札が尽きた最終ラウンド)
	trumpSuit        int   // 切り札スート (-1 = 切り札なし)
	gameEndFlag      bool
	winnerIdx        int
	actionLog        []*ActionLogEntry
}

// NewWizard コンストラクタ
func NewWizard(players []*WizardPlayer, config WizardConfig) *Wizard {
	return &Wizard{
		players:   players,
		config:    config,
		winnerIdx: -1,
		trumpSuit: -1,
	}
}

// NewDefaultWizard returns Wizard with the standard 4-player setup (1 human, 3 CPU)
// and DefaultWizardConfig. Used as the single source of truth for CUI, Web, and Worker
// construction sites.
func NewDefaultWizard() *Wizard {
	players := []*WizardPlayer{
		NewWizardPlayer(true),
		NewWizardPlayer(false),
		NewWizardPlayer(false),
		NewWizardPlayer(false),
	}
	return NewWizard(players, DefaultWizardConfig())
}

// buildWizardDeck 60枚のウィザードデッキを組み立てる (52 + 4 ウィザード + 4 ジェスター)。
func buildWizardDeck() []*Card {
	deck := make([]*Card, 0, WizardDeckSize)
	for d := CardDesignSpade; d <= CardDesignDiamond; d++ {
		for v := 1; v <= 13; v++ {
			deck = append(deck, NewCard(d, v, false))
		}
	}
	for i := 1; i <= 4; i++ {
		deck = append(deck, NewCard(WizardDesignWizard, i, false))
	}
	for i := 1; i <= 4; i++ {
		deck = append(deck, NewCard(WizardDesignJester, i, false))
	}
	return deck
}

// Reset ゲーム初期化
func (o *Wizard) Reset() {
	o.gameEndFlag = false
	o.winnerIdx = -1
	o.dealerIdx = 0
	o.totalRounds = o.calcTotalRounds()
	o.roundNumber = 1
	o.trickNumber = 0
	o.currentTrick = nil
	o.leadPlayerIdx = -1
	o.currentPlayerIdx = -1
	o.actionLog = nil

	for _, p := range o.players {
		p.bid = -1
		p.roundScore = 0
		p.cumulativeScore = 0
		p.tricksTaken = nil
		p.Reset()
		p.SetIsFinished(false)
	}

	o.handSize = o.handSizeForRound(o.roundNumber)
	o.deal()
	o.phase = WizardPhaseBid
	o.bidPlayerIdx = (o.dealerIdx + 1) % WizardPlayerCnt
}

// NextRound 次のラウンドを開始する
func (o *Wizard) NextRound() {
	if o.phase != WizardPhaseRoundEnd {
		return
	}

	o.roundNumber++
	o.dealerIdx = (o.dealerIdx + 1) % WizardPlayerCnt
	o.trickNumber = 0
	o.currentTrick = nil
	o.leadPlayerIdx = -1
	o.currentPlayerIdx = -1

	for _, p := range o.players {
		p.ResetRound()
	}

	o.handSize = o.handSizeForRound(o.roundNumber)
	o.deal()
	o.phase = WizardPhaseBid
	o.bidPlayerIdx = (o.dealerIdx + 1) % WizardPlayerCnt
}

// PlayerBid 人間プレイヤーがビッドする
func (o *Wizard) PlayerBid(bid int) error {
	if o.gameEndFlag {
		return ErrGameEnded
	}
	if o.phase != WizardPhaseBid {
		return ErrWrongPhase
	}

	humanIdx := o.findHumanIdx()
	if humanIdx < 0 {
		return ErrNotHumanTurn
	}
	if o.bidPlayerIdx != humanIdx {
		return ErrNotHumanTurn
	}
	if bid < 0 || bid > o.handSize {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("ビッドは0〜%dで指定してください", o.handSize))
	}
	// Wizardはフックルール（合計制限）を持たない: 合計ビッド≠トリック数を許容する。

	o.players[humanIdx].SetBid(bid)
	o.appendLog(humanIdx, "bid", fmt.Sprintf("%s bids %d", o.playerName(humanIdx), bid), nil)

	o.advanceBid()
	return nil
}

// CpuBid 現在のビッドプレイヤーがCPUの場合にビッドする
func (o *Wizard) CpuBid() {
	if o.gameEndFlag || o.phase != WizardPhaseBid {
		return
	}
	if o.bidPlayerIdx < 0 || o.bidPlayerIdx >= WizardPlayerCnt {
		return
	}
	if o.players[o.bidPlayerIdx].GetIsHuman() {
		return
	}

	bid := o.cpuSelectBid(o.bidPlayerIdx)
	o.players[o.bidPlayerIdx].SetBid(bid)
	o.appendLog(o.bidPlayerIdx, "bid", fmt.Sprintf("%s bids %d", o.playerName(o.bidPlayerIdx), bid), nil)

	o.advanceBid()
}

// PlayerPlay 人間プレイヤーがカードをプレイする
func (o *Wizard) PlayerPlay(cardIndex int) error {
	if o.gameEndFlag {
		return ErrGameEnded
	}
	if o.phase != WizardPhasePlay {
		return ErrWrongPhase
	}
	if !o.players[o.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}

	player := o.players[o.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}

	card := player.GetCard(cardIndex)
	if err := o.validatePlay(o.currentPlayerIdx, card); err != nil {
		return err
	}

	played := player.RemoveCard(cardIndex)
	o.playCard(o.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番がCPUの場合に1ターン実行
func (o *Wizard) CpuPlay() {
	if o.gameEndFlag || o.phase != WizardPhasePlay {
		return
	}
	if o.players[o.currentPlayerIdx].GetIsHuman() {
		return
	}

	player := o.players[o.currentPlayerIdx]
	cardIdx := o.cpuSelectPlayCard(o.currentPlayerIdx)
	played := player.RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	o.playCard(o.currentPlayerIdx, played)
}

// ResolveTrick トリックを解決して勝者を決定する
func (o *Wizard) ResolveTrick() {
	if o.phase != WizardPhaseTrickEnd || len(o.currentTrick) != WizardPlayerCnt {
		return
	}

	winnerIdx := o.trickWinner()
	trickCards := make([]*Card, len(o.currentTrick))
	for i, tc := range o.currentTrick {
		trickCards[i] = tc.Card
	}

	o.players[winnerIdx].AddTrick(trickCards)

	winnerName := o.playerName(winnerIdx)
	o.appendLog(winnerIdx, "trick_win", fmt.Sprintf("%s wins trick %d", winnerName, o.trickNumber), trickCards)

	o.leadPlayerIdx = winnerIdx

	if o.trickNumber >= o.handSize {
		o.phase = WizardPhaseRoundEnd
	} else {
		o.phase = WizardPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する
func (o *Wizard) NextTrick() {
	if o.phase != WizardPhaseTrickEnd {
		return
	}
	o.currentTrick = nil
	o.currentPlayerIdx = o.leadPlayerIdx
	o.trickNumber++
	o.phase = WizardPhasePlay
}

// ScoreRound ラウンドのスコアを確定し、ゲーム終了判定を行う
func (o *Wizard) ScoreRound() {
	if o.phase != WizardPhaseRoundEnd {
		return
	}

	for i := range WizardPlayerCnt {
		p := o.players[i]
		tricks := p.GetTrickCount()
		bid := p.GetBid()

		if tricks == bid {
			// ビッド的中: 20 + 10×bid ポイント
			p.roundScore = 20 + 10*bid
			o.appendLog(i, "bid_success", fmt.Sprintf("%s bid %d, took %d: +%d",
				o.playerName(i), bid, tricks, p.roundScore), nil)
		} else {
			// 外れ: -10×|トリック数 - ビッド|
			diff := tricks - bid
			if diff < 0 {
				diff = -diff
			}
			p.roundScore = -10 * diff
			o.appendLog(i, "bid_fail", fmt.Sprintf("%s bid %d, took %d: %d",
				o.playerName(i), bid, tricks, p.roundScore), nil)
		}
	}

	// 累積スコアに加算
	for i := range WizardPlayerCnt {
		o.players[i].CommitRoundScore()
	}

	// スコアログ
	for i := range WizardPlayerCnt {
		o.appendLog(i, "cumulative_score", fmt.Sprintf("%s: total=%d",
			o.playerName(i), o.players[i].cumulativeScore), nil)
	}

	// ゲーム終了判定
	if o.roundNumber >= o.totalRounds {
		o.gameEndFlag = true
		o.phase = WizardPhaseGameEnd
		o.determineWinner()
	}
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (o *Wizard) GetPhase() WizardPhase { return o.phase }

// SetPhase フェーズ設定 (テスト用)
func (o *Wizard) SetPhase(phase WizardPhase) { o.phase = phase }

// GetRoundNumber 現在のラウンド番号取得
func (o *Wizard) GetRoundNumber() int { return o.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (o *Wizard) SetRoundNumber(n int) { o.roundNumber = n }

// GetTotalRounds 総ラウンド数取得
func (o *Wizard) GetTotalRounds() int { return o.totalRounds }

// SetTotalRounds 総ラウンド数設定 (テスト用)
func (o *Wizard) SetTotalRounds(n int) { o.totalRounds = n }

// GetHandSize 現在のラウンドの手札枚数取得
func (o *Wizard) GetHandSize() int { return o.handSize }

// SetHandSize 手札枚数設定 (テスト用)
func (o *Wizard) SetHandSize(n int) { o.handSize = n }

// GetTrickNumber 現在のトリック番号取得
func (o *Wizard) GetTrickNumber() int { return o.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (o *Wizard) SetTrickNumber(n int) { o.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (o *Wizard) GetCurrentPlayerIdx() int { return o.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (o *Wizard) SetCurrentPlayerIdx(idx int) { o.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (o *Wizard) GetCurrentTrick() []*TrickCard { return o.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (o *Wizard) SetCurrentTrick(trick []*TrickCard) { o.currentTrick = trick }

// GetTrumpCard 切り札カード取得
func (o *Wizard) GetTrumpCard() *Card { return o.trumpCard }

// GetTrumpSuit 切り札スート取得 (-1 = 切り札なし)
func (o *Wizard) GetTrumpSuit() int { return o.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (o *Wizard) SetTrumpSuit(suit int) { o.trumpSuit = suit }

// GetDealerIdx ディーラーインデックス取得
func (o *Wizard) GetDealerIdx() int { return o.dealerIdx }

// SetDealerIdx ディーラーインデックス設定 (テスト用)
func (o *Wizard) SetDealerIdx(idx int) { o.dealerIdx = idx }

// GetGameEndFlag ゲーム終了フラグ取得
func (o *Wizard) GetGameEndFlag() bool { return o.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得 (-1 = 未確定)
func (o *Wizard) GetWinnerIdx() int { return o.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (o *Wizard) GetPlayerCnt() int { return len(o.players) }

// GetPlayer プレイヤー取得
func (o *Wizard) GetPlayer(i int) *WizardPlayer {
	if i < 0 || i >= len(o.players) {
		return nil
	}
	return o.players[i]
}

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (o *Wizard) GetLeadPlayerIdx() int { return o.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (o *Wizard) SetLeadPlayerIdx(idx int) { o.leadPlayerIdx = idx }

// GetBidPlayerIdx ビッドプレイヤーインデックス取得
func (o *Wizard) GetBidPlayerIdx() int { return o.bidPlayerIdx }

// SetBidPlayerIdx ビッドプレイヤーインデックス設定 (テスト用)
func (o *Wizard) SetBidPlayerIdx(idx int) { o.bidPlayerIdx = idx }

// IsHumanTurn 現在の手番が人間かどうか
func (o *Wizard) IsHumanTurn() bool {
	if o.currentPlayerIdx < 0 || o.currentPlayerIdx >= len(o.players) {
		return false
	}
	return o.players[o.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn 現在のビッド手番が人間かどうか
func (o *Wizard) IsHumanBidTurn() bool {
	if o.bidPlayerIdx < 0 || o.bidPlayerIdx >= len(o.players) {
		return false
	}
	return o.players[o.bidPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (o *Wizard) GetConfig() WizardConfig { return o.config }

// SetConfig 設定変更
func (o *Wizard) SetConfig(cfg WizardConfig) { o.config = cfg }

// GetActionLog 棋譜取得
func (o *Wizard) GetActionLog() []*ActionLogEntry { return o.actionLog }

// GetRestrictedBid ディーラーのビッド制限値を返す。
// Wizardはフックルールを持たないため常に -1 (制限なし)。
// トリックテイカー共通インタフェース/DTO との互換のために残している。
func (o *Wizard) GetRestrictedBid() int { return -1 }

// GetValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (o *Wizard) GetValidPlayIndices(playerIdx int) []int {
	return o.getValidPlayIndices(playerIdx)
}

// --- Private methods ---

// findHumanIdx 人間プレイヤーのインデックスを返す (-1=なし)
func (o *Wizard) findHumanIdx() int {
	for i, p := range o.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// calcTotalRounds 総ラウンド数を計算する (60枚 / 4人 = 15ラウンド、昇順のみ)。
func (o *Wizard) calcTotalRounds() int {
	return WizardTotalRounds
}

// handSizeForRound ラウンド番号から手札枚数を算出する。
// Wizardではラウンド r で各プレイヤーに r 枚配る (昇順のみ)。
func (o *Wizard) handSizeForRound(round int) int {
	return round
}

// popDeck 山札の末尾から1枚引く (空なら nil)。
func (o *Wizard) popDeck() *Card {
	if len(o.deck) == 0 {
		return nil
	}
	card := o.deck[len(o.deck)-1]
	o.deck = o.deck[:len(o.deck)-1]
	return card
}

// deal カードを配り、切り札を公開する
func (o *Wizard) deal() {
	o.deck = buildWizardDeck()
	rand.Shuffle(len(o.deck), func(i, j int) {
		o.deck[i], o.deck[j] = o.deck[j], o.deck[i]
	})

	// 各プレイヤーにhandSize枚配る (ディーラーの左から)
	for range o.handSize {
		for j := range WizardPlayerCnt {
			idx := (o.dealerIdx + 1 + j) % WizardPlayerCnt
			card := o.popDeck()
			if card != nil {
				o.players[idx].AddCard(card)
			}
		}
	}

	// 残りの山札から1枚めくって切り札を決定 (最終ラウンドは山札が尽き nil)
	o.trumpCard = o.popDeck()
	o.trumpSuit = o.ComputeTrumpSuit(o.trumpCard)

	o.sortAllHands()
}

// ComputeTrumpSuit めくった切り札カードから切り札スートを決定する。
//
//   - 通常スート (design 1..4) ⇒ そのスートが切り札。
//   - ジェスター (design 6)   ⇒ 切り札なし (-1)。
//   - ウィザード (design 5)   ⇒ 本来はディーラーが切り札を選ぶ。UIフェーズの
//     追加を避けるため、ディーラーが最も多く持つスート (同数なら最小スート
//     インデックス) を自動選択する。スート札を持たない場合は切り札なし (-1)。
//   - nil (最終ラウンドで山札が尽きた)      ⇒ 切り札なし (-1)。
func (o *Wizard) ComputeTrumpSuit(flipped *Card) int {
	if flipped == nil {
		return -1
	}
	switch flipped.GetDesign() {
	case WizardDesignJester:
		return -1
	case WizardDesignWizard:
		return o.dealerMostCommonSuit()
	default:
		return flipped.GetDesign()
	}
}

// dealerMostCommonSuit ディーラーが最も多く持つスート (design 1..4) を返す。
// 同数の場合は最小スートインデックス、スート札が無ければ -1。
func (o *Wizard) dealerMostCommonSuit() int {
	if o.dealerIdx < 0 || o.dealerIdx >= len(o.players) {
		return -1
	}
	dealer := o.players[o.dealerIdx]
	counts := make(map[int]int)
	for i := 0; i < dealer.GetCardsSize(); i++ {
		d := dealer.GetCard(i).GetDesign()
		if d >= CardDesignSpade && d <= CardDesignDiamond {
			counts[d]++
		}
	}
	best := -1
	bestCount := 0
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if counts[suit] > bestCount {
			bestCount = counts[suit]
			best = suit
		}
	}
	return best
}

// advanceBid ビッドプレイヤーを次に進める
func (o *Wizard) advanceBid() {
	// 次のビッドプレイヤーへ（ディーラーの左から時計回り、ディーラーが最後）
	nextIdx := (o.bidPlayerIdx + 1) % WizardPlayerCnt
	bidCount := 0
	for _, p := range o.players {
		if p.GetBid() >= 0 {
			bidCount++
		}
	}
	if bidCount >= WizardPlayerCnt {
		// 全員ビッド完了 → プレイフェーズへ
		o.phase = WizardPhasePlay
		o.startPlayPhase()
	} else {
		o.bidPlayerIdx = nextIdx
	}
}

// startPlayPhase プレイフェーズ開始: ディーラーの左がリード
func (o *Wizard) startPlayPhase() {
	o.trickNumber = 1
	o.currentTrick = nil
	o.leadPlayerIdx = (o.dealerIdx + 1) % WizardPlayerCnt
	o.currentPlayerIdx = o.leadPlayerIdx
}

// playCard カードをプレイする共通処理
func (o *Wizard) playCard(playerIdx int, card *Card) {
	o.currentTrick = append(o.currentTrick, &TrickCard{
		PlayerIdx: playerIdx,
		Card:      card,
	})

	o.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", o.playerName(playerIdx), wizardCardStr(card)), []*Card{card})

	if len(o.currentTrick) == WizardPlayerCnt {
		o.phase = WizardPhaseTrickEnd
	} else {
		o.currentPlayerIdx = (o.currentPlayerIdx + 1) % WizardPlayerCnt
	}
}

// leadSuitOfTrick 現在のトリックのリードスートを返す (-1 = リードスートなし)。
// 先頭から順に見て、ウィザードが最初に出ていればリードスートなし、
// ジェスターは読み飛ばし、最初のスート札 (design 1..4) をリードスートとする。
func (o *Wizard) leadSuitOfTrick() int {
	return WizardLeadSuitOfTrick(o.currentTrick)
}

// WizardLeadSuitOfTrick はトリックのリードスートを返す (-1 = 未確定)。
// ウィザードがリードすればスートは立たず、ジェスターは読み飛ばす。
func WizardLeadSuitOfTrick(trick []*TrickCard) int {
	for _, tc := range trick {
		d := tc.Card.GetDesign()
		if d == WizardDesignWizard {
			return -1
		}
		if d == WizardDesignJester {
			continue
		}
		return d
	}
	return -1
}

// validatePlay カードのプレイが有効か検証する
func (o *Wizard) validatePlay(playerIdx int, card *Card) error {
	if WizardIsLegalPlay(card, o.currentTrick, o.players[playerIdx]) {
		return nil
	}
	return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
}

// WizardIsLegalPlay は card を現在のトリックに出せるかを返す。
//
// リードなら何でも出せる。ウィザード/ジェスターはフォロー義務を免除される。
// リードスートが確定していて、そのスートをまだ持っているなら従わねばならない。
//
// **判定はここ 1 箇所に置く。**presenter が「出せる札」を印付けするのに
// 同じ規則を書き写すと、片方だけ直したときに嘘の案内になる (#4927)。
func WizardIsLegalPlay(card *Card, trick []*TrickCard, hand *WizardPlayer) bool {
	if len(trick) == 0 {
		// リード: ウィザード/ジェスターを含め何でもリード可能。
		return true
	}

	d := card.GetDesign()
	// ウィザード/ジェスターはいつでもプレイ可能 (フォロー義務の免除)。
	if d == WizardDesignWizard || d == WizardDesignJester {
		return true
	}

	leadSuit := WizardLeadSuitOfTrick(trick)
	if leadSuit < 0 {
		// リードスートが未確定 (ウィザードがリード、または全てジェスター)。
		return true
	}
	if d == leadSuit {
		return true
	}
	// リードスートを持っていなければ何を出してもよい。
	for i := 0; i < hand.GetCardsSize(); i++ {
		if c := hand.GetCard(i); c != nil && c.GetDesign() == leadSuit {
			return false
		}
	}
	return true
}

// trickWinner トリックの勝者を決定する
func (o *Wizard) trickWinner() int {
	if len(o.currentTrick) == 0 {
		return 0
	}

	// 1. ウィザードが含まれていれば、最初に出されたウィザードが勝つ。
	for _, tc := range o.currentTrick {
		if tc.Card.GetDesign() == WizardDesignWizard {
			return tc.PlayerIdx
		}
	}

	// 2. リードスート = ウィザードでもジェスターでもない最初の札のスート。
	//    全てジェスターなら最初の札が勝つ。
	leadSuit := -1
	for _, tc := range o.currentTrick {
		d := tc.Card.GetDesign()
		if d != WizardDesignWizard && d != WizardDesignJester {
			leadSuit = d
			break
		}
	}
	if leadSuit < 0 {
		return o.currentTrick[0].PlayerIdx
	}

	// 3. 切り札 (design==trumpSuit) が最も高いものが勝つ。切り札が無ければ
	//    リードスートの最も高いものが勝つ。ジェスターは決して勝たない。
	winnerIdx := -1
	winnerValue := -1
	winnerIsTrump := false
	for _, tc := range o.currentTrick {
		d := tc.Card.GetDesign()
		if d == WizardDesignJester {
			continue
		}
		isTrump := o.trumpSuit >= 0 && d == o.trumpSuit
		isLead := d == leadSuit
		if !isTrump && !isLead {
			continue
		}
		if winnerIdx < 0 {
			winnerIdx = tc.PlayerIdx
			winnerValue = tc.Card.GetValue()
			winnerIsTrump = isTrump
			continue
		}
		if isTrump && !winnerIsTrump {
			winnerIdx = tc.PlayerIdx
			winnerValue = tc.Card.GetValue()
			winnerIsTrump = true
		} else if isTrump == winnerIsTrump && tc.Card.GetValue() > winnerValue {
			winnerIdx = tc.PlayerIdx
			winnerValue = tc.Card.GetValue()
		}
	}
	if winnerIdx < 0 {
		return o.currentTrick[0].PlayerIdx
	}
	return winnerIdx
}

// determineWinner 最終的な勝者を決定する
func (o *Wizard) determineWinner() {
	maxScore := o.players[0].cumulativeScore
	o.winnerIdx = 0
	for i := 1; i < WizardPlayerCnt; i++ {
		if o.players[i].cumulativeScore > maxScore {
			maxScore = o.players[i].cumulativeScore
			o.winnerIdx = i
		}
	}
	o.appendLog(-1, "game_end", fmt.Sprintf("%s wins the game!", o.playerName(o.winnerIdx)), nil)
}

// sortAllHands 全プレイヤーの手札をソートする
func (o *Wizard) sortAllHands() {
	for _, p := range o.players {
		wizardSortHand(p)
	}
}

// wizardSortHand プレイヤーの手札をスート→値の順にソートする
func wizardSortHand(p *WizardPlayer) {
	sortPlayerHand(p, func(ci, cj *Card) bool {
		if ci.GetDesign() != cj.GetDesign() {
			return ci.GetDesign() < cj.GetDesign()
		}
		return ci.GetValue() < cj.GetValue()
	})
}

// wizardCardStr 棋譜表示用のカード文字列 (ウィザード/ジェスターを含む)。
func wizardCardStr(card *Card) string {
	if card == nil {
		return "?"
	}
	switch card.GetDesign() {
	case WizardDesignWizard:
		return "Wizard"
	case WizardDesignJester:
		return "Jester"
	}
	return cardStr(card)
}

// playerName プレイヤー名を返す
func (o *Wizard) playerName(idx int) string {
	if idx < 0 || idx >= len(o.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if o.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// appendLog 棋譜にエントリを追加する
func (o *Wizard) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	o.actionLog = append(o.actionLog, &ActionLogEntry{
		TurnNumber: len(o.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す
func (o *Wizard) getValidPlayIndices(playerIdx int) []int {
	player := o.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return o.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// GetHint ヒントを取得する
func (o *Wizard) GetHint() *WizardHint {
	humanIdx := o.findHumanIdx()
	if humanIdx < 0 {
		return nil
	}
	if o.phase == WizardPhaseBid && o.bidPlayerIdx == humanIdx {
		bid := o.cpuBidHard(humanIdx)
		return &WizardHint{Bid: &bid, Reason: "strategic_bid"}
	}
	if o.phase == WizardPhasePlay && o.currentPlayerIdx == humanIdx {
		validIndices := o.getValidPlayIndices(humanIdx)
		if len(validIndices) == 0 {
			return nil
		}
		idx := o.cpuPlayHard(humanIdx, validIndices)
		return &WizardHint{CardIndex: &idx, Reason: o.playHintReason(humanIdx, idx)}
	}
	return nil
}

// playHintReason プレイヒントの理由を判定する
func (o *Wizard) playHintReason(playerIdx int, chosenIdx int) string {
	player := o.players[playerIdx]
	card := player.GetCard(chosenIdx)

	if card.GetDesign() == WizardDesignWizard {
		return "lead_strong"
	}
	if card.GetDesign() == WizardDesignJester {
		return "discard_high"
	}

	if len(o.currentTrick) == 0 {
		if player.GetTrickCount() < player.GetBid() {
			return "lead_strong"
		}
		return "lead_low"
	}

	leadSuit := o.leadSuitOfTrick()
	if leadSuit >= 0 && card.GetDesign() == leadSuit {
		return "follow_suit"
	}
	if o.trumpSuit >= 0 && card.GetDesign() == o.trumpSuit {
		return "trump_cut"
	}
	return "discard_high"
}

// --- CPU AI ---

// wizardEffRank カードの実効ランク。ウィザードは最強 (1000)、ジェスターは
// 最弱 (-1)、それ以外はカードの値。CPUの比較で使う。
func wizardEffRank(card *Card) int {
	switch card.GetDesign() {
	case WizardDesignWizard:
		return 1000
	case WizardDesignJester:
		return -1
	}
	return card.GetValue()
}

// cpuSelectBid CPUがビッドを選択する
func (o *Wizard) cpuSelectBid(playerIdx int) int {
	switch o.config.CpuDifficulty {
	case WizardCpuDifficultyHard:
		return o.cpuBidHard(playerIdx)
	case WizardCpuDifficultyNormal:
		return o.cpuBidNormal(playerIdx)
	default:
		return o.cpuBidEasy(playerIdx)
	}
}

// cpuBidEasy ランダムなビッド
func (o *Wizard) cpuBidEasy(_ int) int {
	if o.handSize == 0 {
		return 0
	}
	return rand.Intn(o.handSize + 1)
}

// cpuBidNormal カードの強さに基づくビッド (ウィザードは確実な1トリック)
func (o *Wizard) cpuBidNormal(playerIdx int) int {
	player := o.players[playerIdx]
	bid := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		switch card.GetDesign() {
		case WizardDesignWizard:
			bid++
		case WizardDesignJester:
			// ジェスターは0トリック見込み
		default:
			if o.trumpSuit >= 0 && card.GetDesign() == o.trumpSuit {
				if card.GetValue() >= 10 {
					bid++
				}
			} else if card.GetValue() >= 12 {
				bid++
			}
		}
	}
	if bid > o.handSize {
		bid = o.handSize
	}
	return bid
}

// cpuBidHard 戦略的なビッド (ウィザードは確実な1トリック)
func (o *Wizard) cpuBidHard(playerIdx int) int {
	player := o.players[playerIdx]
	bid := 0

	trumpCount := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		card := player.GetCard(i)
		switch card.GetDesign() {
		case WizardDesignWizard:
			bid++
		case WizardDesignJester:
			// 0トリック見込み
		default:
			if o.trumpSuit >= 0 && card.GetDesign() == o.trumpSuit {
				trumpCount++
				if card.GetValue() >= 10 {
					bid++
				}
			} else if card.GetValue() == 1 || card.GetValue() == 13 {
				// A, K は確実なトリック
				bid++
			} else if card.GetValue() == 12 {
				// Qは半確実
				if rand.Intn(2) == 0 {
					bid++
				}
			}
		}
	}

	// トランプが多い場合は追加トリックを見込む
	if trumpCount >= 3 {
		bid++
	}

	if bid > o.handSize {
		bid = o.handSize
	}
	return bid
}

// cpuSelectPlayCard CPUがプレイするカードのインデックスを選択する
func (o *Wizard) cpuSelectPlayCard(playerIdx int) int {
	validIndices := o.getValidPlayIndices(playerIdx)
	if len(validIndices) == 0 {
		return 0
	}
	if len(validIndices) == 1 {
		return validIndices[0]
	}

	switch o.config.CpuDifficulty {
	case WizardCpuDifficultyHard:
		return o.cpuPlayHard(playerIdx, validIndices)
	case WizardCpuDifficultyNormal:
		return o.cpuPlayNormal(playerIdx, validIndices)
	default:
		return o.cpuPlayEasy(validIndices)
	}
}

// cpuPlayEasy ランダムに有効なカードを選択
func (o *Wizard) cpuPlayEasy(validIndices []int) int {
	return validIndices[rand.Intn(len(validIndices))]
}

// cpuPlayNormal ビッドに近づくようにプレイ
func (o *Wizard) cpuPlayNormal(playerIdx int, validIndices []int) int {
	player := o.players[playerIdx]
	bid := player.GetBid()
	tricks := player.GetTrickCount()

	if len(o.currentTrick) == 0 {
		// リード
		if tricks < bid {
			return o.highestCardIdx(player, validIndices)
		}
		return o.lowestCardIdx(player, validIndices)
	}

	leadSuit := o.leadSuitOfTrick()
	if tricks < bid {
		return o.tryWinTrick(player, validIndices, leadSuit)
	}
	return o.tryLoseTrick(player, validIndices, leadSuit)
}

// cpuPlayHard 高度な戦略プレイ
func (o *Wizard) cpuPlayHard(playerIdx int, validIndices []int) int {
	player := o.players[playerIdx]
	bid := player.GetBid()
	tricks := player.GetTrickCount()

	needMore := tricks < bid
	exactBid := tricks == bid

	if len(o.currentTrick) == 0 {
		// リード
		if needMore {
			return o.bestLeadCard(player, validIndices, true)
		}
		return o.bestLeadCard(player, validIndices, false)
	}

	leadSuit := o.leadSuitOfTrick()

	if needMore {
		return o.tryWinTrick(player, validIndices, leadSuit)
	}
	if exactBid {
		return o.tryLoseTrick(player, validIndices, leadSuit)
	}
	// オーバートリック: 最も低いカードを出す
	return o.lowestCardIdx(player, validIndices)
}

// --- CPU helper methods ---

// highestCardIdx 実効ランクが最も高いカードのインデックスを返す
func (o *Wizard) highestCardIdx(player *WizardPlayer, indices []int) int {
	best := indices[0]
	for _, idx := range indices[1:] {
		if wizardEffRank(player.GetCard(idx)) > wizardEffRank(player.GetCard(best)) {
			best = idx
		}
	}
	return best
}

// lowestCardIdx 実効ランクが最も低いカードのインデックスを返す
func (o *Wizard) lowestCardIdx(player *WizardPlayer, indices []int) int {
	best := indices[0]
	for _, idx := range indices[1:] {
		if wizardEffRank(player.GetCard(idx)) < wizardEffRank(player.GetCard(best)) {
			best = idx
		}
	}
	return best
}

// bestLeadCard リード時に最適なカードを選ぶ
func (o *Wizard) bestLeadCard(player *WizardPlayer, indices []int, wantWin bool) int {
	if wantWin {
		// ウィザード or トランプの高札を優先 (実効ランク + トランプボーナス)
		best := indices[0]
		bestScore := -1
		for _, idx := range indices {
			card := player.GetCard(idx)
			score := wizardEffRank(card)
			if o.trumpSuit >= 0 && card.GetDesign() == o.trumpSuit {
				score += 100
			}
			if score > bestScore {
				bestScore = score
				best = idx
			}
		}
		return best
	}
	// 負けたい: 実効ランクが最も低いカード (ジェスター/低札を優先的に処分)
	return o.lowestCardIdx(player, indices)
}

// trickInProgressHasWizard 現在のトリックにウィザードが含まれるか
func (o *Wizard) trickInProgressHasWizard() bool {
	for _, tc := range o.currentTrick {
		if tc.Card.GetDesign() == WizardDesignWizard {
			return true
		}
	}
	return false
}

// firstIndexOfDesign validIndices の中で指定デザインの最初のインデックスを返す (-1=なし)
func firstIndexOfDesign(player *WizardPlayer, validIndices []int, design int) int {
	for _, idx := range validIndices {
		if player.GetCard(idx).GetDesign() == design {
			return idx
		}
	}
	return -1
}

// tryWinTrick 勝ちに行くカードを選ぶ
func (o *Wizard) tryWinTrick(player *WizardPlayer, validIndices []int, leadSuit int) int {
	// トリックに既にウィザードがあれば勝てない → 最も低いカードを処分
	if o.trickInProgressHasWizard() {
		return o.lowestCardIdx(player, validIndices)
	}
	// 自分のウィザードは確実な勝ち
	if wIdx := firstIndexOfDesign(player, validIndices, WizardDesignWizard); wIdx >= 0 {
		return wIdx
	}

	highestInTrick := 0
	highestTrumpInTrick := 0
	hasTrumpInTrick := false
	for _, tc := range o.currentTrick {
		if leadSuit >= 0 && tc.Card.GetDesign() == leadSuit && tc.Card.GetValue() > highestInTrick {
			highestInTrick = tc.Card.GetValue()
		}
		if o.trumpSuit >= 0 && tc.Card.GetDesign() == o.trumpSuit {
			hasTrumpInTrick = true
			if tc.Card.GetValue() > highestTrumpInTrick {
				highestTrumpInTrick = tc.Card.GetValue()
			}
		}
	}

	// リードスートのカードがあれば
	hasLead := false
	if leadSuit >= 0 {
		for _, idx := range validIndices {
			if player.GetCard(idx).GetDesign() == leadSuit {
				hasLead = true
				break
			}
		}
	}

	if hasLead && !hasTrumpInTrick {
		overCards := []int{}
		for _, idx := range validIndices {
			card := player.GetCard(idx)
			if card.GetDesign() == leadSuit && card.GetValue() > highestInTrick {
				overCards = append(overCards, idx)
			}
		}
		if len(overCards) > 0 {
			best := overCards[0]
			for _, idx := range overCards[1:] {
				if player.GetCard(idx).GetValue() < player.GetCard(best).GetValue() {
					best = idx
				}
			}
			return best
		}
	}

	if !hasLead && o.trumpSuit >= 0 {
		// ボイド: トランプでカット
		trumpCards := []int{}
		for _, idx := range validIndices {
			if player.GetCard(idx).GetDesign() == o.trumpSuit {
				trumpCards = append(trumpCards, idx)
			}
		}
		if len(trumpCards) > 0 {
			if hasTrumpInTrick {
				winnable := []int{}
				for _, idx := range trumpCards {
					if player.GetCard(idx).GetValue() > highestTrumpInTrick {
						winnable = append(winnable, idx)
					}
				}
				if len(winnable) > 0 {
					best := winnable[0]
					for _, idx := range winnable[1:] {
						if player.GetCard(idx).GetValue() < player.GetCard(best).GetValue() {
							best = idx
						}
					}
					return best
				}
			} else {
				best := trumpCards[0]
				for _, idx := range trumpCards[1:] {
					if player.GetCard(idx).GetValue() < player.GetCard(best).GetValue() {
						best = idx
					}
				}
				return best
			}
		}
	}

	// 勝てない場合は最も低いカードを出す
	return o.lowestCardIdx(player, validIndices)
}

// tryLoseTrick 負けに行くカードを選ぶ
func (o *Wizard) tryLoseTrick(player *WizardPlayer, validIndices []int, leadSuit int) int {
	// ジェスターは確実な負け
	if jIdx := firstIndexOfDesign(player, validIndices, WizardDesignJester); jIdx >= 0 {
		return jIdx
	}

	highestInTrick := 0
	if leadSuit >= 0 {
		for _, tc := range o.currentTrick {
			if tc.Card.GetDesign() == leadSuit && tc.Card.GetValue() > highestInTrick {
				highestInTrick = tc.Card.GetValue()
			}
		}
	}

	// リードスートのアンダーカード
	underCards := []int{}
	if leadSuit >= 0 {
		for _, idx := range validIndices {
			card := player.GetCard(idx)
			if card.GetDesign() == leadSuit && card.GetValue() < highestInTrick {
				underCards = append(underCards, idx)
			}
		}
	}
	if len(underCards) > 0 {
		// 最も高いアンダーカード（リードスートを温存しつつ負ける）
		best := underCards[0]
		for _, idx := range underCards[1:] {
			if player.GetCard(idx).GetValue() > player.GetCard(best).GetValue() {
				best = idx
			}
		}
		return best
	}

	// アンダーカードがない場合、最も低いカード
	return o.lowestCardIdx(player, validIndices)
}

// --- JSON serialization ---

// wizardJSON is the JSON wire format for Wizard.
type wizardJSON struct {
	Deck             []*Card           `json:"dk"`
	Players          []*WizardPlayer   `json:"ps"`
	Config           WizardConfig      `json:"cf"`
	Phase            WizardPhase       `json:"ph"`
	RoundNumber      int               `json:"rn"`
	TotalRounds      int               `json:"tr"`
	HandSize         int               `json:"hs"`
	TrickNumber      int               `json:"tn"`
	CurrentPlayerIdx int               `json:"ci"`
	CurrentTrick     []*TrickCard      `json:"ct"`
	LeadPlayerIdx    int               `json:"li"`
	BidPlayerIdx     int               `json:"bi"`
	DealerIdx        int               `json:"di"`
	TrumpCard        *Card             `json:"tp"`
	TrumpSuit        int               `json:"ts"`
	GameEndFlag      bool              `json:"ge"`
	WinnerIdx        int               `json:"wi"`
	ActionLog        []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (o *Wizard) MarshalJSON() ([]byte, error) {
	return json.Marshal(wizardJSON{
		Deck:             o.deck,
		Players:          o.players,
		Config:           o.config,
		Phase:            o.phase,
		RoundNumber:      o.roundNumber,
		TotalRounds:      o.totalRounds,
		HandSize:         o.handSize,
		TrickNumber:      o.trickNumber,
		CurrentPlayerIdx: o.currentPlayerIdx,
		CurrentTrick:     o.currentTrick,
		LeadPlayerIdx:    o.leadPlayerIdx,
		BidPlayerIdx:     o.bidPlayerIdx,
		DealerIdx:        o.dealerIdx,
		TrumpCard:        o.trumpCard,
		TrumpSuit:        o.trumpSuit,
		GameEndFlag:      o.gameEndFlag,
		WinnerIdx:        o.winnerIdx,
		ActionLog:        o.actionLog,
	})
}

// wizardMaxSliceLen caps slice sizes during deserialisation to prevent
// excessive memory allocation from malformed input.
const wizardMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (o *Wizard) UnmarshalJSON(data []byte) error {
	var j wizardJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > wizardMaxSliceLen || len(j.CurrentTrick) > wizardMaxSliceLen ||
		len(j.ActionLog) > wizardMaxSliceLen {
		return fmt.Errorf("wizard: input array exceeds maximum allowed size")
	}
	if len(j.Deck) > WizardDeckSize {
		return fmt.Errorf("wizard: deck length %d exceeds maximum %d", len(j.Deck), WizardDeckSize)
	}
	o.deck = j.Deck
	if o.deck == nil {
		o.deck = make([]*Card, 0)
	}
	// プレイヤー数と nil 要素のバリデーション: Wizard は常に WizardPlayerCnt(4)
	// 人で、コードは players[0..3] を直接インデックスするため、数が異なる／nil を
	// 含む状態を復元すると後続処理で範囲外・nil 参照パニックになる。
	if len(j.Players) > 0 {
		if len(j.Players) != WizardPlayerCnt {
			return fmt.Errorf("wizard: players count %d must be %d", len(j.Players), WizardPlayerCnt)
		}
		for i, p := range j.Players {
			if p == nil {
				return fmt.Errorf("wizard: player %d is nil", i)
			}
		}
	}
	o.players = j.Players
	if o.players == nil {
		o.players = make([]*WizardPlayer, 0)
	}
	// インデックスの範囲チェック
	pCnt := len(j.Players)
	if pCnt > 0 {
		if j.CurrentPlayerIdx < -1 || j.CurrentPlayerIdx >= pCnt {
			return fmt.Errorf("wizard: currentPlayerIdx %d out of range [−1, %d)", j.CurrentPlayerIdx, pCnt)
		}
		if j.DealerIdx < 0 || j.DealerIdx >= pCnt {
			return fmt.Errorf("wizard: dealerIdx %d out of range [0, %d)", j.DealerIdx, pCnt)
		}
		if j.BidPlayerIdx < -1 || j.BidPlayerIdx >= pCnt {
			return fmt.Errorf("wizard: bidPlayerIdx %d out of range [−1, %d)", j.BidPlayerIdx, pCnt)
		}
		if j.LeadPlayerIdx < -1 || j.LeadPlayerIdx >= pCnt {
			return fmt.Errorf("wizard: leadPlayerIdx %d out of range [−1, %d)", j.LeadPlayerIdx, pCnt)
		}
	}
	if j.TrickNumber < 0 {
		return fmt.Errorf("wizard: trickNumber %d must be >= 0", j.TrickNumber)
	}
	o.config = j.Config
	o.phase = j.Phase
	o.roundNumber = j.RoundNumber
	o.totalRounds = j.TotalRounds
	o.handSize = j.HandSize
	o.trickNumber = j.TrickNumber
	o.currentPlayerIdx = j.CurrentPlayerIdx
	o.currentTrick = j.CurrentTrick
	if o.currentTrick == nil {
		o.currentTrick = make([]*TrickCard, 0)
	}
	o.leadPlayerIdx = j.LeadPlayerIdx
	o.bidPlayerIdx = j.BidPlayerIdx
	o.dealerIdx = j.DealerIdx
	o.trumpCard = j.TrumpCard
	o.trumpSuit = j.TrumpSuit
	o.gameEndFlag = j.GameEndFlag
	o.winnerIdx = j.WinnerIdx
	o.actionLog = j.ActionLog
	if o.actionLog == nil {
		o.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
