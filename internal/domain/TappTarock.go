//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// 卓とデッキの形。**Königrufen と同じ 54 枚タロックデッキを使う。**
//
// 呼び札の対象が「キング」ではなく「切り札の 20」であること、そして誰も落札
// しなかったときに Trischaken (全員が失点を避け合う契約) になることが、この
// ゲームを Königrufen と分けている。デッキと札の点数は同じものを流用する ──
// 同じタロック札に 2 通りの点数表を持たせると、どちらが本物か分からなくなる。
const (
	// TappTarockPlayerCnt 席数 (人間 1 + CPU 2)。
	TappTarockPlayerCnt = 3
	// TappTarockHandSize 1 人の手札枚数。
	TappTarockHandSize = 16
	// TappTarockTalonSize 場札 (タロン) の枚数。
	TappTarockTalonSize = 6
	// TappTarockTrickCount 1 ディールのトリック数。
	TappTarockTrickCount = 16
	// TappTarockDeckSize デッキ総枚数。
	TappTarockDeckSize = KoenigrufenDeckSize
	// TappTarockMinDeals マッチ最小ディール数。
	TappTarockMinDeals = 1
	// TappTarockMaxDeals マッチ最大ディール数。
	TappTarockMaxDeals = 12
	// TappTarockDefaultDeals 既定のディール数。
	TappTarockDefaultDeals = 4
	// TappTarockBaseGameValue 基本点。
	TappTarockBaseGameValue = 10
	// TappTarockTrullBonus house-rule bonus for taking Pagat, Mond and Skus.
	TappTarockTrullBonus = 3
	// TappTarockTrischakenLoss Trischaken で最多失点者が失う点。
	TappTarockTrischakenLoss = 3
)

// TappTarockBid 入札 (コントラクト) 種別。値が大きいほど高い入札。
type TappTarockBid int

// TappTarock の入札定数
const (
	// TappTarockBidPass パス / 未入札
	TappTarockBidPass TappTarockBid = 0
	// TappTarockBidTrischaken トリシャーケン (全員パス時の契約)
	TappTarockBidTrischaken TappTarockBid = 1
	// TappTarockBidDreier ドライヤー (タロンを交換する契約)。
	TappTarockBidDreier TappTarockBid = 2
	// TappTarockBidSolo ソロ (場札交換なし)。
	TappTarockBidSolo TappTarockBid = 3
)

// TappTarockPhase ゲームフェーズ
type TappTarockPhase int

// TappTarock のフェーズ定数
const (
	// TappTarockPhaseBid 入札フェーズ
	TappTarockPhaseBid TappTarockPhase = 0
	// TappTarockPhaseTalon 場札交換 (discard) フェーズ
	TappTarockPhaseTalon TappTarockPhase = 1
	// TappTarockPhasePlay トリックプレイフェーズ
	TappTarockPhasePlay TappTarockPhase = 2
	// TappTarockPhaseTrickEnd トリック終了フェーズ
	TappTarockPhaseTrickEnd TappTarockPhase = 3
	// TappTarockPhaseRoundEnd ディール終了フェーズ
	TappTarockPhaseRoundEnd TappTarockPhase = 4
	// TappTarockPhaseGameEnd ゲーム終了フェーズ
	TappTarockPhaseGameEnd TappTarockPhase = 5
)

// TappTarockOutcome ディール結果 (デクレアラー側視点)
type TappTarockOutcome int

// TappTarock のディール結果定数
const (
	// TappTarockOutcomeNone 未確定
	TappTarockOutcomeNone TappTarockOutcome = 0
	// TappTarockOutcomeWin デクレアラー側がコントラクトを達成
	TappTarockOutcomeWin TappTarockOutcome = 1
	// TappTarockOutcomeLoss デクレアラー側がコントラクトを失敗
	TappTarockOutcomeLoss TappTarockOutcome = 2
	// TappTarockOutcomeTrischaken トリシャーケン (勝敗ではなく最多失点者を決める)
	TappTarockOutcomeTrischaken TappTarockOutcome = 3
)

// TappTarockBreakdown ディール精算の内訳。
type TappTarockBreakdown struct {
	// Contract 成立した契約。
	Contract TappTarockBid `json:"contract"`
	// TeamPoints デクレアラー側のカードポイント (Trischaken では最多失点者のもの)。
	TeamPoints int `json:"teamPoints"`
	// Threshold 成功に必要な点 (これを超えれば成功)。
	Threshold int `json:"threshold"`
	// Won デクレアラー側が達成したか。
	Won bool `json:"won"`
	// Solo 単独契約か。
	Solo bool `json:"solo"`
	// Base 基本点 (差分を加味した値)。
	Base int `json:"base"`
	// Seats 席ごとの増減 (席番号順)。
	Seats []int `json:"seats"`
	// Loser Trischaken の最多失点者 (それ以外は -1)。
	Loser int `json:"loser"`
}

// TappTarockHint ヒント情報
type TappTarockHint struct {
	// Bid 推奨入札 (入札フェーズ)。nil ならパス推奨。
	Bid *int `json:"bid,omitempty"`
	// DiscardIndices 推奨する伏せ札 (場札交換フェーズ)。
	DiscardIndices []int `json:"discardIndices,omitempty"`
	// CardIndex 推奨する手札インデックス (プレイフェーズ)。
	CardIndex *int `json:"cardIndex,omitempty"`
	// Reason 理由キー。
	Reason string `json:"reason"`
}

