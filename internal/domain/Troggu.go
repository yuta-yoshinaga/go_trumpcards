//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// 卓とデッキの形。**French Tarot と同じ 78 枚タローデッキを使う。**
//
// このゲームを他のタロー系と分けているのは契約のほうで、**契約ごとに勝ち方が
// 変わる** ── 最多得点を狙う契約と、1 トリックだけ取る契約と、1 つも取らない
// 契約が同じ卓に並ぶ。デッキ・アトゥ序列・カードポイントは流用する。同じ
// タロー札に 2 通りの点数表を持たせると、どちらが本物か分からなくなる。
const (
	// TrogguPlayerCnt 席数 (人間 1 + CPU 3)。
	TrogguPlayerCnt = 4
	// TrogguHandSize 1 人の手札枚数。
	TrogguHandSize = FrenchTarotHandSize
	// TrogguTalonSize 場札 (タロン) の枚数。
	TrogguTalonSize = FrenchTarotChienSize
	// TrogguTrickCount 1 ディールのトリック数。
	TrogguTrickCount = FrenchTarotTrickCount
	// TrogguDeckSize デッキ総枚数 (78 枚)。
	TrogguDeckSize = FrenchTarotDeckSize
	// TrogguMinDeals マッチ最小ディール数。
	TrogguMinDeals = 1
	// TrogguMaxDeals マッチ最大ディール数。
	TrogguMaxDeals = 12
	// TrogguDefaultDeals 既定のディール数。
	TrogguDefaultDeals = 4
	// TrogguBaseGamePoints 基本点。
	TrogguBaseGamePoints = 10
)

// TrogguBid 契約 (入札) 種別。値が大きいほど高い入札。
//
// **契約ごとに「勝ち」の意味が違う。** Solo は最多得点、Trois は 3 トリック、
// Piccolo はちょうど 1 トリック、Misère は 1 つも取らないこと ── 同じ卓で
// 目標が正負どちらにも振れるのがこのゲームの骨格。
type TrogguBid int

// Troggu の契約定数
const (
	// TrogguBidPass パス / 未入札
	TrogguBidPass TrogguBid = 0
	// TrogguBidTrois トロワ (単独で 3 トリック以上)
	TrogguBidTrois TrogguBid = 1
	// TrogguBidSolo ソロ (単独でカードポイントの過半)
	TrogguBidSolo TrogguBid = 2
	// TrogguBidPiccolo ピッコロ (単独でちょうど 1 トリック)
	TrogguBidPiccolo TrogguBid = 3
	// TrogguBidMisere ミゼール (単独で 1 トリックも取らない)
	TrogguBidMisere TrogguBid = 4
)

// 契約ごとの目標。
const (
	// TrogguTroisTricks トロワの必要トリック数。
	TrogguTroisTricks = 3
	// TrogguPiccoloTricks ピッコロの必要トリック数 (ちょうどこの数)。
	TrogguPiccoloTricks = 1
)

// TrogguPhase ゲームフェーズ
type TrogguPhase int

// Troggu のフェーズ定数
const (
	// TrogguPhaseBid 入札フェーズ
	TrogguPhaseBid TrogguPhase = 0
	// TrogguPhasePlay トリックプレイフェーズ
	TrogguPhasePlay TrogguPhase = 1
	// TrogguPhaseTrickEnd トリック終了フェーズ
	TrogguPhaseTrickEnd TrogguPhase = 2
	// TrogguPhaseRoundEnd ディール終了フェーズ
	TrogguPhaseRoundEnd TrogguPhase = 3
	// TrogguPhaseGameEnd ゲーム終了フェーズ
	TrogguPhaseGameEnd TrogguPhase = 4
)

// TrogguOutcome ディール結果 (デクレアラー視点)
type TrogguOutcome int

// Troggu のディール結果定数
const (
	// TrogguOutcomeNone 未確定 (誰も落札しなかったディールを含む)
	TrogguOutcomeNone TrogguOutcome = 0
	// TrogguOutcomeWin デクレアラーが契約を達成
	TrogguOutcomeWin TrogguOutcome = 1
	// TrogguOutcomeLoss デクレアラーが契約を失敗
	TrogguOutcomeLoss TrogguOutcome = 2
)

