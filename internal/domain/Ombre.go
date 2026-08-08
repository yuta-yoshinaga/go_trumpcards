//go:build !js || !wasm || extra3

// Package domain オンブル / オンブレ (Ombre / Hombre) のドメインモデル。
//
// Ombre は 17 世紀スペイン発祥の 3 人用トリックテイキングゲームで、1 人の「オンブル (Ombre)」が
// 残り 2 人の連合 (coalition) と対戦する、ソリスト対多数系トリックゲームの祖先。
//
// デッキ: 40 枚 = 標準 52 枚から 8・9・10 を除いたもの (各スート A,2..7,J,Q,K)。
// 各ディールで 9 枚ずつ配り (27 枚)、残り 13 枚は未使用のストック (交換・引き無し=簡略化)。
//
// ビッド: 各プレイヤーは順に Pass / Entrar (オンブルを引き受ける) / Solo (より強い Entrar) を宣言。
// 最高ビッドの宣言者がオンブルとなり、切り札スートを選ぶ。全員パスならディーラーが強制的にオンブル
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
// ディール結果 (オンブルのトリック数 vs 各相手): Sacar=オンブルが各相手より厳密に多い (オンブル勝ち)、
// Puesta=最多に並ばれた (相手のいずれかが同数で誰にも超えられていない、オンブル負け)、
// Codille=相手のいずれかが厳密に多い (連合勝ち・倍額)。
//
// 得点 (累積、TargetRounds ディール): Sacar でオンブル +2 / 各相手 -1、Puesta でオンブル -2 / 各相手 +1、
// Codille でオンブル -4 / 各相手 +2。TargetRounds ディール後、累積点最上位が勝者。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// OmbrePlayerCnt プレイヤー数 (人間 1 + CPU 2)
const OmbrePlayerCnt = 3

// OmbreHandSize 各プレイヤーの配り札枚数
const OmbreHandSize = 9

// OmbreTrickCount 1 ディールのトリック数
const OmbreTrickCount = 9

// OmbreDeckSize デッキ枚数 (52 - 8,9,10)
const OmbreDeckSize = 40

// OmbreWinRounds マッチを構成するディール数 (既定)
const OmbreWinRounds = 5

// OmbreBid ビッド宣言
type OmbreBid int

// Ombre のビッド定数 (数値が大きいほど高い宣言)
const (
	// OmbreBidNone 未宣言 (パス相当の初期値)
	OmbreBidNone OmbreBid = 0
	// OmbreBidEntrar entrar — オンブルを引き受ける宣言。
	OmbreBidEntrar OmbreBid = 1
	// OmbreBidSolo solo — より強い Entrar。
	OmbreBidSolo OmbreBid = 2
)

// OmbrePhase ゲームフェーズ
type OmbrePhase int

// Ombre のフェーズ定数
const (
	// OmbrePhaseBid ビッド (auction) フェーズ
	OmbrePhaseBid OmbrePhase = 0
	// OmbrePhasePlay トリックプレイフェーズ
	OmbrePhasePlay OmbrePhase = 1
	// OmbrePhaseTrickEnd トリック終了フェーズ
	OmbrePhaseTrickEnd OmbrePhase = 2
	// OmbrePhaseRoundEnd ディール終了フェーズ
	OmbrePhaseRoundEnd OmbrePhase = 3
	// OmbrePhaseGameEnd ゲーム終了フェーズ
	OmbrePhaseGameEnd OmbrePhase = 4
)

// OmbrePhaseMin フェーズ下限 (検証用)
const OmbrePhaseMin = int(OmbrePhaseBid)

// OmbrePhaseMax フェーズ上限 (検証用)
const OmbrePhaseMax = int(OmbrePhaseGameEnd)

// OmbreOutcome ディール結果
type OmbreOutcome int

// Ombre のディール結果定数
const (
	// OmbreOutcomeNone 未確定
	OmbreOutcomeNone OmbreOutcome = 0
	// OmbreOutcomeSacar オンブルが各相手より厳密に多く取り勝利
	OmbreOutcomeSacar OmbreOutcome = 1
	// OmbreOutcomePuesta オンブルが最多に並ばれ敗北 (軽い罰)
	OmbreOutcomePuesta OmbreOutcome = 2
	// OmbreOutcomeCodille 相手が厳密に多く取り連合勝ち (倍の罰)
	OmbreOutcomeCodille OmbreOutcome = 3
)

// OmbreResult 人間視点のマッチ結果
type OmbreResult int

// Ombre のマッチ結果定数
const (
	// OmbreResultLose 敗北
	OmbreResultLose OmbreResult = -1
	// OmbreResultNone 未確定 / 引き分け
	OmbreResultNone OmbreResult = 0
	// OmbreResultWin 勝利
	OmbreResultWin OmbreResult = 1
)

// Ombre のスコア増減表 (1 ディールで移動する点)
const (
	// ombreScoreSacar Sacar でのオンブル側の増分 (各相手は -1)
	ombreScoreSacar = 2
	// ombreScorePuesta Puesta でのオンブル側の減分 (各相手は +1)
	ombreScorePuesta = 2
	// ombreScoreCodille Codille でのオンブル側の減分 (各相手は +2)
	ombreScoreCodille = 4
)

// OmbreHint ヒント情報
type OmbreHint struct {
	CardIndices []int  // 推奨カードインデックス (play フェーズ)
	Reason      string // ヒント理由キー
}