// TappTarock はタップ・タロックの集約ルート。
type TappTarock struct {
	deck        []*Card
	deckDrawCnt int
	players     []*TappTarockPlayer
	config      TappTarockConfig

	phase            TappTarockPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int

	// --- 入札 ---
	bidPlayerIdx  int
	bidActedCnt   int
	highestBid    TappTarockBid
	highestBidder int
	passed        [TappTarockPlayerCnt]bool

	// --- 契約 ---
	declarerIdx int
	contract    TappTarockBid

	// --- 場札 ---
	talon []*Card
	stash []*Card // 得点計上のため脇に置いた札
	// stashOwner は stash を数える側。0=デクレアラー側, 1=防御側, -1=最終トリックの勝者へ。
	stashOwner int

	// --- 精算 ---
	playerScores    [TappTarockPlayerCnt]int
	lastTrickWinner int
	lastTrickCards  []*Card
	outcome         TappTarockOutcome
	breakdown       *TappTarockBreakdown
	scored          bool
	gameEndFlag     bool
	winnerPlayer    int

	actionLogBase
}

// NewTappTarock コンストラクタ。
func NewTappTarock(players []*TappTarockPlayer, config TappTarockConfig) *TappTarock {
	return &TappTarock{
		players:         players,
		config:          config,
		winnerPlayer:    -1,
		lastTrickWinner: -1,
		declarerIdx:     -1,
		highestBidder:   -1,
		contract:        TappTarockBidPass,
		stashOwner:      0,
	}
}

// NewDefaultTappTarock 標準の 3 人構成 (人間 1, CPU 2) と既定設定で生成する。
func NewDefaultTappTarock() *TappTarock {
	players := make([]*TappTarockPlayer, TappTarockPlayerCnt)
	players[0] = NewTappTarockPlayer(true)
	for i := 1; i < TappTarockPlayerCnt; i++ {
		players[i] = NewTappTarockPlayer(false)
	}
	return NewTappTarock(players, DefaultTappTarockConfig())
}

// --- ゲーム進行 ---

