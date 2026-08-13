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
	// ZwanzigerrufenPlayerCnt 席数 (人間 1 + CPU 3)。
	ZwanzigerrufenPlayerCnt = 4
	// ZwanzigerrufenHandSize 1 人の手札枚数。
	ZwanzigerrufenHandSize = 12
	// ZwanzigerrufenTalonSize 場札 (タロン) の枚数。
	ZwanzigerrufenTalonSize = 6
	// ZwanzigerrufenTrickCount 1 ディールのトリック数。
	ZwanzigerrufenTrickCount = 12
	// ZwanzigerrufenDeckSize デッキ総枚数。
	ZwanzigerrufenDeckSize = KoenigrufenDeckSize
	// ZwanzigerrufenMinDeals マッチ最小ディール数。
	ZwanzigerrufenMinDeals = 1
	// ZwanzigerrufenMaxDeals マッチ最大ディール数。
	ZwanzigerrufenMaxDeals = 12
	// ZwanzigerrufenDefaultDeals 既定のディール数。
	ZwanzigerrufenDefaultDeals = 4
	// ZwanzigerrufenBaseGameValue 基本点。
	ZwanzigerrufenBaseGameValue = 10
	// ZwanzigerrufenTrischakenLoss Trischaken で最多失点者が失う点。
	ZwanzigerrufenTrischakenLoss = 3
)

// 呼び札。**呼ぶのは切り札の 20 番。**
const (
	// ZwanzigerrufenCallTrump 呼び札の切り札番号 (XX)。
	ZwanzigerrufenCallTrump = 20
	// ZwanzigerrufenMinCallTrump 呼び下げの下限。デクレアラーが XX・XIX・XVIII を
	// すべて持っていることは 12 枚の手札では起こりうるので、下限まで下げても
	// 呼べないときは単独扱いにする。
	ZwanzigerrufenMinCallTrump = 18
)

// ZwanzigerrufenBid 入札 (コントラクト) 種別。値が大きいほど高い入札。
type ZwanzigerrufenBid int

// Zwanzigerrufen の入札定数
const (
	// ZwanzigerrufenBidPass パス / 未入札
	ZwanzigerrufenBidPass ZwanzigerrufenBid = 0
	// ZwanzigerrufenBidTrischaken トリシャーケン (全員パス時の契約)
	ZwanzigerrufenBidTrischaken ZwanzigerrufenBid = 1
	// ZwanzigerrufenBidRufer ツヴァンツィガールーフ (20 番呼び)
	ZwanzigerrufenBidRufer ZwanzigerrufenBid = 2
	// ZwanzigerrufenBidSolo ソロ (1 対 3、場札交換なし)
	ZwanzigerrufenBidSolo ZwanzigerrufenBid = 3
)

// ZwanzigerrufenPhase ゲームフェーズ
type ZwanzigerrufenPhase int

// Zwanzigerrufen のフェーズ定数
const (
	// ZwanzigerrufenPhaseBid 入札フェーズ
	ZwanzigerrufenPhaseBid ZwanzigerrufenPhase = 0
	// ZwanzigerrufenPhaseTalon 場札交換 (discard) フェーズ
	ZwanzigerrufenPhaseTalon ZwanzigerrufenPhase = 1
	// ZwanzigerrufenPhasePlay トリックプレイフェーズ
	ZwanzigerrufenPhasePlay ZwanzigerrufenPhase = 2
	// ZwanzigerrufenPhaseTrickEnd トリック終了フェーズ
	ZwanzigerrufenPhaseTrickEnd ZwanzigerrufenPhase = 3
	// ZwanzigerrufenPhaseRoundEnd ディール終了フェーズ
	ZwanzigerrufenPhaseRoundEnd ZwanzigerrufenPhase = 4
	// ZwanzigerrufenPhaseGameEnd ゲーム終了フェーズ
	ZwanzigerrufenPhaseGameEnd ZwanzigerrufenPhase = 5
)

// ZwanzigerrufenOutcome ディール結果 (デクレアラー側視点)
type ZwanzigerrufenOutcome int

// Zwanzigerrufen のディール結果定数
const (
	// ZwanzigerrufenOutcomeNone 未確定
	ZwanzigerrufenOutcomeNone ZwanzigerrufenOutcome = 0
	// ZwanzigerrufenOutcomeWin デクレアラー側がコントラクトを達成
	ZwanzigerrufenOutcomeWin ZwanzigerrufenOutcome = 1
	// ZwanzigerrufenOutcomeLoss デクレアラー側がコントラクトを失敗
	ZwanzigerrufenOutcomeLoss ZwanzigerrufenOutcome = 2
	// ZwanzigerrufenOutcomeTrischaken トリシャーケン (勝敗ではなく最多失点者を決める)
	ZwanzigerrufenOutcomeTrischaken ZwanzigerrufenOutcome = 3
)

