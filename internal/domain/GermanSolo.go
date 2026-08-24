//go:build !js || !wasm || classic

// Package domain ジャーマン・ソロ (German Solo) のドメインモデル。
//
// German Solo は 18 世紀ドイツで遊ばれた 4 人用トリックテイキングゲームで、スペインの
// オンブル → フランスのカドリールと伝わった「ソリスト対多数」系のドイツ版。同じ血統の
// `quadrille` が 40 枚のスペイン式デッキを使うのに対し、こちらは **32 枚のスカート・パック
// (A,K,Q,J,10,9,8,7)** を 4 人に 8 枚ずつ配り切る。
//
// 三大マタドール (契約に関わらず常時最強の切り札 3 枚):
//
//  1. Spadille = ♣Q (クラブが切り札でなくても常に第 1 切り札)
//  2. Manille  = 切り札スートの 7 (常に第 2)
//  3. Basta    = ♠Q (スペードが切り札でなくても常に第 3)
//
// **黒の 2 枚のクイーンは常に切り札**なので、♠♣ が平札のときそのスートに Q は現れない。
// 切り札スートの残りは A > K > (Q) > J > 10 > 9 > 8。切り札の Q が普通の切り札として
// 残るのは切り札が赤 (♥♦) のときだけなので、切り札は赤 10 枚 / 黒 9 枚になる。
// 平札は A > K > Q > J > 10 > 9 > 8 > 7。
//
// ビッドの階段 (低い順、数値が大きいほど高い宣言):
//
//   - Frage (フラーゲ): 切り札を決め、**自分が持っていないエース**を 1 枚指名する。
//     そのエースの持ち主が伏せられた味方となり、2 対 2 で 8 トリック中 5 トリックを狙う。
//   - Solo (ソロ): 単独で残り 3 人を相手に 5 トリック。
//   - Tout (トゥー): 単独で **8 トリック全部**。
//
// 全員パスなら Spadille (♣Q) を持つ席が **Mussfrage** (強制の Frage) を引き受ける。
// 配り直しにすると CPU がビッドしない手を引き続ける限りディールが終わらないので、
// 卓の誰かが必ず引き受ける方式にしてある。
//
// 得点: 契約値 V (Mussfrage 1 / Frage 2 / Solo 4 / Tout 8) が宣言側と守備側の間で動く。
// 成功なら守備側の各席が V を失い、その合計を宣言側で山分けする (単独なら +3V、味方ありなら
// 1 人あたり +V)。失敗なら符号が反転する。TargetRounds ディール後、累積点最上位が勝者。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"sort"
)

// GermanSoloPlayerCnt プレイヤー数 (人間 1 + CPU 3)
const GermanSoloPlayerCnt = 4

// GermanSoloHandSize 各プレイヤーの配り札枚数。
//
// **32 枚を 4 人で配り切る。** 4 × 8 = 32 でストックは残らないので、
// 8 枚が 8 トリックにそのまま対応する。
const GermanSoloHandSize = 8

// GermanSoloTrickCount 1 ディールのトリック数 (手札 8 枚を出し切る)
const GermanSoloTrickCount = 8

// GermanSoloDeckSize デッキ枚数 (52 - 2,3,4,5,6)
const GermanSoloDeckSize = 32

// GermanSoloMakeTricks Frage / Mussfrage / Solo の成功に必要なトリック数 (8 の過半数)
const GermanSoloMakeTricks = 5

// GermanSoloWinRounds マッチを構成するディール数 (既定)
const GermanSoloWinRounds = 5

// GermanSoloBid ビッド宣言
type GermanSoloBid int

// GermanSolo のビッド定数 (数値が大きいほど高い宣言)
const (
	// GermanSoloBidNone 未宣言 (パス相当の初期値)
	GermanSoloBidNone GermanSoloBid = 0
	// GermanSoloBidMussfrage Mussfrage — 全員パスのときに Spadille (♣Q) の持ち主が
	// **強制的に**引き受ける Frage。競りでは宣言できない (isBidLegal が弾く)。
	GermanSoloBidMussfrage GermanSoloBid = 1
	// GermanSoloBidFrage Frage — 切り札を決め、持っていないエースを呼んで味方を作る。
	GermanSoloBidFrage GermanSoloBid = 2
	// GermanSoloBidSolo Solo — 単独で 5 トリック。
	GermanSoloBidSolo GermanSoloBid = 3
	// GermanSoloBidTout Tout — 単独で 8 トリック全部。
	GermanSoloBidTout GermanSoloBid = 4
)

// germanSoloBidValue は契約値 V を返す。得点はこの V が宣言側と守備側の間で動く。
func germanSoloBidValue(bid GermanSoloBid) int {
	switch bid {
	case GermanSoloBidMussfrage:
		return 1
	case GermanSoloBidFrage:
		return 2
	case GermanSoloBidSolo:
		return 4
	case GermanSoloBidTout:
		return 8
	default:
		return 0
	}
}

// germanSoloRequiredTricks は bid の成功に必要なトリック数を返す。
//
// **Tout だけ全トリック。** ここを 5 のままにすると、8 点払う最上位契約が
// Solo と同じ条件で成功してしまい、階段が意味を失う。
func germanSoloRequiredTricks(bid GermanSoloBid) int {
	if bid == GermanSoloBidTout {
		return GermanSoloTrickCount
	}
	return GermanSoloMakeTricks
}

// germanSoloIsPartnerBid は bid が味方を呼ぶ契約 (Frage / Mussfrage) かを返す。
func germanSoloIsPartnerBid(bid GermanSoloBid) bool {
	return bid == GermanSoloBidFrage || bid == GermanSoloBidMussfrage
}

// GermanSoloPhase ゲームフェーズ
type GermanSoloPhase int

// GermanSolo のフェーズ定数
const (
	// GermanSoloPhaseBid ビッド (auction) フェーズ
	GermanSoloPhaseBid GermanSoloPhase = 0
	// GermanSoloPhaseAceCall エース呼びフェーズ。Frage / Mussfrage のときだけ通る。
	// 落札者が手札に無いエースを 1 枚指名し、そのエースを持つ席が伏せられた味方になる。
	// Solo / Tout は単独で戦う契約なのでこのフェーズを飛ばす。
	GermanSoloPhaseAceCall GermanSoloPhase = 1
	// GermanSoloPhasePlay トリックプレイフェーズ
	GermanSoloPhasePlay GermanSoloPhase = 2
	// GermanSoloPhaseTrickEnd トリック終了フェーズ
	GermanSoloPhaseTrickEnd GermanSoloPhase = 3
	// GermanSoloPhaseRoundEnd ディール終了フェーズ
	GermanSoloPhaseRoundEnd GermanSoloPhase = 4
	// GermanSoloPhaseGameEnd ゲーム終了フェーズ
	GermanSoloPhaseGameEnd GermanSoloPhase = 5
)

// GermanSoloPhaseMin フェーズ下限 (検証用)
const GermanSoloPhaseMin = int(GermanSoloPhaseBid)

// GermanSoloPhaseMax フェーズ上限 (検証用)
const GermanSoloPhaseMax = int(GermanSoloPhaseGameEnd)

// GermanSoloOutcome ディール結果
type GermanSoloOutcome int

// GermanSolo のディール結果定数
const (
	// GermanSoloOutcomeNone 未確定
	GermanSoloOutcomeNone GermanSoloOutcome = 0
	// GermanSoloOutcomeMade 宣言側が必要トリック数に届き契約成功
	GermanSoloOutcomeMade GermanSoloOutcome = 1
	// GermanSoloOutcomeFailed 宣言側が必要トリック数に届かず契約失敗
	GermanSoloOutcomeFailed GermanSoloOutcome = 2
)

// GermanSoloResult 人間視点のマッチ結果
type GermanSoloResult int

// GermanSolo のマッチ結果定数
const (
	// GermanSoloResultLose 敗北
	GermanSoloResultLose GermanSoloResult = -1
	// GermanSoloResultNone 未確定 / 引き分け
	GermanSoloResultNone GermanSoloResult = 0
	// GermanSoloResultWin 勝利
	GermanSoloResultWin GermanSoloResult = 1
)

// GermanSoloHint ヒント情報
type GermanSoloHint struct {
	CardIndices []int  // 推奨カードインデックス (play フェーズ)
	Reason      string // ヒント理由キー
	SuitHint    int    // 推奨スート (ace call フェーズ, 0=なし)
}