// Reset ゲーム初期化。
func (g *TappTarock) Reset() {
	if err := g.config.Validate(); err != nil {
		g.config = DefaultTappTarockConfig()
	}
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.playerScores = [TappTarockPlayerCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のディールを開始する。
func (g *TappTarock) NextRound() {
	if g.gameEndFlag || g.phase != TappTarockPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % TappTarockPlayerCnt
	g.startRound()
}

// startRound 手札を配り、入札フェーズを開始する。
func (g *TappTarock) startRound() {
	g.trickNumber = 0
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.lastTrickWinner = -1
	g.lastTrickCards = nil
	g.declarerIdx = -1
	g.contract = TappTarockBidPass
	g.talon = nil
	g.stash = nil
	g.stashOwner = 0
	g.outcome = TappTarockOutcomeNone
	g.breakdown = nil
	g.scored = false
	g.passed = [TappTarockPlayerCnt]bool{}
	g.highestBid = TappTarockBidPass
	g.highestBidder = -1
	g.bidActedCnt = 0
	for _, p := range g.players {
		p.ResetRound()
	}
	g.deal()
	g.sortAllHands()
	g.bidPlayerIdx = (g.dealerIdx + 1) % TappTarockPlayerCnt
	g.currentPlayerIdx = g.bidPlayerIdx
	g.phase = TappTarockPhaseBid
	g.appendLog(-1, "deal", fmt.Sprintf("deal %d: 16 cards each, talon %d", g.roundNumber, len(g.talon)), nil)
}

// deal 3 枚パケットで各プレイヤーへ 16 枚を配り、場札 6 枚を脇に置く。
func (g *TappTarock) deal() {
	g.deck = buildKoenigrufenDeck()
	rand.Shuffle(len(g.deck), func(i, j int) { g.deck[i], g.deck[j] = g.deck[j], g.deck[i] })
	g.deckDrawCnt = 0
	g.talon = make([]*Card, 0, TappTarockTalonSize)
	for range TappTarockHandSize / 3 {
		for j := range TappTarockPlayerCnt {
			idx := (g.dealerIdx + 1 + j) % TappTarockPlayerCnt
			for range 3 {
				if c := drawFromDeck(g.deck, &g.deckDrawCnt); c != nil {
					g.players[idx].AddCard(c)
				}
			}
		}
	}
	for j := range TappTarockPlayerCnt {
		idx := (g.dealerIdx + 1 + j) % TappTarockPlayerCnt
		if c := drawFromDeck(g.deck, &g.deckDrawCnt); c != nil {
			g.players[idx].AddCard(c)
		}
	}
	for range TappTarockTalonSize {
		if c := drawFromDeck(g.deck, &g.deckDrawCnt); c != nil {
			g.talon = append(g.talon, c)
		}
	}
}

// --- 入札 ---

// PlayerBid 人間プレイヤーが入札する。
func (g *TappTarock) PlayerBid(bid TappTarockBid) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TappTarockPhaseBid {
		return NewDomainError(ErrWrongPhase, "not in bidding phase")
	}
	if !g.players[g.bidPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if err := g.validateBid(bid); err != nil {
		return err
	}
	g.applyBid(g.bidPlayerIdx, bid)
	return nil
}

// validateBid 入札が現在の最高入札を上回るかを検証する。
//
// **Trischaken は入札できない。** 誰も落札しなかったときにだけ成立する契約なので、
// 宣言できてしまうと「全員パス」との区別が付かなくなる。
func (g *TappTarock) validateBid(bid TappTarockBid) error {
	switch bid {
	case TappTarockBidDreier, TappTarockBidSolo:
	default:
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("bid %d cannot be declared", bid))
	}
	if bid <= g.highestBid {
		return NewDomainError(ErrInvalidPlay,
			fmt.Sprintf("bid %d does not beat the current bid %d", bid, g.highestBid))
	}
	return nil
}

// PlayerPass 人間プレイヤーがパスする。
func (g *TappTarock) PlayerPass() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TappTarockPhaseBid {
		return NewDomainError(ErrWrongPhase, "not in bidding phase")
	}
	if !g.players[g.bidPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.applyPass(g.bidPlayerIdx)
	return nil
}

// CpuBid CPU の入札を 1 手進める。
func (g *TappTarock) CpuBid() {
	if g.gameEndFlag || g.phase != TappTarockPhaseBid {
		return
	}
	idx := g.bidPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	if bid, ok := g.cpuSelectBid(idx); ok {
		g.applyBid(idx, bid)
		return
	}
	g.applyPass(idx)
}

// applyBid 入札を記録して手番を進める。
func (g *TappTarock) applyBid(idx int, bid TappTarockBid) {
	g.highestBid = bid
	g.highestBidder = idx
	g.bidActedCnt++
	g.appendLog(idx, "bid", fmt.Sprintf("%s bids %s", g.playerName(idx), TappTarockBidName(bid)), nil)
	g.advanceBid()
}

// applyPass パスを記録して手番を進める。
func (g *TappTarock) applyPass(idx int) {
	g.passed[idx] = true
	g.bidActedCnt++
	g.appendLog(idx, "pass", fmt.Sprintf("%s passes", g.playerName(idx)), nil)
	g.advanceBid()
}

// advanceBid 次の入札者を選び、全員が表明したら契約を確定する。
func (g *TappTarock) advanceBid() {
	if g.bidActedCnt >= TappTarockPlayerCnt {
		g.finalizeBid()
		return
	}
	for range TappTarockPlayerCnt {
		g.bidPlayerIdx = (g.bidPlayerIdx + 1) % TappTarockPlayerCnt
		if !g.passed[g.bidPlayerIdx] {
			g.currentPlayerIdx = g.bidPlayerIdx
			return
		}
	}
	g.finalizeBid()
}

// finalizeBid 落札者を決め、契約に応じた次のフェーズへ進む。
//
// **誰も落札しなければ Trischaken。** デクレアラーも場札交換も無く、全員が自分の
// ためだけに打ち、最も多くカードポイントを取った人が負ける。
func (g *TappTarock) finalizeBid() {
	if g.highestBidder < 0 {
		g.contract = TappTarockBidTrischaken
		g.declarerIdx = -1
		// **場札は脇へ移し、最終トリックの勝者が引き取る。** 移し忘れると
		// assignStash が何も足さず、6 枚ぶんの点が黙って消える。
		g.stash = append([]*Card(nil), g.talon...)
		g.talon = nil
		g.stashOwner = -1
		g.appendLog(-1, "contract", "everyone passed: Trischaken", nil)
		g.startPlay()
		return
	}
	g.declarerIdx = g.highestBidder
	g.contract = g.highestBid
	if g.contract == TappTarockBidSolo {
		// ソロでは場札は防御側の得点になる。
		g.stash = append([]*Card(nil), g.talon...)
		g.talon = nil
		g.stashOwner = 1
		g.appendLog(g.declarerIdx, "contract",
			fmt.Sprintf("%s plays solo", g.playerName(g.declarerIdx)), nil)
		g.startPlay()
		return
	}
	g.phase = TappTarockPhaseTalon
	g.currentPlayerIdx = g.declarerIdx
	// 場札はデクレアラーの手札に加わる。伏せる 6 枚は本人が選ぶ。
	for _, c := range g.talon {
		g.players[g.declarerIdx].AddCard(c)
	}
	g.talon = nil
	g.sortAllHands()
}

// --- 場札交換 ---

// PlayerDiscard 人間のデクレアラーが 6 枚を伏せる。
func (g *TappTarock) PlayerDiscard(indices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TappTarockPhaseTalon {
		return NewDomainError(ErrWrongPhase, "not in talon phase")
	}
	if g.declarerIdx < 0 || !g.players[g.declarerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if err := g.validateDiscards(indices); err != nil {
		return err
	}
	g.applyDiscard(indices)
	return nil
}

// validateDiscards 伏せ札の枚数・重複・種別を検証する。
//
// **キングとトゥルルは伏せられない。** 5 点札を黙って自分の得点に移せてしまうと、
// 場札交換が「点を配る操作」になり、契約の難度が消える。
func (g *TappTarock) validateDiscards(indices []int) error {
	if len(indices) != TappTarockTalonSize {
		return NewDomainError(ErrInvalidPlay,
			fmt.Sprintf("must discard exactly %d cards, got %d", TappTarockTalonSize, len(indices)))
	}
	p := g.players[g.declarerIdx]
	// **添字の妥当性を先に全部見る。** 札の種別より先に範囲と重複を弾かないと、
	// 返る理由が「何番目の札がキングだったか」という配り次第の話になる。
	seen := map[int]bool{}
	for _, idx := range indices {
		if idx < 0 || idx >= p.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", idx))
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("hand index %d listed twice", idx))
		}
		seen[idx] = true
	}
	for _, idx := range indices {
		c := p.GetCard(idx)
		if koenigrufenIsKing(c) {
			return NewDomainError(ErrInvalidPlay, "a king cannot be discarded")
		}
		if koenigrufenIsTrull(c) {
			return NewDomainError(ErrInvalidPlay, "a trull card cannot be discarded")
		}
	}
	return nil
}

// applyDiscard 伏せ札を脇へ移し、プレイフェーズを始める。
func (g *TappTarock) applyDiscard(indices []int) {
	p := g.players[g.declarerIdx]
	g.stash = p.RemoveCards(indices)
	g.stashOwner = 0
	g.sortAllHands()
	g.appendLog(g.declarerIdx, "discard",
		fmt.Sprintf("%s buries %d cards", g.playerName(g.declarerIdx), len(g.stash)), nil)
	g.startPlay()
}

// CpuDiscard CPU のデクレアラーに伏せ札を選ばせる。
func (g *TappTarock) CpuDiscard() {
	if g.gameEndFlag || g.phase != TappTarockPhaseTalon {
		return
	}
	if g.declarerIdx < 0 || g.players[g.declarerIdx].GetIsHuman() {
		return
	}
	g.applyDiscard(g.cpuSelectDiscards(g.declarerIdx))
}

