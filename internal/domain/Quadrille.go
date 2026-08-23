//go:build !js || !wasm || extra4

// Package domain カドリール (Quadrille) のドメインモデル。
//
// Quadrille は 17 世紀スペイン発祥の 3 人用トリックテイキングゲームで、1 人の「カドリール (Quadrille)」が
// 残り 2 人の連合 (coalition) と対戦する、ソリスト対多数系トリックゲームの祖先。
//
// デッキ: 40 枚 = 標準 52 枚から 8・9・10 を除いたもの (各スート A,2..7,J,Q,K)。
// 各ディールで 9 枚ずつ配り (27 枚)、残り 13 枚は未使用のストック (交換・引き無し=簡略化)。
//
// ビッド: 各プレイヤーは順に Pass / Entrar (カドリールを引き受ける) / Solo (より強い Entrar) を宣言。
// 最高ビッドの宣言者がカドリールとなり、切り札スートを選ぶ。全員パスならディーラーが強制的にカドリール
// となる (切り札はディーラーの最良スートを自動選択)。残り 2 人は連合。ビッド序列は Solo > Entrar > Pass、
// 同値は手番順で決着。
//
// 切り札グループとランク (この実装の中核・強い順):
//
//  1. Spadille = ♠A (常に第1切り札、スペードが切り札でなくても)
//  2. Manille  = 切り札スートの 7 (常に第2)
//  3. Basto    = ♣A (常に第3、クラブが切り札でなくても)
//  4. Punto    = 切り札スートの A (第4、切り札が赤=ハート/ダイヤのときのみ存在)
//  5. 切り札スートの残り札 降順: K > Q > J > 6 > 5 > 4 > 3 > 2
//
// 平札 (非切り札) のランク 降順: K > Q > J > 7 > 6 > 5 > 4 > 3 > 2 > A (A は最弱)。
// ♠A と ♣A は常に切り札で平札には現れない。切り札の 7 は常に Manille。
//
// トリック: 9 トリック、マストフォロー。フォロー判定上、♠A・♣A・切り札の 7・切り札スートの全札は
// すべて「切り札」として扱う。切り札が場に出れば最強切り札が、無ければリードスートの最強札が勝つ。
//
// ディール結果 (カドリールのトリック数 vs 各相手): Sacar=カドリールが各相手より厳密に多い (カドリール勝ち)、
// Puesta=最多に並ばれた (相手のいずれかが同数で誰にも超えられていない、カドリール負け)、
// Codille=相手のいずれかが厳密に多い (連合勝ち・倍額)。
//
// 得点 (累積、TargetRounds ディール): Sacar でカドリール +2 / 各相手 -1、Puesta でカドリール -2 / 各相手 +1、
// Codille でカドリール -4 / 各相手 +2。TargetRounds ディール後、累積点最上位が勝者。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"sort"
)

// QuadrillePlayerCnt プレイヤー数 (人間 1 + CPU 3)
const QuadrillePlayerCnt = 4

// QuadrilleHandSize 各プレイヤーの配り札枚数。
//
// **40 枚を 4 人で配り切る。** クローン元のオンブルは 3 人 × 9 枚で 13 枚を
// 配り残していたが、こちらは 4 × 10 = 40 でちょうど尽きる。
const QuadrilleHandSize = 10

// QuadrilleTrickCount 1 ディールのトリック数
const QuadrilleTrickCount = 9

// QuadrilleDeckSize デッキ枚数 (52 - 8,9,10)
const QuadrilleDeckSize = 40

// QuadrilleWinRounds マッチを構成するディール数 (既定)
const QuadrilleWinRounds = 5

// QuadrilleBid ビッド宣言
type QuadrilleBid int

// Quadrille のビッド定数 (数値が大きいほど高い宣言)
const (
	// QuadrilleBidNone 未宣言 (パス相当の初期値)
	QuadrilleBidNone QuadrilleBid = 0
	// QuadrilleBidEntrar entrar — カドリールを引き受ける宣言。
	QuadrilleBidEntrar QuadrilleBid = 1
	// QuadrilleBidSolo solo — より強い Entrar。
	QuadrilleBidSolo QuadrilleBid = 2
)

// QuadrillePhase ゲームフェーズ
type QuadrillePhase int

// Quadrille のフェーズ定数
const (
	// QuadrillePhaseBid ビッド (auction) フェーズ
	QuadrillePhaseBid QuadrillePhase = 0
	// QuadrillePhaseKingCall 王呼びフェーズ。**オンブルには無い、この
	// ゲーム固有のフェーズ。** 落札者が手札に無い王を 1 枚指名し、
	// その王を持つ席が一時的な味方になる。
	QuadrillePhaseKingCall QuadrillePhase = 1
	// QuadrillePhasePlay トリックプレイフェーズ
	QuadrillePhasePlay QuadrillePhase = 2
	// QuadrillePhaseTrickEnd トリック終了フェーズ
	QuadrillePhaseTrickEnd QuadrillePhase = 3
	// QuadrillePhaseRoundEnd ディール終了フェーズ
	QuadrillePhaseRoundEnd QuadrillePhase = 4
	// QuadrillePhaseGameEnd ゲーム終了フェーズ
	QuadrillePhaseGameEnd QuadrillePhase = 5
)

// QuadrillePhaseMin フェーズ下限 (検証用)
const QuadrillePhaseMin = int(QuadrillePhaseBid)

// QuadrillePhaseMax フェーズ上限 (検証用)
const QuadrillePhaseMax = int(QuadrillePhaseGameEnd)

// QuadrilleOutcome ディール結果
type QuadrilleOutcome int

// Quadrille のディール結果定数
const (
	// QuadrilleOutcomeNone 未確定
	QuadrilleOutcomeNone QuadrilleOutcome = 0
	// QuadrilleOutcomeSacar カドリールが各相手より厳密に多く取り勝利
	QuadrilleOutcomeSacar QuadrilleOutcome = 1
	// QuadrilleOutcomePuesta カドリールが最多に並ばれ敗北 (軽い罰)
	QuadrilleOutcomePuesta QuadrilleOutcome = 2
	// QuadrilleOutcomeCodille 相手が厳密に多く取り連合勝ち (倍の罰)
	QuadrilleOutcomeCodille QuadrilleOutcome = 3
)

// QuadrilleResult 人間視点のマッチ結果
type QuadrilleResult int

// Quadrille のマッチ結果定数
const (
	// QuadrilleResultLose 敗北
	QuadrilleResultLose QuadrilleResult = -1
	// QuadrilleResultNone 未確定 / 引き分け
	QuadrilleResultNone QuadrilleResult = 0
	// QuadrilleResultWin 勝利
	QuadrilleResultWin QuadrilleResult = 1
)

// Quadrille のスコア増減表 (1 ディールで移動する点)
const (
	// quadrilleScoreSacar Sacar でのカドリール側の増分 (各相手は -1)
	quadrilleScoreSacar = 2
	// quadrilleScorePuesta Puesta でのカドリール側の減分 (各相手は +1)
	quadrilleScorePuesta = 2
	// quadrilleScoreCodille Codille でのカドリール側の減分 (各相手は +2)
	quadrilleScoreCodille = 4
)

// QuadrilleHint ヒント情報
type QuadrilleHint struct {
	CardIndices []int  // 推奨カードインデックス (play フェーズ)
	Reason      string // ヒント理由キー
}