// GermanSolo ジャーマン・ソロのゲームクラス
type GermanSolo struct {
	trumpCards       *TrumpCards
	players          []*GermanSoloPlayer
	config           GermanSoloConfig
	phase            GermanSoloPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	forehandIdx      int                                // ディーラーの左隣 (ビッド開始)
	declarerIdx      int                                // ジャーマン・ソロ (-1=未確定)
	winningBid       GermanSoloBid                      // 確定したビッド (ジャーマン・ソロの宣言)
	trumpSuit        int                                // 切り札スート (-1=未確定, 1..4)
	currentBidderIdx int                                // 現在ビッド中のプレイヤー (bid フェーズ)
	bids             [GermanSoloPlayerCnt]GermanSoloBid // 各プレイヤーの宣言
	bidTrump         [GermanSoloPlayerCnt]int           // 各プレイヤーが宣言時に選んだ切り札 (-1=なし)
	bidActed         [GermanSoloPlayerCnt]bool          // 各プレイヤーが宣言済みか
	playerScores     [GermanSoloPlayerCnt]int           // 累積ゲーム点
	lastTrickWinner  int                                // 最終トリック勝者 (-1=未確定)

	// エース呼び。calledAceSuit は指名されたエースのスート (-1=未指名)。
	// **呼び声は卓で聞こえるのでエース自体は公開情報**だが、誰が持っているかは
	// そのエースが場に出るまで伏せる —— そこがこのゲームの緊張感そのもの。
	// playsAlone は単独契約 (Solo / Tout) と、呼べるエースが 1 枚も無かった Frage。
	calledAceSuit   int
	partnerIdx      int // 呼ばれたエースの持ち主 (-1=未確定 / 単独)
	partnerRevealed bool
	playsAlone      bool
	// **同じトリックを二度精算しない。** ResolveTrick はトリック終了フェーズで
	// 呼ばれるが、そのフェーズは NextTrick まで続く。二度呼ぶと勝者に同じ札束が
	// 二度積まれ、8 トリックのディールで 9 トリック取った席が出る。
	trickResolved bool
	outcome       GermanSoloOutcome // 直近ディールの結果
	result        GermanSoloResult  // 人間視点のマッチ結果
	scored        bool              // 当該ディールの得点計算済みか (RoundEnd 突入時に一度だけ)
	gameEndFlag   bool
	winnerPlayer  int // -1=未確定
	actionLogBase
}

// NewGermanSolo コンストラクタ
func NewGermanSolo(trumpCards *TrumpCards, players []*GermanSoloPlayer, config GermanSoloConfig) *GermanSolo {
	return &GermanSolo{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		winnerPlayer:    -1,
		lastTrickWinner: -1,
		declarerIdx:     -1,
		trumpSuit:       -1,
	}
}

// NewDefaultGermanSolo 標準の 3 人構成 (人間 1, CPU 2) と既定設定で生成する。
func NewDefaultGermanSolo() *GermanSolo {
	players := make([]*GermanSoloPlayer, GermanSoloPlayerCnt)
	players[0] = NewGermanSoloPlayer(true)
	for i := 1; i < GermanSoloPlayerCnt; i++ {
		players[i] = NewGermanSoloPlayer(false)
	}
	return NewGermanSolo(newGermanSoloDeck(), players, DefaultGermanSoloConfig())
}

// newGermanSoloDeck GermanSolo 用 32 枚のスカート・パックを生成する。
// 標準 52 枚 (NewTrumpCards(0)) から 2,3,4,5,6 を除外し、A,K,Q,J,10,9,8,7 を残す。
// domain パッケージ内なので TrumpCards の内部 deck を直接フィルタできる。build-tag 無しの
// NewTrumpCards はどのワーカーからも到達可能。
func newGermanSoloDeck() *TrumpCards {
	full := NewTrumpCards(0)
	t := new(TrumpCards)
	t.deck = make([]*Card, 0, GermanSoloDeckSize)
	for _, c := range full.deck {
		v := c.GetValue()
		if v >= 2 && v <= 6 {
			continue
		}
		t.deck = append(t.deck, NewCard(c.GetDesign(), v, false))
	}
	t.deckCnt = len(t.deck)
	t.deckInit()
	return t
}

// Reset ゲーム初期化
func (g *GermanSolo) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	// **ディーラーは人間の右隣。** forehand = dealer+1 が最初に宣言するので、
	// dealerIdx=0 にすると人間 (席 0) は毎回最後に喋る番になり、開幕ディールで
	// 一度も高い契約を選べない。ここを最終席にしておくと人間が forehand になる。
	g.dealerIdx = GermanSoloPlayerCnt - 1
	g.playerScores = [GermanSoloPlayerCnt]int{}
	g.result = GermanSoloResultNone
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のディールを開始する
func (g *GermanSolo) NextRound() {
	if g.phase != GermanSoloPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % GermanSoloPlayerCnt
	g.startRound()
}

// startRound 手札を配り、ビッドフェーズを開始する。
func (g *GermanSolo) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.trickResolved = false
	g.lastTrickWinner = -1
	g.declarerIdx = -1
	g.winningBid = GermanSoloBidNone
	g.trumpSuit = -1
	g.outcome = GermanSoloOutcomeNone
	g.scored = false
	// **エース呼びはディールごとにやり直す。** 持ち越すと前のディールの相方が
	// そのまま味方として残り、勝敗の集計が静かにずれる。
	g.calledAceSuit = -1
	g.partnerIdx = -1
	g.partnerRevealed = false
	g.playsAlone = false
	g.bids = [GermanSoloPlayerCnt]GermanSoloBid{}
	g.bidActed = [GermanSoloPlayerCnt]bool{}
	for i := range g.bidTrump {
		g.bidTrump[i] = -1
	}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()

	g.forehandIdx = (g.dealerIdx + 1) % GermanSoloPlayerCnt
	g.sortAllHands()

	g.currentBidderIdx = g.forehandIdx
	g.phase = GermanSoloPhaseBid
}