// cpuSelectDiscards 伏せられる札のうち、点数の低い 6 枚を選ぶ。
func (g *TappTarock) cpuSelectDiscards(idx int) []int {
	p := g.players[idx]
	type entry struct{ idx, pts int }
	cand := make([]entry, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if koenigrufenIsKing(c) || koenigrufenIsTrull(c) || koenigrufenIsTrumpLike(c) {
			// 切り札は残す (伏せられるが、手放すと勝てない)。
			continue
		}
		cand = append(cand, entry{i, koenigrufenCardPoints(c)})
	}
	// 点数の低い順、同点なら手札の並び順。
	for i := 1; i < len(cand); i++ {
		for j := i; j > 0 && cand[j].pts < cand[j-1].pts; j-- {
			cand[j], cand[j-1] = cand[j-1], cand[j]
		}
	}
	out := make([]int, 0, TappTarockTalonSize)
	for _, e := range cand {
		if len(out) == TappTarockTalonSize {
			break
		}
		out = append(out, e.idx)
	}
	// 足りなければ、伏せられる残りの札から補う (切り札しか残っていない場合)。
	for i := range p.GetCardsSize() {
		if len(out) == TappTarockTalonSize {
			break
		}
		c := p.GetCard(i)
		if koenigrufenIsKing(c) || koenigrufenIsTrull(c) || tapptarockContains(out, i) {
			continue
		}
		out = append(out, i)
	}
	return out
}

// tapptarockContains スライスに値が含まれるか。
func tapptarockContains(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// --- プレイ ---

// startPlay トリックプレイを開始する。
func (g *TappTarock) startPlay() {
	g.phase = TappTarockPhasePlay
	g.trickNumber = 1
	g.leadPlayerIdx = (g.dealerIdx + 1) % TappTarockPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.currentTrick = nil
}

// PlayerPlayCard 人間が手札を 1 枚出す。
func (g *TappTarock) PlayerPlayCard(handIdx int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TappTarockPhasePlay {
		return NewDomainError(ErrWrongPhase, "not in play phase")
	}
	p := g.players[g.currentPlayerIdx]
	if !p.GetIsHuman() {
		return ErrNotHumanTurn
	}
	if handIdx < 0 || handIdx >= p.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", handIdx))
	}
	if !tapptarockContains(g.GetValidPlayIndices(g.currentPlayerIdx), handIdx) {
		return NewDomainError(ErrInvalidPlay, "that card does not follow the lead")
	}
	g.playCard(g.currentPlayerIdx, handIdx)
	return nil
}

// CpuPlayCard CPU の 1 手を進める。
func (g *TappTarock) CpuPlayCard() {
	if g.gameEndFlag || g.phase != TappTarockPhasePlay {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	g.playCard(g.currentPlayerIdx, g.cpuSelectPlayCard(g.currentPlayerIdx))
}

// playCard 1 枚を場に出し、トリックが揃ったら決着させる。
func (g *TappTarock) playCard(playerIdx, handIdx int) {
	card := g.players[playerIdx].RemoveCard(handIdx)
	if card == nil {
		return
	}
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", g.playerName(playerIdx), koenigrufenCardStr(card)), []*Card{card})

	if len(g.currentTrick) < TappTarockPlayerCnt {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % TappTarockPlayerCnt
		return
	}
	g.finishTrick()
}

// finishTrick トリックの勝者を決め、次のトリックへ進む。
func (g *TappTarock) finishTrick() {
	winner := g.trickWinner()
	cards := make([]*Card, 0, len(g.currentTrick))
	for _, tc := range g.currentTrick {
		cards = append(cards, tc.Card)
	}
	g.players[winner].AddTrick(cards)
	g.lastTrickWinner = winner
	g.lastTrickCards = cards
	g.appendLog(winner, "trick",
		fmt.Sprintf("%s takes trick %d", g.playerName(winner), g.trickNumber), cards)
	g.currentTrick = nil
	g.phase = TappTarockPhaseTrickEnd
	g.currentPlayerIdx = winner
	g.leadPlayerIdx = winner
}

// NextTrick トリック終了フェーズから次のトリックへ進む。
func (g *TappTarock) NextTrick() {
	if g.gameEndFlag || g.phase != TappTarockPhaseTrickEnd {
		return
	}
	if g.trickNumber >= TappTarockTrickCount {
		g.finishRound()
		return
	}
	g.trickNumber++
	g.phase = TappTarockPhasePlay
	g.currentPlayerIdx = g.leadPlayerIdx
}

// trickWinner 現在のトリックの勝者を返す。
func (g *TappTarock) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return g.leadPlayerIdx
	}
	led := g.ledSuit()
	best, bestRank := g.currentTrick[0].PlayerIdx, -1
	for _, tc := range g.currentTrick {
		if r := koenigrufenWinRank(tc.Card, led); r > bestRank {
			best, bestRank = tc.PlayerIdx, r
		}
	}
	return best
}

// ledSuit 現在のトリックの実効リードスートを返す (空なら -1)。
func (g *TappTarock) ledSuit() int {
	if len(g.currentTrick) == 0 {
		return -1
	}
	first := g.currentTrick[0].Card
	if koenigrufenIsTrumpLike(first) {
		return KoenigrufenTrumpDesign
	}
	return first.GetDesign()
}

// GetDiscardableIndices はデクレアラーがタロンフェーズで伏せられる手札のインデックスを返す。
// タロンフェーズでない場合や、デクレアラーが無効な場合は空スライスを返す。
//
// キングとトゥルルは伏せられないため除外される。
func (g *TappTarock) GetDiscardableIndices() []int {
	if g.phase != TappTarockPhaseTalon || g.declarerIdx < 0 || g.declarerIdx >= len(g.players) {
		return []int{}
	}
	p := g.players[g.declarerIdx]
	if p == nil {
		return []int{}
	}
	idxs := make([]int, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if koenigrufenIsKing(c) || koenigrufenIsTrull(c) {
			continue
		}
		idxs = append(idxs, i)
	}
	return idxs
}