// Quadrille カドリールのゲームクラス
type Quadrille struct {
	trumpCards       *TrumpCards
	players          []*QuadrillePlayer
	config           QuadrilleConfig
	phase            QuadrillePhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	forehandIdx      int                              // ディーラーの左隣 (ビッド開始)
	quadrilleIdx     int                              // カドリール (-1=未確定)
	winningBid       QuadrilleBid                     // 確定したビッド (カドリールの宣言)
	trumpSuit        int                              // 切り札スート (-1=未確定, 1..4)
	currentBidderIdx int                              // 現在ビッド中のプレイヤー (bid フェーズ)
	bids             [QuadrillePlayerCnt]QuadrilleBid // 各プレイヤーの宣言
	bidTrump         [QuadrillePlayerCnt]int          // 各プレイヤーが宣言時に選んだ切り札 (-1=なし)
	bidActed         [QuadrillePlayerCnt]bool         // 各プレイヤーが宣言済みか
	playerScores     [QuadrillePlayerCnt]int          // 累積ゲーム点
	lastTrickWinner  int                              // 最終トリック勝者 (-1=未確定)

	// 王呼び。calledKingSuit は指名された王のスート (-1=未指名)。
	// **呼び声は卓で聞こえるので王自体は公開情報**だが、誰が持っているかは
	// その王が場に出るまで伏せる —— そこがこのゲームの緊張感そのもの。
	// roiSeul は落札者が王 4 枚を全部持っていた場合の単独プレイ。
	calledKingSuit  int
	partnerIdx      int // 呼ばれた王の持ち主 (-1=未確定 / 単独)
	partnerRevealed bool
	roiSeul         bool
	outcome         QuadrilleOutcome // 直近ディールの結果
	result          QuadrilleResult  // 人間視点のマッチ結果
	scored          bool             // 当該ディールの得点計算済みか (RoundEnd 突入時に一度だけ)
	gameEndFlag     bool
	winnerPlayer    int // -1=未確定
	actionLogBase
}

// NewQuadrille コンストラクタ
func NewQuadrille(trumpCards *TrumpCards, players []*QuadrillePlayer, config QuadrilleConfig) *Quadrille {
	return &Quadrille{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		winnerPlayer:    -1,
		lastTrickWinner: -1,
		quadrilleIdx:    -1,
		trumpSuit:       -1,
	}
}

// NewDefaultQuadrille 標準の 3 人構成 (人間 1, CPU 2) と既定設定で生成する。
func NewDefaultQuadrille() *Quadrille {
	players := make([]*QuadrillePlayer, QuadrillePlayerCnt)
	players[0] = NewQuadrillePlayer(true)
	for i := 1; i < QuadrillePlayerCnt; i++ {
		players[i] = NewQuadrillePlayer(false)
	}
	return NewQuadrille(newQuadrilleDeck(), players, DefaultQuadrilleConfig())
}

// newQuadrilleDeck Quadrille 用 40 枚デッキを生成する。標準 52 枚 (NewTrumpCards(0)) から 8,9,10 を除外する。
// domain パッケージ内なので TrumpCards の内部 deck を直接フィルタできる。build-tag 無しの
// NewTrumpCards は extra ワーカーからも到達可能。
func newQuadrilleDeck() *TrumpCards {
	full := NewTrumpCards(0)
	t := new(TrumpCards)
	t.deck = make([]*Card, 0, QuadrilleDeckSize)
	for _, c := range full.deck {
		v := c.GetValue()
		if v == 8 || v == 9 || v == 10 {
			continue
		}
		t.deck = append(t.deck, NewCard(c.GetDesign(), v, false))
	}
	t.deckCnt = len(t.deck)
	t.deckInit()
	return t
}

// Reset ゲーム初期化
func (g *Quadrille) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.playerScores = [QuadrillePlayerCnt]int{}
	g.result = QuadrilleResultNone
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のディールを開始する
func (g *Quadrille) NextRound() {
	if g.phase != QuadrillePhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % QuadrillePlayerCnt
	g.startRound()
}

// startRound 手札を配り、ビッドフェーズを開始する。
func (g *Quadrille) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.lastTrickWinner = -1
	g.quadrilleIdx = -1
	g.winningBid = QuadrilleBidNone
	g.trumpSuit = -1
	g.outcome = QuadrilleOutcomeNone
	g.scored = false
	// **王呼びはディールごとにやり直す。** 持ち越すと前のディールの相方が
	// そのまま味方として残り、勝敗の集計が静かにずれる。
	g.calledKingSuit = -1
	g.partnerIdx = -1
	g.partnerRevealed = false
	g.roiSeul = false
	g.bids = [QuadrillePlayerCnt]QuadrilleBid{}
	g.bidActed = [QuadrillePlayerCnt]bool{}
	for i := range g.bidTrump {
		g.bidTrump[i] = -1
	}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()

	g.forehandIdx = (g.dealerIdx + 1) % QuadrillePlayerCnt
	g.sortAllHands()

	g.currentBidderIdx = g.forehandIdx
	g.phase = QuadrillePhaseBid
}