// Ombre オンブルのゲームクラス
type Ombre struct {
	trumpCards       *TrumpCards
	players          []*OmbrePlayer
	config           OmbreConfig
	phase            OmbrePhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	forehandIdx      int                      // ディーラーの左隣 (ビッド開始)
	ombreIdx         int                      // オンブル (-1=未確定)
	winningBid       OmbreBid                 // 確定したビッド (オンブルの宣言)
	trumpSuit        int                      // 切り札スート (-1=未確定, 1..4)
	currentBidderIdx int                      // 現在ビッド中のプレイヤー (bid フェーズ)
	bids             [OmbrePlayerCnt]OmbreBid // 各プレイヤーの宣言
	bidTrump         [OmbrePlayerCnt]int      // 各プレイヤーが宣言時に選んだ切り札 (-1=なし)
	bidActed         [OmbrePlayerCnt]bool     // 各プレイヤーが宣言済みか
	playerScores     [OmbrePlayerCnt]int      // 累積ゲーム点
	lastTrickWinner  int                      // 最終トリック勝者 (-1=未確定)
	outcome          OmbreOutcome             // 直近ディールの結果
	result           OmbreResult              // 人間視点のマッチ結果
	scored           bool                     // 当該ディールの得点計算済みか (RoundEnd 突入時に一度だけ)
	gameEndFlag      bool
	winnerPlayer     int // -1=未確定
	actionLogBase
}

// NewOmbre コンストラクタ
func NewOmbre(trumpCards *TrumpCards, players []*OmbrePlayer, config OmbreConfig) *Ombre {
	return &Ombre{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		winnerPlayer:    -1,
		lastTrickWinner: -1,
		ombreIdx:        -1,
		trumpSuit:       -1,
	}
}

// NewDefaultOmbre 標準の 3 人構成 (人間 1, CPU 2) と既定設定で生成する。
func NewDefaultOmbre() *Ombre {
	players := make([]*OmbrePlayer, OmbrePlayerCnt)
	players[0] = NewOmbrePlayer(true)
	for i := 1; i < OmbrePlayerCnt; i++ {
		players[i] = NewOmbrePlayer(false)
	}
	return NewOmbre(newOmbreDeck(), players, DefaultOmbreConfig())
}

// newOmbreDeck Ombre 用 40 枚デッキを生成する。標準 52 枚 (NewTrumpCards(0)) から 8,9,10 を除外する。
// domain パッケージ内なので TrumpCards の内部 deck を直接フィルタできる。build-tag 無しの
// NewTrumpCards は extra ワーカーからも到達可能。
func newOmbreDeck() *TrumpCards {
	full := NewTrumpCards(0)
	t := new(TrumpCards)
	t.deck = make([]*Card, 0, OmbreDeckSize)
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
func (g *Ombre) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.playerScores = [OmbrePlayerCnt]int{}
	g.result = OmbreResultNone
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のディールを開始する
func (g *Ombre) NextRound() {
	if g.phase != OmbrePhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % OmbrePlayerCnt
	g.startRound()
}

// startRound 手札を配り、ビッドフェーズを開始する。
func (g *Ombre) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.lastTrickWinner = -1
	g.ombreIdx = -1
	g.winningBid = OmbreBidNone
	g.trumpSuit = -1
	g.outcome = OmbreOutcomeNone
	g.scored = false
	g.bids = [OmbrePlayerCnt]OmbreBid{}
	g.bidActed = [OmbrePlayerCnt]bool{}
	for i := range g.bidTrump {
		g.bidTrump[i] = -1
	}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()

	g.forehandIdx = (g.dealerIdx + 1) % OmbrePlayerCnt
	g.sortAllHands()

	g.currentBidderIdx = g.forehandIdx
	g.phase = OmbrePhaseBid
}