// deal 各プレイヤーへ GermanSoloHandSize 枚を配る。32 枚をちょうど配り切るのでストックは残らない。
func (g *GermanSolo) deal() {
	for i := 0; i < GermanSoloHandSize; i++ {
		for j := 0; j < GermanSoloPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % GermanSoloPlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
}

// --- Bidding ---

// PlayerBid 人間がビッドする。bid は Pass(None)/Frage/Solo/Tout。trumpSuit は宣言時の切り札 (1..4)。
func (g *GermanSolo) PlayerBid(bid GermanSoloBid, trumpSuit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != GermanSoloPhaseBid {
		return ErrWrongPhase
	}
	if !g.players[g.currentBidderIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if !g.isBidLegal(bid) {
		return NewDomainError(ErrInvalidPlay, "現在の最高ビッドを上回る宣言が必要です")
	}
	if bid != GermanSoloBidNone && !germanSoloValidSuit(trumpSuit) {
		return NewDomainError(ErrInvalidPlay, "切り札スートを選んでください (1..4)")
	}
	g.applyBid(g.currentBidderIdx, bid, trumpSuit)
	return nil
}

// CpuBid 現在のビッド手番が CPU の場合に 1 回ビッドする。
func (g *GermanSolo) CpuBid() {
	if g.gameEndFlag || g.phase != GermanSoloPhaseBid {
		return
	}
	idx := g.currentBidderIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	bid := g.cpuChooseBid(idx)
	trump := -1
	if bid != GermanSoloBidNone {
		trump = g.cpuChooseTrump(idx)
	}
	g.applyBid(idx, bid, trump)
}

// isBidLegal bid が合法か (pass は常に可、Frage/Solo/Tout は現最高ビッドを上回る必要)。
//
// **Mussfrage は競りでは選べない。** 全員パスの後に卓が押し付ける契約であって、
// 手を挙げて取りに行く契約ではない。ここを通すと 1 点しか動かない最下位契約を
// 宣言して降りるより安く逃げられてしまう。
func (g *GermanSolo) isBidLegal(bid GermanSoloBid) bool {
	if bid == GermanSoloBidNone {
		return true
	}
	if bid != GermanSoloBidFrage && bid != GermanSoloBidSolo && bid != GermanSoloBidTout {
		return false
	}
	return bid > g.highestBid()
}

// highestBid これまでに宣言された最高ビッドを返す。
func (g *GermanSolo) highestBid() GermanSoloBid {
	best := GermanSoloBidNone
	for _, b := range g.bids {
		if b > best {
			best = b
		}
	}
	return best
}

// applyBid ビッドを適用し、全員が宣言し終えたら auction を締める。
func (g *GermanSolo) applyBid(playerIdx int, bid GermanSoloBid, trumpSuit int) {
	g.bids[playerIdx] = bid
	g.bidActed[playerIdx] = true
	if bid == GermanSoloBidNone {
		g.bidTrump[playerIdx] = -1
		g.appendLog(playerIdx, "bid_pass",
			fmt.Sprintf("%s passes", playerName(g.players, playerIdx)), nil)
	} else {
		g.bidTrump[playerIdx] = trumpSuit
		g.appendLog(playerIdx, "bid",
			fmt.Sprintf("%s bids %s (trump %s)", playerName(g.players, playerIdx), germanSoloBidName(bid), germanSoloSuitName(trumpSuit)), nil)
	}

	if g.allBidsActed() {
		g.finalizeAuction()
		return
	}
	g.currentBidderIdx = g.nextBidder(playerIdx)
}

// allBidsActed 全員が宣言済みか。
func (g *GermanSolo) allBidsActed() bool {
	for _, acted := range g.bidActed {
		if !acted {
			return false
		}
	}
	return true
}

// nextBidder playerIdx の次でまだ宣言していないプレイヤーを返す。
func (g *GermanSolo) nextBidder(playerIdx int) int {
	for i := 1; i <= GermanSoloPlayerCnt; i++ {
		cand := (playerIdx + i) % GermanSoloPlayerCnt
		if !g.bidActed[cand] {
			return cand
		}
	}
	return playerIdx
}

// finalizeAuction 最高ビッドの宣言者を落札者に確定し、次のフェーズへ進む。
// 同値は forehand から時計回りで最初に宣言したプレイヤーが優先される。
// 全員パスなら Spadille (♣Q) の持ち主が Mussfrage を強制的に引き受ける。
func (g *GermanSolo) finalizeAuction() {
	best := g.highestBid()
	declarer := -1
	if best > GermanSoloBidNone {
		for i := 0; i < GermanSoloPlayerCnt; i++ {
			cand := (g.forehandIdx + i) % GermanSoloPlayerCnt
			if g.bids[cand] == best {
				declarer = cand
				break
			}
		}
	}
	if declarer < 0 {
		declarer, best = g.forceMussfrage()
	}
	g.declarerIdx = declarer
	g.winningBid = best
	g.trumpSuit = g.bidTrump[declarer]
	if !germanSoloValidSuit(g.trumpSuit) {
		g.trumpSuit = g.cpuChooseTrump(declarer)
	}
	// **切り札が決まった時点で並べ替える。** プレイ開始まで待つと、エース呼びの
	// 画面だけ配られたままの並びになり、どれが切り札か読めないまま指名させられる。
	g.sortAllHands()
	g.appendLog(declarer, "declarer",
		fmt.Sprintf("%s declares %s (trump %s)", playerName(g.players, declarer), germanSoloBidName(best), germanSoloSuitName(g.trumpSuit)), nil)
	if germanSoloIsPartnerBid(best) {
		g.startAceCall()
		return
	}
	// Solo / Tout は単独契約。呼ぶエースが無いのでそのままプレイへ入る。
	g.playsAlone = true
	g.partnerIdx = -1
	g.startPlay()
}

// forceMussfrage は全員パスのディールで Mussfrage を押し付ける席を決める。
//
// **配り直しはしない。** CPU がビッドしない手を引き続ければディールが永久に
// 終わらないので、卓の誰かが必ず引き受ける。伝統どおり Spadille (♣Q) を持つ
// 席が引き受け、♣Q が誰の手にも無い盤面 (テストが手札を差し替えた場合など) では
// forehand が引き受ける。
func (g *GermanSolo) forceMussfrage() (int, GermanSoloBid) {
	declarer := g.spadilleHolder()
	if declarer < 0 {
		declarer = g.forehandIdx
	}
	g.bidTrump[declarer] = g.cpuChooseTrump(declarer)
	g.appendLog(declarer, "mussfrage",
		fmt.Sprintf("all pass: %s must take the Mussfrage", playerName(g.players, declarer)), nil)
	return declarer, GermanSoloBidMussfrage
}

// spadilleHolder は Spadille (♣Q) を持つ席を返す (-1=誰も持っていない)。
func (g *GermanSolo) spadilleHolder() int {
	for i, p := range g.players {
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			if c != nil && c.GetDesign() == CardDesignClover && c.GetValue() == germanSoloQueenValue {
				return i
			}
		}
	}
	return -1
}

// startAceCall はエース呼びフェーズに入る。
//
// 呼べるエースが 1 枚も無ければ (落札者が非切り札のエースを全部握っている盤面)
// 相手を呼びようがないので、Frage のまま **単独で** 5 トリックを狙う。
func (g *GermanSolo) startAceCall() {
	if len(g.callableAceSuits(g.declarerIdx)) == 0 {
		g.playsAlone = true
		g.partnerIdx = -1
		g.appendLog(g.declarerIdx, "plays_alone",
			fmt.Sprintf("%s holds every callable ace and plays the Frage alone",
				playerName(g.players, g.declarerIdx)), nil)
		g.startPlay()
		return
	}
	g.currentPlayerIdx = g.declarerIdx
	g.phase = GermanSoloPhaseAceCall
}

// callableAceSuits は落札者が**呼べる**エースのスートを返す。
//
// 除くのは 2 種類:
//
//   - 自分が持っているエース。呼んでも味方が増えない。
//   - **切り札スートのエース。** 切り札の A は三大マタドールに次ぐ 4 番目の切り札で、
//     これを呼べるとただの「上から 4 枚目の切り札を引き寄せる手」になり、味方探しでなく
//     切り札の補充になってしまう。伝統的にも呼ぶエースは切り札以外から選ぶ。
func (g *GermanSolo) callableAceSuits(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	held := map[int]bool{}
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if c := p.GetCard(i); c != nil && c.GetValue() == germanSoloAceValue {
			held[c.GetDesign()] = true
		}
	}
	out := make([]int, 0, CardDesignMax)
	for suit := CardDesignSpade; suit <= CardDesignMax; suit++ {
		if held[suit] || suit == g.trumpSuit {
			continue
		}
		out = append(out, suit)
	}
	return out
}

// DeclareAce は落札者が味方を呼ぶエースを指名する。
func (g *GermanSolo) DeclareAce(playerIdx, suit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != GermanSoloPhaseAceCall {
		return ErrWrongPhase
	}
	if playerIdx != g.declarerIdx {
		return NewDomainError(ErrInvalidPlay, "エースを呼べるのは落札者だけです")
	}
	if !germanSoloValidSuit(suit) {
		return NewDomainError(ErrInvalidCard, "スートを選んでください (1..4)")
	}
	if !slices.Contains(g.callableAceSuits(playerIdx), suit) {
		// **自分が持っているエース・切り札のエースは呼べない。** 前者を通すと
		// 味方が増えないまま「4 人卓の 1 対 3」が黙って成立し、後者を通すと
		// 味方探しでなく切り札の補充になる。
		return NewDomainError(ErrInvalidPlay, "自分が持っているエースと切り札のエースは呼べません")
	}

	g.calledAceSuit = suit
	g.partnerIdx = g.findAceHolder(suit)
	g.partnerRevealed = false
	g.appendLog(playerIdx, "call_ace",
		fmt.Sprintf("%s calls the Ace of %s", playerName(g.players, playerIdx), germanSoloSuitName(suit)), nil)
	g.startPlay()
	return nil
}

// CpuDeclareAce は CPU の落札者にエースを呼ばせる。
//
// 手札に一番少ないスートのエースを選ぶ: そのスートは自分では取れないので、
// 味方に持たせる価値が一番高い。
func (g *GermanSolo) CpuDeclareAce() {
	if g.phase != GermanSoloPhaseAceCall {
		return
	}
	idx := g.declarerIdx
	if idx < 0 || g.players[idx].GetIsHuman() {
		return
	}
	if suit := g.cpuPickAceSuit(idx); germanSoloValidSuit(suit) {
		_ = g.DeclareAce(idx, suit)
	}
}

// cpuPickAceSuit は呼ぶエースのスートを選ぶ (-1=呼べるエースが無い)。
//
// **CPU もヒントもここを読む。** 別々に書くと、助言に従った人間だけが
// 違うスートを呼ぶことになる。
func (g *GermanSolo) cpuPickAceSuit(playerIdx int) int {
	callable := g.callableAceSuits(playerIdx)
	if len(callable) == 0 {
		return -1
	}
	counts := map[int]int{}
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if c := p.GetCard(i); c != nil {
			counts[c.GetDesign()]++
		}
	}
	best := callable[0]
	for _, suit := range callable[1:] {
		if counts[suit] < counts[best] {
			best = suit
		}
	}
	return best
}

// findAceHolder は指定スートのエースを持つ席を返す (-1=誰も持っていない)。
//
// 40 枚を配り切るので通常は必ず誰かが持っているが、テストが手札を差し替えた
// 盤面では見つからないことがある。その場合は単独プレイと同じ扱いになる。
func (g *GermanSolo) findAceHolder(suit int) int {
	for i, p := range g.players {
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			if c != nil && c.GetValue() == germanSoloAceValue && c.GetDesign() == suit {
				return i
			}
		}
	}
	return -1
}

// IsHumanAceCallTurn は人間のエース呼び待ちかを返す。
func (g *GermanSolo) IsHumanAceCallTurn() bool {
	return g.phase == GermanSoloPhaseAceCall &&
		g.declarerIdx >= 0 && g.players[g.declarerIdx].GetIsHuman()
}

// GetCalledAceSuit は呼ばれたエースのスートを返す (-1=未指名)。
func (g *GermanSolo) GetCalledAceSuit() int { return g.calledAceSuit }

// GetPartnerIdx は味方の席を返す。**まだエースが場に出ていなければ -1。**
// 誰が味方かはそのエースが出るまで伏せる。
func (g *GermanSolo) GetPartnerIdx() int {
	if !g.partnerRevealed {
		return -1
	}
	return g.partnerIdx
}

// IsPlayingAlone は落札者が単独で戦っているかを返す。
func (g *GermanSolo) IsPlayingAlone() bool { return g.playsAlone }

// GetCallableAceSuits は落札者が呼べるエースのスートを返す (画面の選択肢)。
func (g *GermanSolo) GetCallableAceSuits() []int {
	if g.phase != GermanSoloPhaseAceCall {
		return nil
	}
	return g.callableAceSuits(g.declarerIdx)
}

// germanSoloSideOf は席 idx が落札者側かを返す。
//
// **単独プレイでは落札者だけ。** 味方が伏せられている間も内部の判定には
// 実際の相方を使う —— 伏せているのは表示であって、勝敗の計算ではない。
func (g *GermanSolo) germanSoloSideOf(idx int) bool {
	if idx == g.declarerIdx {
		return true
	}
	return !g.playsAlone && g.partnerIdx >= 0 && idx == g.partnerIdx
}