// deal 各プレイヤーへ QuadrilleHandSize 枚を配る。残り 13 枚はストックとして未使用のまま残す。
func (g *Quadrille) deal() {
	for i := 0; i < QuadrilleHandSize; i++ {
		for j := 0; j < QuadrillePlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % QuadrillePlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
}

// --- Bidding ---

// PlayerBid 人間がビッドする。bid は Pass(None)/Entrar/Solo。trumpSuit は Entrar/Solo 時の切り札 (1..4)。
func (g *Quadrille) PlayerBid(bid QuadrilleBid, trumpSuit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != QuadrillePhaseBid {
		return ErrWrongPhase
	}
	if !g.players[g.currentBidderIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if !g.isBidLegal(bid) {
		return NewDomainError(ErrInvalidPlay, "現在の最高ビッドを上回る宣言が必要です")
	}
	if bid != QuadrilleBidNone && !quadrilleValidSuit(trumpSuit) {
		return NewDomainError(ErrInvalidPlay, "切り札スートを選んでください (1..4)")
	}
	g.applyBid(g.currentBidderIdx, bid, trumpSuit)
	return nil
}

// CpuBid 現在のビッド手番が CPU の場合に 1 回ビッドする。
func (g *Quadrille) CpuBid() {
	if g.gameEndFlag || g.phase != QuadrillePhaseBid {
		return
	}
	idx := g.currentBidderIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	bid := g.cpuChooseBid(idx)
	trump := -1
	if bid != QuadrilleBidNone {
		trump = g.cpuChooseTrump(idx)
	}
	g.applyBid(idx, bid, trump)
}

// isBidLegal bid が合法か (pass は常に可、entrar/solo は現最高ビッドを上回る必要)。
func (g *Quadrille) isBidLegal(bid QuadrilleBid) bool {
	if bid == QuadrilleBidNone {
		return true
	}
	if bid != QuadrilleBidEntrar && bid != QuadrilleBidSolo {
		return false
	}
	return bid > g.highestBid()
}

// highestBid これまでに宣言された最高ビッドを返す。
func (g *Quadrille) highestBid() QuadrilleBid {
	best := QuadrilleBidNone
	for _, b := range g.bids {
		if b > best {
			best = b
		}
	}
	return best
}

// applyBid ビッドを適用し、全員が宣言し終えたら auction を締める。
func (g *Quadrille) applyBid(playerIdx int, bid QuadrilleBid, trumpSuit int) {
	g.bids[playerIdx] = bid
	g.bidActed[playerIdx] = true
	if bid == QuadrilleBidNone {
		g.bidTrump[playerIdx] = -1
		g.appendLog(playerIdx, "bid_pass",
			fmt.Sprintf("%s passes", playerName(g.players, playerIdx)), nil)
	} else {
		g.bidTrump[playerIdx] = trumpSuit
		g.appendLog(playerIdx, "bid",
			fmt.Sprintf("%s bids %s (trump %s)", playerName(g.players, playerIdx), quadrilleBidName(bid), quadrilleSuitName(trumpSuit)), nil)
	}

	if g.allBidsActed() {
		g.finalizeAuction()
		return
	}
	g.currentBidderIdx = g.nextBidder(playerIdx)
}

// allBidsActed 全員が宣言済みか。
func (g *Quadrille) allBidsActed() bool {
	for _, acted := range g.bidActed {
		if !acted {
			return false
		}
	}
	return true
}

// nextBidder playerIdx の次でまだ宣言していないプレイヤーを返す。
func (g *Quadrille) nextBidder(playerIdx int) int {
	for i := 1; i <= QuadrillePlayerCnt; i++ {
		cand := (playerIdx + i) % QuadrillePlayerCnt
		if !g.bidActed[cand] {
			return cand
		}
	}
	return playerIdx
}

// finalizeAuction 最高ビッドの宣言者をカドリールに確定し、プレイフェーズへ進む。
// 同値は forehand から時計回りで最初に宣言したプレイヤーが優先される。
// 全員パスならディーラーが強制的にカドリール (切り札はディーラーの最良スートを自動選択)。
func (g *Quadrille) finalizeAuction() {
	best := g.highestBid()
	quadrille := -1
	if best > QuadrilleBidNone {
		for i := 0; i < QuadrillePlayerCnt; i++ {
			cand := (g.forehandIdx + i) % QuadrillePlayerCnt
			if g.bids[cand] == best {
				quadrille = cand
				break
			}
		}
	}
	if quadrille < 0 {
		// 全員パス: ディーラーが強制的に Entrar を引き受け、切り札を自動選択する。
		quadrille = g.dealerIdx
		best = QuadrilleBidEntrar
		g.bidTrump[quadrille] = g.cpuChooseTrump(quadrille)
	}
	g.quadrilleIdx = quadrille
	g.winningBid = best
	g.trumpSuit = g.bidTrump[quadrille]
	if !quadrilleValidSuit(g.trumpSuit) {
		g.trumpSuit = g.cpuChooseTrump(quadrille)
	}
	g.appendLog(quadrille, "quadrille",
		fmt.Sprintf("%s is Quadrille with %s (trump %s)", playerName(g.players, quadrille), quadrilleBidName(best), quadrilleSuitName(g.trumpSuit)), nil)
	g.startKingCall()
}

// startKingCall は王呼びフェーズに入る。
//
// 落札者が王 4 枚を全部持っていたら呼ぶ相手がいないので **Roi seul**
// (単独プレイ) として、そのままプレイへ進む。
func (g *Quadrille) startKingCall() {
	if len(g.callableKingSuits(g.quadrilleIdx)) == 0 {
		g.roiSeul = true
		g.partnerIdx = -1
		g.appendLog(g.quadrilleIdx, "roi_seul",
			fmt.Sprintf("%s holds every king and plays alone (Roi seul)",
				playerName(g.players, g.quadrilleIdx)), nil)
		g.startPlay()
		return
	}
	g.currentPlayerIdx = g.quadrilleIdx
	g.phase = QuadrillePhaseKingCall
}

// callableKingSuits は落札者が**呼べる**王のスートを返す。
//
// 自分が持っている王は呼べない (呼んでも味方が増えない) ので除く。
func (g *Quadrille) callableKingSuits(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) {
		return nil
	}
	held := map[int]bool{}
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if c := p.GetCard(i); c != nil && c.GetValue() == quadrilleKingValue {
			held[c.GetDesign()] = true
		}
	}
	out := make([]int, 0, CardDesignMax)
	for suit := CardDesignSpade; suit <= CardDesignMax; suit++ {
		if !held[suit] {
			out = append(out, suit)
		}
	}
	return out
}

// DeclareKing は落札者が味方を呼ぶ王を指名する。
func (g *Quadrille) DeclareKing(playerIdx, suit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != QuadrillePhaseKingCall {
		return ErrWrongPhase
	}
	if playerIdx != g.quadrilleIdx {
		return NewDomainError(ErrInvalidPlay, "王を呼べるのは落札者だけです")
	}
	if !quadrilleValidSuit(suit) {
		return NewDomainError(ErrInvalidCard, "スートを選んでください (1..4)")
	}
	if !slices.Contains(g.callableKingSuits(playerIdx), suit) {
		// **自分が持っている王は呼べない。** 呼べてしまうと味方が増えず、
		// 単独プレイが「4 人卓の 1 対 3」ではなく黙って成立する。
		return NewDomainError(ErrInvalidPlay, "自分が持っている王は呼べません")
	}

	g.calledKingSuit = suit
	g.partnerIdx = g.findKingHolder(suit)
	g.partnerRevealed = false
	g.appendLog(playerIdx, "call_king",
		fmt.Sprintf("%s calls the King of %s", playerName(g.players, playerIdx), quadrilleSuitName(suit)), nil)
	g.startPlay()
	return nil
}

// CpuDeclareKing は CPU の落札者に王を呼ばせる。
//
// 一番長いスート**以外**の王を選ぶ: 自分の長いスートの王は自分で引き当てる
// 可能性が高く、味方を呼ぶ意味が薄い。
func (g *Quadrille) CpuDeclareKing() {
	if g.phase != QuadrillePhaseKingCall {
		return
	}
	idx := g.quadrilleIdx
	if idx < 0 || g.players[idx].GetIsHuman() {
		return
	}
	callable := g.callableKingSuits(idx)
	if len(callable) == 0 {
		return
	}
	counts := map[int]int{}
	p := g.players[idx]
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
	_ = g.DeclareKing(idx, best)
}

// findKingHolder は指定スートの王を持つ席を返す (-1=誰も持っていない)。
//
// 40 枚を配り切るので通常は必ず誰かが持っているが、テストが手札を差し替えた
// 盤面では見つからないことがある。その場合は単独プレイと同じ扱いになる。
func (g *Quadrille) findKingHolder(suit int) int {
	for i, p := range g.players {
		for j := 0; j < p.GetCardsSize(); j++ {
			c := p.GetCard(j)
			if c != nil && c.GetValue() == quadrilleKingValue && c.GetDesign() == suit {
				return i
			}
		}
	}
	return -1
}

// IsHumanKingCallTurn は人間の王呼び待ちかを返す。
func (g *Quadrille) IsHumanKingCallTurn() bool {
	return g.phase == QuadrillePhaseKingCall &&
		g.quadrilleIdx >= 0 && g.players[g.quadrilleIdx].GetIsHuman()
}

// GetCalledKingSuit は呼ばれた王のスートを返す (-1=未指名)。
func (g *Quadrille) GetCalledKingSuit() int { return g.calledKingSuit }

// GetPartnerIdx は味方の席を返す。**まだ王が場に出ていなければ -1。**
// 誰が味方かはその王が出るまで伏せる。
func (g *Quadrille) GetPartnerIdx() int {
	if !g.partnerRevealed {
		return -1
	}
	return g.partnerIdx
}

// IsRoiSeul は落札者が単独で戦っているかを返す。
func (g *Quadrille) IsRoiSeul() bool { return g.roiSeul }