// ZwanzigerrufenBreakdown ディール精算の内訳。
type ZwanzigerrufenBreakdown struct {
	// Contract 成立した契約。
	Contract ZwanzigerrufenBid `json:"contract"`
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

// ZwanzigerrufenHint ヒント情報
type ZwanzigerrufenHint struct {
	// Bid 推奨入札 (入札フェーズ)。nil ならパス推奨。
	Bid *int `json:"bid,omitempty"`
	// DiscardIndices 推奨する伏せ札 (場札交換フェーズ)。
	DiscardIndices []int `json:"discardIndices,omitempty"`
	// CardIndex 推奨する手札インデックス (プレイフェーズ)。
	CardIndex *int `json:"cardIndex,omitempty"`
	// Reason 理由キー。
	Reason string `json:"reason"`
}

// Zwanzigerrufen はツヴァンツィガールーフェンの集約ルート。
type Zwanzigerrufen struct {
	deck        []*Card
	deckDrawCnt int
	players     []*ZwanzigerrufenPlayer
	config      ZwanzigerrufenConfig

	phase            ZwanzigerrufenPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int

	// --- 入札 ---
	bidPlayerIdx  int
	bidActedCnt   int
	highestBid    ZwanzigerrufenBid
	highestBidder int
	passed        [ZwanzigerrufenPlayerCnt]bool

	// --- 契約 ---
	declarerIdx int
	contract    ZwanzigerrufenBid

	// --- 呼び札 (partnerIdx はサーバー側のみ。明かすまで Web 出力に出さない) ---
	calledTrump     int // 呼んだ切り札の番号 (18..20)、未呼び/-1
	partnerIdx      int // 秘密のパートナー (-1 = 単独)
	partnerRevealed bool

	// --- 場札 ---
	talon []*Card
	stash []*Card // 得点計上のため脇に置いた札
	// stashOwner は stash を数える側。0=デクレアラー側, 1=防御側, -1=最終トリックの勝者へ。
	stashOwner int

	// --- 精算 ---
	playerScores    [ZwanzigerrufenPlayerCnt]int
	lastTrickWinner int
	lastTrickCards  []*Card
	outcome         ZwanzigerrufenOutcome
	breakdown       *ZwanzigerrufenBreakdown
	scored          bool
	gameEndFlag     bool
	winnerPlayer    int

	actionLogBase
}

// NewZwanzigerrufen コンストラクタ。
func NewZwanzigerrufen(players []*ZwanzigerrufenPlayer, config ZwanzigerrufenConfig) *Zwanzigerrufen {
	return &Zwanzigerrufen{
		players:         players,
		config:          config,
		winnerPlayer:    -1,
		lastTrickWinner: -1,
		declarerIdx:     -1,
		highestBidder:   -1,
		contract:        ZwanzigerrufenBidPass,
		calledTrump:     -1,
		partnerIdx:      -1,
		stashOwner:      0,
	}
}

// NewDefaultZwanzigerrufen 標準の 4 人構成 (人間 1, CPU 3) と既定設定で生成する。
func NewDefaultZwanzigerrufen() *Zwanzigerrufen {
	players := make([]*ZwanzigerrufenPlayer, ZwanzigerrufenPlayerCnt)
	players[0] = NewZwanzigerrufenPlayer(true)
	for i := 1; i < ZwanzigerrufenPlayerCnt; i++ {
		players[i] = NewZwanzigerrufenPlayer(false)
	}
	return NewZwanzigerrufen(players, DefaultZwanzigerrufenConfig())
}

// --- ゲーム進行 ---

// Reset ゲーム初期化。
func (g *Zwanzigerrufen) Reset() {
	if err := g.config.Validate(); err != nil {
		g.config = DefaultZwanzigerrufenConfig()
	}
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.playerScores = [ZwanzigerrufenPlayerCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のディールを開始する。
func (g *Zwanzigerrufen) NextRound() {
	if g.gameEndFlag || g.phase != ZwanzigerrufenPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % ZwanzigerrufenPlayerCnt
	g.startRound()
}

// startRound 手札を配り、入札フェーズを開始する。
func (g *Zwanzigerrufen) startRound() {
	g.trickNumber = 0
	g.currentTrick = nil
	g.leadPlayerIdx = -1
	g.lastTrickWinner = -1
	g.lastTrickCards = nil
	g.declarerIdx = -1
	g.contract = ZwanzigerrufenBidPass
	g.calledTrump = -1
	g.partnerIdx = -1
	g.partnerRevealed = false
	g.talon = nil
	g.stash = nil
	g.stashOwner = 0
	g.outcome = ZwanzigerrufenOutcomeNone
	g.breakdown = nil
	g.scored = false
	g.passed = [ZwanzigerrufenPlayerCnt]bool{}
	g.highestBid = ZwanzigerrufenBidPass
	g.highestBidder = -1
	g.bidActedCnt = 0
	for _, p := range g.players {
		p.ResetRound()
	}
	g.deal()
	g.sortAllHands()
	g.bidPlayerIdx = (g.dealerIdx + 1) % ZwanzigerrufenPlayerCnt
	g.currentPlayerIdx = g.bidPlayerIdx
	g.phase = ZwanzigerrufenPhaseBid
	g.appendLog(-1, "deal", fmt.Sprintf("deal %d: 12 cards each, talon %d", g.roundNumber, len(g.talon)), nil)
}

// deal 3 枚パケットで各プレイヤーへ 12 枚を配り、場札 6 枚を脇に置く。
func (g *Zwanzigerrufen) deal() {
	g.deck = buildKoenigrufenDeck()
	rand.Shuffle(len(g.deck), func(i, j int) { g.deck[i], g.deck[j] = g.deck[j], g.deck[i] })
	g.deckDrawCnt = 0
	g.talon = make([]*Card, 0, ZwanzigerrufenTalonSize)
	for range ZwanzigerrufenHandSize / 3 {
		for j := range ZwanzigerrufenPlayerCnt {
			idx := (g.dealerIdx + 1 + j) % ZwanzigerrufenPlayerCnt
			for range 3 {
				if c := drawFromDeck(g.deck, &g.deckDrawCnt); c != nil {
					g.players[idx].AddCard(c)
				}
			}
		}
	}
	for range ZwanzigerrufenTalonSize {
		if c := drawFromDeck(g.deck, &g.deckDrawCnt); c != nil {
			g.talon = append(g.talon, c)
		}
	}
}

// --- 入札 ---

// PlayerBid 人間プレイヤーが入札する。
func (g *Zwanzigerrufen) PlayerBid(bid ZwanzigerrufenBid) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ZwanzigerrufenPhaseBid {
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
func (g *Zwanzigerrufen) validateBid(bid ZwanzigerrufenBid) error {
	switch bid {
	case ZwanzigerrufenBidRufer, ZwanzigerrufenBidSolo:
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
func (g *Zwanzigerrufen) PlayerPass() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ZwanzigerrufenPhaseBid {
		return NewDomainError(ErrWrongPhase, "not in bidding phase")
	}
	if !g.players[g.bidPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.applyPass(g.bidPlayerIdx)
	return nil
}

// CpuBid CPU の入札を 1 手進める。
func (g *Zwanzigerrufen) CpuBid() {
	if g.gameEndFlag || g.phase != ZwanzigerrufenPhaseBid {
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
func (g *Zwanzigerrufen) applyBid(idx int, bid ZwanzigerrufenBid) {
	g.highestBid = bid
	g.highestBidder = idx
	g.bidActedCnt++
	g.appendLog(idx, "bid", fmt.Sprintf("%s bids %s", g.playerName(idx), ZwanzigerrufenBidName(bid)), nil)
	g.advanceBid()
}

// applyPass パスを記録して手番を進める。
func (g *Zwanzigerrufen) applyPass(idx int) {
	g.passed[idx] = true
	g.bidActedCnt++
	g.appendLog(idx, "pass", fmt.Sprintf("%s passes", g.playerName(idx)), nil)
	g.advanceBid()
}

// advanceBid 次の入札者を選び、全員が表明したら契約を確定する。
func (g *Zwanzigerrufen) advanceBid() {
	if g.bidActedCnt >= ZwanzigerrufenPlayerCnt {
		g.finalizeBid()
		return
	}
	for range ZwanzigerrufenPlayerCnt {
		g.bidPlayerIdx = (g.bidPlayerIdx + 1) % ZwanzigerrufenPlayerCnt
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
func (g *Zwanzigerrufen) finalizeBid() {
	if g.highestBidder < 0 {
		g.contract = ZwanzigerrufenBidTrischaken
		g.declarerIdx = -1
		g.partnerIdx = -1
		g.calledTrump = -1
		// 場札は最終トリックの勝者が引き取る。
		g.stashOwner = -1
		g.appendLog(-1, "contract", "everyone passed: Trischaken", nil)
		g.startPlay()
		return
	}
	g.declarerIdx = g.highestBidder
	g.contract = g.highestBid
	if g.contract == ZwanzigerrufenBidSolo {
		g.partnerIdx = -1
		g.calledTrump = -1
		// ソロでは場札は防御側の得点になる。
		g.stash = append([]*Card(nil), g.talon...)
		g.talon = nil
		g.stashOwner = 1
		g.appendLog(g.declarerIdx, "contract",
			fmt.Sprintf("%s plays solo", g.playerName(g.declarerIdx)), nil)
		g.startPlay()
		return
	}
	g.resolveCall()
	g.phase = ZwanzigerrufenPhaseTalon
	g.currentPlayerIdx = g.declarerIdx
	// 場札はデクレアラーの手札に加わる。伏せる 6 枚は本人が選ぶ。
	for _, c := range g.talon {
		g.players[g.declarerIdx].AddCard(c)
	}
	g.talon = nil
	g.sortAllHands()
}

// resolveCall 呼ぶ切り札を決め、その持ち主を秘密のパートナーにする。
//
// **自分が持っている札は呼べない。** 20 番を自分で抱えていたら 19 番、それも
// 持っていたら 18 番へと下げる ── 下げ切っても呼べなければ単独で戦う。呼んだ
// 札を自分で持てると、パートナーが自分自身になり「秘密の味方」が消える。
func (g *Zwanzigerrufen) resolveCall() {
	declarer := g.players[g.declarerIdx]
	for t := ZwanzigerrufenCallTrump; t >= ZwanzigerrufenMinCallTrump; t-- {
		if handHasTrump(declarer, t) {
			continue
		}
		g.calledTrump = t
		g.partnerIdx = g.findTrumpHolder(t)
		g.appendLog(g.declarerIdx, "call",
			fmt.Sprintf("%s calls trump %d", g.playerName(g.declarerIdx), t), nil)
		return
	}
	// 呼べる札が手元に全部ある: 単独で戦う。
	g.calledTrump = -1
	g.partnerIdx = -1
	g.appendLog(g.declarerIdx, "call",
		fmt.Sprintf("%s holds every callable trump and plays alone", g.playerName(g.declarerIdx)), nil)
}

// handHasTrump プレイヤーが指定番号の切り札を持っているか。
func handHasTrump(p *ZwanzigerrufenPlayer, trump int) bool {
	for i := range p.GetCardsSize() {
		c := p.GetCard(i)
		if koenigrufenIsTrump(c) && c.GetValue() == trump {
			return true
		}
	}
	return false
}

// findTrumpHolder 指定番号の切り札を持つ席を返す (見つからなければ -1)。
func (g *Zwanzigerrufen) findTrumpHolder(trump int) int {
	for i, p := range g.players {
		if handHasTrump(p, trump) {
			return i
		}
	}
	return -1
}

// --- 場札交換 ---

// PlayerDiscard 人間のデクレアラーが 6 枚を伏せる。
func (g *Zwanzigerrufen) PlayerDiscard(indices []int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ZwanzigerrufenPhaseTalon {
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
func (g *Zwanzigerrufen) validateDiscards(indices []int) error {
	if len(indices) != ZwanzigerrufenTalonSize {
		return NewDomainError(ErrInvalidPlay,
			fmt.Sprintf("must discard exactly %d cards, got %d", ZwanzigerrufenTalonSize, len(indices)))
	}
	p := g.players[g.declarerIdx]
	seen := map[int]bool{}
	for _, idx := range indices {
		if idx < 0 || idx >= p.GetCardsSize() {
			return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", idx))
		}
		if seen[idx] {
			return NewDomainError(ErrInvalidPlay, fmt.Sprintf("hand index %d listed twice", idx))
		}
		seen[idx] = true
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
func (g *Zwanzigerrufen) applyDiscard(indices []int) {
	p := g.players[g.declarerIdx]
	g.stash = p.RemoveCards(indices)
	g.stashOwner = 0
	g.sortAllHands()
	g.appendLog(g.declarerIdx, "discard",
		fmt.Sprintf("%s buries %d cards", g.playerName(g.declarerIdx), len(g.stash)), nil)
	g.startPlay()
}

// CpuDiscard CPU のデクレアラーに伏せ札を選ばせる。
func (g *Zwanzigerrufen) CpuDiscard() {
	if g.gameEndFlag || g.phase != ZwanzigerrufenPhaseTalon {
		return
	}
	if g.declarerIdx < 0 || g.players[g.declarerIdx].GetIsHuman() {
		return
	}
	g.applyDiscard(g.cpuSelectDiscards(g.declarerIdx))
}

// cpuSelectDiscards 伏せられる札のうち、点数の低い 6 枚を選ぶ。
func (g *Zwanzigerrufen) cpuSelectDiscards(idx int) []int {
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
	out := make([]int, 0, ZwanzigerrufenTalonSize)
	for _, e := range cand {
		if len(out) == ZwanzigerrufenTalonSize {
			break
		}
		out = append(out, e.idx)
	}
	// 足りなければ、伏せられる残りの札から補う (切り札しか残っていない場合)。
	for i := range p.GetCardsSize() {
		if len(out) == ZwanzigerrufenTalonSize {
			break
		}
		c := p.GetCard(i)
		if koenigrufenIsKing(c) || koenigrufenIsTrull(c) || zwanzigerrufenContains(out, i) {
			continue
		}
		out = append(out, i)
	}
	return out
}

// zwanzigerrufenContains スライスに値が含まれるか。
func zwanzigerrufenContains(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// --- プレイ ---

// startPlay トリックプレイを開始する。
func (g *Zwanzigerrufen) startPlay() {
	g.phase = ZwanzigerrufenPhasePlay
	g.trickNumber = 1
	g.leadPlayerIdx = (g.dealerIdx + 1) % ZwanzigerrufenPlayerCnt
	g.currentPlayerIdx = g.leadPlayerIdx
	g.currentTrick = nil
}

// PlayerPlayCard 人間が手札を 1 枚出す。
func (g *Zwanzigerrufen) PlayerPlayCard(handIdx int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != ZwanzigerrufenPhasePlay {
		return NewDomainError(ErrWrongPhase, "not in play phase")
	}
	p := g.players[g.currentPlayerIdx]
	if !p.GetIsHuman() {
		return ErrNotHumanTurn
	}
	if handIdx < 0 || handIdx >= p.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, fmt.Sprintf("hand index %d out of range", handIdx))
	}
	if !zwanzigerrufenContains(g.GetValidPlayIndices(g.currentPlayerIdx), handIdx) {
		return NewDomainError(ErrInvalidPlay, "that card does not follow the lead")
	}
	g.playCard(g.currentPlayerIdx, handIdx)
	return nil
}

// CpuPlayCard CPU の 1 手を進める。
func (g *Zwanzigerrufen) CpuPlayCard() {
	if g.gameEndFlag || g.phase != ZwanzigerrufenPhasePlay {
		return
	}
	if g.players[g.currentPlayerIdx].GetIsHuman() {
		return
	}
	g.playCard(g.currentPlayerIdx, g.cpuSelectPlayCard(g.currentPlayerIdx))
}

// playCard 1 枚を場に出し、トリックが揃ったら決着させる。
func (g *Zwanzigerrufen) playCard(playerIdx, handIdx int) {
	card := g.players[playerIdx].RemoveCard(handIdx)
	if card == nil {
		return
	}
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play",
		fmt.Sprintf("%s plays %s", g.playerName(playerIdx), koenigrufenCardStr(card)), []*Card{card})

	// 呼ばれた切り札が出た瞬間にパートナーが明らかになる。
	if g.calledTrump > 0 && koenigrufenIsTrump(card) && card.GetValue() == g.calledTrump {
		g.partnerRevealed = true
		g.appendLog(playerIdx, "reveal",
			fmt.Sprintf("%s holds the called trump", g.playerName(playerIdx)), nil)
	}

	if len(g.currentTrick) < ZwanzigerrufenPlayerCnt {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % ZwanzigerrufenPlayerCnt
		return
	}
	g.finishTrick()
}

// finishTrick トリックの勝者を決め、次のトリックへ進む。
func (g *Zwanzigerrufen) finishTrick() {
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
	g.phase = ZwanzigerrufenPhaseTrickEnd
	g.currentPlayerIdx = winner
	g.leadPlayerIdx = winner
}

// NextTrick トリック終了フェーズから次のトリックへ進む。
func (g *Zwanzigerrufen) NextTrick() {
	if g.gameEndFlag || g.phase != ZwanzigerrufenPhaseTrickEnd {
		return
	}
	if g.trickNumber >= ZwanzigerrufenTrickCount {
		g.finishRound()
		return
	}
	g.trickNumber++
	g.phase = ZwanzigerrufenPhasePlay
	g.currentPlayerIdx = g.leadPlayerIdx
}

// trickWinner 現在のトリックの勝者を返す。
func (g *Zwanzigerrufen) trickWinner() int {
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
func (g *Zwanzigerrufen) ledSuit() int {
	if len(g.currentTrick) == 0 {
		return -1
	}
	first := g.currentTrick[0].Card
	if koenigrufenIsTrumpLike(first) {
		return KoenigrufenTrumpDesign
	}
	return first.GetDesign()
}

// GetValidPlayIndices 出せる手札のインデックスを返す。
//
// リードスートに従う義務、持っていなければ切り札を出す義務、どちらも無ければ自由。
func (g *Zwanzigerrufen) GetValidPlayIndices(playerIdx int) []int {
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
func (g *Zwanzigerrufen) cpuSelectPlayCard(playerIdx int) int {
	valid := g.GetValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	p := g.players[playerIdx]
	if g.config.CpuDifficulty == ZwanzigerrufenCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	pick, bestScore := valid[0], 0
	for i, idx := range valid {
		pts := koenigrufenCardPoints(p.GetCard(idx))
		score := -pts // 既定では安い札から落とす
		if g.contract != ZwanzigerrufenBidTrischaken && len(g.currentTrick) == ZwanzigerrufenPlayerCnt-1 {
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
func (g *Zwanzigerrufen) cpuSelectBid(playerIdx int) (ZwanzigerrufenBid, bool) {
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
	case trumps >= 9 && honours >= 3 && ZwanzigerrufenBidSolo > g.highestBid:
		return ZwanzigerrufenBidSolo, true
	case trumps >= 5 && ZwanzigerrufenBidRufer > g.highestBid:
		return ZwanzigerrufenBidRufer, true
	default:
		return ZwanzigerrufenBidPass, false
	}
}

// --- 精算 ---

// finishRound ディールを精算する。
func (g *Zwanzigerrufen) finishRound() {
	if g.scored {
		return
	}
	g.scored = true
	g.assignStash()
	if g.contract == ZwanzigerrufenBidTrischaken {
		g.breakdown = g.scoreTrischaken()
		g.outcome = ZwanzigerrufenOutcomeTrischaken
	} else {
		g.breakdown = g.scoreContract()
		if g.breakdown.Won {
			g.outcome = ZwanzigerrufenOutcomeWin
		} else {
			g.outcome = ZwanzigerrufenOutcomeLoss
		}
	}
	for i, delta := range g.breakdown.Seats {
		g.playerScores[i] += delta
	}
	g.phase = ZwanzigerrufenPhaseRoundEnd
	g.appendLog(-1, "score",
		fmt.Sprintf("deal %d scored (%s)", g.roundNumber, ZwanzigerrufenBidName(g.contract)), nil)
	if g.roundNumber >= g.config.TargetDeals {
		g.finishGame()
	}
}

// assignStash 脇に置いた札を、契約に応じた引き取り手のトリックに加える。
//
// **Trischaken の場札は最終トリックの勝者が引き取る。** どこにも足さないと、
// 場札 6 枚ぶんの点が消えて総点が合わなくなる。
func (g *Zwanzigerrufen) assignStash() {
	if len(g.stash) == 0 {
		return
	}
	target := -1
	switch g.stashOwner {
	case 0:
		target = g.declarerIdx
	case 1:
		// 防御側。デクレアラー・パートナー以外の最初の席へ渡す。
		for i := range g.players {
			if !g.isDeclarerSide(i) {
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
func (g *Zwanzigerrufen) cardPointsOf(idx int) int {
	total := 0
	for _, trick := range g.players[idx].GetTricksTaken() {
		for _, c := range trick {
			total += koenigrufenCardPoints(c)
		}
	}
	return total
}

// isDeclarerSide 指定席がデクレアラー側か。
func (g *Zwanzigerrufen) isDeclarerSide(idx int) bool {
	if g.declarerIdx < 0 {
		return false
	}
	return idx == g.declarerIdx || (g.partnerIdx >= 0 && idx == g.partnerIdx)
}

// zwanzigerrufenTotalPoints デッキ全体のカードポイントを返す。
//
// **表を書き写さず数える。** 閾値を定数で持つと、点数表を変えたときに片方だけが
// ずれる。
func zwanzigerrufenTotalPoints() int {
	total := 0
	for _, c := range buildKoenigrufenDeck() {
		total += koenigrufenCardPoints(c)
	}
	return total
}

// scoreContract Rufer / Solo の精算を行う (ゼロサム)。
func (g *Zwanzigerrufen) scoreContract() *ZwanzigerrufenBreakdown {
	team := 0
	for i := range g.players {
		if g.isDeclarerSide(i) {
			team += g.cardPointsOf(i)
		}
	}
	total := zwanzigerrufenTotalPoints()
	threshold := total / 2
	won := 2*team > total
	diff := team - threshold
	if diff < 0 {
		diff = -diff
	}
	solo := g.partnerIdx < 0
	base := ZwanzigerrufenBaseGameValue + diff
	if g.contract == ZwanzigerrufenBidSolo {
		base *= 2
	}
	sign := 1
	if !won {
		sign = -1
	}
	bd := &ZwanzigerrufenBreakdown{
		Contract: g.contract, TeamPoints: team, Threshold: threshold,
		Won: won, Solo: solo, Base: base,
		Seats: make([]int, ZwanzigerrufenPlayerCnt), Loser: -1,
	}
	// **単独なら 1 人で 3 人ぶんを背負う。** 組と同額だと、ソロを宣言する
	// 理由も抑止も無くなる。
	declarerShare := base
	if solo {
		declarerShare = 3 * base
	}
	for i := range bd.Seats {
		switch {
		case i == g.declarerIdx:
			bd.Seats[i] = sign * declarerShare
		case g.isDeclarerSide(i):
			bd.Seats[i] = sign * base
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
func (g *Zwanzigerrufen) scoreTrischaken() *ZwanzigerrufenBreakdown {
	loser, worst := 0, -1
	for i := range g.players {
		if pts := g.cardPointsOf(i); pts > worst {
			loser, worst = i, pts
		}
	}
	bd := &ZwanzigerrufenBreakdown{
		Contract: ZwanzigerrufenBidTrischaken, TeamPoints: worst,
		Threshold: 0, Won: false, Solo: false,
		Base:  ZwanzigerrufenTrischakenLoss,
		Seats: make([]int, ZwanzigerrufenPlayerCnt), Loser: loser,
	}
	for i := range bd.Seats {
		if i == loser {
			bd.Seats[i] = -ZwanzigerrufenTrischakenLoss
			continue
		}
		bd.Seats[i] = 1
	}
	return bd
}

// finishGame 通算得点で勝者を決める。
func (g *Zwanzigerrufen) finishGame() {
	best, tied := 0, false
	for i := 1; i < ZwanzigerrufenPlayerCnt; i++ {
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
	g.phase = ZwanzigerrufenPhaseGameEnd
	g.appendLog(g.winnerPlayer, "gameEnd", "match over", nil)
}

// --- 補助 ---

// sortAllHands 全員の手札を並べ替える。
func (g *Zwanzigerrufen) sortAllHands() {
	for _, p := range g.players {
		zwanzigerrufenSortHand(p)
	}
}

// zwanzigerrufenSortHand 手札をスート順・切り札は番号順に並べる。
func zwanzigerrufenSortHand(p *ZwanzigerrufenPlayer) {
	cards := make([]*Card, 0, p.GetCardsSize())
	for i := range p.GetCardsSize() {
		cards = append(cards, p.GetCard(i))
	}
	for i := 1; i < len(cards); i++ {
		for j := i; j > 0 && zwanzigerrufenSortKey(cards[j]) < zwanzigerrufenSortKey(cards[j-1]); j-- {
			cards[j], cards[j-1] = cards[j-1], cards[j]
		}
	}
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// zwanzigerrufenSortKey 手札の並び順キー。
func zwanzigerrufenSortKey(c *Card) int {
	if c == nil {
		return 1 << 20
	}
	if koenigrufenIsTrumpLike(c) {
		return 1000 + koenigrufenTrumpValue(c)
	}
	return c.GetDesign()*100 + c.GetValue()
}

// playerName 席の表示名を返す。
func (g *Zwanzigerrufen) playerName(i int) string {
	if i < 0 || i >= len(g.players) {
		return "-"
	}
	if g.players[i].GetIsHuman() {
		return "YOU"
	}
	return fmt.Sprintf("CPU%d", i)
}

// ZwanzigerrufenBidName 入札の識別子を返す (i18n キーの一部に使う)。
func ZwanzigerrufenBidName(bid ZwanzigerrufenBid) string {
	switch bid {
	case ZwanzigerrufenBidTrischaken:
		return "trischaken"
	case ZwanzigerrufenBidRufer:
		return "rufer"
	case ZwanzigerrufenBidSolo:
		return "solo"
	default:
		return "pass"
	}
}

// --- アクセサ ---

// GetPlayers 席の一覧を返す。
func (g *Zwanzigerrufen) GetPlayers() []*ZwanzigerrufenPlayer { return g.players }

// GetPlayer 指定した席を返す (範囲外なら nil)。
func (g *Zwanzigerrufen) GetPlayer(i int) *ZwanzigerrufenPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// GetPlayerCnt 席数を返す。
func (g *Zwanzigerrufen) GetPlayerCnt() int { return len(g.players) }

// GetConfig 設定を返す。
func (g *Zwanzigerrufen) GetConfig() ZwanzigerrufenConfig { return g.config }

// SetConfig 設定を差し替える (次の Reset で反映)。
func (g *Zwanzigerrufen) SetConfig(c ZwanzigerrufenConfig) { g.config = c }

// GetPhase 現在のフェーズを返す。
func (g *Zwanzigerrufen) GetPhase() ZwanzigerrufenPhase { return g.phase }

// GetRoundNumber 現在のディール番号を返す。
func (g *Zwanzigerrufen) GetRoundNumber() int { return g.roundNumber }

// GetTrickNumber 現在のトリック番号を返す。
func (g *Zwanzigerrufen) GetTrickNumber() int { return g.trickNumber }

// GetCurrentPlayerIdx 手番の席を返す。
func (g *Zwanzigerrufen) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// GetCurrentTrick 場に出ている札を返す。
func (g *Zwanzigerrufen) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// GetDealerIdx 親の席を返す。
func (g *Zwanzigerrufen) GetDealerIdx() int { return g.dealerIdx }

// GetBidPlayerIdx 入札の手番を返す。
func (g *Zwanzigerrufen) GetBidPlayerIdx() int { return g.bidPlayerIdx }

// GetHighestBid 現在の最高入札を返す。
func (g *Zwanzigerrufen) GetHighestBid() ZwanzigerrufenBid { return g.highestBid }

// GetDeclarerIdx デクレアラーの席を返す (-1 = 未確定 / Trischaken)。
func (g *Zwanzigerrufen) GetDeclarerIdx() int { return g.declarerIdx }

// GetContract 成立した契約を返す。
func (g *Zwanzigerrufen) GetContract() ZwanzigerrufenBid { return g.contract }

// GetCalledTrump 呼んだ切り札の番号を返す (-1 = 呼んでいない)。
func (g *Zwanzigerrufen) GetCalledTrump() int { return g.calledTrump }

// GetPartnerRevealed パートナーが判明したかを返す。
func (g *Zwanzigerrufen) GetPartnerRevealed() bool { return g.partnerRevealed }

// GetPartnerIdx パートナーの席を返す。**判明するまでは -1 を返す。**
//
// 呼び札が場に出るまで正体を隠すのがこのゲームの骨格なので、判明前に答える
// アクセサを用意しない。
func (g *Zwanzigerrufen) GetPartnerIdx() int {
	if !g.partnerRevealed {
		return -1
	}
	return g.partnerIdx
}

// GetTalonSize 手札に加える前の場札の枚数を返す。
func (g *Zwanzigerrufen) GetTalonSize() int { return len(g.talon) }

// GetLastTrickWinner 直前のトリックの勝者を返す。
func (g *Zwanzigerrufen) GetLastTrickWinner() int { return g.lastTrickWinner }

// GetLastTrickCards 直前のトリックの札を返す。
func (g *Zwanzigerrufen) GetLastTrickCards() []*Card { return g.lastTrickCards }

// GetOutcome ディール結果を返す。
func (g *Zwanzigerrufen) GetOutcome() ZwanzigerrufenOutcome { return g.outcome }

// GetBreakdown 直近ディールの精算内訳を返す。
func (g *Zwanzigerrufen) GetBreakdown() *ZwanzigerrufenBreakdown { return g.breakdown }

// GetPlayerScore 指定席の通算得点を返す。
func (g *Zwanzigerrufen) GetPlayerScore(i int) int {
	if i < 0 || i >= ZwanzigerrufenPlayerCnt {
		return 0
	}
	return g.playerScores[i]
}

// GetCardPoints 指定席が獲得したカードポイントを返す。
func (g *Zwanzigerrufen) GetCardPoints(i int) int {
	if i < 0 || i >= len(g.players) {
		return 0
	}
	return g.cardPointsOf(i)
}

// GetGameEndFlag 終局したかを返す。
func (g *Zwanzigerrufen) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 終局時の勝者を返す (-1 = 引き分け/未決)。
func (g *Zwanzigerrufen) GetWinnerPlayer() int { return g.winnerPlayer }

// HumanSeat 人間の席を返す (居なければ 0)。
func (g *Zwanzigerrufen) HumanSeat() int {
	if i := findHumanIdx(g.players); i >= 0 {
		return i
	}
	return 0
}

// IsHumanTurn 人間の入力を待っているかを返す。
func (g *Zwanzigerrufen) IsHumanTurn() bool {
	if g.gameEndFlag {
		return false
	}
	switch g.phase {
	case ZwanzigerrufenPhaseBid:
		return g.players[g.bidPlayerIdx].GetIsHuman()
	case ZwanzigerrufenPhaseTalon:
		return g.declarerIdx >= 0 && g.players[g.declarerIdx].GetIsHuman()
	case ZwanzigerrufenPhasePlay:
		return g.players[g.currentPlayerIdx].GetIsHuman()
	default:
		return false
	}
}

// GetHint 人間の手番における推奨手を返す。
func (g *Zwanzigerrufen) GetHint() *ZwanzigerrufenHint {
	if g.gameEndFlag || !g.IsHumanTurn() {
		return nil
	}
	seat := g.HumanSeat()
	switch g.phase {
	case ZwanzigerrufenPhaseBid:
		bid, ok := g.cpuSelectBid(seat)
		if !ok {
			return &ZwanzigerrufenHint{Reason: "pass_weak_hand"}
		}
		v := int(bid)
		return &ZwanzigerrufenHint{Bid: &v, Reason: "bid_strong_trumps"}
	case ZwanzigerrufenPhaseTalon:
		return &ZwanzigerrufenHint{DiscardIndices: g.cpuSelectDiscards(seat), Reason: "bury_cheap_cards"}
	case ZwanzigerrufenPhasePlay:
		idx := g.cpuSelectPlayCard(seat)
		reason := "play_low"
		if g.contract == ZwanzigerrufenBidTrischaken {
			reason = "avoid_points"
		}
		return &ZwanzigerrufenHint{CardIndex: &idx, Reason: reason}
	default:
		return nil
	}
}

// --- JSON ---

// zwanzigerrufenJSON は Zwanzigerrufen の JSON 表現。
type zwanzigerrufenJSON struct {
	Players         []*ZwanzigerrufenPlayer  `json:"pl"`
	Config          ZwanzigerrufenConfig     `json:"cf"`
	Phase           ZwanzigerrufenPhase      `json:"ph"`
	RoundNumber     int                      `json:"rn"`
	TrickNumber     int                      `json:"tn"`
	CurrentPlayer   int                      `json:"cp"`
	CurrentTrick    []*TrickCard             `json:"ct"`
	LeadPlayer      int                      `json:"lp"`
	DealerIdx       int                      `json:"di"`
	BidPlayerIdx    int                      `json:"bp"`
	BidActedCnt     int                      `json:"ba"`
	HighestBid      ZwanzigerrufenBid        `json:"hb"`
	HighestBidder   int                      `json:"hr"`
	Passed          []bool                   `json:"ps"`
	DeclarerIdx     int                      `json:"dc"`
	Contract        ZwanzigerrufenBid        `json:"co"`
	CalledTrump     int                      `json:"cl"`
	PartnerIdx      int                      `json:"pi"`
	PartnerRevealed bool                     `json:"pr"`
	Talon           []*Card                  `json:"tl"`
	Stash           []*Card                  `json:"st"`
	StashOwner      int                      `json:"so"`
	PlayerScores    []int                    `json:"sc"`
	LastTrickWinner int                      `json:"lw"`
	LastTrickCards  []*Card                  `json:"lc"`
	Outcome         ZwanzigerrufenOutcome    `json:"oc"`
	Breakdown       *ZwanzigerrufenBreakdown `json:"bd"`
	Scored          bool                     `json:"sd"`
	GameEndFlag     bool                     `json:"ge"`
	WinnerPlayer    int                      `json:"wp"`
	ActionLog       []*ActionLogEntry        `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Zwanzigerrufen) MarshalJSON() ([]byte, error) {
	return json.Marshal(zwanzigerrufenJSON{
		Players: g.players, Config: g.config, Phase: g.phase,
		RoundNumber: g.roundNumber, TrickNumber: g.trickNumber,
		CurrentPlayer: g.currentPlayerIdx, CurrentTrick: g.currentTrick,
		LeadPlayer: g.leadPlayerIdx, DealerIdx: g.dealerIdx,
		BidPlayerIdx: g.bidPlayerIdx, BidActedCnt: g.bidActedCnt,
		HighestBid: g.highestBid, HighestBidder: g.highestBidder,
		Passed: g.passed[:], DeclarerIdx: g.declarerIdx, Contract: g.contract,
		CalledTrump: g.calledTrump, PartnerIdx: g.partnerIdx, PartnerRevealed: g.partnerRevealed,
		Talon: g.talon, Stash: g.stash, StashOwner: g.stashOwner,
		PlayerScores: g.playerScores[:], LastTrickWinner: g.lastTrickWinner,
		LastTrickCards: g.lastTrickCards, Outcome: g.outcome, Breakdown: g.breakdown,
		Scored: g.scored, GameEndFlag: g.gameEndFlag, WinnerPlayer: g.winnerPlayer,
		ActionLog: g.actionLog,
	})
}

// zwanzigerrufenMaxSliceLen 復元時のスライス長上限。
const zwanzigerrufenMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
//
// **呼んだ切り札は 18..20 か「呼んでいない」しか無い。** 範囲を見ないと、保存を
// 書き換えるだけで「1 番を呼んだ」ことにでき、パートナーの席を好きに選べる。
// 契約との整合も見る ── Trischaken にデクレアラーが居る保存は、範囲検査だけなら
// 通ってしまう。
func (g *Zwanzigerrufen) UnmarshalJSON(data []byte) error {
	var j zwanzigerrufenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > zwanzigerrufenMaxSliceLen || len(j.Talon) > zwanzigerrufenMaxSliceLen ||
		len(j.Stash) > zwanzigerrufenMaxSliceLen || len(j.ActionLog) > zwanzigerrufenMaxSliceLen ||
		len(j.CurrentTrick) > zwanzigerrufenMaxSliceLen {
		return fmt.Errorf("zwanzigerrufen: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("zwanzigerrufen: invalid config: %w", err)
	}
	if len(j.Players) != ZwanzigerrufenPlayerCnt {
		return fmt.Errorf("zwanzigerrufen: expected %d players, got %d", ZwanzigerrufenPlayerCnt, len(j.Players))
	}
	for _, p := range j.Players {
		if p == nil {
			return fmt.Errorf("zwanzigerrufen: nil player in state")
		}
	}
	if j.Phase < ZwanzigerrufenPhaseBid || j.Phase > ZwanzigerrufenPhaseGameEnd {
		return fmt.Errorf("zwanzigerrufen: invalid phase %d", j.Phase)
	}
	if j.RoundNumber < 1 || j.RoundNumber > j.Config.TargetDeals {
		return fmt.Errorf("zwanzigerrufen: round %d out of range", j.RoundNumber)
	}
	if j.TrickNumber < 0 || j.TrickNumber > ZwanzigerrufenTrickCount {
		return fmt.Errorf("zwanzigerrufen: trick %d out of range", j.TrickNumber)
	}
	for name, idx := range map[string]int{"current player": j.CurrentPlayer, "bid player": j.BidPlayerIdx, "dealer": j.DealerIdx} {
		if idx < 0 || idx >= ZwanzigerrufenPlayerCnt {
			return fmt.Errorf("zwanzigerrufen: %s out of range", name)
		}
	}
	for name, idx := range map[string]int{"declarer": j.DeclarerIdx, "partner": j.PartnerIdx, "winner": j.WinnerPlayer, "highest bidder": j.HighestBidder} {
		if idx < -1 || idx >= ZwanzigerrufenPlayerCnt {
			return fmt.Errorf("zwanzigerrufen: %s out of range", name)
		}
	}
	if err := zwanzigerrufenValidateContract(&j); err != nil {
		return err
	}
	if err := zwanzigerrufenValidateCards(j.Talon); err != nil {
		return err
	}
	if err := zwanzigerrufenValidateCards(j.Stash); err != nil {
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
	g.passed = [ZwanzigerrufenPlayerCnt]bool{}
	copy(g.passed[:], j.Passed)
	g.declarerIdx = j.DeclarerIdx
	g.contract = j.Contract
	g.calledTrump = j.CalledTrump
	g.partnerIdx = j.PartnerIdx
	g.partnerRevealed = j.PartnerRevealed
	g.talon = j.Talon
	g.stash = j.Stash
	g.stashOwner = j.StashOwner
	g.playerScores = [ZwanzigerrufenPlayerCnt]int{}
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

// zwanzigerrufenValidateContract 契約・呼び札・デクレアラーの整合を検証する。
func zwanzigerrufenValidateContract(j *zwanzigerrufenJSON) error {
	switch j.Contract {
	case ZwanzigerrufenBidPass, ZwanzigerrufenBidTrischaken,
		ZwanzigerrufenBidRufer, ZwanzigerrufenBidSolo:
	default:
		return fmt.Errorf("zwanzigerrufen: invalid contract %d", j.Contract)
	}
	if j.CalledTrump != -1 &&
		(j.CalledTrump < ZwanzigerrufenMinCallTrump || j.CalledTrump > ZwanzigerrufenCallTrump) {
		return fmt.Errorf("zwanzigerrufen: called trump %d out of range", j.CalledTrump)
	}
	// **呼び札は 20 番呼び契約でしか存在しない。**
	if j.CalledTrump != -1 && j.Contract != ZwanzigerrufenBidRufer {
		return fmt.Errorf("zwanzigerrufen: a called trump requires the rufer contract")
	}
	// **Trischaken にデクレアラーは居ない。**
	if j.Contract == ZwanzigerrufenBidTrischaken && j.DeclarerIdx != -1 {
		return fmt.Errorf("zwanzigerrufen: trischaken has no declarer")
	}
	// **パートナーは呼び札があるときにしか居ない。**
	if j.PartnerIdx != -1 && j.CalledTrump == -1 {
		return fmt.Errorf("zwanzigerrufen: a partner requires a called trump")
	}
	if j.PartnerIdx != -1 && j.PartnerIdx == j.DeclarerIdx {
		return fmt.Errorf("zwanzigerrufen: the declarer cannot be their own partner")
	}
	return nil
}

// zwanzigerrufenValidateCards 札に nil や範囲外の design/value が無いか検証する。
func zwanzigerrufenValidateCards(cards []*Card) error {
	for _, c := range cards {
		if c == nil {
			return fmt.Errorf("zwanzigerrufen: nil card in state")
		}
		d, v := c.GetDesign(), c.GetValue()
		switch {
		case d >= 1 && d <= KoenigrufenSuitCnt && v >= 1 && v <= KoenigrufenSuitMaxValue:
		case d == KoenigrufenTrumpDesign && v >= 1 && v <= KoenigrufenMaxTrump:
		case d == KoenigrufenSkusDesign && v == KoenigrufenSkusValue:
		default:
			return fmt.Errorf("zwanzigerrufen: card out of range (design %d, value %d)", d, v)
		}
	}
	return nil
}