// GetValidPlayIndices 出せる手札のインデックスを返す。
//
// リードスートに従う義務、持っていなければ切り札を出す義務、どちらも無ければ自由。
func (g *TappTarock) GetValidPlayIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	p := g.players[playerIdx]
	all := make([]int, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		all = append(all, i)
	}
	led := g.ledSuit()
	if led < 0 {
		return all
	}
	follow := make([]int, 0, len(all))
	for _, i := range all {
		c := p.GetCard(i)
		if led == KoenigrufenTrumpDesign {
			if koenigrufenIsTrumpLike(c) {
				follow = append(follow, i)
			}
			continue
		}
		if !koenigrufenIsTrumpLike(c) && c.GetDesign() == led {
			follow = append(follow, i)
		}
	}
	if len(follow) > 0 {
		return follow
	}
	// リードスートが無ければ切り札を切る義務がある。
	trumps := make([]int, 0, len(all))
	for _, i := range all {
		if koenigrufenIsTrumpLike(p.GetCard(i)) {
			trumps = append(trumps, i)
		}
	}
	if len(trumps) > 0 {
		return trumps
	}
	return all
}

// cpuSelectPlayCard CPU の 1 手を選ぶ。
//
// Trischaken では点を取らないほうが良いので、勝ちにいかず安い札を落とす。
func (g *TappTarock) cpuSelectPlayCard(playerIdx int) int {
	valid := g.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := g.players[playerIdx]
	if g.config.CpuDifficulty == TappTarockCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	pick, bestScore := valid[0], 0
	for i, idx := range valid {
		pts := koenigrufenCardPoints(p.GetCard(idx))
		score := -pts // 既定では安い札から落とす
		if g.contract != TappTarockBidTrischaken && len(g.currentTrick) == TappTarockPlayerCnt-1 {
			// 最後の 1 枚なら、勝てるときだけ高い札を使う。
			score = koenigrufenWinRank(p.GetCard(idx), g.ledSuit())
		}
		if i == 0 || score > bestScore {
			pick, bestScore = idx, score
		}
	}
	return pick
}

// cpuSelectBid CPU の入札を選ぶ。切り札の枚数で判断する。
func (g *TappTarock) cpuSelectBid(playerIdx int) (TappTarockBid, bool) {
	p := g.players[playerIdx]
	trumps, honours := 0, 0
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if koenigrufenIsTrumpLike(c) {
			trumps++
		}
		if koenigrufenIsTrull(c) || koenigrufenIsKing(c) {
			honours++
		}
	}
	switch {
	case trumps >= 9 && honours >= 3 && TappTarockBidSolo > g.highestBid:
		return TappTarockBidSolo, true
	case trumps >= 5 && TappTarockBidDreier > g.highestBid:
		return TappTarockBidDreier, true
	default:
		return TappTarockBidPass, false
	}
}

// --- 精算 ---

// finishRound ディールを精算する。
func (g *TappTarock) finishRound() {
	if g.scored {
		return
	}
	g.scored = true
	g.assignStash()
	if g.contract == TappTarockBidTrischaken {
		g.breakdown = g.scoreTrischaken()
		g.outcome = TappTarockOutcomeTrischaken
	} else {
		g.breakdown = g.scoreContract()
		if g.breakdown.Won {
			g.outcome = TappTarockOutcomeWin
		} else {
			g.outcome = TappTarockOutcomeLoss
		}
	}
	for i, delta := range g.breakdown.Seats {
		g.playerScores[i] += delta
	}
	g.phase = TappTarockPhaseRoundEnd
	g.appendLog(-1, "score",
		fmt.Sprintf("deal %d scored (%s)", g.roundNumber, TappTarockBidName(g.contract)), nil)
	if g.roundNumber >= g.config.TargetDeals {
		g.finishGame()
	}
}

// assignStash 脇に置いた札を、契約に応じた引き取り手のトリックに加える。
//
// **Trischaken の場札は最終トリックの勝者が引き取る。** どこにも足さないと、
// 場札 6 枚ぶんの点が消えて総点が合わなくなる。
func (g *TappTarock) assignStash() {
	if len(g.stash) == 0 {
		return
	}
	target := -1
	switch g.stashOwner {
	case 0:
		target = g.declarerIdx
	case 1:
		// 防御側へ渡す。
		for i := range g.players {
			if i != g.declarerIdx {
				target = i
				break
			}
		}
	default:
		target = g.lastTrickWinner
	}
	if target < 0 || target >= len(g.players) {
		target = 0
	}
	g.players[target].AddTrick(g.stash)
	g.stash = nil
}

// cardPointsOf 席が獲得した札のカードポイント合計を返す。
func (g *TappTarock) cardPointsOf(idx int) int {
	total := 0
	trull := 0
	for _, trick := range g.players[idx].GetTricksTaken() {
		for _, c := range trick {
			total += koenigrufenCardPoints(c)
			if koenigrufenIsTrull(c) {
				trull++
			}
		}
	}
	if trull == 3 {
		total += TappTarockTrullBonus
	}
	return total
}

// tapptarockTotalPoints デッキ全体のカードポイントを返す。
//
// **表を書き写さず数える。** 閾値を定数で持つと、点数表を変えたときに片方だけが
// ずれる。
func tapptarockTotalPoints() int {
	total := 0
	for _, c := range buildKoenigrufenDeck() {
		total += koenigrufenCardPoints(c)
	}
	return total
}