// GetCallableKingSuits は落札者が呼べる王のスートを返す (画面の選択肢)。
func (g *Quadrille) GetCallableKingSuits() []int {
	if g.phase != QuadrillePhaseKingCall {
		return nil
	}
	return g.callableKingSuits(g.quadrilleIdx)
}

// quadrilleSideOf は席 idx が落札者側かを返す。
//
// **単独プレイでは落札者だけ。** 味方が伏せられている間も内部の判定には
// 実際の相方を使う —— 伏せているのは表示であって、勝敗の計算ではない。
func (g *Quadrille) quadrilleSideOf(idx int) bool {
	if idx == g.quadrilleIdx {
		return true
	}
	return !g.roiSeul && g.partnerIdx >= 0 && idx == g.partnerIdx
}

// cpuChooseBid CPU が手札強度からビッドを選ぶ。
func (g *Quadrille) cpuChooseBid(playerIdx int) QuadrilleBid {
	if g.config.CpuDifficulty == QuadrilleCpuDifficultyEasy {
		return QuadrilleBidNone
	}
	_, strength := g.handBestTrump(playerIdx)
	highest := g.highestBid()
	if strength >= 20 && QuadrilleBidSolo > highest {
		return QuadrilleBidSolo
	}
	if strength >= 13 && QuadrilleBidEntrar > highest {
		return QuadrilleBidEntrar
	}
	return QuadrilleBidNone
}

// cpuChooseTrump CPU がカドリールとして選ぶ切り札スート (最良スート) を返す。
func (g *Quadrille) cpuChooseTrump(playerIdx int) int {
	suit, _ := g.handBestTrump(playerIdx)
	return suit
}

// handBestTrump 各スートを切り札とした場合の手札強度を評価し、最良スートとその評価値を返す。
func (g *Quadrille) handBestTrump(playerIdx int) (int, int) {
	bestSuit, bestScore := CardDesignSpade, -1
	for _, s := range []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond} {
		if sc := g.evalTrump(playerIdx, s); sc > bestScore {
			bestScore = sc
			bestSuit = s
		}
	}
	return bestSuit, bestScore
}

// evalTrump trump を切り札としたときの手札強度を見積もる。切り札枚数・マタドール・高位平札を加点する。
func (g *Quadrille) evalTrump(playerIdx, trump int) int {
	p := g.players[playerIdx]
	score := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if quadrilleIsTrump(c, trump) {
			score += 2
			if quadrilleCardStrength(c, trump) >= quadrilleStrPunto {
				score += 3 // マタドール / Punto
			}
		} else if c.GetValue() == 13 || c.GetValue() == 12 {
			score++ // 高位平札
		}
	}
	return score
}

// --- Play ---

// startPlay ビッド確定後、プレイフェーズを開始する (カドリールがリード)。
func (g *Quadrille) startPlay() {
	g.sortAllHands()
	g.leadPlayerIdx = g.quadrilleIdx
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = QuadrillePhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Quadrille) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != QuadrillePhasePlay {
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
func (g *Quadrille) CpuPlay() {
	if g.gameEndFlag || g.phase != QuadrillePhasePlay {
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
func (g *Quadrille) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})
	g.revealPartnerIfCalledKing(playerIdx, card)

	if len(g.currentTrick) == QuadrillePlayerCnt {
		g.phase = QuadrillePhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % QuadrillePlayerCnt
	}
}

// revealPartnerIfCalledKing は呼ばれた王が場に出た時点で味方を公開する。
//
// **誰が味方かはそれまで伏せる** —— 落札者以外は自分が呼ばれたことすら
// 知らないまま打つ、というのが Quadrille の緊張感そのもの。
func (g *Quadrille) revealPartnerIfCalledKing(playerIdx int, card *Card) {
	if g.partnerRevealed || g.roiSeul || card == nil {
		return
	}
	if card.GetValue() != quadrilleKingValue || card.GetDesign() != g.calledKingSuit {
		return
	}
	g.partnerRevealed = true
	g.appendLog(playerIdx, "partner_revealed",
		fmt.Sprintf("%s holds the called King of %s and is the Quadrille's partner",
			playerName(g.players, playerIdx), quadrilleSuitName(g.calledKingSuit)), nil)
}

// ResolveTrick トリックを解決して勝者を決定する。最終トリックなら RoundEnd に入り、得点計算を発火する。
func (g *Quadrille) ResolveTrick() {
	if g.phase != QuadrillePhaseTrickEnd || len(g.currentTrick) != QuadrillePlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", playerName(g.players, winnerIdx), g.trickNumber), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= QuadrilleTrickCount {
		g.lastTrickWinner = winnerIdx
		g.phase = QuadrillePhaseRoundEnd
		g.enterRoundEnd()
	} else {
		g.phase = QuadrillePhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Quadrille) NextTrick() {
	if g.phase != QuadrillePhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = QuadrillePhasePlay
}

// enterRoundEnd RoundEnd 突入時に一度だけ結果判定と得点計算を行う (scored フラグでガード)。
func (g *Quadrille) enterRoundEnd() {
	if g.scored {
		return
	}
	g.scored = true
	g.outcome = g.evalOutcome()
	g.applyScores(g.outcome)
	stake := 1
	if g.winningBid == QuadrilleBidSolo {
		stake = 2
	}
	g.appendLog(-1, "round_score",
		fmt.Sprintf("round %d: Quadrille(%s) %s (stake=%d)",
			g.roundNumber, playerName(g.players, g.quadrilleIdx), quadrilleOutcomeName(g.outcome), stake), nil)
	g.checkGameEnd()
}

// evalOutcome カドリールのトリック数と 2 人の相手のトリック数から Sacar/Puesta/Codille を判定する。
// evalOutcome は落札者**側**と相手**側**のトリック数を比べる。
//
// **提携の単位と判定の単位を一致させる。** クローン元のオンブルは 3 人卓で
// 味方がいないので「落札者 1 人 対 他席の最大」で足りたが、こちらは
// 呼ばれた王の持ち主が味方に付く。落札者 1 人分だけ数えると、味方が取った
// トリックが相手側に計上されて勝敗が反転する。単独プレイ (Roi seul) なら
// 1 席 対 3 席。
func (g *Quadrille) evalOutcome() QuadrilleOutcome {
	ours, theirs := g.sideTrickCounts()
	switch {
	case ours > theirs:
		return QuadrilleOutcomeSacar
	case ours < theirs:
		return QuadrilleOutcomeCodille
	default:
		return QuadrilleOutcomePuesta
	}
}

// sideTrickCounts は (落札者側, 相手側) の獲得トリック数を返す。
func (g *Quadrille) sideTrickCounts() (int, int) {
	ours, theirs := 0, 0
	for i := 0; i < QuadrillePlayerCnt; i++ {
		n := g.players[i].GetTrickCount()
		if g.quadrilleSideOf(i) {
			ours += n
		} else {
			theirs += n
		}
	}
	return ours, theirs
}

// GetSideTrickCounts は (落札者側, 相手側) の獲得トリック数を返す。
func (g *Quadrille) GetSideTrickCounts() (int, int) { return g.sideTrickCounts() }