// GermanSolo の CPU がビッドに踏み切る手札強度のしきい値 (evalTrump の評価値)。
const (
	// germanSoloBidThresholdFrage Frage (味方を呼んで 5 トリック) に必要な強度。
	germanSoloBidThresholdFrage = 11
	// germanSoloBidThresholdSolo Solo (単独で 5 トリック) に必要な強度。
	germanSoloBidThresholdSolo = 17
)

// germanSoloCanSweep は playerIdx の手札が trump を切り札として **8 トリック全部を
// 確実に取れる**かを返す。Tout を宣言してよい唯一の条件。
//
// **Tout は強度では測れない。** 8 トリック全部が条件なので、「切り札が多い」では
// なく「1 枚も落とさない」かが問われる。評価値のしきい値で代用すると、実測で
// 5 回に 1 回しか成立しない手が 24 点を賭ける契約を宣言し、CPU もヒントも同じ
// 負け方を勧める (しきい値 22 で 111 回宣言し 19.8% しか成立しなかった)。
//
// 確実と言い切れるのは、**手札が切り札序列の上から連続した札だけでできている**とき。
// このとき手札より強い札は 1 枚も残っていないので、リードを取られても自分が
// フォローすれば必ず勝つ。平札のエースを数に入れないのは、切り札で殺されるから。
func (g *GermanSolo) germanSoloCanSweep(playerIdx, trump int) bool {
	p := g.players[playerIdx]
	if p.GetCardsSize() == 0 {
		return false
	}
	held := func(design, value int) bool {
		for i := 0; i < p.GetCardsSize(); i++ {
			if c := p.GetCard(i); c != nil && c.GetDesign() == design && c.GetValue() == value {
				return true
			}
		}
		return false
	}
	prefix := 0
	for _, c := range germanSoloTrumpOrder(trump) {
		if !held(c.design, c.value) {
			break
		}
		prefix++
	}
	return prefix >= p.GetCardsSize()
}

// germanSoloTrumpCard は切り札序列の 1 枚。
type germanSoloTrumpCard struct{ design, value int }

// germanSoloTrumpOrder は trump を切り札としたときの切り札を強い順に返す。
//
// **序列は germanSoloCardStrength と同じ順に並べる。** 別に書くと、片方を
// 変えたときに「確実に取れる札」の数え方だけが古いままになる。
func germanSoloTrumpOrder(trump int) []germanSoloTrumpCard {
	out := []germanSoloTrumpCard{
		{CardDesignClover, germanSoloQueenValue}, // Spadille
		{trump, germanSoloManilleValue},          // Manille
		{CardDesignSpade, germanSoloQueenValue},  // Basta
	}
	for _, v := range []int{germanSoloAceValue, 13, 12, 11, 10, 9, 8} {
		if v == germanSoloQueenValue && (trump == CardDesignSpade || trump == CardDesignClover) {
			continue // その Q は上で Spadille / Basta として数えている
		}
		out = append(out, germanSoloTrumpCard{trump, v})
	}
	return out
}

// cpuChooseBid CPU が手札強度からビッドを選ぶ。
func (g *GermanSolo) cpuChooseBid(playerIdx int) GermanSoloBid {
	if g.config.CpuDifficulty == GermanSoloCpuDifficultyEasy {
		return GermanSoloBidNone
	}
	return g.bidForStrength(playerIdx)
}

// bidForStrength は手札強度と現在の最高宣言から、選べる中で一番高い宣言を返す。
//
// **手札が支えられない段には登らない。** 上回るためだけに一段上を返すと、
// Frage がやっとの手札が Tout を掴む。届かないなら降りる (パス) が正解で、
// これは CPU とヒントの両方が読む唯一の判断なので、片方だけ甘くしてはいけない。
func (g *GermanSolo) bidForStrength(playerIdx int) GermanSoloBid {
	suit, strength := g.handBestTrump(playerIdx)
	highest := g.highestBid()
	want := GermanSoloBidNone
	switch {
	case g.germanSoloCanSweep(playerIdx, suit):
		want = GermanSoloBidTout
	case strength >= germanSoloBidThresholdSolo:
		want = GermanSoloBidSolo
	case strength >= germanSoloBidThresholdFrage:
		want = GermanSoloBidFrage
	}
	if want <= highest {
		return GermanSoloBidNone
	}
	return want
}

// cpuChooseTrump CPU がジャーマン・ソロとして選ぶ切り札スート (最良スート) を返す。
func (g *GermanSolo) cpuChooseTrump(playerIdx int) int {
	suit, _ := g.handBestTrump(playerIdx)
	return suit
}

// handBestTrump 各スートを切り札とした場合の手札強度を評価し、最良スートとその評価値を返す。
func (g *GermanSolo) handBestTrump(playerIdx int) (int, int) {
	bestSuit, bestScore := CardDesignSpade, -1
	for _, s := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if sc := g.evalTrump(playerIdx, s); sc > bestScore {
			bestScore = sc
			bestSuit = s
		}
	}
	return bestSuit, bestScore
}

// evalTrump trump を切り札としたときの手札強度を見積もる。
//
// 切り札 1 枚 = 2 点、三大マタドールはさらに +3。平札はエース +2 / キング +1。
// 8 枚の手札なので、切り札 4 枚とマタドール 2 枚で 17 前後 = Solo の目安になる。
func (g *GermanSolo) evalTrump(playerIdx, trump int) int {
	p := g.players[playerIdx]
	score := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if germanSoloIsTrump(c, trump) {
			score += 2
			if germanSoloCardStrength(c, trump) >= germanSoloStrBasta {
				score += 3 // 三大マタドール
			}
			continue
		}
		switch c.GetValue() {
		case germanSoloAceValue:
			score += 2 // 平札のエースはそのスートの最強札
		case germanSoloKingValue:
			score++
		}
	}
	return score
}

// --- Play ---

// startPlay ビッド確定後、プレイフェーズを開始する。
//
// **リードするのは落札者ではなく forehand (ディーラーの左隣)。** 落札者が
// 先に打てると、単独契約でも自分の切り札から好きな順に叩ける。伝統どおり
// 前席が先に出すことで、守備側が最初の 1 枚で探りを入れられる。
func (g *GermanSolo) startPlay() {
	g.sortAllHands()
	g.leadPlayerIdx = g.forehandIdx
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = GermanSoloPhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *GermanSolo) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != GermanSoloPhasePlay {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	player := g.players[g.currentPlayerIdx]
	if cardIndex < 0 || cardIndex >= player.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	card := player.GetCard(cardIndex)
	if err := g.validatePlay(g.currentPlayerIdx, card); err != nil {
		return err
	}
	played := player.RemoveCard(cardIndex)
	g.playCard(g.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行する。
func (g *GermanSolo) CpuPlay() {
	if g.gameEndFlag || g.phase != GermanSoloPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	cardIdx := g.cpuSelectPlayCard(idx)
	played := g.players[idx].RemoveCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら RemoveCard(0) は nil を返す。それを playCard に渡すと
	// nil デリファレンスで HTTP ハンドラごと落ちる (#4606)。
	if played == nil {
		return
	}
	g.playCard(idx, played)
}

// playCard カードをプレイする共通処理。
func (g *GermanSolo) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})
	g.revealPartnerIfCalledAce(playerIdx, card)

	if len(g.currentTrick) == GermanSoloPlayerCnt {
		g.phase = GermanSoloPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % GermanSoloPlayerCnt
	}
}

// revealPartnerIfCalledAce は呼ばれたエースが場に出た時点で味方を公開する。
//
// **誰が味方かはそれまで伏せる** —— 落札者以外は自分が呼ばれたことすら
// 知らないまま打つ、というのが GermanSolo の緊張感そのもの。
func (g *GermanSolo) revealPartnerIfCalledAce(playerIdx int, card *Card) {
	if g.partnerRevealed || g.playsAlone || card == nil {
		return
	}
	if card.GetValue() != germanSoloAceValue || card.GetDesign() != g.calledAceSuit {
		return
	}
	g.partnerRevealed = true
	g.appendLog(playerIdx, "partner_revealed",
		fmt.Sprintf("%s holds the called Ace of %s and is the GermanSolo's partner",
			playerName(g.players, playerIdx), germanSoloSuitName(g.calledAceSuit)), nil)
}

// ResolveTrick トリックを解決して勝者を決定する。最終トリックなら RoundEnd に入り、得点計算を発火する。
func (g *GermanSolo) ResolveTrick() {
	if g.phase != GermanSoloPhaseTrickEnd || len(g.currentTrick) != GermanSoloPlayerCnt {
		return
	}
	if g.trickResolved {
		return
	}
	g.trickResolved = true
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", playerName(g.players, winnerIdx), g.trickNumber), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= GermanSoloTrickCount {
		g.lastTrickWinner = winnerIdx
		g.phase = GermanSoloPhaseRoundEnd
		g.enterRoundEnd()
	} else {
		g.phase = GermanSoloPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *GermanSolo) NextTrick() {
	if g.phase != GermanSoloPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.trickResolved = false
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = GermanSoloPhasePlay
}