// TrogguBreakdown ディール精算の内訳。
type TrogguBreakdown struct {
	// Contract 成立した契約。
	Contract TrogguBid `json:"contract"`
	// ContractName 契約の識別子 (i18n キーに使う)。
	ContractName string `json:"contractName"`
	// DeclarerPoints デクレアラーのカードポイント (ハーフポイント表記の実点)。
	DeclarerPoints int `json:"declarerPoints"`
	// DeclarerTricks デクレアラーが取ったトリック数。
	DeclarerTricks int `json:"declarerTricks"`
	// Target 目標値 (Solo は必要点、それ以外は必要トリック数)。
	Target int `json:"target"`
	// TargetIsTricks 目標がトリック数か (false なら点数)。
	TargetIsTricks bool `json:"targetIsTricks"`
	// Won 契約を達成したか。
	Won bool `json:"won"`
	// Base 基本点。
	Base int `json:"base"`
	// Seats 席ごとの増減 (席番号順)。合計は常に 0。
	Seats []int `json:"seats"`
}

// TrogguHint ヒント情報
type TrogguHint struct {
	// Bid 推奨入札 (入札フェーズ)。nil ならパス推奨。
	Bid *int `json:"bid,omitempty"`
	// CardIndex 推奨する手札インデックス (プレイフェーズ)。
	CardIndex *int `json:"cardIndex,omitempty"`
	// Reason 理由キー。
	Reason string `json:"reason"`
}

// Troggu はトロッグの集約ルート。
type Troggu struct {
	deck        []*Card
	deckDrawCnt int
	players     []*TrogguPlayer
	config      TrogguConfig

	phase            TrogguPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int

	// --- 入札 ---
	bidPlayerIdx  int
	bidActedCnt   int
	highestBid    TrogguBid
	highestBidder int
	passed        [TrogguPlayerCnt]bool

	// --- 契約 ---
	declarerIdx int
	contract    TrogguBid

	// --- 場札 ---
	talon []*Card

	// --- 精算 ---
	playerScores    [TrogguPlayerCnt]int
	lastTrickWinner int
	lastTrickCards  []*Card
	outcome         TrogguOutcome
	breakdown       *TrogguBreakdown
	scored          bool
	gameEndFlag     bool
	winnerPlayer    int

	actionLogBase
}

// NewTroggu コンストラクタ。
func NewTroggu(players []*TrogguPlayer, config TrogguConfig) *Troggu {
	return &Troggu{
		players:         players,
		config:          config,
		winnerPlayer:    -1,
		lastTrickWinner: -1,
		declarerIdx:     -1,
		highestBidder:   -1,
		contract:        TrogguBidPass,
	}
}

// NewDefaultTroggu 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultTroggu() *Troggu {
	players := make([]*TrogguPlayer, TrogguPlayerCnt)
	players[0] = NewTrogguPlayer(true)
	for i := 1; i < TrogguPlayerCnt; i++ {
		players[i] = NewTrogguPlayer(false)
	}
	return NewTroggu(players, DefaultTrogguConfig())
}

// --- ゲーム進行 ---