// applyScores ディール結果に応じて累積点を更新する。カドリールと連合 2 人の間で点が移動する。
func (g *Quadrille) applyScores(outcome QuadrilleOutcome) {
	var quadrilleDelta, oppDelta int
	switch outcome {
	case QuadrilleOutcomeSacar:
		quadrilleDelta, oppDelta = +quadrilleScoreSacar, -1
	case QuadrilleOutcomePuesta:
		quadrilleDelta, oppDelta = -quadrilleScorePuesta, +1
	case QuadrilleOutcomeCodille:
		quadrilleDelta, oppDelta = -quadrilleScoreCodille, +2
	default:
		return
	}
	for i := 0; i < QuadrillePlayerCnt; i++ {
		// **味方も落札者と同じ側の点を受け取る。** 落札者だけに配ると、
		// 呼ばれた王を持っていた席は自分が勝った側にいながら相手側の点を貰う。
		if g.quadrilleSideOf(i) {
			g.playerScores[i] += quadrilleDelta
		} else {
			g.playerScores[i] += oppDelta
		}
	}
}

// checkGameEnd 規定ディール数を終えたらマッチ終了を判定し、累積点最上位を勝者とする。
func (g *Quadrille) checkGameEnd() {
	if g.roundNumber < g.config.TargetRounds {
		return
	}
	leader, best := 0, g.playerScores[0]
	tie := false
	for i := 1; i < QuadrillePlayerCnt; i++ {
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
	g.phase = QuadrillePhaseGameEnd
	g.result = g.humanResult(leader, tie)
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", playerName(g.players, leader)), nil)
}

// humanResult 人間 (seat 0) の視点でマッチ結果を返す。単独トップなら Win、トップ同点なら None、他は Lose。
func (g *Quadrille) humanResult(leader int, tie bool) QuadrilleResult {
	human := findHumanIdx(g.players)
	if human < 0 {
		return QuadrilleResultNone
	}
	if g.playerScores[human] == g.playerScores[leader] {
		if tie {
			return QuadrilleResultNone
		}
		return QuadrilleResultWin
	}
	return QuadrilleResultLose
}

// ScoreRound RoundEnd フェーズでの得点計算を行う (enterRoundEnd を idempotent に呼ぶ、インタフェース互換)。
func (g *Quadrille) ScoreRound() {
	if g.phase != QuadrillePhaseRoundEnd {
		return
	}
	g.enterRoundEnd()
}

// --- Trick / play helpers ---

// validatePlay マストフォローを検証する (切り札グループを 1 スートとして扱う)。
func (g *Quadrille) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadEff := quadrilleEffectiveSuit(g.currentTrick[0].Card, g.trumpSuit)
	if quadrilleEffectiveSuit(card, g.trumpSuit) != leadEff && g.playerHasEffSuit(playerIdx, leadEff) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasEffSuit プレイヤーが指定の実効スート (平札スート or 切り札グループ) を持っているか。
func (g *Quadrille) playerHasEffSuit(playerIdx, eff int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if quadrilleEffectiveSuit(p.GetCard(i), g.trumpSuit) == eff {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する。切り札があれば最強切り札、無ければリードスートの最強札が勝つ。
func (g *Quadrille) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadEff := quadrilleEffectiveSuit(g.currentTrick[0].Card, g.trumpSuit)
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerStr := quadrilleCardStrength(g.currentTrick[0].Card, g.trumpSuit)
	for _, tc := range g.currentTrick[1:] {
		eff := quadrilleEffectiveSuit(tc.Card, g.trumpSuit)
		if eff != quadrilleTrumpGroup && eff != leadEff {
			continue // 場外の平札は勝てない
		}
		if s := quadrilleCardStrength(tc.Card, g.trumpSuit); s > winnerStr {
			winnerIdx = tc.PlayerIdx
			winnerStr = s
		}
	}
	return winnerIdx
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Quadrille) getValidPlayIndices(playerIdx int) []int {
	return validPlayIndices(g.players[playerIdx], func(c *Card) bool { return g.validatePlay(playerIdx, c) == nil })
}

// isCoalition playerIdx が連合 (非カドリール) 側か。
func (g *Quadrille) isCoalition(playerIdx int) bool {
	return playerIdx != g.quadrilleIdx
}

// sameSide a と b が同じ陣営 (両方カドリール or 両方連合) か。
func (g *Quadrille) sameSide(a, b int) bool {
	return g.isCoalition(a) == g.isCoalition(b)
}

// --- Card ranking (Quadrille matador ranking, inline) ---

// quadrilleTrumpGroup 実効スートとしての「切り札グループ」を表す番号 (Joker=0 は 40 枚デッキに現れないため流用)。
const quadrilleTrumpGroup = 0

// 切り札グループ内の強さ (数値が大きいほど強い)。平札の最大 (10) を大きく上回るよう基準を高く取る。
const (
	quadrilleStrSpadille = 200 // ♠A
	quadrilleStrManille  = 199 // 切り札の 7
	quadrilleStrBasto    = 198 // ♣A
	quadrilleStrPunto    = 197 // 赤切り札の A
)

// quadrilleIsTrump card が (trump を切り札としたとき) 切り札か。♠A・♣A は常に切り札。
func quadrilleIsTrump(c *Card, trump int) bool {
	d, v := c.GetDesign(), c.GetValue()
	if d == CardDesignSpade && v == 1 {
		return true // Spadille
	}
	if d == CardDesignClover && v == 1 {
		return true // Basto
	}
	return d == trump
}

// quadrilleEffectiveSuit フォロー判定用の実効スート。切り札は quadrilleTrumpGroup、それ以外は本来のスート。
func quadrilleEffectiveSuit(c *Card, trump int) int {
	if quadrilleIsTrump(c, trump) {
		return quadrilleTrumpGroup
	}
	return c.GetDesign()
}

// quadrilleCardStrength trump を切り札としたときのカード強さ (全カードの全順序)。任意の切り札 > 任意の平札。
func quadrilleCardStrength(c *Card, trump int) int {
	d, v := c.GetDesign(), c.GetValue()
	if d == CardDesignSpade && v == 1 {
		return quadrilleStrSpadille
	}
	if d == trump && v == 7 {
		return quadrilleStrManille
	}
	if d == CardDesignClover && v == 1 {
		return quadrilleStrBasto
	}
	if d == trump {
		if v == 1 {
			return quadrilleStrPunto // 赤切り札の A (黒切り札の A は上で捕捉済み)
		}
		return quadrilleTrumpSuitRank(v)
	}
	return quadrillePlainRank(d, v)
}

// QuadrilleMatadorRank は札が三大マタドールのどれかを返す。
//
// 1 = スパディーユ (♠A)、2 = マニーユ (切り札の 7)、3 = バスト (♣A)、
// 0 = マタドールでない。**切り札が未確定なら 0。**マニーユは切り札スート
// 次第で決まるので、確定前に一部だけ示すと不揃いな案内になる。
//
// **判定は quadrilleCardStrength をそのまま読む。**別に条件を書くと、序列を
// 変えたときに表示だけ古いままになる。
func QuadrilleMatadorRank(c *Card, trump int) int {
	if c == nil || !quadrilleValidSuit(trump) {
		return 0
	}
	switch quadrilleCardStrength(c, trump) {
	case quadrilleStrSpadille:
		return 1
	case quadrilleStrManille:
		return 2
	case quadrilleStrBasto:
		return 3
	}
	return 0
}

// quadrilleTrumpSuitRank 切り札スートの残り札 K>Q>J>6>5>4>3>2 の強さ。
func quadrilleTrumpSuitRank(v int) int {
	switch v {
	case 13:
		return 196
	case 12:
		return 195
	case 11:
		return 194
	case 6:
		return 193
	case 5:
		return 192
	case 4:
		return 191
	case 3:
		return 190
	default: // 2
		return 189
	}
}

// quadrillePlainRank 平札 (非切り札) の強さ。伝統的カドリールの非対称ランクを再現する:
// 黒スート (♠♣): K>Q>J>7>6>5>4>3>2>A (A が最弱)。
// 赤スート (♥♦): K>Q>J>A>2>3>4>5>6>7 (A が 4 番目、7 が最弱)。
func quadrillePlainRank(d, v int) int {
	if d == CardDesignHeart || d == CardDesignDiamond {
		switch v {
		case 13: // K
			return 10
		case 12: // Q
			return 9
		case 11: // J
			return 8
		case 1: // A
			return 7
		case 2:
			return 6
		case 3:
			return 5
		case 4:
			return 4
		case 5:
			return 3
		case 6:
			return 2
		default: // 7
			return 1
		}
	}
	switch v {
	case 13:
		return 10
	case 12:
		return 9
	case 11:
		return 8
	case 7:
		return 7
	case 6:
		return 6
	case 5:
		return 5
	case 4:
		return 4
	case 3:
		return 3
	case 2:
		return 2
	default: // Ace
		return 1
	}
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Quadrille) sortAllHands() {
	for _, p := range g.players {
		quadrilleSortHand(p, g.trumpSuit)
	}
}

// quadrilleSortHand 手札を実効スート→強さ順にソートする。trump=-1 (未確定) の場合はスート→値で並べる。
func quadrilleSortHand(p *QuadrillePlayer, trump int) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if !quadrilleValidSuit(trump) {
			if cards[i].GetDesign() != cards[j].GetDesign() {
				return cards[i].GetDesign() < cards[j].GetDesign()
			}
			return cards[i].GetValue() > cards[j].GetValue()
		}
		ei, ej := quadrilleEffectiveSuit(cards[i], trump), quadrilleEffectiveSuit(cards[j], trump)
		if ei != ej {
			return ei < ej
		}
		return quadrilleCardStrength(cards[i], trump) > quadrilleCardStrength(cards[j], trump)
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// quadrilleBidName ビッドの表示名を返す。
func quadrilleBidName(bid QuadrilleBid) string {
	switch bid {
	case QuadrilleBidEntrar:
		return "entrar"
	case QuadrilleBidSolo:
		return "solo"
	default:
		return "pass"
	}
}

// quadrilleSuitName スートの表示名を返す。
func quadrilleSuitName(suit int) string {
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

// quadrilleOutcomeName 結果の表示名を返す。
func quadrilleOutcomeName(o QuadrilleOutcome) string {
	switch o {
	case QuadrilleOutcomeSacar:
		return "sacar"
	case QuadrilleOutcomePuesta:
		return "puesta"
	case QuadrilleOutcomeCodille:
		return "codille"
	default:
		return "-"
	}
}

// quadrilleKingValue 王の札位 (呼び声の対象)。
const quadrilleKingValue = 13

// quadrilleValidSuit suit が有効なスート (1..4) か。
func quadrilleValidSuit(suit int) bool {
	return suit >= CardDesignSpade && suit <= CardDesignDiamond
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Quadrille) indexOfPlayerInTrick(playerIdx int) int {
	return indexOfPlayerInTrick(g.currentTrick, playerIdx)
}

// --- CPU AI (play) ---

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Quadrille) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == QuadrilleCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 陣営 (カドリール vs 連合) を意識した戦略プレイ。
func (g *Quadrille) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	trump := g.trumpSuit

	// リード: カドリールは強い札で主導、連合は弱い札で温存する。
	if len(g.currentTrick) == 0 {
		if playerIdx == g.quadrilleIdx {
			return g.maxByStrength(player, valid)
		}
		return g.minByStrength(player, valid)
	}

	leadEff := quadrilleEffectiveSuit(g.currentTrick[0].Card, trump)
	winnerIdx := g.trickWinner()
	topStr := quadrilleCardStrength(g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card, trump)
	partnerWinning := g.sameSide(winnerIdx, playerIdx)

	var follows []int
	for _, idx := range valid {
		if quadrilleEffectiveSuit(player.GetCard(idx), trump) == leadEff {
			follows = append(follows, idx)
		}
	}

	if len(follows) == 0 {
		// ボイド: 最も弱い札を捨てる。
		return g.minByStrength(player, valid)
	}

	winners := filterIndices(follows, func(idx int) bool {
		return quadrilleCardStrength(player.GetCard(idx), trump) > topStr
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
func (g *Quadrille) minByStrength(player *QuadrillePlayer, indices []int) int {
	best := indices[0]
	bestScore := quadrilleCardStrength(player.GetCard(best), g.trumpSuit)
	for _, idx := range indices[1:] {
		if s := quadrilleCardStrength(player.GetCard(idx), g.trumpSuit); s < bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// maxByStrength 強さが最大となるインデックスを返す。
func (g *Quadrille) maxByStrength(player *QuadrillePlayer, indices []int) int {
	best := indices[0]
	bestScore := quadrilleCardStrength(player.GetCard(best), g.trumpSuit)
	for _, idx := range indices[1:] {
		if s := quadrilleCardStrength(player.GetCard(idx), g.trumpSuit); s > bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨アクションを返す。
func (g *Quadrille) GetHint() *QuadrilleHint {
	human := findHumanIdx(g.players)
	if human < 0 {
		return nil
	}
	switch g.phase {
	case QuadrillePhaseBid:
		if g.currentBidderIdx != human {
			return nil
		}
		bid := g.cpuChooseBidForHint(human)
		return &QuadrilleHint{Reason: quadrilleBidHintReason(bid)}
	case QuadrillePhasePlay:
		if g.currentPlayerIdx != human {
			return nil
		}
		valid := g.getValidPlayIndices(human)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuPlaySmart(human, valid)
		return &QuadrilleHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
	}
	return nil
}

// cpuChooseBidForHint ヒント用にビッド推奨を計算する (Easy 難易度でも強度から推奨する)。
func (g *Quadrille) cpuChooseBidForHint(playerIdx int) QuadrilleBid {
	_, strength := g.handBestTrump(playerIdx)
	highest := g.highestBid()
	if strength >= 20 && QuadrilleBidSolo > highest {
		return QuadrilleBidSolo
	}
	if strength >= 13 && QuadrilleBidEntrar > highest {
		return QuadrilleBidEntrar
	}
	return QuadrilleBidNone
}

// quadrilleBidHintReason ビッド推奨に対応するヒント理由キーを返す。
func quadrilleBidHintReason(bid QuadrilleBid) string {
	switch bid {
	case QuadrilleBidSolo:
		return "bid_solo"
	case QuadrilleBidEntrar:
		return "bid_entrar"
	default:
		return "bid_pass"
	}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Quadrille) playHintReason(playerIdx, chosenIdx int) string {
	trump := g.trumpSuit
	if len(g.currentTrick) == 0 {
		if playerIdx == g.quadrilleIdx {
			return "lead_high"
		}
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadEff := quadrilleEffectiveSuit(g.currentTrick[0].Card, trump)
	if quadrilleEffectiveSuit(card, trump) != leadEff {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	topStr := quadrilleCardStrength(g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card, trump)
	if quadrilleCardStrength(card, trump) > topStr {
		return "follow_win"
	}
	if g.sameSide(winnerIdx, playerIdx) {
		return "give_partner"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Quadrille) GetPhase() QuadrillePhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Quadrille) SetPhase(phase QuadrillePhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Quadrille) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Quadrille) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Quadrille) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Quadrille) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Quadrille) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Quadrille) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Quadrille) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Quadrille) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Quadrille) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Quadrille) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Quadrille) GetDealerIdx() int { return g.dealerIdx }