// enterRoundEnd RoundEnd 突入時に一度だけ結果判定と得点計算を行う (scored フラグでガード)。
func (g *GermanSolo) enterRoundEnd() {
	if g.scored {
		return
	}
	g.scored = true
	g.outcome = g.evalOutcome()
	g.applyScores(g.outcome)
	ours, _ := g.sideTrickCounts()
	g.appendLog(-1, "round_score",
		fmt.Sprintf("round %d: %s %s with %d/%d tricks (needed %d, stake=%d)",
			g.roundNumber, playerName(g.players, g.declarerIdx), germanSoloOutcomeName(g.outcome),
			ours, GermanSoloTrickCount, g.RequiredTricks(), germanSoloBidValue(g.winningBid)), nil)
	g.checkGameEnd()
}

// evalOutcome は宣言側が契約の必要トリック数に届いたかを判定する。
//
// **相手より多く取ったかではなく、宣言した数に届いたか。** 5 トリック契約で
// 4 対 4 に分かれたら、相手に負けていなくても契約は落ちている。Tout は 8 が必要なので
// 7 でも失敗。
//
// **提携の単位と判定の単位を一致させる。** 呼ばれたエースの持ち主は味方なので、
// 落札者 1 人分だけ数えると味方が取ったトリックが守備側に計上され、勝敗が反転する。
func (g *GermanSolo) evalOutcome() GermanSoloOutcome {
	ours, _ := g.sideTrickCounts()
	if ours >= g.RequiredTricks() {
		return GermanSoloOutcomeMade
	}
	return GermanSoloOutcomeFailed
}

// RequiredTricks は確定した契約の成功に必要なトリック数を返す (Tout=8, それ以外=5)。
func (g *GermanSolo) RequiredTricks() int { return germanSoloRequiredTricks(g.winningBid) }

// GetDeclarerSideSize は宣言側の人数 (単独=1, 味方あり=2) を返す。
func (g *GermanSolo) GetDeclarerSideSize() int {
	n := 0
	for i := 0; i < GermanSoloPlayerCnt; i++ {
		if g.germanSoloSideOf(i) {
			n++
		}
	}
	return n
}

// sideTrickCounts は (落札者側, 相手側) の獲得トリック数を返す。
func (g *GermanSolo) sideTrickCounts() (int, int) {
	ours, theirs := 0, 0
	for i := 0; i < GermanSoloPlayerCnt; i++ {
		n := g.players[i].GetTrickCount()
		if g.germanSoloSideOf(i) {
			ours += n
		} else {
			theirs += n
		}
	}
	return ours, theirs
}

// GetSideTrickCounts は (落札者側, 相手側) の獲得トリック数を返す。
func (g *GermanSolo) GetSideTrickCounts() (int, int) { return g.sideTrickCounts() }

// applyScores ディール結果に応じて累積点を更新する。
//
// **動く点はゼロ和にする。** 守備側の各席が契約値 V を払い、その合計を宣言側で
// 山分けする。単独 (1 対 3) なら落札者 +3V / 各守備 -V、味方あり (2 対 2) なら
// 宣言側の 2 人が +V ずつ / 守備側の 2 人が -V ずつ。失敗なら符号が反転する。
// 割り切れる組み合わせしか無い (3/1 と 2/2) ので、端数は出ない。
func (g *GermanSolo) applyScores(outcome GermanSoloOutcome) {
	value := germanSoloBidValue(g.winningBid)
	if value == 0 || outcome == GermanSoloOutcomeNone {
		return
	}
	side := g.GetDeclarerSideSize()
	if side <= 0 || side >= GermanSoloPlayerCnt {
		return
	}
	opps := GermanSoloPlayerCnt - side
	germanSoloDelta, oppDelta := value*opps/side, -value
	if outcome == GermanSoloOutcomeFailed {
		germanSoloDelta, oppDelta = -germanSoloDelta, -oppDelta
	}
	for i := 0; i < GermanSoloPlayerCnt; i++ {
		// **味方も落札者と同じ側の点を受け取る。** 落札者だけに配ると、
		// 呼ばれたエースを持っていた席は自分が勝った側にいながら相手側の点を貰う。
		if g.germanSoloSideOf(i) {
			g.playerScores[i] += germanSoloDelta
		} else {
			g.playerScores[i] += oppDelta
		}
	}
}

// checkGameEnd 規定ディール数を終えたらマッチ終了を判定し、累積点最上位を勝者とする。
func (g *GermanSolo) checkGameEnd() {
	if g.roundNumber < g.config.TargetRounds {
		return
	}
	leader, best := 0, g.playerScores[0]
	tie := false
	for i := 1; i < GermanSoloPlayerCnt; i++ {
		if g.playerScores[i] > best {
			best = g.playerScores[i]
			leader = i
			tie = false
		} else if g.playerScores[i] == best {
			tie = true
		}
	}
	g.gameEndFlag = true
	g.winnerPlayer = leader
	g.phase = GermanSoloPhaseGameEnd
	g.result = g.humanResult(leader, tie)
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", playerName(g.players, leader)), nil)
}

// humanResult 人間 (seat 0) の視点でマッチ結果を返す。単独トップなら Win、トップ同点なら None、他は Lose。
func (g *GermanSolo) humanResult(leader int, tie bool) GermanSoloResult {
	human := findHumanIdx(g.players)
	if human < 0 {
		return GermanSoloResultNone
	}
	if g.playerScores[human] == g.playerScores[leader] {
		if tie {
			return GermanSoloResultNone
		}
		return GermanSoloResultWin
	}
	return GermanSoloResultLose
}

// ScoreRound RoundEnd フェーズでの得点計算を行う (enterRoundEnd を idempotent に呼ぶ、インタフェース互換)。
func (g *GermanSolo) ScoreRound() {
	if g.phase != GermanSoloPhaseRoundEnd {
		return
	}
	g.enterRoundEnd()
}

// --- Trick / play helpers ---

// validatePlay マストフォローを検証する (切り札グループを 1 スートとして扱う)。
func (g *GermanSolo) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadEff := germanSoloEffectiveSuit(g.currentTrick[0].Card, g.trumpSuit)
	if germanSoloEffectiveSuit(card, g.trumpSuit) != leadEff && g.playerHasEffSuit(playerIdx, leadEff) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasEffSuit プレイヤーが指定の実効スート (平札スート or 切り札グループ) を持っているか。
func (g *GermanSolo) playerHasEffSuit(playerIdx, eff int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if germanSoloEffectiveSuit(p.GetCard(i), g.trumpSuit) == eff {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する。切り札があれば最強切り札、無ければリードスートの最強札が勝つ。
func (g *GermanSolo) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadEff := germanSoloEffectiveSuit(g.currentTrick[0].Card, g.trumpSuit)
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerStr := germanSoloCardStrength(g.currentTrick[0].Card, g.trumpSuit)
	for _, tc := range g.currentTrick[1:] {
		eff := germanSoloEffectiveSuit(tc.Card, g.trumpSuit)
		if eff != germanSoloTrumpGroup && eff != leadEff {
			continue // 場外の平札は勝てない
		}
		if s := germanSoloCardStrength(tc.Card, g.trumpSuit); s > winnerStr {
			winnerIdx = tc.PlayerIdx
			winnerStr = s
		}
	}
	return winnerIdx
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *GermanSolo) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.validatePlay(playerIdx, c) == nil })
}

// sameSide a と b が同じ陣営 (両方宣言側 or 両方守備側) か。
//
// **味方を数える。** 落札者かどうかだけで見ると、呼ばれたエースを持つ席は
// 落札者と別陣営に見えるので、CPU が味方の勝ちトリックを上から取りに行く。
func (g *GermanSolo) sameSide(a, b int) bool {
	return g.germanSoloSideOf(a) == g.germanSoloSideOf(b)
}

// --- Card ranking (GermanSolo matador ranking, inline) ---

// germanSoloTrumpGroup 実効スートとしての「切り札グループ」を表す番号 (Joker=0 は 32 枚デッキに現れないため流用)。
const germanSoloTrumpGroup = 0

// germanSoloAceValue エースの札位 (呼び声の対象、平札の最強)。
const germanSoloAceValue = 1

// germanSoloQueenValue クイーンの札位 (♣Q=Spadille / ♠Q=Basta)。
const germanSoloQueenValue = 12

// germanSoloKingValue キングの札位 (手札評価で加点する高位平札)。
const germanSoloKingValue = 13

// germanSoloManilleValue Manille の札位 (切り札スートの 7)。
const germanSoloManilleValue = 7

// 切り札グループ内の強さ (数値が大きいほど強い)。平札の最大 (8) を大きく上回るよう基準を高く取る。
const (
	germanSoloStrSpadille = 200 // ♣Q (常に第 1 切り札)
	germanSoloStrManille  = 199 // 切り札スートの 7 (常に第 2)
	germanSoloStrBasta    = 198 // ♠Q (常に第 3)
)

// germanSoloIsTrump card が (trump を切り札としたとき) 切り札か。♣Q・♠Q は常に切り札。
func germanSoloIsTrump(c *Card, trump int) bool {
	d, v := c.GetDesign(), c.GetValue()
	if v == germanSoloQueenValue && (d == CardDesignClover || d == CardDesignSpade) {
		return true // Spadille / Basta
	}
	return d == trump
}