// Reset ゲーム初期化。
func (g *Troggu) Reset() {
	if err := g.config.Validate(); err != nil {
		g.config = DefaultTrogguConfig()
	}
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.playerScores = [TrogguPlayerCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のディールを開始する。
func (g *Troggu) NextRound() {
	if g.gameEndFlag || g.phase != TrogguPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % TrogguPlayerCnt
	g.startRound()
}

// startRound 手札を配り、入札フェーズを開始する。
func (g *Troggu) startRound() {
	g.trickNumber = 0
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.lastTrickWinner = -1
	g.lastTrickCards = nil
	g.declarerIdx = -1
	g.contract = TrogguBidPass
	g.talon = nil
	g.outcome = TrogguOutcomeNone
	g.breakdown = nil
	g.scored = false
	g.passed = [TrogguPlayerCnt]bool{}
	g.highestBid = TrogguBidPass
	g.highestBidder = -1
	g.bidActedCnt = 0
	for _, p := range g.players {
		p.ResetRound()
	}
	g.deal()
	g.sortAllHands()
	g.bidPlayerIdx = (g.dealerIdx + 1) % TrogguPlayerCnt
	g.currentPlayerIdx = g.bidPlayerIdx
	g.phase = TrogguPhaseBid
	g.appendLog(-1, "deal",
		fmt.Sprintf("deal %d: %d cards each, talon %d", g.roundNumber, TrogguHandSize, len(g.talon)), nil)
}

// deal 3 枚パケットで各プレイヤーへ 18 枚を配り、場札 6 枚を脇に置く。
func (g *Troggu) deal() {
	g.deck = buildFrenchTarotDeck()
	rand.Shuffle(len(g.deck), func(i, j int) { g.deck[i], g.deck[j] = g.deck[j], g.deck[i] })
	g.deckDrawCnt = 0
	g.talon = make([]*Card, 0, TrogguTalonSize)
	for range TrogguHandSize / 3 {
		for j := range TrogguPlayerCnt {
			idx := (g.dealerIdx + 1 + j) % TrogguPlayerCnt
			for range 3 {
				if c := drawFromDeck(g.deck, &g.deckDrawCnt); c != nil {
					g.players[idx].AddCard(c)
				}
			}
		}
	}
	for range TrogguTalonSize {
		if c := drawFromDeck(g.deck, &g.deckDrawCnt); c != nil {
			g.talon = append(g.talon, c)
		}
	}
}

// --- 入札 ---

// PlayerBid 人間プレイヤーが入札する。
func (g *Troggu) PlayerBid(bid TrogguBid) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TrogguPhaseBid {
		return NewDomainError(ErrWrongPhase, "not in bidding phase")
	}
	if !g.players[g.bidPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if !TrogguValidBid(bid) {
		return NewDomainError(ErrInvalidPlay, fmt.Sprintf("bid %d cannot be declared", bid))
	}
	if bid <= g.highestBid {
		return NewDomainError(ErrInvalidPlay,
			fmt.Sprintf("bid %d does not beat the current bid %d", bid, g.highestBid))
	}
	g.applyBid(g.bidPlayerIdx, bid)
	return nil
}

// TrogguValidBid 宣言できる契約かを返す (パスは入札ではない)。
func TrogguValidBid(bid TrogguBid) bool {
	return bid >= TrogguBidTrois && bid <= TrogguBidMisere
}

// PlayerPass 人間プレイヤーがパスする。
func (g *Troggu) PlayerPass() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TrogguPhaseBid {
		return NewDomainError(ErrWrongPhase, "not in bidding phase")
	}
	if !g.players[g.bidPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.applyPass(g.bidPlayerIdx)
	return nil
}

// CpuBid CPU の入札を 1 手進める。
func (g *Troggu) CpuBid() {
	if g.gameEndFlag || g.phase != TrogguPhaseBid {
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
func (g *Troggu) applyBid(idx int, bid TrogguBid) {
	g.highestBid = bid
	g.highestBidder = idx
	g.bidActedCnt++
	g.appendLog(idx, "bid", fmt.Sprintf("%s bids %s", g.playerName(idx), TrogguBidName(bid)), nil)
	g.advanceBid()
}

// applyPass パスを記録して手番を進める。
func (g *Troggu) applyPass(idx int) {
	g.passed[idx] = true
	g.bidActedCnt++
	g.appendLog(idx, "pass", fmt.Sprintf("%s passes", g.playerName(idx)), nil)
	g.advanceBid()
}

// advanceBid 次の入札者を選び、全員が表明したら契約を確定する。
func (g *Troggu) advanceBid() {
	if g.bidActedCnt >= TrogguPlayerCnt {
		g.finalizeBid()
		return
	}
	for range TrogguPlayerCnt {
		g.bidPlayerIdx = (g.bidPlayerIdx + 1) % TrogguPlayerCnt
		if !g.passed[g.bidPlayerIdx] {
			g.currentPlayerIdx = g.bidPlayerIdx
			return
		}
	}
	g.finalizeBid()
}

// finalizeBid 落札者を決めてプレイフェーズへ進む。
//
// **誰も落札しなければ配り直す。** Troggu は契約が無いと勝ち方そのものが
// 決まらない ── 目標の無いディールを 18 トリック打たせても誰の得にもならない。
func (g *Troggu) finalizeBid() {
	if g.highestBidder < 0 {
		g.appendLog(-1, "contract", "everyone passed: the deal is thrown in", nil)
		g.outcome = TrogguOutcomeNone
		g.breakdown = nil
		g.scored = true
		g.phase = TrogguPhaseRoundEnd
		if g.roundNumber >= g.config.TargetDeals {
			g.finishGame()
		}
		return
	}
	g.declarerIdx = g.highestBidder
	g.contract = g.highestBid
	g.appendLog(g.declarerIdx, "contract",
		fmt.Sprintf("%s plays %s", g.playerName(g.declarerIdx), TrogguBidName(g.contract)), nil)
	g.startPlay()
}

// --- プレイ ---

// startPlay トリックプレイを開始する。
func (g *Troggu) startPlay() {
	g.phase = TrogguPhasePlay
	g.trickNumber = 1
	g.leadPlayerIdx = g.declarerIdx
	g.currentPlayerIdx = g.leadPlayerIdx
	g.currentTrick = nil
}

// PlayerPlayCard 人間が手札を 1 枚出す。
func (g *Troggu) PlayerPlayCard(handIdx int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TrogguPhasePlay {
		return NewDomainError(ErrWrongPhase, "not in play phase")
	}
	p := g.players[g.currentPlayerIdx]
	if !p.GetIsHuman() {
		return ErrNotHumanTurn
	}
	if handIdx < 0 || handIdx >= p.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", handIdx))
	}
	if !trogguContains(g.GetValidPlayIndices(g.currentPlayerIdx), handIdx) {
		return NewDomainError(ErrInvalidPlay, "that card does not follow the lead")
	}
	g.playCard(g.currentPlayerIdx, handIdx)
	return nil
}

// CpuPlayCard CPU の 1 手を進める。
func (g *Troggu) CpuPlayCard() {
	if g.gameEndFlag || g.phase != TrogguPhasePlay {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	g.playCard(g.currentPlayerIdx, g.cpuSelectPlayCard(g.currentPlayerIdx))
}

// playCard 1 枚を場に出し、トリックが揃ったら決着させる。
func (g *Troggu) playCard(playerIdx, handIdx int) {
	card := g.players[playerIdx].RemoveCard(handIdx)
	if card == nil {
		return
	}
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", g.playerName(playerIdx), frenchTarotCardStr(card)), []*Card{card})

	if len(g.currentTrick) < TrogguPlayerCnt {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % TrogguPlayerCnt
		return
	}
	g.finishTrick()
}

// finishTrick トリックの勝者を決め、次のトリックへ進む。
func (g *Troggu) finishTrick() {
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
	g.phase = TrogguPhaseTrickEnd
	g.currentPlayerIdx = winner
	g.leadPlayerIdx = winner
}

// NextTrick トリック終了フェーズから次のトリックへ進む。
func (g *Troggu) NextTrick() {
	if g.gameEndFlag || g.phase != TrogguPhaseTrickEnd {
		return
	}
	if g.trickNumber >= TrogguTrickCount {
		g.finishRound()
		return
	}
	g.trickNumber++
	g.phase = TrogguPhasePlay
	g.currentPlayerIdx = g.leadPlayerIdx
}

// trickWinner 現在のトリックの勝者を返す。
func (g *Troggu) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return g.leadPlayerIdx
	}
	led := g.ledSuit()
	best, bestRank := g.currentTrick[0].PlayerIdx, -2
	for _, tc := range g.currentTrick {
		if r := frenchTarotWinRank(tc.Card, led); r > bestRank {
			best, bestRank = tc.PlayerIdx, r
		}
	}
	return best
}

// ledSuit 現在のトリックの実効リードスートを返す (空なら -1)。
//
// **エクスキューズはリードスートを決めない。** 序列の外にある札なので、
// 先頭に出されたときは次の札がスートを決める。
func (g *Troggu) ledSuit() int {
	for _, tc := range g.currentTrick {
		if frenchTarotIsExcuse(tc.Card) {
			continue
		}
		if frenchTarotIsTrump(tc.Card) {
			return FrenchTarotTrumpDesign
		}
		return tc.Card.GetDesign()
	}
	return -1
}

// GetValidPlayIndices 出せる手札のインデックスを返す。
//
// リードスートに従う義務、持っていなければアトゥを出す義務。エクスキューズは
// いつでも出せる。
func (g *Troggu) GetValidPlayIndices(playerIdx int) []int {
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
	excuse := make([]int, 0, 1)
	follow := make([]int, 0, len(all))
	trumps := make([]int, 0, len(all))
	for _, i := range all {
		c := p.GetCard(i)
		switch {
		case frenchTarotIsExcuse(c):
			excuse = append(excuse, i)
		case frenchTarotIsTrump(c):
			trumps = append(trumps, i)
			if led == FrenchTarotTrumpDesign {
				follow = append(follow, i)
			}
		case c.GetDesign() == led:
			follow = append(follow, i)
		}
	}
	if len(follow) > 0 {
		return append(follow, excuse...)
	}
	if len(trumps) > 0 {
		return append(trumps, excuse...)
	}
	return all
}

// cpuSelectPlayCard CPU の 1 手を選ぶ。
//
// **契約によって狙いが逆になる。** ミゼール／ピッコロのデクレアラーは取っては
// いけないので、勝ちにいかず安い札を落とす。
func (g *Troggu) cpuSelectPlayCard(playerIdx int) int {
	valid := g.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := g.players[playerIdx]
	if g.config.CpuDifficulty == TrogguCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	avoid := g.avoidsTricks(playerIdx)
	pick, best := valid[0], 0
	for i, idx := range valid {
		rank := frenchTarotWinRank(p.GetCard(idx), g.ledSuit())
		score := rank
		if avoid {
			score = -rank
		}
		if i == 0 || score > best {
			pick, best = idx, score
		}
	}
	return pick
}

// avoidsTricks その席がトリックを避けたいかを返す。
func (g *Troggu) avoidsTricks(playerIdx int) bool {
	switch g.contract {
	case TrogguBidMisere, TrogguBidPiccolo:
		// デクレアラーは取りたくない。防御側はむしろ押し付けたい。
		return playerIdx == g.declarerIdx
	default:
		return false
	}
}

// cpuSelectBid CPU の入札を選ぶ。
//
// 強い手はソロ、アトゥも絵札も無い手はミゼールが狙える ── どちらでもなければ
// パスする。
func (g *Troggu) cpuSelectBid(playerIdx int) (TrogguBid, bool) {
	p := g.players[playerIdx]
	trumps, honours, highTrumps := 0, 0, 0
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if frenchTarotIsTrump(c) || frenchTarotIsExcuse(c) {
			trumps++
			if frenchTarotIsTrump(c) && c.GetValue() >= 15 {
				highTrumps++
			}
		}
		if frenchTarotCardHalfPoints(c) >= 7 {
			honours++
		}
	}
	switch {
	case trumps >= 9 && highTrumps >= 4 && TrogguBidSolo > g.highestBid:
		return TrogguBidSolo, true
	case honours == 0 && highTrumps == 0 && TrogguBidMisere > g.highestBid:
		return TrogguBidMisere, true
	case trumps >= 5 && TrogguBidTrois > g.highestBid:
		return TrogguBidTrois, true
	default:
		return TrogguBidPass, false
	}
}

// --- 精算 ---

// finishRound ディールを精算する。
func (g *Troggu) finishRound() {
	if g.scored {
		return
	}
	g.scored = true
	// 場札はデクレアラーの得点に加える。
	if len(g.talon) > 0 && g.declarerIdx >= 0 {
		g.players[g.declarerIdx].AddTrick(g.talon)
		g.talon = nil
	}
	g.breakdown = g.scoreDeal()
	if g.breakdown.Won {
		g.outcome = TrogguOutcomeWin
	} else {
		g.outcome = TrogguOutcomeLoss
	}
	for i, delta := range g.breakdown.Seats {
		g.playerScores[i] += delta
	}
	g.phase = TrogguPhaseRoundEnd
	g.appendLog(-1, "score",
		fmt.Sprintf("deal %d scored (%s)", g.roundNumber, TrogguBidName(g.contract)), nil)
	if g.roundNumber >= g.config.TargetDeals {
		g.finishGame()
	}
}

// trogguTotalHalfPoints デッキ全体のハーフポイント合計を返す。
//
// **表を書き写さず数える。** 閾値を定数で持つと、点数表を変えたときに片方だけが
// ずれる。
func trogguTotalHalfPoints() int {
	total := 0
	for _, c := range buildFrenchTarotDeck() {
		total += frenchTarotCardHalfPoints(c)
	}
	return total
}

// cardHalfPointsOf 席が獲得した札のハーフポイント合計を返す。
func (g *Troggu) cardHalfPointsOf(idx int) int {
	total := 0
	for _, trick := range g.players[idx].GetTricksTaken() {
		for _, c := range trick {
			total += frenchTarotCardHalfPoints(c)
		}
	}
	return total
}

// scoreDeal 契約ごとの成否を判定して精算する (ゼロサム)。
//
// **契約ごとに見るものが違う。** ソロだけがカードポイント、他の 3 つは
// トリック数 ── ピッコロは「ちょうど 1 つ」なので、多すぎても失敗する。
func (g *Troggu) scoreDeal() *TrogguBreakdown {
	declPoints := g.cardHalfPointsOf(g.declarerIdx)
	declTricks := g.players[g.declarerIdx].GetTrickCount()
	total := trogguTotalHalfPoints()

	bd := &TrogguBreakdown{
		Contract:       g.contract,
		ContractName:   TrogguBidName(g.contract),
		DeclarerPoints: declPoints,
		DeclarerTricks: declTricks,
		Seats:          make([]int, TrogguPlayerCnt),
	}
	switch g.contract {
	case TrogguBidSolo:
		bd.Target = total/2 + 1
		bd.TargetIsTricks = false
		bd.Won = 2*declPoints > total
	case TrogguBidTrois:
		bd.Target = TrogguTroisTricks
		bd.TargetIsTricks = true
		bd.Won = declTricks >= TrogguTroisTricks
	case TrogguBidPiccolo:
		bd.Target = TrogguPiccoloTricks
		bd.TargetIsTricks = true
		// **ちょうど 1 つ。** 多すぎても失敗する。
		bd.Won = declTricks == TrogguPiccoloTricks
	default: // Misère
		bd.Target = 0
		bd.TargetIsTricks = true
		bd.Won = declTricks == 0
	}

	bd.Base = TrogguBaseGamePoints * TrogguBidMultiplier(g.contract)
	sign := 1
	if !bd.Won {
		sign = -1
	}
	// 単独契約なので、デクレアラーは 3 人ぶんを受け取るか払う。
	for i := range bd.Seats {
		if i == g.declarerIdx {
			bd.Seats[i] = sign * 3 * bd.Base
			continue
		}
		bd.Seats[i] = -sign * bd.Base
	}
	return bd
}

// TrogguBidMultiplier 契約の倍率を返す。
//
// そろえにくい契約ほど高い ── ミゼール (1 つも取らない) が最も難しく、
// トロワ (3 トリック) が最も易しい。
func TrogguBidMultiplier(bid TrogguBid) int {
	switch bid {
	case TrogguBidMisere:
		return 4
	case TrogguBidPiccolo:
		return 3
	case TrogguBidSolo:
		return 2
	default:
		return 1
	}
}

// finishGame 通算得点で勝者を決める。
func (g *Troggu) finishGame() {
	best, tied := 0, false
	for i := 1; i < TrogguPlayerCnt; i++ {
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
	g.phase = TrogguPhaseGameEnd
	g.appendLog(g.winnerPlayer, "gameEnd", "match over", nil)
}

// --- 補助 ---

// trogguContains スライスに値が含まれるか。
func trogguContains(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// sortAllHands 全員の手札を並べ替える。
func (g *Troggu) sortAllHands() {
	for _, p := range g.players {
		trogguSortHand(p)
	}
}

// trogguSortHand 手札をスート順・アトゥは番号順に並べる。
func trogguSortHand(p *TrogguPlayer) {
	cards := make([]*Card, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		cards = append(cards, p.GetCard(i))
	}
	for i := 1; i < len(cards); i++ {
		for j := i; j > 0 && trogguSortKey(cards[j]) < trogguSortKey(cards[j-1]); j-- {
			cards[j], cards[j-1] = cards[j-1], cards[j]
		}
	}
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// trogguSortKey 手札の並び順キー。
func trogguSortKey(c *Card) int {
	switch {
	case c == nil:
		return 1 << 20
	case frenchTarotIsExcuse(c):
		return 2000
	case frenchTarotIsTrump(c):
		return 1000 + c.GetValue()
	default:
		return c.GetDesign()*100 + c.GetValue()
	}
}

// playerName 席の表示名を返す。
func (g *Troggu) playerName(i int) string {
	if i < 0 || i >= len(g.players) {
		return "-"
	}
	if g.players[i].GetIsHuman() {
		return "YOU"
	}
	return fmt.Sprintf("CPU%d", i)
}

// TrogguBidName 契約の識別子を返す (i18n キーの一部に使う)。
func TrogguBidName(bid TrogguBid) string {
	switch bid {
	case TrogguBidTrois:
		return "trois"
	case TrogguBidSolo:
		return "solo"
	case TrogguBidPiccolo:
		return "piccolo"
	case TrogguBidMisere:
		return "misere"
	default:
		return "pass"
	}
}

// --- アクセサ ---

// GetPlayers 席の一覧を返す。
func (g *Troggu) GetPlayers() []*TrogguPlayer { return g.players }

// GetPlayer 指定した席を返す (範囲外なら nil)。
func (g *Troggu) GetPlayer(i int) *TrogguPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetPlayerCnt 席数を返す。
func (g *Troggu) GetPlayerCnt() int { return len(g.players) }

// GetConfig 設定を返す。
func (g *Troggu) GetConfig() TrogguConfig { return g.config }

// SetConfig 設定を差し替える (次の Reset で反映)。
func (g *Troggu) SetConfig(c TrogguConfig) { g.config = c }

// GetPhase 現在のフェーズを返す。
func (g *Troggu) GetPhase() TrogguPhase { return g.phase }

// GetRoundNumber 現在のディール番号を返す。
func (g *Troggu) GetRoundNumber() int { return g.roundNumber }

// GetTrickNumber 現在のトリック番号を返す。
func (g *Troggu) GetTrickNumber() int { return g.trickNumber }

// GetCurrentPlayerIdx 手番の席を返す。
func (g *Troggu) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// GetCurrentTrick 場に出ている札を返す。
func (g *Troggu) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// GetDealerIdx 親の席を返す。
func (g *Troggu) GetDealerIdx() int { return g.dealerIdx }

// GetBidPlayerIdx 入札の手番を返す。
func (g *Troggu) GetBidPlayerIdx() int { return g.bidPlayerIdx }

// GetHighestBid 現在の最高入札を返す。
func (g *Troggu) GetHighestBid() TrogguBid { return g.highestBid }

// GetDeclarerIdx デクレアラーの席を返す (-1 = 未確定 / 流局)。
func (g *Troggu) GetDeclarerIdx() int { return g.declarerIdx }

// GetContract 成立した契約を返す。
func (g *Troggu) GetContract() TrogguBid { return g.contract }

// GetTalonSize 場札の枚数を返す。
func (g *Troggu) GetTalonSize() int { return len(g.talon) }

// GetLastTrickWinner 直前のトリックの勝者を返す。
func (g *Troggu) GetLastTrickWinner() int { return g.lastTrickWinner }

// GetLastTrickCards 直前のトリックの札を返す。
func (g *Troggu) GetLastTrickCards() []*Card { return g.lastTrickCards }

// GetOutcome ディール結果を返す。
func (g *Troggu) GetOutcome() TrogguOutcome { return g.outcome }

// GetBreakdown 直近ディールの精算内訳を返す。
func (g *Troggu) GetBreakdown() *TrogguBreakdown { return g.breakdown }

// GetPlayerScore 指定席の通算得点を返す。
func (g *Troggu) GetPlayerScore(i int) int {
	if i < 0 || i >= TrogguPlayerCnt {
		return 0
	}
	return g.playerScores[i]
}

// GetCardPoints 指定席が獲得したカードポイント (ハーフポイント) を返す。
func (g *Troggu) GetCardPoints(i int) int {
	if i < 0 || i >= len(g.players) {
		return 0
	}
	return g.cardHalfPointsOf(i)
}

// GetGameEndFlag 終局したかを返す。
func (g *Troggu) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 終局時の勝者を返す (-1 = 引き分け/未決)。
func (g *Troggu) GetWinnerPlayer() int { return g.winnerPlayer }

// HumanSeat 人間の席を返す (居なければ 0)。
func (g *Troggu) HumanSeat() int {
	if i := findHumanIdx(g.players); i >= 0 {
		return i
	}
	return 0
}

// IsHumanTurn 人間の入力を待っているかを返す。
func (g *Troggu) IsHumanTurn() bool {
	if g.gameEndFlag {
		return false
	}
	switch g.phase {
	case TrogguPhaseBid:
		return g.players[g.bidPlayerIdx].GetIsHuman()
	case TrogguPhasePlay:
		return g.players[g.currentPlayerIdx].GetIsHuman()
	default:
		return false
	}
}

// GetHint 人間の手番における推奨手を返す。
func (g *Troggu) GetHint() *TrogguHint {
	if g.gameEndFlag || !g.IsHumanTurn() {
		return nil
	}
	seat := g.HumanSeat()
	switch g.phase {
	case TrogguPhaseBid:
		bid, ok := g.cpuSelectBid(seat)
		if !ok {
			return &TrogguHint{Reason: "pass_weak_hand"}
		}
		v := int(bid)
		return &TrogguHint{Bid: &v, Reason: "bid_" + TrogguBidName(bid)}
	case TrogguPhasePlay:
		idx := g.cpuSelectPlayCard(seat)
		reason := "play_win"
		if g.avoidsTricks(seat) {
			reason = "play_duck"
		}
		return &TrogguHint{CardIndex: &idx, Reason: reason}
	default:
		return nil
	}
}

// --- JSON ---

// trogguJSON は Troggu の JSON 表現。
type trogguJSON struct {
	Players         []*TrogguPlayer   `json:"pl"`
	Config          TrogguConfig      `json:"cf"`
	Phase           TrogguPhase       `json:"ph"`
	RoundNumber     int               `json:"rn"`
	TrickNumber     int               `json:"tn"`
	CurrentPlayer   int               `json:"cp"`
	CurrentTrick    []*TrickCard      `json:"ct"`
	LeadPlayer      int               `json:"lp"`
	DealerIdx       int               `json:"di"`
	BidPlayerIdx    int               `json:"bp"`
	BidActedCnt     int               `json:"ba"`
	HighestBid      TrogguBid         `json:"hb"`
	HighestBidder   int               `json:"hr"`
	Passed          []bool            `json:"ps"`
	DeclarerIdx     int               `json:"dc"`
	Contract        TrogguBid         `json:"co"`
	Talon           []*Card           `json:"tl"`
	PlayerScores    []int             `json:"sc"`
	LastTrickWinner int               `json:"lw"`
	LastTrickCards  []*Card           `json:"lc"`
	Outcome         TrogguOutcome     `json:"oc"`
	Breakdown       *TrogguBreakdown  `json:"bd"`
	Scored          bool              `json:"sd"`
	GameEndFlag     bool              `json:"ge"`
	WinnerPlayer    int               `json:"wp"`
	ActionLog       []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Troggu) MarshalJSON() ([]byte, error) {
	return json.Marshal(trogguJSON{
		Players: g.players, Config: g.config, Phase: g.phase,
		RoundNumber: g.roundNumber, TrickNumber: g.trickNumber,
		CurrentPlayer: g.currentPlayerIdx, CurrentTrick: g.currentTrick,
		LeadPlayer: g.leadPlayerIdx, DealerIdx: g.dealerIdx,
		BidPlayerIdx: g.bidPlayerIdx, BidActedCnt: g.bidActedCnt,
		HighestBid: g.highestBid, HighestBidder: g.highestBidder,
		Passed: g.passed[:], DeclarerIdx: g.declarerIdx, Contract: g.contract,
		Talon: g.talon, PlayerScores: g.playerScores[:],
		LastTrickWinner: g.lastTrickWinner, LastTrickCards: g.lastTrickCards,
		Outcome: g.outcome, Breakdown: g.breakdown, Scored: g.scored,
		GameEndFlag: g.gameEndFlag, WinnerPlayer: g.winnerPlayer, ActionLog: g.actionLog,
	})
}

// trogguMaxSliceLen 復元時のスライス長上限。
const trogguMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
//
// **契約とデクレアラーの整合まで見る。** 「契約が決まっているのにデクレアラーが
// 居ない」保存は範囲検査を通ってしまい、精算のときに誰の得点でもない札を数える
// ことになる。
func (g *Troggu) UnmarshalJSON(data []byte) error {
	var j trogguJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > trogguMaxSliceLen || len(j.Talon) > trogguMaxSliceLen ||
		len(j.ActionLog) > trogguMaxSliceLen || len(j.CurrentTrick) > trogguMaxSliceLen {
		return fmt.Errorf("troggu: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("troggu: invalid config: %w", err)
	}
	if len(j.Players) != TrogguPlayerCnt {
		return fmt.Errorf("troggu: expected %d players, got %d", TrogguPlayerCnt, len(j.Players))
	}
	for _, p := range j.Players {
		if p == nil {
			return fmt.Errorf("troggu: nil player in state")
		}
	}
	if j.Phase < TrogguPhaseBid || j.Phase > TrogguPhaseGameEnd {
		return fmt.Errorf("troggu: invalid phase %d", j.Phase)
	}
	if j.RoundNumber < 1 || j.RoundNumber > j.Config.TargetDeals {
		return fmt.Errorf("troggu: round %d out of range", j.RoundNumber)
	}
	if j.TrickNumber < 0 || j.TrickNumber > TrogguTrickCount {
		return fmt.Errorf("troggu: trick %d out of range", j.TrickNumber)
	}
	for name, idx := range map[string]int{"current player": j.CurrentPlayer, "bid player": j.BidPlayerIdx, "dealer": j.DealerIdx} {
		if idx < 0 || idx >= TrogguPlayerCnt {
			return fmt.Errorf("troggu: %s out of range", name)
		}
	}
	for name, idx := range map[string]int{"declarer": j.DeclarerIdx, "winner": j.WinnerPlayer, "highest bidder": j.HighestBidder} {
		if idx < -1 || idx >= TrogguPlayerCnt {
			return fmt.Errorf("troggu: %s out of range", name)
		}
	}
	if j.Contract < TrogguBidPass || j.Contract > TrogguBidMisere {
		return fmt.Errorf("troggu: invalid contract %d", j.Contract)
	}
	// **契約があるならデクレアラーが居る。**
	if j.Contract != TrogguBidPass && j.DeclarerIdx < 0 {
		return fmt.Errorf("troggu: a contract requires a declarer")
	}
	if j.Contract == TrogguBidPass && j.DeclarerIdx >= 0 {
		return fmt.Errorf("troggu: a declarer requires a contract")
	}
	if err := trogguValidateCards(j.Talon); err != nil {
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
	g.passed = [TrogguPlayerCnt]bool{}
	copy(g.passed[:], j.Passed)
	g.declarerIdx = j.DeclarerIdx
	g.contract = j.Contract
	g.talon = j.Talon
	g.playerScores = [TrogguPlayerCnt]int{}
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

// trogguValidateCards 札に nil や範囲外の design/value が無いか検証する。
func trogguValidateCards(cards []*Card) error {
	for _, c := range cards {
		if c == nil {
			return fmt.Errorf("troggu: nil card in state")
		}
		if !frenchTarotValidCard(c) {
			return fmt.Errorf("troggu: card out of range (design %d, value %d)", c.GetDesign(), c.GetValue())
		}
	}
	return nil
}