// GetForehandIdx forehand インデックス取得
func (g *Quadrille) GetForehandIdx() int { return g.forehandIdx }

// GetQuadrilleIdx カドリールインデックス取得 (-1=未確定)
func (g *Quadrille) GetQuadrilleIdx() int { return g.quadrilleIdx }

// GetCallableKingSuitsForTest はテスト用に席 idx が呼べる王のスートを返す。
// GetCallableKingSuits と違いフェーズを問わない。
func (g *Quadrille) GetCallableKingSuitsForTest(idx int) []int {
	return g.callableKingSuits(idx)
}

// SetPartnerForTest はテスト用に味方の席と公開状態を設定する。
func (g *Quadrille) SetPartnerForTest(idx int, revealed bool) {
	g.partnerIdx = idx
	g.partnerRevealed = revealed
}

// SetRoiSeulForTest はテスト用に単独プレイを設定する。
func (g *Quadrille) SetRoiSeulForTest(v bool) { g.roiSeul = v }

// SetCalledKingSuitForTest はテスト用に呼ばれた王のスートを設定する。
func (g *Quadrille) SetCalledKingSuitForTest(suit int) { g.calledKingSuit = suit }

// SetQuadrilleIdx カドリールインデックス設定 (テスト用)
func (g *Quadrille) SetQuadrilleIdx(idx int) { g.quadrilleIdx = idx }