// germanSoloEffectiveSuit フォロー判定用の実効スート。切り札は germanSoloTrumpGroup、それ以外は本来のスート。
func germanSoloEffectiveSuit(c *Card, trump int) int {
	if germanSoloIsTrump(c, trump) {
		return germanSoloTrumpGroup
	}
	return c.GetDesign()
}

// germanSoloCardStrength trump を切り札としたときのカード強さ (全カードの全順序)。任意の切り札 > 任意の平札。
func germanSoloCardStrength(c *Card, trump int) int {
	d, v := c.GetDesign(), c.GetValue()
	if d == CardDesignClover && v == germanSoloQueenValue {
		return germanSoloStrSpadille
	}
	if d == trump && v == germanSoloManilleValue {
		return germanSoloStrManille
	}
	if d == CardDesignSpade && v == germanSoloQueenValue {
		return germanSoloStrBasta
	}
	if d == trump {
		return germanSoloTrumpSuitRank(v)
	}
	return germanSoloPlainRank(v)
}

// GermanSoloMatadorRank は札が三大マタドールのどれかを返す。
//
// 1 = スパディーユ (♣Q)、2 = マニーユ (切り札の 7)、3 = バスタ (♠Q)、
// 0 = マタドールでない。**切り札が未確定なら 0。** マニーユは切り札スート次第で
// 決まるので、確定前に一部だけ示すと不揃いな案内になる。
//
// **判定は germanSoloCardStrength をそのまま読む。** 別に条件を書くと、序列を
// 変えたときに表示だけ古いままになる。
func GermanSoloMatadorRank(c *Card, trump int) int {
	if c == nil || !germanSoloValidSuit(trump) {
		return 0
	}
	switch germanSoloCardStrength(c, trump) {
	case germanSoloStrSpadille:
		return 1
	case germanSoloStrManille:
		return 2
	case germanSoloStrBasta:
		return 3
	}
	return 0
}

// germanSoloTrumpSuitRank 切り札スートの残り札 A>K>Q>J>10>9>8 の強さ。
//
// **Q がここに残るのは切り札が赤 (♥♦) のときだけ。** ♠♣ が切り札なら、その Q は
// 上で Spadille / Basta として捕まえられているのでここには来ない。だから切り札の
// 枚数は赤 10 枚 / 黒 9 枚になる。7 は Manille なので同じくここには来ない。
func germanSoloTrumpSuitRank(v int) int {
	switch v {
	case germanSoloAceValue: // A
		return 197
	case 13: // K
		return 196
	case 12: // Q (赤切り札のみ)
		return 195
	case 11: // J
		return 194
	case 10:
		return 193
	case 9:
		return 192
	default: // 8
		return 191
	}
}

// germanSoloPlainRank 平札 (非切り札) の強さ A>K>Q>J>10>9>8>7。
//
// ♣Q・♠Q は常に切り札なのでここには来ない。切り札スートの 7 も Manille なので来ない。
func germanSoloPlainRank(v int) int {
	switch v {
	case germanSoloAceValue: // A
		return 8
	case 13: // K
		return 7
	case 12: // Q (♥♦ のみ)
		return 6
	case 11: // J
		return 5
	case 10:
		return 4
	case 9:
		return 3
	case 8:
		return 2
	default: // 7
		return 1
	}
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *GermanSolo) sortAllHands() {
	for _, p := range g.players {
		germanSoloSortHand(p, g.trumpSuit)
	}
}

// germanSoloSortHand 手札を実効スート→強さ順にソートする。trump=-1 (未確定) の場合はスート→値で並べる。
func germanSoloSortHand(p *GermanSoloPlayer, trump int) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if !germanSoloValidSuit(trump) {
			if cards[i].GetDesign() != cards[j].GetDesign() {
				return cards[i].GetDesign() < cards[j].GetDesign()
			}
			return cards[i].GetValue() > cards[j].GetValue()
		}
		ei, ej := germanSoloEffectiveSuit(cards[i], trump), germanSoloEffectiveSuit(cards[j], trump)
		if ei != ej {
			return ei < ej
		}
		return germanSoloCardStrength(cards[i], trump) > germanSoloCardStrength(cards[j], trump)
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// germanSoloBidName ビッドの表示名を返す。
func germanSoloBidName(bid GermanSoloBid) string {
	switch bid {
	case GermanSoloBidMussfrage:
		return "mussfrage"
	case GermanSoloBidFrage:
		return "frage"
	case GermanSoloBidSolo:
		return "solo"
	case GermanSoloBidTout:
		return "tout"
	default:
		return "pass"
	}
}

// germanSoloSuitName スートの表示名を返す。
func germanSoloSuitName(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "spades"
	case CardDesignClover:
		return "clubs"
	case CardDesignHeart:
		return "hearts"
	case CardDesignDiamond:
		return "diamonds"
	default:
		return "-"
	}
}

// germanSoloOutcomeName 結果の表示名を返す。
func germanSoloOutcomeName(o GermanSoloOutcome) string {
	switch o {
	case GermanSoloOutcomeMade:
		return "made"
	case GermanSoloOutcomeFailed:
		return "failed"
	default:
		return "-"
	}
}

// germanSoloValidSuit suit が有効なスート (1..4) か。
func germanSoloValidSuit(suit int) bool {
	return suit >= CardDesignSpade && suit <= CardDesignDiamond
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *GermanSolo) indexOfPlayerInTrick(playerIdx int) int {
	return indexOfPlayerInTrick(g.currentTrick, playerIdx)
}

// --- CPU AI (play) ---

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *GermanSolo) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == GermanSoloCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 陣営 (ジャーマン・ソロ vs 連合) を意識した戦略プレイ。
func (g *GermanSolo) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	trump := g.trumpSuit

	// リード: ジャーマン・ソロは強い札で主導、連合は弱い札で温存する。
	if len(g.currentTrick) == 0 {
		if playerIdx == g.declarerIdx {
			return g.maxByStrength(player, valid)
		}
		return g.minByStrength(player, valid)
	}

	leadEff := germanSoloEffectiveSuit(g.currentTrick[0].Card, trump)
	winnerIdx := g.trickWinner()
	topStr := germanSoloCardStrength(g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card, trump)
	partnerWinning := g.sameSide(winnerIdx, playerIdx)

	var follows []int
	for _, idx := range valid {
		if germanSoloEffectiveSuit(player.GetCard(idx), trump) == leadEff {
			follows = append(follows, idx)
		}
	}

	if len(follows) == 0 {
		// ボイド: 最も弱い札を捨てる。
		return g.minByStrength(player, valid)
	}

	winners := filterIndices(follows, func(idx int) bool {
		return germanSoloCardStrength(player.GetCard(idx), trump) > topStr
	})

	if partnerWinning {
		// 味方が勝っている: 上書きせず最も弱い札を出す。
		return g.minByStrength(player, follows)
	}
	// 相手が勝っている: 勝てるなら最小限の勝ち札で取りに行く。
	if len(winners) > 0 {
		return g.minByStrength(player, winners)
	}
	return g.minByStrength(player, follows)
}

// minByStrength 強さが最小となるインデックスを返す。
func (g *GermanSolo) minByStrength(player *GermanSoloPlayer, indices []int) int {
	best := indices[0]
	bestScore := germanSoloCardStrength(player.GetCard(best), g.trumpSuit)
	for _, idx := range indices[1:] {
		if s := germanSoloCardStrength(player.GetCard(idx), g.trumpSuit); s < bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// maxByStrength 強さが最大となるインデックスを返す。
func (g *GermanSolo) maxByStrength(player *GermanSoloPlayer, indices []int) int {
	best := indices[0]
	bestScore := germanSoloCardStrength(player.GetCard(best), g.trumpSuit)
	for _, idx := range indices[1:] {
		if s := germanSoloCardStrength(player.GetCard(idx), g.trumpSuit); s > bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨アクションを返す。
func (g *GermanSolo) GetHint() *GermanSoloHint {
	human := findHumanIdx(g.players)
	if human < 0 {
		return nil
	}
	switch g.phase {
	case GermanSoloPhaseBid:
		if g.currentBidderIdx != human {
			return nil
		}
		bid := g.cpuChooseBidForHint(human)
		return &GermanSoloHint{Reason: germanSoloBidHintReason(bid)}
	case GermanSoloPhaseAceCall:
		if g.declarerIdx != human {
			return nil
		}
		if suit := g.cpuPickAceSuit(human); germanSoloValidSuit(suit) {
			return &GermanSoloHint{Reason: "call_ace", SuitHint: suit}
		}
		return nil
	case GermanSoloPhasePlay:
		if g.currentPlayerIdx != human {
			return nil
		}
		valid := g.getValidPlayIndices(human)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuPlaySmart(human, valid)
		return &GermanSoloHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
	}
	return nil
}

// cpuChooseBidForHint ヒント用にビッド推奨を計算する (Easy 難易度でも強度から推奨する)。
//
// **CPU と同じ関数を読む。** ヒントだけ別の条件を書くと、助言に従った人間が
// CPU なら降りている契約を掴む。
func (g *GermanSolo) cpuChooseBidForHint(playerIdx int) GermanSoloBid {
	return g.bidForStrength(playerIdx)
}

// germanSoloBidHintReason ビッド推奨に対応するヒント理由キーを返す。
func germanSoloBidHintReason(bid GermanSoloBid) string {
	switch bid {
	case GermanSoloBidTout:
		return "bid_tout"
	case GermanSoloBidSolo:
		return "bid_solo"
	case GermanSoloBidFrage:
		return "bid_frage"
	default:
		return "bid_pass"
	}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *GermanSolo) playHintReason(playerIdx, chosenIdx int) string {
	trump := g.trumpSuit
	if len(g.currentTrick) == 0 {
		if playerIdx == g.declarerIdx {
			return "lead_high"
		}
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadEff := germanSoloEffectiveSuit(g.currentTrick[0].Card, trump)
	if germanSoloEffectiveSuit(card, trump) != leadEff {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	topStr := germanSoloCardStrength(g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card, trump)
	if germanSoloCardStrength(card, trump) > topStr {
		return "follow_win"
	}
	if g.sameSide(winnerIdx, playerIdx) {
		return "give_partner"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *GermanSolo) GetPhase() GermanSoloPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *GermanSolo) SetPhase(phase GermanSoloPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *GermanSolo) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *GermanSolo) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *GermanSolo) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *GermanSolo) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *GermanSolo) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *GermanSolo) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *GermanSolo) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *GermanSolo) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *GermanSolo) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *GermanSolo) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *GermanSolo) GetDealerIdx() int { return g.dealerIdx }