// scoreContract Dreier / Solo の精算を行う (ゼロサム)。
func (g *TappTarock) scoreContract() *TappTarockBreakdown {
	team := 0
	for i := range g.players {
		if i == g.declarerIdx {
			team += g.cardPointsOf(i)
		}
	}
	total := tapptarockTotalPoints()
	threshold := total / 2
	won := 2*team > total
	diff := team - threshold
	if diff < 0 {
		diff = -diff
	}
	solo := g.contract == TappTarockBidSolo
	base := TappTarockBaseGameValue + diff
	if g.contract == TappTarockBidSolo {
		base *= 2
	}
	sign := 1
	if !won {
		sign = -1
	}
	bd := &TappTarockBreakdown{
		Contract: g.contract, TeamPoints: team, Threshold: threshold,
		Won: won, Solo: solo, Base: base,
		Seats: make([]int, TappTarockPlayerCnt), Loser: -1,
	}
	declarerShare := base
	if solo {
		declarerShare = (TappTarockPlayerCnt - 1) * base
	}
	for i := range bd.Seats {
		switch i {
		case g.declarerIdx:
			bd.Seats[i] = sign * declarerShare
		default:
			bd.Seats[i] = -sign * base
		}
	}
	return bd
}

// scoreTrischaken Trischaken の精算を行う。
//
// **最も多くカードポイントを取った席が負ける。** 取らないことが目的の契約なので、
// 勝ち負けの向きが他の契約と逆になる。同点なら先に到達した席 (席番号の小さいほう)
// が負けを引き受ける ── 引き分けにすると、誰も失点しないディールができてしまう。
func (g *TappTarock) scoreTrischaken() *TappTarockBreakdown {
	loser, worst := 0, -1
	for i := range g.players {
		if pts := g.cardPointsOf(i); pts > worst {
			loser, worst = i, pts
		}
	}
	bd := &TappTarockBreakdown{
		Contract: TappTarockBidTrischaken, TeamPoints: worst,
		Threshold: 0, Won: false, Solo: false,
		Base:  TappTarockTrischakenLoss,
		Seats: make([]int, TappTarockPlayerCnt), Loser: loser,
	}
	for i := range bd.Seats {
		if i == loser {
			bd.Seats[i] = -TappTarockTrischakenLoss
			continue
		}
		bd.Seats[i] = 1
	}
	return bd
}

// finishGame 通算得点で勝者を決める。
func (g *TappTarock) finishGame() {
	best, tied := 0, false
	for i := 1; i < TappTarockPlayerCnt; i++ {
		switch {
		case g.playerScores[i] > g.playerScores[best]:
			best, tied = i, false
		case g.playerScores[i] == g.playerScores[best]:
			tied = true
		}
	}
	g.winnerPlayer = -1
	if !tied {
		g.winnerPlayer = best
	}
	g.gameEndFlag = true
	g.phase = TappTarockPhaseGameEnd
	g.appendLog(g.winnerPlayer, "gameEnd", "match over", nil)
}

// --- 補助 ---

// sortAllHands 全員の手札を並べ替える。
func (g *TappTarock) sortAllHands() {
	for _, p := range g.players {
		tapptarockSortHand(p)
	}
}