// GetWinningBid 確定ビッド取得
func (g *Quadrille) GetWinningBid() QuadrilleBid { return g.winningBid }

// GetHighestBid は**競り中の**最高宣言を返す。
//
// GetWinningBid は落札が確定するまで None のままなので、競っている最中の
// 画面がそれを出すと「最高ビッド: -」と表示されたまま Entrar が弾かれる。
// 何を上回ればよいのかが読めないので、競り中はこちらを出す。
func (g *Quadrille) GetHighestBid() QuadrilleBid { return g.highestBid() }

// SetWinningBid 確定ビッド設定 (テスト用)
func (g *Quadrille) SetWinningBid(b QuadrilleBid) { g.winningBid = b }

// GetTrumpSuit 切り札スート取得 (-1=未確定, 1..4)
func (g *Quadrille) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *Quadrille) SetTrumpSuit(s int) { g.trumpSuit = s }

// GetCurrentBidderIdx 現在のビッド手番インデックス取得
func (g *Quadrille) GetCurrentBidderIdx() int { return g.currentBidderIdx }

// GetPlayerScores プレイヤー別累積点取得
func (g *Quadrille) GetPlayerScores() [QuadrillePlayerCnt]int { return g.playerScores }

// SetPlayerScores プレイヤー別累積点設定 (テスト用)
func (g *Quadrille) SetPlayerScores(s [QuadrillePlayerCnt]int) { g.playerScores = s }

// GetOutcome 直近ディールの結果取得
func (g *Quadrille) GetOutcome() QuadrilleOutcome { return g.outcome }

// GetResult 人間視点のマッチ結果取得
func (g *Quadrille) GetResult() QuadrilleResult { return g.result }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Quadrille) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *Quadrille) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayerCnt プレイヤー数取得
func (g *Quadrille) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Quadrille) GetPlayer(i int) *QuadrillePlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番 (プレイ) が人間か。
func (g *Quadrille) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// IsHumanBidTurn 現在のビッド手番が人間か。
func (g *Quadrille) IsHumanBidTurn() bool {
	if g.phase != QuadrillePhaseBid {
		return false
	}
	if g.currentBidderIdx < 0 || g.currentBidderIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentBidderIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Quadrille) GetConfig() QuadrilleConfig { return g.config }

// SetConfig 設定変更
func (g *Quadrille) SetConfig(cfg QuadrilleConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Quadrille) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != QuadrillePhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// quadrilleJSON is the JSON wire format for Quadrille.
type quadrilleJSON struct {
	TrumpCards       *TrumpCards                      `json:"tc"`
	Players          []*QuadrillePlayer               `json:"ps"`
	Config           QuadrilleConfig                  `json:"cf"`
	Phase            QuadrillePhase                   `json:"ph"`
	RoundNumber      int                              `json:"rn"`
	TrickNumber      int                              `json:"tn"`
	CurrentPlayerIdx int                              `json:"ci"`
	CurrentTrick     []*TrickCard                     `json:"ct"`
	LeadPlayerIdx    int                              `json:"li"`
	DealerIdx        int                              `json:"di"`
	ForehandIdx      int                              `json:"fh"`
	QuadrilleIdx     int                              `json:"om"`
	WinningBid       QuadrilleBid                     `json:"wb"`
	TrumpSuit        int                              `json:"ts"`
	CurrentBidderIdx int                              `json:"cbi"`
	Bids             [QuadrillePlayerCnt]QuadrilleBid `json:"bd"`
	BidTrump         [QuadrillePlayerCnt]int          `json:"bt"`
	BidActed         [QuadrillePlayerCnt]bool         `json:"ba"`
	PlayerScores     [QuadrillePlayerCnt]int          `json:"sc"`
	LastTrickWinner  int                              `json:"lt"`
	// 王呼びは盤面の一部。落とすと復元後に**味方が席 0 になる** (ゼロ値)。
	// 呼ばれた王も単独プレイの区別も消えるので、勝敗の集計が静かに変わる。
	CalledKingSuit  int               `json:"ck"`
	PartnerIdx      int               `json:"pi"`
	PartnerRevealed bool              `json:"pr"`
	RoiSeul         bool              `json:"rq"`
	Outcome         QuadrilleOutcome  `json:"oc"`
	Result          QuadrilleResult   `json:"rs"`
	Scored          bool              `json:"sd"`
	GameEndFlag     bool              `json:"ge"`
	WinnerPlayer    int               `json:"wp"`
	ActionLog       []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Quadrille) MarshalJSON() ([]byte, error) {
	return json.Marshal(quadrilleJSON{
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
		QuadrilleIdx:     g.quadrilleIdx,
		WinningBid:       g.winningBid,
		TrumpSuit:        g.trumpSuit,
		CurrentBidderIdx: g.currentBidderIdx,
		Bids:             g.bids,
		BidTrump:         g.bidTrump,
		BidActed:         g.bidActed,
		PlayerScores:     g.playerScores,
		LastTrickWinner:  g.lastTrickWinner,
		CalledKingSuit:   g.calledKingSuit,
		PartnerIdx:       g.partnerIdx,
		PartnerRevealed:  g.partnerRevealed,
		RoiSeul:          g.roiSeul,
		Outcome:          g.outcome,
		Result:           g.result,
		Scored:           g.scored,
		GameEndFlag:      g.gameEndFlag,
		WinnerPlayer:     g.winnerPlayer,
		ActionLog:        g.actionLog,
	})
}

// quadrilleMaxSliceLen caps slice sizes during deserialisation.
const quadrilleMaxSliceLen = 5000

// errQuadrilleOversized is the single sentinel error for oversized input arrays.
var errQuadrilleOversized = errors.New("quadrille: input array exceeds maximum allowed size")

// errQuadrilleInvalidPlayers is returned when restored state lacks exactly QuadrillePlayerCnt players.
var errQuadrilleInvalidPlayers = errors.New("quadrille: invalid player count")

// errQuadrilleInvalidTrick is returned when a restored trick card or its card is nil / out of range.
var errQuadrilleInvalidTrick = errors.New("quadrille: invalid trick card")

// errQuadrilleInvalidIndex is returned when a restored index field is out of range.
var errQuadrilleInvalidIndex = errors.New("quadrille: index field out of range")