// GetForehandIdx forehand インデックス取得
func (g *GermanSolo) GetForehandIdx() int { return g.forehandIdx }

// GetDeclarerIdx ジャーマン・ソロインデックス取得 (-1=未確定)
func (g *GermanSolo) GetDeclarerIdx() int { return g.declarerIdx }

// GetCallableAceSuitsForTest はテスト用に席 idx が呼べるエースのスートを返す。
// GetCallableAceSuits と違いフェーズを問わない。
func (g *GermanSolo) GetCallableAceSuitsForTest(idx int) []int {
	return g.callableAceSuits(idx)
}

// SetPartnerForTest はテスト用に味方の席と公開状態を設定する。
func (g *GermanSolo) SetPartnerForTest(idx int, revealed bool) {
	g.partnerIdx = idx
	g.partnerRevealed = revealed
}

// SetPlaysAloneForTest はテスト用に単独プレイを設定する。
func (g *GermanSolo) SetPlaysAloneForTest(v bool) { g.playsAlone = v }

// SetCalledAceSuitForTest はテスト用に呼ばれたエースのスートを設定する。
func (g *GermanSolo) SetCalledAceSuitForTest(suit int) { g.calledAceSuit = suit }

// CpuBidForTest はテスト用に席 idx に 1 回ビッドさせる (手番を問わない)。
func (g *GermanSolo) CpuBidForTest(idx int) {
	bid := g.cpuChooseBid(idx)
	trump := -1
	if bid != GermanSoloBidNone {
		trump = g.cpuChooseTrump(idx)
	}
	g.bids[idx] = bid
	g.bidActed[idx] = true
	g.bidTrump[idx] = trump
}

// StartPlayForTest はテスト用にプレイフェーズを開始する。
func (g *GermanSolo) StartPlayForTest() { g.startPlay() }

// SetDeclarerIdx ジャーマン・ソロインデックス設定 (テスト用)
func (g *GermanSolo) SetDeclarerIdx(idx int) { g.declarerIdx = idx }

// GetWinningBid 確定ビッド取得
func (g *GermanSolo) GetWinningBid() GermanSoloBid { return g.winningBid }

// GetBiddableBids は今の競り状況で人間が宣言できるビッドを返す (画面の選択肢)。
//
// **サーバが弾く選択肢を画面に出さない。** 既に Solo が出ている卓で Frage の
// ボタンが押せると、押した人間だけが「上回る宣言が必要です」を読むことになる。
func (g *GermanSolo) GetBiddableBids() []int {
	if g.phase != GermanSoloPhaseBid {
		return nil
	}
	out := make([]int, 0, 3)
	for _, b := range []GermanSoloBid{GermanSoloBidFrage, GermanSoloBidSolo, GermanSoloBidTout} {
		if g.isBidLegal(b) {
			out = append(out, int(b))
		}
	}
	return out
}

// GetHighestBid は**競り中の**最高宣言を返す。
//
// GetWinningBid は落札が確定するまで None のままなので、競っている最中の
// 画面がそれを出すと「最高ビッド: -」と表示されたまま Frage が弾かれる。
// 何を上回ればよいのかが読めないので、競り中はこちらを出す。
func (g *GermanSolo) GetHighestBid() GermanSoloBid { return g.highestBid() }

// SetWinningBid 確定ビッド設定 (テスト用)
func (g *GermanSolo) SetWinningBid(b GermanSoloBid) { g.winningBid = b }

// GetTrumpSuit 切り札スート取得 (-1=未確定, 1..4)
func (g *GermanSolo) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *GermanSolo) SetTrumpSuit(s int) { g.trumpSuit = s }

// GetCurrentBidderIdx 現在のビッド手番インデックス取得
func (g *GermanSolo) GetCurrentBidderIdx() int { return g.currentBidderIdx }

// GetPlayerScores プレイヤー別累積点取得
func (g *GermanSolo) GetPlayerScores() [GermanSoloPlayerCnt]int { return g.playerScores }

// SetPlayerScores プレイヤー別累積点設定 (テスト用)
func (g *GermanSolo) SetPlayerScores(s [GermanSoloPlayerCnt]int) { g.playerScores = s }

// GetOutcome 直近ディールの結果取得
func (g *GermanSolo) GetOutcome() GermanSoloOutcome { return g.outcome }

// GetResult 人間視点のマッチ結果取得
func (g *GermanSolo) GetResult() GermanSoloResult { return g.result }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *GermanSolo) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *GermanSolo) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayerCnt プレイヤー数取得
func (g *GermanSolo) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *GermanSolo) GetPlayer(i int) *GermanSoloPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番 (プレイ) が人間か。
func (g *GermanSolo) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// IsHumanBidTurn 現在のビッド手番が人間か。
func (g *GermanSolo) IsHumanBidTurn() bool {
	if g.phase != GermanSoloPhaseBid {
		return false
	}
	if g.currentBidderIdx < 0 || g.currentBidderIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentBidderIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *GermanSolo) GetConfig() GermanSoloConfig { return g.config }

// SetConfig 設定変更
func (g *GermanSolo) SetConfig(cfg GermanSoloConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *GermanSolo) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != GermanSoloPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// germanSoloJSON is the JSON wire format for GermanSolo.
type germanSoloJSON struct {
	TrumpCards       *TrumpCards                        `json:"tc"`
	Players          []*GermanSoloPlayer                `json:"ps"`
	Config           GermanSoloConfig                   `json:"cf"`
	Phase            GermanSoloPhase                    `json:"ph"`
	RoundNumber      int                                `json:"rn"`
	TrickNumber      int                                `json:"tn"`
	CurrentPlayerIdx int                                `json:"ci"`
	CurrentTrick     []*TrickCard                       `json:"ct"`
	LeadPlayerIdx    int                                `json:"li"`
	DealerIdx        int                                `json:"di"`
	ForehandIdx      int                                `json:"fh"`
	DeclarerIdx      int                                `json:"om"`
	WinningBid       GermanSoloBid                      `json:"wb"`
	TrumpSuit        int                                `json:"ts"`
	CurrentBidderIdx int                                `json:"cbi"`
	Bids             [GermanSoloPlayerCnt]GermanSoloBid `json:"bd"`
	BidTrump         [GermanSoloPlayerCnt]int           `json:"bt"`
	BidActed         [GermanSoloPlayerCnt]bool          `json:"ba"`
	PlayerScores     [GermanSoloPlayerCnt]int           `json:"sc"`
	LastTrickWinner  int                                `json:"lt"`
	// エース呼びは盤面の一部。落とすと復元後に**味方が席 0 になる** (ゼロ値)。
	// 呼ばれたエースも単独プレイの区別も消えるので、勝敗の集計が静かに変わる。
	CalledAceSuit   int               `json:"ca"`
	PartnerIdx      int               `json:"pi"`
	PartnerRevealed bool              `json:"pr"`
	PlaysAlone      bool              `json:"pa"`
	TrickResolved   bool              `json:"tr"`
	Outcome         GermanSoloOutcome `json:"oc"`
	Result          GermanSoloResult  `json:"rs"`
	Scored          bool              `json:"sd"`
	GameEndFlag     bool              `json:"ge"`
	WinnerPlayer    int               `json:"wp"`
	ActionLog       []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *GermanSolo) MarshalJSON() ([]byte, error) {
	return json.Marshal(germanSoloJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		RoundNumber:      g.roundNumber,
		TrickNumber:      g.trickNumber,
		CurrentPlayerIdx: g.currentPlayerIdx,
		CurrentTrick:     g.currentTrick,
		LeadPlayerIdx:    g.leadPlayerIdx,
		DealerIdx:        g.dealerIdx,
		ForehandIdx:      g.forehandIdx,
		DeclarerIdx:      g.declarerIdx,
		WinningBid:       g.winningBid,
		TrumpSuit:        g.trumpSuit,
		CurrentBidderIdx: g.currentBidderIdx,
		Bids:             g.bids,
		BidTrump:         g.bidTrump,
		BidActed:         g.bidActed,
		PlayerScores:     g.playerScores,
		LastTrickWinner:  g.lastTrickWinner,
		CalledAceSuit:    g.calledAceSuit,
		PartnerIdx:       g.partnerIdx,
		PartnerRevealed:  g.partnerRevealed,
		PlaysAlone:       g.playsAlone,
		TrickResolved:    g.trickResolved,
		Outcome:          g.outcome,
		Result:           g.result,
		Scored:           g.scored,
		GameEndFlag:      g.gameEndFlag,
		WinnerPlayer:     g.winnerPlayer,
		ActionLog:        g.actionLog,
	})
}