// tapptarockSortHand 手札をスート順・切り札は番号順に並べる。
func tapptarockSortHand(p *TappTarockPlayer) {
	cards := make([]*Card, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		cards = append(cards, p.GetCard(i))
	}
	for i := 1; i < len(cards); i++ {
		for j := i; j > 0 && tapptarockSortKey(cards[j]) < tapptarockSortKey(cards[j-1]); j-- {
			cards[j], cards[j-1] = cards[j-1], cards[j]
		}
	}
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// tapptarockSortKey 手札の並び順キー。
func tapptarockSortKey(c *Card) int {
	if c == nil {
		return 1 << 20
	}
	if koenigrufenIsTrumpLike(c) {
		return 1000 + koenigrufenTrumpValue(c)
	}
	return c.GetDesign()*100 + c.GetValue()
}

// playerName 席の表示名を返す。
func (g *TappTarock) playerName(i int) string {
	if i < 0 || i >= len(g.players) {
		return "-"
	}
	if g.players[i].GetIsHuman() {
		return "YOU"
	}
	return fmt.Sprintf("CPU%d", i)
}

// TappTarockBidName 入札の識別子を返す (i18n キーの一部に使う)。
func TappTarockBidName(bid TappTarockBid) string {
	switch bid {
	case TappTarockBidTrischaken:
		return "trischaken"
	case TappTarockBidDreier:
		return "dreier"
	case TappTarockBidSolo:
		return "solo"
	default:
		return "pass"
	}
}

// --- アクセサ ---

// GetPlayers 席の一覧を返す。
func (g *TappTarock) GetPlayers() []*TappTarockPlayer { return g.players }

// GetPlayer 指定した席を返す (範囲外なら nil)。
func (g *TappTarock) GetPlayer(i int) *TappTarockPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetPlayerCnt 席数を返す。
func (g *TappTarock) GetPlayerCnt() int { return len(g.players) }

// GetConfig 設定を返す。
func (g *TappTarock) GetConfig() TappTarockConfig { return g.config }

// SetConfig 設定を差し替える (次の Reset で反映)。
func (g *TappTarock) SetConfig(c TappTarockConfig) { g.config = c }

// GetPhase 現在のフェーズを返す。
func (g *TappTarock) GetPhase() TappTarockPhase { return g.phase }

// GetRoundNumber 現在のディール番号を返す。
func (g *TappTarock) GetRoundNumber() int { return g.roundNumber }

// GetTrickNumber 現在のトリック番号を返す。
func (g *TappTarock) GetTrickNumber() int { return g.trickNumber }

// GetCurrentPlayerIdx 手番の席を返す。
func (g *TappTarock) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// GetCurrentTrick 場に出ている札を返す。
func (g *TappTarock) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// GetDealerIdx 親の席を返す。
func (g *TappTarock) GetDealerIdx() int { return g.dealerIdx }

// GetBidPlayerIdx 入札の手番を返す。
func (g *TappTarock) GetBidPlayerIdx() int { return g.bidPlayerIdx }

// GetHighestBid 現在の最高入札を返す。
func (g *TappTarock) GetHighestBid() TappTarockBid { return g.highestBid }

// GetDeclarerIdx デクレアラーの席を返す (-1 = 未確定 / Trischaken)。
func (g *TappTarock) GetDeclarerIdx() int { return g.declarerIdx }

// GetContract 成立した契約を返す。
func (g *TappTarock) GetContract() TappTarockBid { return g.contract }

// GetTalonSize 手札に加える前の場札の枚数を返す。
func (g *TappTarock) GetTalonSize() int { return len(g.talon) }

// GetLastTrickWinner 直前のトリックの勝者を返す。
func (g *TappTarock) GetLastTrickWinner() int { return g.lastTrickWinner }

// GetLastTrickCards 直前のトリックの札を返す。
func (g *TappTarock) GetLastTrickCards() []*Card { return g.lastTrickCards }

// GetOutcome ディール結果を返す。
func (g *TappTarock) GetOutcome() TappTarockOutcome { return g.outcome }

// GetBreakdown 直近ディールの精算内訳を返す。
func (g *TappTarock) GetBreakdown() *TappTarockBreakdown { return g.breakdown }

// GetPlayerScore 指定席の通算得点を返す。
func (g *TappTarock) GetPlayerScore(i int) int {
	if i < 0 || i >= TappTarockPlayerCnt {
		return 0
	}
	return g.playerScores[i]
}

// GetCardPoints 指定席が獲得したカードポイントを返す。
func (g *TappTarock) GetCardPoints(i int) int {
	if i < 0 || i >= len(g.players) {
		return 0
	}
	return g.cardPointsOf(i)
}

// GetGameEndFlag 終局したかを返す。
func (g *TappTarock) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 終局時の勝者を返す (-1 = 引き分け/未決)。
func (g *TappTarock) GetWinnerPlayer() int { return g.winnerPlayer }

// HumanSeat 人間の席を返す (居なければ 0)。
func (g *TappTarock) HumanSeat() int {
	if i := findHumanIdx(g.players); i >= 0 {
		return i
	}
	return 0
}

// IsHumanTurn 人間の入力を待っているかを返す。
func (g *TappTarock) IsHumanTurn() bool {
	if g.gameEndFlag {
		return false
	}
	switch g.phase {
	case TappTarockPhaseBid:
		return g.players[g.bidPlayerIdx].GetIsHuman()
	case TappTarockPhaseTalon:
		return g.declarerIdx >= 0 && g.players[g.declarerIdx].GetIsHuman()
	case TappTarockPhasePlay:
		return g.players[g.currentPlayerIdx].GetIsHuman()
	default:
		return false
	}
}

// GetHint 人間の手番における推奨手を返す。
func (g *TappTarock) GetHint() *TappTarockHint {
	if g.gameEndFlag || !g.IsHumanTurn() {
		return nil
	}
	seat := g.HumanSeat()
	switch g.phase {
	case TappTarockPhaseBid:
		bid, ok := g.cpuSelectBid(seat)
		if !ok {
			return &TappTarockHint{Reason: "pass_weak_hand"}
		}
		v := int(bid)
		return &TappTarockHint{Bid: &v, Reason: "bid_strong_trumps"}
	case TappTarockPhaseTalon:
		return &TappTarockHint{DiscardIndices: g.cpuSelectDiscards(seat), Reason: "bury_cheap_cards"}
	case TappTarockPhasePlay:
		idx := g.cpuSelectPlayCard(seat)
		reason := "play_low"
		if g.contract == TappTarockBidTrischaken {
			reason = "avoid_points"
		}
		return &TappTarockHint{CardIndex: &idx, Reason: reason}
	default:
		return nil
	}
}

// --- JSON ---

// tapptarockJSON は TappTarock の JSON 表現。
type tapptarockJSON struct {
	Players         []*TappTarockPlayer  `json:"pl"`
	Config          TappTarockConfig     `json:"cf"`
	Phase           TappTarockPhase      `json:"ph"`
	RoundNumber     int                  `json:"rn"`
	TrickNumber     int                  `json:"tn"`
	CurrentPlayer   int                  `json:"cp"`
	CurrentTrick    []*TrickCard         `json:"ct"`
	LeadPlayer      int                  `json:"lp"`
	DealerIdx       int                  `json:"di"`
	BidPlayerIdx    int                  `json:"bp"`
	BidActedCnt     int                  `json:"ba"`
	HighestBid      TappTarockBid        `json:"hb"`
	HighestBidder   int                  `json:"hr"`
	Passed          []bool               `json:"ps"`
	DeclarerIdx     int                  `json:"dc"`
	Contract        TappTarockBid        `json:"co"`
	Talon           []*Card              `json:"tl"`
	Stash           []*Card              `json:"st"`
	StashOwner      int                  `json:"so"`
	PlayerScores    []int                `json:"sc"`
	LastTrickWinner int                  `json:"lw"`
	LastTrickCards  []*Card              `json:"lc"`
	Outcome         TappTarockOutcome    `json:"oc"`
	Breakdown       *TappTarockBreakdown `json:"bd"`
	Scored          bool                 `json:"sd"`
	GameEndFlag     bool                 `json:"ge"`
	WinnerPlayer    int                  `json:"wp"`
	ActionLog       []*ActionLogEntry    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *TappTarock) MarshalJSON() ([]byte, error) {
	return json.Marshal(tapptarockJSON{
		Players: g.players, Config: g.config, Phase: g.phase,
		RoundNumber: g.roundNumber, TrickNumber: g.trickNumber,
		CurrentPlayer: g.currentPlayerIdx, CurrentTrick: g.currentTrick,
		LeadPlayer: g.leadPlayerIdx, DealerIdx: g.dealerIdx,
		BidPlayerIdx: g.bidPlayerIdx, BidActedCnt: g.bidActedCnt,
		HighestBid: g.highestBid, HighestBidder: g.highestBidder,
		Passed: g.passed[:], DeclarerIdx: g.declarerIdx, Contract: g.contract,
		Talon: g.talon, Stash: g.stash, StashOwner: g.stashOwner,
		PlayerScores: g.playerScores[:], LastTrickWinner: g.lastTrickWinner,
		LastTrickCards: g.lastTrickCards, Outcome: g.outcome, Breakdown: g.breakdown,
		Scored: g.scored, GameEndFlag: g.gameEndFlag, WinnerPlayer: g.winnerPlayer,
		ActionLog: g.actionLog,
	})
}

// tapptarockMaxSliceLen 復元時のスライス長上限。
const tapptarockMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
//
// 保存状態の設定・人数・フェーズ・ラウンド／トリック番号・プレイヤーの添字・契約・
// タロン／脇札の整合性を検証してから復元する。
func (g *TappTarock) UnmarshalJSON(data []byte) error {
	var j tapptarockJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > tapptarockMaxSliceLen || len(j.Talon) > tapptarockMaxSliceLen ||
		len(j.Stash) > tapptarockMaxSliceLen || len(j.ActionLog) > tapptarockMaxSliceLen ||
		len(j.CurrentTrick) > tapptarockMaxSliceLen {
		return fmt.Errorf("tapptarock: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("tapptarock: invalid config: %w", err)
	}
	if len(j.Players) != TappTarockPlayerCnt {
		return fmt.Errorf("tapptarock: expected %d players, got %d", TappTarockPlayerCnt, len(j.Players))
	}
	for _, p := range j.Players {
		if p == nil {
			return fmt.Errorf("tapptarock: nil player in state")
		}
	}
	if j.Phase < TappTarockPhaseBid || j.Phase > TappTarockPhaseGameEnd {
		return fmt.Errorf("tapptarock: invalid phase %d", j.Phase)
	}
	if j.RoundNumber < 1 || j.RoundNumber > j.Config.TargetDeals {
		return fmt.Errorf("tapptarock: round %d out of range", j.RoundNumber)
	}
	if j.TrickNumber < 0 || j.TrickNumber > TappTarockTrickCount {
		return fmt.Errorf("tapptarock: trick %d out of range", j.TrickNumber)
	}
	for name, idx := range map[string]int{"current player": j.CurrentPlayer, "bid player": j.BidPlayerIdx, "dealer": j.DealerIdx} {
		if idx < 0 || idx >= TappTarockPlayerCnt {
			return fmt.Errorf("tapptarock: %s out of range", name)
		}
	}
	for name, idx := range map[string]int{"declarer": j.DeclarerIdx, "winner": j.WinnerPlayer, "highest bidder": j.HighestBidder} {
		if idx < -1 || idx >= TappTarockPlayerCnt {
			return fmt.Errorf("tapptarock: %s out of range", name)
		}
	}
	if err := tapptarockValidateContract(&j); err != nil {
		return err
	}
	if err := tapptarockValidateCards(j.Talon); err != nil {
		return err
	}
	if err := tapptarockValidateCards(j.Stash); err != nil {
		return err
	}

	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayer
	g.currentTrick = j.CurrentTrick
	g.leadPlayerIdx = j.LeadPlayer
	g.dealerIdx = j.DealerIdx
	g.bidPlayerIdx = j.BidPlayerIdx
	g.bidActedCnt = j.BidActedCnt
	g.highestBid = j.HighestBid
	g.highestBidder = j.HighestBidder
	g.passed = [TappTarockPlayerCnt]bool{}
	copy(g.passed[:], j.Passed)
	g.declarerIdx = j.DeclarerIdx
	g.contract = j.Contract
	g.talon = j.Talon
	g.stash = j.Stash
	g.stashOwner = j.StashOwner
	g.playerScores = [TappTarockPlayerCnt]int{}
	copy(g.playerScores[:], j.PlayerScores)
	g.lastTrickWinner = j.LastTrickWinner
	g.lastTrickCards = j.LastTrickCards
	g.outcome = j.Outcome
	g.breakdown = j.Breakdown
	g.scored = j.Scored
	g.gameEndFlag = j.GameEndFlag
	g.winnerPlayer = j.WinnerPlayer
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}

// tapptarockValidateContract 契約・呼び札・デクレアラーの整合を検証する。
func tapptarockValidateContract(j *tapptarockJSON) error {
	switch j.Contract {
	case TappTarockBidPass, TappTarockBidTrischaken,
		TappTarockBidDreier, TappTarockBidSolo:
	default:
		return fmt.Errorf("tapptarock: invalid contract %d", j.Contract)
	}
	// **Trischaken にデクレアラーは居ない。**
	if j.Contract == TappTarockBidTrischaken && j.DeclarerIdx != -1 {
		return fmt.Errorf("tapptarock: trischaken has no declarer")
	}
	return nil
}

// tapptarockValidateCards 札に nil や範囲外の design/value が無いか検証する。
func tapptarockValidateCards(cards []*Card) error {
	for _, c := range cards {
		if c == nil {
			return fmt.Errorf("tapptarock: nil card in state")
		}
		d, v := c.GetDesign(), c.GetValue()
		switch {
		case d >= 1 && d <= KoenigrufenSuitCnt && v >= 1 && v <= KoenigrufenSuitMaxValue:
		case d == KoenigrufenTrumpDesign && v >= 1 && v <= KoenigrufenMaxTrump:
		case d == KoenigrufenSkusDesign && v == KoenigrufenSkusValue:
		default:
			return fmt.Errorf("tapptarock: card out of range (design %d, value %d)", d, v)
		}
	}
	return nil
}