// errQuadrilleInvalidPhase is returned when a restored phase is out of range.
var errQuadrilleInvalidPhase = errors.New("quadrille: phase out of range")

// errQuadrilleInvalidBid is returned when a restored bid value is out of range.
var errQuadrilleInvalidBid = errors.New("quadrille: bid value out of range")

// errQuadrilleInvalidTrump is returned when a restored trump suit is out of range.
var errQuadrilleInvalidTrump = errors.New("quadrille: trump suit out of range")

// errQuadrilleInvalidOutcome is returned when a restored outcome or result value is out of range.
var errQuadrilleInvalidOutcome = errors.New("quadrille: outcome/result value out of range")

// quadrilleInRange reports whether v is in [0, QuadrillePlayerCnt).
func quadrilleInRange(v int) bool { return v >= 0 && v < QuadrillePlayerCnt }

// quadrilleInRangeOrUnset reports whether v is -1 (unset) or in [0, QuadrillePlayerCnt).
func quadrilleInRangeOrUnset(v int) bool { return v == -1 || quadrilleInRange(v) }

// quadrilleValidBid reports whether b is a defined bid value.
func quadrilleValidBid(b QuadrilleBid) bool {
	return b >= QuadrilleBidNone && b <= QuadrilleBidSolo
}

// quadrilleTrumpInRangeOrUnset reports whether s is -1 (unset) or a valid suit (1..4).
func quadrilleTrumpInRangeOrUnset(s int) bool { return s == -1 || quadrilleValidSuit(s) }

// UnmarshalJSON implements json.Unmarshaler.
func (g *Quadrille) UnmarshalJSON(data []byte) error {
	var j quadrilleJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > quadrilleMaxSliceLen || len(j.CurrentTrick) > quadrilleMaxSliceLen ||
		len(j.ActionLog) > quadrilleMaxSliceLen {
		return errQuadrilleOversized
	}
	if len(j.Players) != QuadrillePlayerCnt {
		return errQuadrilleInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errQuadrilleInvalidPlayers
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return errQuadrilleInvalidTrick
		}
		if !quadrilleInRange(tc.PlayerIdx) {
			return errQuadrilleInvalidTrick
		}
	}
	// 範囲必須のインデックス [0, PlayerCnt)。
	if !quadrilleInRange(j.CurrentPlayerIdx) || !quadrilleInRange(j.DealerIdx) ||
		!quadrilleInRange(j.ForehandIdx) || !quadrilleInRange(j.CurrentBidderIdx) {
		return errQuadrilleInvalidIndex
	}
	// -1 (未設定) 許容のインデックス。
	if !quadrilleInRangeOrUnset(j.LeadPlayerIdx) || !quadrilleInRangeOrUnset(j.QuadrilleIdx) ||
		!quadrilleInRangeOrUnset(j.LastTrickWinner) || !quadrilleInRangeOrUnset(j.WinnerPlayer) {
		return errQuadrilleInvalidIndex
	}
	// フェーズ依存の厳格化: play 以降は quadrille・lead・trump が確定していなければ
	// 後続処理で g.players[-1] / g.trumpSuit を参照して panic するため確定を要求する。
	if j.Phase >= QuadrillePhasePlay {
		if !quadrilleInRange(j.QuadrilleIdx) || !quadrilleInRange(j.LeadPlayerIdx) {
			return errQuadrilleInvalidIndex
		}
		if !quadrilleValidSuit(j.TrumpSuit) {
			return errQuadrilleInvalidTrump
		}
	}
	if int(j.Phase) < QuadrillePhaseMin || int(j.Phase) > QuadrillePhaseMax {
		return errQuadrilleInvalidPhase
	}
	if !quadrilleValidBid(j.WinningBid) {
		return errQuadrilleInvalidBid
	}
	for _, b := range j.Bids {
		if !quadrilleValidBid(b) {
			return errQuadrilleInvalidBid
		}
	}
	// 王呼びも他のインデックス同様に検査する。**壊れた値を素通しすると
	// quadrilleSideOf が誤った席を味方として数え、勝敗が変わる。**
	//
	// ただし 0 は「未指定」と「席 0 / スート 0」の区別が付かない ——
	// JSON にフィールドが無ければ Go は 0 を入れるので、素直に信じると
	// **席 0 が誰の味方でもないのに落札者側として数えられる**。
	// 検査は範囲だけにして、整合は下の quadrilleNormaliseKingCall で取る。
	if j.CalledKingSuit != 0 && !quadrilleTrumpInRangeOrUnset(j.CalledKingSuit) {
		return errQuadrilleInvalidTrump
	}
	if !quadrilleInRangeOrUnset(j.PartnerIdx) {
		return errQuadrilleInvalidIndex
	}
	if !quadrilleTrumpInRangeOrUnset(j.TrumpSuit) {
		return errQuadrilleInvalidTrump
	}
	// bidTrump は -1/0 (未設定) または有効スート (1..4)。範囲外のみ拒否する
	// (finalizeAuction は無効な bidTrump を自動で選び直すため厳格化は不要)。
	for _, t := range j.BidTrump {
		if t < -1 || t > CardDesignDiamond {
			return errQuadrilleInvalidTrump
		}
	}
	if j.Outcome < QuadrilleOutcomeNone || j.Outcome > QuadrilleOutcomeCodille {
		return errQuadrilleInvalidOutcome
	}
	if j.Result < QuadrilleResultLose || j.Result > QuadrilleResultWin {
		return errQuadrilleInvalidOutcome
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newQuadrilleDeck()
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
	g.quadrilleIdx = j.QuadrilleIdx
	g.winningBid = j.WinningBid
	g.trumpSuit = j.TrumpSuit
	g.currentBidderIdx = j.CurrentBidderIdx
	g.bids = j.Bids
	g.bidTrump = j.BidTrump
	g.bidActed = j.BidActed
	g.playerScores = j.PlayerScores
	g.lastTrickWinner = j.LastTrickWinner
	g.calledKingSuit, g.partnerIdx, g.partnerRevealed, g.roiSeul =
		quadrilleNormaliseKingCall(j.CalledKingSuit, j.PartnerIdx, j.PartnerRevealed, j.RoiSeul)
	g.outcome = j.Outcome
	g.result = j.Result
	g.scored = j.Scored
	g.gameEndFlag = j.GameEndFlag
	g.winnerPlayer = j.WinnerPlayer
	g.actionLog = j.ActionLog
	return nil
}

// quadrilleNormaliseKingCall は復元した王呼びの状態を整合させる。
//
// **0 は「未指定」と区別が付かない。** JSON にフィールドが無ければ Go は 0 を
// 入れるので、そのまま信じると calledKingSuit=0 (どの札とも一致しないので
// 味方が永久に公開されない) と partnerIdx=0 (席 0 が味方として数えられ、
// 勝敗の集計が変わる) が同時に起きる。
//
// 王が呼ばれていない盤面に味方は存在しない、という不変条件で揃える。
func quadrilleNormaliseKingCall(suit, partner int, revealed, roiSeul bool) (int, int, bool, bool) {
	if roiSeul {
		// 単独プレイに味方はいない。
		return -1, -1, false, true
	}
	if !quadrilleValidSuit(suit) {
		// 王が呼ばれていないので、味方も公開状態も無い。
		return -1, -1, false, false
	}
	if partner < 0 || partner >= QuadrillePlayerCnt {
		return suit, -1, false, false
	}
	return suit, partner, revealed, false
}