// deal 各プレイヤーへ OmbreHandSize 枚を配る。残り 13 枚はストックとして未使用のまま残す。
func (g *Ombre) deal() {
	for i := 0; i < OmbreHandSize; i++ {
		for j := 0; j < OmbrePlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % OmbrePlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
}

// --- Bidding ---

// PlayerBid 人間がビッドする。bid は Pass(None)/Entrar/Solo。trumpSuit は Entrar/Solo 時の切り札 (1..4)。
func (g *Ombre) PlayerBid(bid OmbreBid, trumpSuit int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != OmbrePhaseBid {
		return ErrWrongPhase
	}
	if !g.players[g.currentBidderIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if !g.isBidLegal(bid) {
		return NewDomainError(ErrInvalidPlay, "現在の最高ビッドを上回る宣言が必要です")
	}
	if bid != OmbreBidNone && !ombreValidSuit(trumpSuit) {
		return NewDomainError(ErrInvalidPlay, "切り札スートを選んでください (1..4)")
	}
	g.applyBid(g.currentBidderIdx, bid, trumpSuit)
	return nil
}

// CpuBid 現在のビッド手番が CPU の場合に 1 回ビッドする。
func (g *Ombre) CpuBid() {
	if g.gameEndFlag || g.phase != OmbrePhaseBid {
		return
	}
	idx := g.currentBidderIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	bid := g.cpuChooseBid(idx)
	trump := -1
	if bid != OmbreBidNone {
		trump = g.cpuChooseTrump(idx)
	}
	g.applyBid(idx, bid, trump)
}

// isBidLegal bid が合法か (pass は常に可、entrar/solo は現最高ビッドを上回る必要)。
func (g *Ombre) isBidLegal(bid OmbreBid) bool {
	if bid == OmbreBidNone {
		return true
	}
	if bid != OmbreBidEntrar && bid != OmbreBidSolo {
		return false
	}
	return bid > g.highestBid()
}

// highestBid これまでに宣言された最高ビッドを返す。
func (g *Ombre) highestBid() OmbreBid {
	best := OmbreBidNone
	for _, b := range g.bids {
		if b > best {
			best = b
		}
	}
	return best
}

// applyBid ビッドを適用し、全員が宣言し終えたら auction を締める。
func (g *Ombre) applyBid(playerIdx int, bid OmbreBid, trumpSuit int) {
	g.bids[playerIdx] = bid
	g.bidActed[playerIdx] = true
	if bid == OmbreBidNone {
		g.bidTrump[playerIdx] = -1
		g.appendLog(playerIdx, "bid_pass",
			fmt.Sprintf("%s passes", g.playerName(playerIdx)), nil)
	} else {
		g.bidTrump[playerIdx] = trumpSuit
		g.appendLog(playerIdx, "bid",
			fmt.Sprintf("%s bids %s (trump %s)", g.playerName(playerIdx), ombreBidName(bid), ombreSuitName(trumpSuit)), nil)
	}

	if g.allBidsActed() {
		g.finalizeAuction()
		return
	}
	g.currentBidderIdx = g.nextBidder(playerIdx)
}

// allBidsActed 全員が宣言済みか。
func (g *Ombre) allBidsActed() bool {
	for _, acted := range g.bidActed {
		if !acted {
			return false
		}
	}
	return true
}

// nextBidder playerIdx の次でまだ宣言していないプレイヤーを返す。
func (g *Ombre) nextBidder(playerIdx int) int {
	for i := 1; i <= OmbrePlayerCnt; i++ {
		cand := (playerIdx + i) % OmbrePlayerCnt
		if !g.bidActed[cand] {
			return cand
		}
	}
	return playerIdx
}

// finalizeAuction 最高ビッドの宣言者をオンブルに確定し、プレイフェーズへ進む。
// 同値は forehand から時計回りで最初に宣言したプレイヤーが優先される。
// 全員パスならディーラーが強制的にオンブル (切り札はディーラーの最良スートを自動選択)。
func (g *Ombre) finalizeAuction() {
	best := g.highestBid()
	ombre := -1
	if best > OmbreBidNone {
		for i := 0; i < OmbrePlayerCnt; i++ {
			cand := (g.forehandIdx + i) % OmbrePlayerCnt
			if g.bids[cand] == best {
				ombre = cand
				break
			}
		}
	}
	if ombre < 0 {
		// 全員パス: ディーラーが強制的に Entrar を引き受け、切り札を自動選択する。
		ombre = g.dealerIdx
		best = OmbreBidEntrar
		g.bidTrump[ombre] = g.cpuChooseTrump(ombre)
	}
	g.ombreIdx = ombre
	g.winningBid = best
	g.trumpSuit = g.bidTrump[ombre]
	if !ombreValidSuit(g.trumpSuit) {
		g.trumpSuit = g.cpuChooseTrump(ombre)
	}
	g.appendLog(ombre, "ombre",
		fmt.Sprintf("%s is Ombre with %s (trump %s)", g.playerName(ombre), ombreBidName(best), ombreSuitName(g.trumpSuit)), nil)
	g.startPlay()
}

// cpuChooseBid CPU が手札強度からビッドを選ぶ。
func (g *Ombre) cpuChooseBid(playerIdx int) OmbreBid {
	if g.config.CpuDifficulty == OmbreCpuDifficultyEasy {
		return OmbreBidNone
	}
	_, strength := g.handBestTrump(playerIdx)
	highest := g.highestBid()
	if strength >= 20 && OmbreBidSolo > highest {
		return OmbreBidSolo
	}
	if strength >= 13 && OmbreBidEntrar > highest {
		return OmbreBidEntrar
	}
	return OmbreBidNone
}

// cpuChooseTrump CPU がオンブルとして選ぶ切り札スート (最良スート) を返す。
func (g *Ombre) cpuChooseTrump(playerIdx int) int {
	suit, _ := g.handBestTrump(playerIdx)
	return suit
}

// handBestTrump 各スートを切り札とした場合の手札強度を評価し、最良スートとその評価値を返す。
func (g *Ombre) handBestTrump(playerIdx int) (int, int) {
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
func (g *Ombre) evalTrump(playerIdx, trump int) int {
	p := g.players[playerIdx]
	score := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if ombreIsTrump(c, trump) {
			score += 2
			if ombreCardStrength(c, trump) >= ombreStrPunto {
				score += 3 // マタドール / Punto
			}
		} else if c.GetValue() == 13 || c.GetValue() == 12 {
			score++ // 高位平札
		}
	}
	return score
}

// --- Play ---

// startPlay ビッド確定後、プレイフェーズを開始する (オンブルがリード)。
func (g *Ombre) startPlay() {
	g.sortAllHands()
	g.leadPlayerIdx = g.ombreIdx
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = OmbrePhasePlay
}

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Ombre) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != OmbrePhasePlay {
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
func (g *Ombre) CpuPlay() {
	if g.gameEndFlag || g.phase != OmbrePhasePlay {
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
func (g *Ombre) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", g.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == OmbrePlayerCnt {
		g.phase = OmbrePhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % OmbrePlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定する。最終トリックなら RoundEnd に入り、得点計算を発火する。
func (g *Ombre) ResolveTrick() {
	if g.phase != OmbrePhaseTrickEnd || len(g.currentTrick) != OmbrePlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d", g.playerName(winnerIdx), g.trickNumber), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= OmbreTrickCount {
		g.lastTrickWinner = winnerIdx
		g.phase = OmbrePhaseRoundEnd
		g.enterRoundEnd()
	} else {
		g.phase = OmbrePhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Ombre) NextTrick() {
	if g.phase != OmbrePhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = OmbrePhasePlay
}

// enterRoundEnd RoundEnd 突入時に一度だけ結果判定と得点計算を行う (scored フラグでガード)。
func (g *Ombre) enterRoundEnd() {
	if g.scored {
		return
	}
	g.scored = true
	g.outcome = g.evalOutcome()
	g.applyScores(g.outcome)
	stake := 1
	if g.winningBid == OmbreBidSolo {
		stake = 2
	}
	g.appendLog(-1, "round_score",
		fmt.Sprintf("round %d: Ombre(%s) %s (stake=%d)",
			g.roundNumber, g.playerName(g.ombreIdx), ombreOutcomeName(g.outcome), stake), nil)
	g.checkGameEnd()
}

// evalOutcome オンブルのトリック数と 2 人の相手のトリック数から Sacar/Puesta/Codille を判定する。
func (g *Ombre) evalOutcome() OmbreOutcome {
	ombreTricks := g.players[g.ombreIdx].GetTrickCount()
	maxOpp := -1
	for i := 0; i < OmbrePlayerCnt; i++ {
		if i == g.ombreIdx {
			continue
		}
		if t := g.players[i].GetTrickCount(); t > maxOpp {
			maxOpp = t
		}
	}
	switch {
	case ombreTricks > maxOpp:
		return OmbreOutcomeSacar
	case ombreTricks < maxOpp:
		return OmbreOutcomeCodille
	default:
		return OmbreOutcomePuesta
	}
}

// applyScores ディール結果に応じて累積点を更新する。オンブルと連合 2 人の間で点が移動する。
func (g *Ombre) applyScores(outcome OmbreOutcome) {
	var ombreDelta, oppDelta int
	switch outcome {
	case OmbreOutcomeSacar:
		ombreDelta, oppDelta = +ombreScoreSacar, -1
	case OmbreOutcomePuesta:
		ombreDelta, oppDelta = -ombreScorePuesta, +1
	case OmbreOutcomeCodille:
		ombreDelta, oppDelta = -ombreScoreCodille, +2
	default:
		return
	}
	for i := 0; i < OmbrePlayerCnt; i++ {
		if i == g.ombreIdx {
			g.playerScores[i] += ombreDelta
		} else {
			g.playerScores[i] += oppDelta
		}
	}
}

// checkGameEnd 規定ディール数を終えたらマッチ終了を判定し、累積点最上位を勝者とする。
func (g *Ombre) checkGameEnd() {
	if g.roundNumber < g.config.TargetRounds {
		return
	}
	leader, best := 0, g.playerScores[0]
	tie := false
	for i := 1; i < OmbrePlayerCnt; i++ {
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
	g.phase = OmbrePhaseGameEnd
	g.result = g.humanResult(leader, tie)
	g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", g.playerName(leader)), nil)
}

// humanResult 人間 (seat 0) の視点でマッチ結果を返す。単独トップなら Win、トップ同点なら None、他は Lose。
func (g *Ombre) humanResult(leader int, tie bool) OmbreResult {
	human := g.findHumanIdx()
	if human < 0 {
		return OmbreResultNone
	}
	if g.playerScores[human] == g.playerScores[leader] {
		if tie {
			return OmbreResultNone
		}
		return OmbreResultWin
	}
	return OmbreResultLose
}

// ScoreRound RoundEnd フェーズでの得点計算を行う (enterRoundEnd を idempotent に呼ぶ、インタフェース互換)。
func (g *Ombre) ScoreRound() {
	if g.phase != OmbrePhaseRoundEnd {
		return
	}
	g.enterRoundEnd()
}

// --- Trick / play helpers ---

// validatePlay マストフォローを検証する (切り札グループを 1 スートとして扱う)。
func (g *Ombre) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadEff := ombreEffectiveSuit(g.currentTrick[0].Card, g.trumpSuit)
	if ombreEffectiveSuit(card, g.trumpSuit) != leadEff && g.playerHasEffSuit(playerIdx, leadEff) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasEffSuit プレイヤーが指定の実効スート (平札スート or 切り札グループ) を持っているか。
func (g *Ombre) playerHasEffSuit(playerIdx, eff int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if ombreEffectiveSuit(p.GetCard(i), g.trumpSuit) == eff {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する。切り札があれば最強切り札、無ければリードスートの最強札が勝つ。
func (g *Ombre) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadEff := ombreEffectiveSuit(g.currentTrick[0].Card, g.trumpSuit)
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerStr := ombreCardStrength(g.currentTrick[0].Card, g.trumpSuit)
	for _, tc := range g.currentTrick[1:] {
		eff := ombreEffectiveSuit(tc.Card, g.trumpSuit)
		if eff != ombreTrumpGroup && eff != leadEff {
			continue // 場外の平札は勝てない
		}
		if s := ombreCardStrength(tc.Card, g.trumpSuit); s > winnerStr {
			winnerIdx = tc.PlayerIdx
			winnerStr = s
		}
	}
	return winnerIdx
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Ombre) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// isCoalition playerIdx が連合 (非オンブル) 側か。
func (g *Ombre) isCoalition(playerIdx int) bool {
	return playerIdx != g.ombreIdx
}

// sameSide a と b が同じ陣営 (両方オンブル or 両方連合) か。
func (g *Ombre) sameSide(a, b int) bool {
	return g.isCoalition(a) == g.isCoalition(b)
}

// --- Card ranking (Ombre matador ranking, inline) ---

// ombreTrumpGroup 実効スートとしての「切り札グループ」を表す番号 (Joker=0 は 40 枚デッキに現れないため流用)。
const ombreTrumpGroup = 0

// 切り札グループ内の強さ (数値が大きいほど強い)。平札の最大 (10) を大きく上回るよう基準を高く取る。
const (
	ombreStrSpadille = 200 // ♠A
	ombreStrManille  = 199 // 切り札の 7
	ombreStrBasto    = 198 // ♣A
	ombreStrPunto    = 197 // 赤切り札の A
)

// ombreIsTrump card が (trump を切り札としたとき) 切り札か。♠A・♣A は常に切り札。
func ombreIsTrump(c *Card, trump int) bool {
	d, v := c.GetDesign(), c.GetValue()
	if d == CardDesignSpade && v == 1 {
		return true // Spadille
	}
	if d == CardDesignClover && v == 1 {
		return true // Basto
	}
	return d == trump
}

// ombreEffectiveSuit フォロー判定用の実効スート。切り札は ombreTrumpGroup、それ以外は本来のスート。
func ombreEffectiveSuit(c *Card, trump int) int {
	if ombreIsTrump(c, trump) {
		return ombreTrumpGroup
	}
	return c.GetDesign()
}

// ombreCardStrength trump を切り札としたときのカード強さ (全カードの全順序)。任意の切り札 > 任意の平札。
func ombreCardStrength(c *Card, trump int) int {
	d, v := c.GetDesign(), c.GetValue()
	if d == CardDesignSpade && v == 1 {
		return ombreStrSpadille
	}
	if d == trump && v == 7 {
		return ombreStrManille
	}
	if d == CardDesignClover && v == 1 {
		return ombreStrBasto
	}
	if d == trump {
		if v == 1 {
			return ombreStrPunto // 赤切り札の A (黒切り札の A は上で捕捉済み)
		}
		return ombreTrumpSuitRank(v)
	}
	return ombrePlainRank(d, v)
}

// OmbreMatadorRank は札が三大マタドールのどれかを返す。
//
// 1 = スパディーユ (♠A)、2 = マニーユ (切り札の 7)、3 = バスト (♣A)、
// 0 = マタドールでない。**切り札が未確定なら 0。**マニーユは切り札スート
// 次第で決まるので、確定前に一部だけ示すと不揃いな案内になる。
//
// **判定は ombreCardStrength をそのまま読む。**別に条件を書くと、序列を
// 変えたときに表示だけ古いままになる。
func OmbreMatadorRank(c *Card, trump int) int {
	if c == nil || !ombreValidSuit(trump) {
		return 0
	}
	switch ombreCardStrength(c, trump) {
	case ombreStrSpadille:
		return 1
	case ombreStrManille:
		return 2
	case ombreStrBasto:
		return 3
	}
	return 0
}

// ombreTrumpSuitRank 切り札スートの残り札 K>Q>J>6>5>4>3>2 の強さ。
func ombreTrumpSuitRank(v int) int {
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

// ombrePlainRank 平札 (非切り札) の強さ。伝統的オンブルの非対称ランクを再現する:
// 黒スート (♠♣): K>Q>J>7>6>5>4>3>2>A (A が最弱)。
// 赤スート (♥♦): K>Q>J>A>2>3>4>5>6>7 (A が 4 番目、7 が最弱)。
func ombrePlainRank(d, v int) int {
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
func (g *Ombre) sortAllHands() {
	for _, p := range g.players {
		ombreSortHand(p, g.trumpSuit)
	}
}

// ombreSortHand 手札を実効スート→強さ順にソートする。trump=-1 (未確定) の場合はスート→値で並べる。
func ombreSortHand(p *OmbrePlayer, trump int) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if !ombreValidSuit(trump) {
			if cards[i].GetDesign() != cards[j].GetDesign() {
				return cards[i].GetDesign() < cards[j].GetDesign()
			}
			return cards[i].GetValue() > cards[j].GetValue()
		}
		ei, ej := ombreEffectiveSuit(cards[i], trump), ombreEffectiveSuit(cards[j], trump)
		if ei != ej {
			return ei < ej
		}
		return ombreCardStrength(cards[i], trump) > ombreCardStrength(cards[j], trump)
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー名を返す。
func (g *Ombre) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// ombreBidName ビッドの表示名を返す。
func ombreBidName(bid OmbreBid) string {
	switch bid {
	case OmbreBidEntrar:
		return "entrar"
	case OmbreBidSolo:
		return "solo"
	default:
		return "pass"
	}
}

// ombreSuitName スートの表示名を返す。
func ombreSuitName(suit int) string {
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

// ombreOutcomeName 結果の表示名を返す。
func ombreOutcomeName(o OmbreOutcome) string {
	switch o {
	case OmbreOutcomeSacar:
		return "sacar"
	case OmbreOutcomePuesta:
		return "puesta"
	case OmbreOutcomeCodille:
		return "codille"
	default:
		return "-"
	}
}

// ombreValidSuit suit が有効なスート (1..4) か。
func ombreValidSuit(suit int) bool {
	return suit >= CardDesignSpade && suit <= CardDesignDiamond
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Ombre) indexOfPlayerInTrick(playerIdx int) int {
	for i, tc := range g.currentTrick {
		if tc.PlayerIdx == playerIdx {
			return i
		}
	}
	return -1
}

// findHumanIdx 人間プレイヤーのインデックス (-1=なし)。
func (g *Ombre) findHumanIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// --- CPU AI (play) ---

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Ombre) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == OmbreCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 陣営 (オンブル vs 連合) を意識した戦略プレイ。
func (g *Ombre) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	trump := g.trumpSuit

	// リード: オンブルは強い札で主導、連合は弱い札で温存する。
	if len(g.currentTrick) == 0 {
		if playerIdx == g.ombreIdx {
			return g.maxByStrength(player, valid)
		}
		return g.minByStrength(player, valid)
	}

	leadEff := ombreEffectiveSuit(g.currentTrick[0].Card, trump)
	winnerIdx := g.trickWinner()
	topStr := ombreCardStrength(g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card, trump)
	partnerWinning := g.sameSide(winnerIdx, playerIdx)

	var follows []int
	for _, idx := range valid {
		if ombreEffectiveSuit(player.GetCard(idx), trump) == leadEff {
			follows = append(follows, idx)
		}
	}

	if len(follows) == 0 {
		// ボイド: 最も弱い札を捨てる。
		return g.minByStrength(player, valid)
	}

	winners := ombreFilter(follows, func(idx int) bool {
		return ombreCardStrength(player.GetCard(idx), trump) > topStr
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
func (g *Ombre) minByStrength(player *OmbrePlayer, indices []int) int {
	best := indices[0]
	bestScore := ombreCardStrength(player.GetCard(best), g.trumpSuit)
	for _, idx := range indices[1:] {
		if s := ombreCardStrength(player.GetCard(idx), g.trumpSuit); s < bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// maxByStrength 強さが最大となるインデックスを返す。
func (g *Ombre) maxByStrength(player *OmbrePlayer, indices []int) int {
	best := indices[0]
	bestScore := ombreCardStrength(player.GetCard(best), g.trumpSuit)
	for _, idx := range indices[1:] {
		if s := ombreCardStrength(player.GetCard(idx), g.trumpSuit); s > bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// ombreFilter 述語を満たすインデックスを抽出する。
func ombreFilter(indices []int, pred func(int) bool) []int {
	var out []int
	for _, idx := range indices {
		if pred(idx) {
			out = append(out, idx)
		}
	}
	return out
}

// --- Hint ---

// GetHint 人間プレイヤーの手番における推奨アクションを返す。
func (g *Ombre) GetHint() *OmbreHint {
	human := g.findHumanIdx()
	if human < 0 {
		return nil
	}
	switch g.phase {
	case OmbrePhaseBid:
		if g.currentBidderIdx != human {
			return nil
		}
		bid := g.cpuChooseBidForHint(human)
		return &OmbreHint{Reason: ombreBidHintReason(bid)}
	case OmbrePhasePlay:
		if g.currentPlayerIdx != human {
			return nil
		}
		valid := g.getValidPlayIndices(human)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuPlaySmart(human, valid)
		return &OmbreHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
	}
	return nil
}

// cpuChooseBidForHint ヒント用にビッド推奨を計算する (Easy 難易度でも強度から推奨する)。
func (g *Ombre) cpuChooseBidForHint(playerIdx int) OmbreBid {
	_, strength := g.handBestTrump(playerIdx)
	highest := g.highestBid()
	if strength >= 20 && OmbreBidSolo > highest {
		return OmbreBidSolo
	}
	if strength >= 13 && OmbreBidEntrar > highest {
		return OmbreBidEntrar
	}
	return OmbreBidNone
}

// ombreBidHintReason ビッド推奨に対応するヒント理由キーを返す。
func ombreBidHintReason(bid OmbreBid) string {
	switch bid {
	case OmbreBidSolo:
		return "bid_solo"
	case OmbreBidEntrar:
		return "bid_entrar"
	default:
		return "bid_pass"
	}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Ombre) playHintReason(playerIdx, chosenIdx int) string {
	trump := g.trumpSuit
	if len(g.currentTrick) == 0 {
		if playerIdx == g.ombreIdx {
			return "lead_high"
		}
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadEff := ombreEffectiveSuit(g.currentTrick[0].Card, trump)
	if ombreEffectiveSuit(card, trump) != leadEff {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	topStr := ombreCardStrength(g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card, trump)
	if ombreCardStrength(card, trump) > topStr {
		return "follow_win"
	}
	if g.sameSide(winnerIdx, playerIdx) {
		return "give_partner"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Ombre) GetPhase() OmbrePhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Ombre) SetPhase(phase OmbrePhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Ombre) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Ombre) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Ombre) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Ombre) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Ombre) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Ombre) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Ombre) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Ombre) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Ombre) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Ombre) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Ombre) GetDealerIdx() int { return g.dealerIdx }

// GetForehandIdx forehand インデックス取得
func (g *Ombre) GetForehandIdx() int { return g.forehandIdx }

// GetOmbreIdx オンブルインデックス取得 (-1=未確定)
func (g *Ombre) GetOmbreIdx() int { return g.ombreIdx }

// SetOmbreIdx オンブルインデックス設定 (テスト用)
func (g *Ombre) SetOmbreIdx(idx int) { g.ombreIdx = idx }

// GetWinningBid 確定ビッド取得
func (g *Ombre) GetWinningBid() OmbreBid { return g.winningBid }

// SetWinningBid 確定ビッド設定 (テスト用)
func (g *Ombre) SetWinningBid(b OmbreBid) { g.winningBid = b }

// GetTrumpSuit 切り札スート取得 (-1=未確定, 1..4)
func (g *Ombre) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *Ombre) SetTrumpSuit(s int) { g.trumpSuit = s }

// GetCurrentBidderIdx 現在のビッド手番インデックス取得
func (g *Ombre) GetCurrentBidderIdx() int { return g.currentBidderIdx }

// GetPlayerScores プレイヤー別累積点取得
func (g *Ombre) GetPlayerScores() [OmbrePlayerCnt]int { return g.playerScores }

// SetPlayerScores プレイヤー別累積点設定 (テスト用)
func (g *Ombre) SetPlayerScores(s [OmbrePlayerCnt]int) { g.playerScores = s }

// GetOutcome 直近ディールの結果取得
func (g *Ombre) GetOutcome() OmbreOutcome { return g.outcome }

// GetResult 人間視点のマッチ結果取得
func (g *Ombre) GetResult() OmbreResult { return g.result }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Ombre) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *Ombre) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayerCnt プレイヤー数取得
func (g *Ombre) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Ombre) GetPlayer(i int) *OmbrePlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番 (プレイ) が人間か。
func (g *Ombre) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn 現在のビッド手番が人間か。
func (g *Ombre) IsHumanBidTurn() bool {
	if g.phase != OmbrePhaseBid {
		return false
	}
	if g.currentBidderIdx < 0 || g.currentBidderIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentBidderIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Ombre) GetConfig() OmbreConfig { return g.config }

// SetConfig 設定変更
func (g *Ombre) SetConfig(cfg OmbreConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Ombre) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != OmbrePhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// ombreJSON is the JSON wire format for Ombre.
type ombreJSON struct {
	TrumpCards       *TrumpCards              `json:"tc"`
	Players          []*OmbrePlayer           `json:"ps"`
	Config           OmbreConfig              `json:"cf"`
	Phase            OmbrePhase               `json:"ph"`
	RoundNumber      int                      `json:"rn"`
	TrickNumber      int                      `json:"tn"`
	CurrentPlayerIdx int                      `json:"ci"`
	CurrentTrick     []*TrickCard             `json:"ct"`
	LeadPlayerIdx    int                      `json:"li"`
	DealerIdx        int                      `json:"di"`
	ForehandIdx      int                      `json:"fh"`
	OmbreIdx         int                      `json:"om"`
	WinningBid       OmbreBid                 `json:"wb"`
	TrumpSuit        int                      `json:"ts"`
	CurrentBidderIdx int                      `json:"cbi"`
	Bids             [OmbrePlayerCnt]OmbreBid `json:"bd"`
	BidTrump         [OmbrePlayerCnt]int      `json:"bt"`
	BidActed         [OmbrePlayerCnt]bool     `json:"ba"`
	PlayerScores     [OmbrePlayerCnt]int      `json:"sc"`
	LastTrickWinner  int                      `json:"lt"`
	Outcome          OmbreOutcome             `json:"oc"`
	Result           OmbreResult              `json:"rs"`
	Scored           bool                     `json:"sd"`
	GameEndFlag      bool                     `json:"ge"`
	WinnerPlayer     int                      `json:"wp"`
	ActionLog        []*ActionLogEntry        `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Ombre) MarshalJSON() ([]byte, error) {
	return json.Marshal(ombreJSON{
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
		OmbreIdx:         g.ombreIdx,
		WinningBid:       g.winningBid,
		TrumpSuit:        g.trumpSuit,
		CurrentBidderIdx: g.currentBidderIdx,
		Bids:             g.bids,
		BidTrump:         g.bidTrump,
		BidActed:         g.bidActed,
		PlayerScores:     g.playerScores,
		LastTrickWinner:  g.lastTrickWinner,
		Outcome:          g.outcome,
		Result:           g.result,
		Scored:           g.scored,
		GameEndFlag:      g.gameEndFlag,
		WinnerPlayer:     g.winnerPlayer,
		ActionLog:        g.actionLog,
	})
}

// ombreMaxSliceLen caps slice sizes during deserialisation.
const ombreMaxSliceLen = 5000

// errOmbreOversized is the single sentinel error for oversized input arrays.
var errOmbreOversized = errors.New("ombre: input array exceeds maximum allowed size")

// errOmbreInvalidPlayers is returned when restored state lacks exactly OmbrePlayerCnt players.
var errOmbreInvalidPlayers = errors.New("ombre: invalid player count")

// errOmbreInvalidTrick is returned when a restored trick card or its card is nil / out of range.
var errOmbreInvalidTrick = errors.New("ombre: invalid trick card")

// errOmbreInvalidIndex is returned when a restored index field is out of range.
var errOmbreInvalidIndex = errors.New("ombre: index field out of range")

// errOmbreInvalidPhase is returned when a restored phase is out of range.
var errOmbreInvalidPhase = errors.New("ombre: phase out of range")

// errOmbreInvalidBid is returned when a restored bid value is out of range.
var errOmbreInvalidBid = errors.New("ombre: bid value out of range")

// errOmbreInvalidTrump is returned when a restored trump suit is out of range.
var errOmbreInvalidTrump = errors.New("ombre: trump suit out of range")

// errOmbreInvalidOutcome is returned when a restored outcome or result value is out of range.
var errOmbreInvalidOutcome = errors.New("ombre: outcome/result value out of range")

// ombreInRange reports whether v is in [0, OmbrePlayerCnt).
func ombreInRange(v int) bool { return v >= 0 && v < OmbrePlayerCnt }

// ombreInRangeOrUnset reports whether v is -1 (unset) or in [0, OmbrePlayerCnt).
func ombreInRangeOrUnset(v int) bool { return v == -1 || ombreInRange(v) }

// ombreValidBid reports whether b is a defined bid value.
func ombreValidBid(b OmbreBid) bool {
	return b >= OmbreBidNone && b <= OmbreBidSolo
}

// ombreTrumpInRangeOrUnset reports whether s is -1 (unset) or a valid suit (1..4).
func ombreTrumpInRangeOrUnset(s int) bool { return s == -1 || ombreValidSuit(s) }

// UnmarshalJSON implements json.Unmarshaler.
func (g *Ombre) UnmarshalJSON(data []byte) error {
	var j ombreJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > ombreMaxSliceLen || len(j.CurrentTrick) > ombreMaxSliceLen ||
		len(j.ActionLog) > ombreMaxSliceLen {
		return errOmbreOversized
	}
	if len(j.Players) != OmbrePlayerCnt {
		return errOmbreInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errOmbreInvalidPlayers
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return errOmbreInvalidTrick
		}
		if !ombreInRange(tc.PlayerIdx) {
			return errOmbreInvalidTrick
		}
	}
	// 範囲必須のインデックス [0, PlayerCnt)。
	if !ombreInRange(j.CurrentPlayerIdx) || !ombreInRange(j.DealerIdx) ||
		!ombreInRange(j.ForehandIdx) || !ombreInRange(j.CurrentBidderIdx) {
		return errOmbreInvalidIndex
	}
	// -1 (未設定) 許容のインデックス。
	if !ombreInRangeOrUnset(j.LeadPlayerIdx) || !ombreInRangeOrUnset(j.OmbreIdx) ||
		!ombreInRangeOrUnset(j.LastTrickWinner) || !ombreInRangeOrUnset(j.WinnerPlayer) {
		return errOmbreInvalidIndex
	}
	// フェーズ依存の厳格化: play 以降は ombre・lead・trump が確定していなければ
	// 後続処理で g.players[-1] / g.trumpSuit を参照して panic するため確定を要求する。
	if j.Phase >= OmbrePhasePlay {
		if !ombreInRange(j.OmbreIdx) || !ombreInRange(j.LeadPlayerIdx) {
			return errOmbreInvalidIndex
		}
		if !ombreValidSuit(j.TrumpSuit) {
			return errOmbreInvalidTrump
		}
	}
	if int(j.Phase) < OmbrePhaseMin || int(j.Phase) > OmbrePhaseMax {
		return errOmbreInvalidPhase
	}
	if !ombreValidBid(j.WinningBid) {
		return errOmbreInvalidBid
	}
	for _, b := range j.Bids {
		if !ombreValidBid(b) {
			return errOmbreInvalidBid
		}
	}
	if !ombreTrumpInRangeOrUnset(j.TrumpSuit) {
		return errOmbreInvalidTrump
	}
	// bidTrump は -1/0 (未設定) または有効スート (1..4)。範囲外のみ拒否する
	// (finalizeAuction は無効な bidTrump を自動で選び直すため厳格化は不要)。
	for _, t := range j.BidTrump {
		if t < -1 || t > CardDesignDiamond {
			return errOmbreInvalidTrump
		}
	}
	if j.Outcome < OmbreOutcomeNone || j.Outcome > OmbreOutcomeCodille {
		return errOmbreInvalidOutcome
	}
	if j.Result < OmbreResultLose || j.Result > OmbreResultWin {
		return errOmbreInvalidOutcome
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newOmbreDeck()
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
	g.ombreIdx = j.OmbreIdx
	g.winningBid = j.WinningBid
	g.trumpSuit = j.TrumpSuit
	g.currentBidderIdx = j.CurrentBidderIdx
	g.bids = j.Bids
	g.bidTrump = j.BidTrump
	g.bidActed = j.BidActed
	g.playerScores = j.PlayerScores
	g.lastTrickWinner = j.LastTrickWinner
	g.outcome = j.Outcome
	g.result = j.Result
	g.scored = j.Scored
	g.gameEndFlag = j.GameEndFlag
	g.winnerPlayer = j.WinnerPlayer
	g.actionLog = j.ActionLog
	return nil
}