// germanSoloMaxSliceLen caps slice sizes during deserialisation.
const germanSoloMaxSliceLen = 5000

// errGermanSoloOversized is the single sentinel error for oversized input arrays.
var errGermanSoloOversized = errors.New("germansolo: input array exceeds maximum allowed size")

// errGermanSoloInvalidPlayers is returned when restored state lacks exactly GermanSoloPlayerCnt players.
var errGermanSoloInvalidPlayers = errors.New("germansolo: invalid player count")

// errGermanSoloInvalidTrick is returned when a restored trick card or its card is nil / out of range.
var errGermanSoloInvalidTrick = errors.New("germansolo: invalid trick card")

// errGermanSoloInvalidIndex is returned when a restored index field is out of range.
var errGermanSoloInvalidIndex = errors.New("germansolo: index field out of range")

// errGermanSoloInvalidPhase is returned when a restored phase is out of range.
var errGermanSoloInvalidPhase = errors.New("germansolo: phase out of range")

// errGermanSoloInvalidBid is returned when a restored bid value is out of range.
var errGermanSoloInvalidBid = errors.New("germansolo: bid value out of range")

// errGermanSoloInvalidTrump is returned when a restored trump suit is out of range.
var errGermanSoloInvalidTrump = errors.New("germansolo: trump suit out of range")

// errGermanSoloInvalidOutcome is returned when a restored outcome or result value is out of range.
var errGermanSoloInvalidOutcome = errors.New("germansolo: outcome/result value out of range")

// germanSoloInRange reports whether v is in [0, GermanSoloPlayerCnt).
func germanSoloInRange(v int) bool { return v >= 0 && v < GermanSoloPlayerCnt }

// germanSoloInRangeOrUnset reports whether v is -1 (unset) or in [0, GermanSoloPlayerCnt).
func germanSoloInRangeOrUnset(v int) bool { return v == -1 || germanSoloInRange(v) }

// germanSoloValidBid reports whether b is a defined bid value.
func germanSoloValidBid(b GermanSoloBid) bool {
	return b >= GermanSoloBidNone && b <= GermanSoloBidTout
}

// germanSoloTrumpInRangeOrUnset reports whether s is -1 (unset) or a valid suit (1..4).
func germanSoloTrumpInRangeOrUnset(s int) bool { return s == -1 || germanSoloValidSuit(s) }

// UnmarshalJSON implements json.Unmarshaler.
func (g *GermanSolo) UnmarshalJSON(data []byte) error {
	var j germanSoloJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > germanSoloMaxSliceLen || len(j.CurrentTrick) > germanSoloMaxSliceLen ||
		len(j.ActionLog) > germanSoloMaxSliceLen {
		return errGermanSoloOversized
	}
	if len(j.Players) != GermanSoloPlayerCnt {
		return errGermanSoloInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errGermanSoloInvalidPlayers
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return errGermanSoloInvalidTrick
		}
		if !germanSoloInRange(tc.PlayerIdx) {
			return errGermanSoloInvalidTrick
		}
	}
	// 範囲必須のインデックス [0, PlayerCnt)。
	if !germanSoloInRange(j.CurrentPlayerIdx) || !germanSoloInRange(j.DealerIdx) ||
		!germanSoloInRange(j.ForehandIdx) || !germanSoloInRange(j.CurrentBidderIdx) {
		return errGermanSoloInvalidIndex
	}
	// -1 (未設定) 許容のインデックス。
	if !germanSoloInRangeOrUnset(j.LeadPlayerIdx) || !germanSoloInRangeOrUnset(j.DeclarerIdx) ||
		!germanSoloInRangeOrUnset(j.LastTrickWinner) || !germanSoloInRangeOrUnset(j.WinnerPlayer) {
		return errGermanSoloInvalidIndex
	}
	// フェーズ依存の厳格化: play 以降は germanSolo・lead・trump が確定していなければ
	// 後続処理で g.players[-1] / g.trumpSuit を参照して panic するため確定を要求する。
	if j.Phase >= GermanSoloPhasePlay {
		if !germanSoloInRange(j.DeclarerIdx) || !germanSoloInRange(j.LeadPlayerIdx) {
			return errGermanSoloInvalidIndex
		}
		if !germanSoloValidSuit(j.TrumpSuit) {
			return errGermanSoloInvalidTrump
		}
	}
	if int(j.Phase) < GermanSoloPhaseMin || int(j.Phase) > GermanSoloPhaseMax {
		return errGermanSoloInvalidPhase
	}
	if !germanSoloValidBid(j.WinningBid) {
		return errGermanSoloInvalidBid
	}
	for _, b := range j.Bids {
		if !germanSoloValidBid(b) {
			return errGermanSoloInvalidBid
		}
	}
	// エース呼びも他のインデックス同様に検査する。**壊れた値を素通しすると
	// germanSoloSideOf が誤った席を味方として数え、勝敗が変わる。**
	//
	// ただし 0 は「未指定」と「席 0 / スート 0」の区別が付かない ——
	// JSON にフィールドが無ければ Go は 0 を入れるので、素直に信じると
	// **席 0 が誰の味方でもないのに落札者側として数えられる**。
	// 検査は範囲だけにして、整合は下の germanSoloNormaliseAceCall で取る。
	if j.CalledAceSuit != 0 && !germanSoloTrumpInRangeOrUnset(j.CalledAceSuit) {
		return errGermanSoloInvalidTrump
	}
	if !germanSoloInRangeOrUnset(j.PartnerIdx) {
		return errGermanSoloInvalidIndex
	}
	if !germanSoloTrumpInRangeOrUnset(j.TrumpSuit) {
		return errGermanSoloInvalidTrump
	}
	// bidTrump は -1/0 (未設定) または有効スート (1..4)。範囲外のみ拒否する
	// (finalizeAuction は無効な bidTrump を自動で選び直すため厳格化は不要)。
	for _, t := range j.BidTrump {
		if t < -1 || t > CardDesignDiamond {
			return errGermanSoloInvalidTrump
		}
	}
	if j.Outcome < GermanSoloOutcomeNone || j.Outcome > GermanSoloOutcomeFailed {
		return errGermanSoloInvalidOutcome
	}
	if j.Result < GermanSoloResultLose || j.Result > GermanSoloResultWin {
		return errGermanSoloInvalidOutcome
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newGermanSoloDeck()
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*TrickCard, 0)
	}
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.forehandIdx = j.ForehandIdx
	g.declarerIdx = j.DeclarerIdx
	g.winningBid = j.WinningBid
	g.trumpSuit = j.TrumpSuit
	g.currentBidderIdx = j.CurrentBidderIdx
	g.bids = j.Bids
	g.bidTrump = j.BidTrump
	g.bidActed = j.BidActed
	g.playerScores = j.PlayerScores
	g.lastTrickWinner = j.LastTrickWinner
	g.calledAceSuit, g.partnerIdx, g.partnerRevealed, g.playsAlone =
		germanSoloNormaliseAceCall(j.CalledAceSuit, j.PartnerIdx, j.PartnerRevealed, j.PlaysAlone)
	g.trickResolved = j.TrickResolved
	g.outcome = j.Outcome
	g.result = j.Result
	g.scored = j.Scored
	g.gameEndFlag = j.GameEndFlag
	g.winnerPlayer = j.WinnerPlayer
	g.actionLog = j.ActionLog
	return nil
}

// germanSoloNormaliseAceCall は復元したエース呼びの状態を整合させる。
//
// **0 は「未指定」と区別が付かない。** JSON にフィールドが無ければ Go は 0 を
// 入れるので、そのまま信じると calledAceSuit=0 (どの札とも一致しないので
// 味方が永久に公開されない) と partnerIdx=0 (席 0 が味方として数えられ、
// 勝敗の集計が変わる) が同時に起きる。
//
// エースが呼ばれていない盤面に味方は存在しない、という不変条件で揃える。
func germanSoloNormaliseAceCall(suit, partner int, revealed, playsAlone bool) (int, int, bool, bool) {
	if playsAlone {
		// 単独プレイに味方はいない。
		return -1, -1, false, true
	}
	if !germanSoloValidSuit(suit) {
		// エースが呼ばれていないので、味方も公開状態も無い。
		return -1, -1, false, false
	}
	if partner < 0 || partner >= GermanSoloPlayerCnt {
		return suit, -1, false, false
	}
	return suit, partner, revealed, false
}
