//go:build !js || !wasm || extra

// Package domain カラブレセッラ / テルツィーリオ (Calabresella / Terziglio) のドメインモデル。
//
// Calabresella はイタリア・カラブリア地方の 3 人用トリックテイキングゲームで、
// トレセッテ (Tressette) 系の切り札なしゲーム。1 人のソリスト (soloist) が 2 人の
// 連合 (coalition) と対戦する。40 枚デッキ (8,9,10 を除く A,2..7,J,Q,K × 4 スート、
// トレセッテと同一) から各自 12 枚を配り (36 枚)、残り 4 枚を伏せて monte (widow) とする。
//
// ビッド: 各プレイヤーは順に pass / chiamo (call, ステーク 1) / solo (ステーク 2) を宣言でき、
// 最高の宣言をしたプレイヤーがソリストとなる。全員パスなら forehand が chiamo を引き受ける。
// chiamo または solo でソリストが確定すると、ソリストは monte 4 枚を取り (16 枚)、12 枚まで
// 4 枚を捨てる (discard)。
//
// トリック: トレセッテのランク・得点。切り札なし・マストフォロー。リードスートの最強札が勝つ。
//
//	ランク (強い順): 3 > 2 > A > K > Q > J > 7 > 6 > 5 > 4
//	得点 (1/3 点単位): A=3、2/3/J/Q/K=1、その他=0。最終トリック勝者に ultima ボーナス 1 (=1/3 点)。
//	1 ラウンドの合計は 33 (=11 点)。
//
// 結果: ソリストが過半 (18/33 以上、= 6 点以上) を獲得すればソリストの勝ち。ステーク分の点数が
// ソリストと連合の間で移動する (ソリスト勝ち = +stake、負け = -stake、連合側は逆符号)。累積点が
// 目標点に達したプレイヤーが勝者となる。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// CalabresellaPlayerCnt プレイヤー数 (人間 1 + CPU 2)
const CalabresellaPlayerCnt = 3

// CalabresellaHandSize 各プレイヤーの配り札枚数 (monte 交換前)
const CalabresellaHandSize = 12

// CalabresellaMonteSize monte (widow) の枚数
const CalabresellaMonteSize = 4

// CalabresellaTrickCount 1 ラウンドのトリック数 (12 枚 × 3 人 = 36 枚 / 3 = 12)
const CalabresellaTrickCount = 12

// CalabresellaUltimaThirds 最終トリック勝者へのボーナス (1/3 点 = 1)
const CalabresellaUltimaThirds = 1

// CalabresellaRoundThirds 1 ラウンドで奪い合う得点の総和 (1/3 点単位 = 11 点)
const CalabresellaRoundThirds = 33

// CalabresellaWinThirds ソリストが勝つために必要な 1/3 点 (過半 = 18/33)
const CalabresellaWinThirds = 18

// CalabresellaWinTarget マッチ勝利に必要な累積点 (既定)
const CalabresellaWinTarget = 21

// CalabresellaBid ビッド宣言
type CalabresellaBid int

// Calabresella のビッド定数 (数値が大きいほど高い宣言)
const (
	// CalabresellaBidNone 未宣言 (パス相当の初期値)
	CalabresellaBidNone CalabresellaBid = 0
	// CalabresellaBidChiamo chiamo (call) — monte を交換できる。ステーク 1。
	CalabresellaBidChiamo CalabresellaBid = 1
	// CalabresellaBidSolo solo — より高いステーク。ステーク 2。
	CalabresellaBidSolo CalabresellaBid = 2
)

// CalabresellaPhase ゲームフェーズ
type CalabresellaPhase int

// Calabresella のフェーズ定数
const (
	// CalabresellaPhaseBid ビッド (auction) フェーズ
	CalabresellaPhaseBid CalabresellaPhase = 0
	// CalabresellaPhaseDiscard monte 交換 (捨て札) フェーズ
	CalabresellaPhaseDiscard CalabresellaPhase = 1
	// CalabresellaPhasePlay トリックプレイフェーズ
	CalabresellaPhasePlay CalabresellaPhase = 2
	// CalabresellaPhaseTrickEnd トリック終了フェーズ
	CalabresellaPhaseTrickEnd CalabresellaPhase = 3
	// CalabresellaPhaseRoundEnd ラウンド終了フェーズ
	CalabresellaPhaseRoundEnd CalabresellaPhase = 4
	// CalabresellaPhaseGameEnd ゲーム終了フェーズ
	CalabresellaPhaseGameEnd CalabresellaPhase = 5
)

// CalabresellaPhaseMin フェーズ下限 (検証用)
const CalabresellaPhaseMin = int(CalabresellaPhaseBid)

// CalabresellaPhaseMax フェーズ上限 (検証用)
const CalabresellaPhaseMax = int(CalabresellaPhaseGameEnd)

// CalabresellaHint ヒント情報
type CalabresellaHint struct {
	CardIndices []int  // 推奨カードインデックス (discard / play フェーズ)
	Reason      string // ヒント理由キー
}

// CalabresellaTrickCard トリック中の 1 枚
type CalabresellaTrickCard struct {
	PlayerIdx int   `json:"pi"`
	Card      *Card `json:"c"`
}

// Calabresella カラブレセッラのゲームクラス
type Calabresella struct {
	trumpCards       *TrumpCards
	players          []*CalabresellaPlayer
	config           CalabresellaConfig
	phase            CalabresellaPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*CalabresellaTrickCard
	leadPlayerIdx    int
	dealerIdx        int
	forehandIdx      int                                    // ディーラーの左隣 (ビッド開始 & 最初のリード)
	soloistIdx       int                                    // ソリスト (-1=未確定)
	winningBid       CalabresellaBid                        // 確定したビッド (ソリストの宣言)
	currentBidderIdx int                                    // 現在ビッド中のプレイヤー (bid フェーズ)
	bids             [CalabresellaPlayerCnt]CalabresellaBid // 各プレイヤーの宣言
	bidActed         [CalabresellaPlayerCnt]bool            // 各プレイヤーが宣言済みか
	monte            []*Card                                // monte (widow) 4 枚
	monteTaken       bool                                   // ソリストが monte を取得済みか
	discardCount     int                                    // discard で捨てた枚数 (0..CalabresellaMonteSize)
	playerScores     [CalabresellaPlayerCnt]int             // 累積ゲーム点
	roundThirds      [CalabresellaPlayerCnt]int             // 現ラウンドのプレイヤー別 1/3 点
	lastTrickWinner  int                                    // 最終トリック勝者 (-1=未確定)
	gameEndFlag      bool
	winnerPlayer     int // -1=未確定
	actionLog        []*ActionLogEntry
}

// NewCalabresella コンストラクタ
func NewCalabresella(trumpCards *TrumpCards, players []*CalabresellaPlayer, config CalabresellaConfig) *Calabresella {
	return &Calabresella{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		winnerPlayer:    -1,
		lastTrickWinner: -1,
		soloistIdx:      -1,
	}
}

// NewDefaultCalabresella 標準の 3 人構成 (人間 1, CPU 2) と既定設定で生成する。
func NewDefaultCalabresella() *Calabresella {
	players := make([]*CalabresellaPlayer, CalabresellaPlayerCnt)
	players[0] = NewCalabresellaPlayer(true)
	for i := 1; i < CalabresellaPlayerCnt; i++ {
		players[i] = NewCalabresellaPlayer(false)
	}
	return NewCalabresella(newCalabresellaDeck(), players, DefaultCalabresellaConfig())
}

// newCalabresellaDeck Calabresella 用 40 枚デッキ (A,2..7,J,Q,K × 4 スート、8,9,10 除外) を生成する。
// トレセッテと同一構成。NewTrumpCardsBriscola は build-tag 無しの TrumpCards.go にあり extra ワーカー
// からも到達可能。casino タグの Tressette ヘルパーは extra からは到達不能なため利用しない。
func newCalabresellaDeck() *TrumpCards {
	return NewTrumpCardsBriscola()
}

// Reset ゲーム初期化
func (g *Calabresella) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.playerScores = [CalabresellaPlayerCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Calabresella) NextRound() {
	if g.phase != CalabresellaPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % CalabresellaPlayerCnt
	g.startRound()
}

// startRound 手札を配り、monte を伏せ、ビッドフェーズを開始する。
func (g *Calabresella) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.roundThirds = [CalabresellaPlayerCnt]int{}
	g.lastTrickWinner = -1
	g.soloistIdx = -1
	g.winningBid = CalabresellaBidNone
	g.monteTaken = false
	g.discardCount = 0
	g.monte = nil
	g.bids = [CalabresellaPlayerCnt]CalabresellaBid{}
	g.bidActed = [CalabresellaPlayerCnt]bool{}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()

	g.forehandIdx = (g.dealerIdx + 1) % CalabresellaPlayerCnt
	g.sortAllHands()

	// ビッド開始: forehand から。
	g.currentBidderIdx = g.forehandIdx
	g.phase = CalabresellaPhaseBid
}

// deal 各プレイヤーへ 12 枚を配り、残り 4 枚を monte にする。
func (g *Calabresella) deal() {
	for i := 0; i < CalabresellaHandSize; i++ {
		for j := 0; j < CalabresellaPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % CalabresellaPlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
	g.monte = make([]*Card, 0, CalabresellaMonteSize)
	for i := 0; i < CalabresellaMonteSize; i++ {
		if c := g.trumpCards.DrawCard(); c != nil {
			g.monte = append(g.monte, c)
		}
	}
}

// --- Bidding ---

// PlayerBid 人間がビッドする。bid は pass(None)/chiamo/solo のいずれか。
func (g *Calabresella) PlayerBid(bid CalabresellaBid) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CalabresellaPhaseBid {
		return ErrWrongPhase
	}
	if !g.players[g.currentBidderIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	if !g.isBidLegal(bid) {
		return NewDomainError(ErrInvalidPlay, "現在の最高ビッドを上回る宣言が必要です")
	}
	g.applyBid(g.currentBidderIdx, bid)
	return nil
}

// CpuBid 現在のビッド手番が CPU の場合に 1 回ビッドする。
func (g *Calabresella) CpuBid() {
	if g.gameEndFlag || g.phase != CalabresellaPhaseBid {
		return
	}
	idx := g.currentBidderIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	g.applyBid(idx, g.cpuChooseBid(idx))
}

// isBidLegal bid が合法か (pass は常に可、chiamo/solo は現最高ビッドを上回る必要)。
func (g *Calabresella) isBidLegal(bid CalabresellaBid) bool {
	if bid == CalabresellaBidNone {
		return true
	}
	if bid != CalabresellaBidChiamo && bid != CalabresellaBidSolo {
		return false
	}
	return bid > g.highestBid()
}

// highestBid これまでに宣言された最高ビッドを返す。
func (g *Calabresella) highestBid() CalabresellaBid {
	best := CalabresellaBidNone
	for _, b := range g.bids {
		if b > best {
			best = b
		}
	}
	return best
}

// applyBid ビッドを適用し、全員が宣言し終えたら auction を締める。
func (g *Calabresella) applyBid(playerIdx int, bid CalabresellaBid) {
	g.bids[playerIdx] = bid
	g.bidActed[playerIdx] = true
	if bid == CalabresellaBidNone {
		g.appendLog(playerIdx, "bid_pass",
			fmt.Sprintf("%s passes", g.playerName(playerIdx)), nil)
	} else {
		g.appendLog(playerIdx, "bid",
			fmt.Sprintf("%s bids %s", g.playerName(playerIdx), calabresellaBidName(bid)), nil)
	}

	if g.allBidsActed() {
		g.finalizeAuction()
		return
	}
	g.currentBidderIdx = g.nextBidder(playerIdx)
}

// allBidsActed 全員が宣言済みか。
func (g *Calabresella) allBidsActed() bool {
	for _, acted := range g.bidActed {
		if !acted {
			return false
		}
	}
	return true
}

// nextBidder playerIdx の次でまだ宣言していないプレイヤーを返す。
func (g *Calabresella) nextBidder(playerIdx int) int {
	for i := 1; i <= CalabresellaPlayerCnt; i++ {
		cand := (playerIdx + i) % CalabresellaPlayerCnt
		if !g.bidActed[cand] {
			return cand
		}
	}
	return playerIdx
}

// finalizeAuction 最高ビッドを宣言したプレイヤーをソリストに確定し、discard フェーズへ進む。
// 同点は forehand から時計回りで最初に宣言したプレイヤーが優先される。
func (g *Calabresella) finalizeAuction() {
	best := g.highestBid()
	soloist := -1
	if best > CalabresellaBidNone {
		for i := 0; i < CalabresellaPlayerCnt; i++ {
			cand := (g.forehandIdx + i) % CalabresellaPlayerCnt
			if g.bids[cand] == best {
				soloist = cand
				break
			}
		}
	}
	if soloist < 0 {
		// 全員パス: forehand が最低ビッド (chiamo) を引き受ける。
		soloist = g.forehandIdx
		best = CalabresellaBidChiamo
	}
	g.soloistIdx = soloist
	g.winningBid = best
	g.appendLog(soloist, "soloist",
		fmt.Sprintf("%s is soloist with %s", g.playerName(soloist), calabresellaBidName(best)), nil)

	g.startDiscard()
}

// cpuChooseBid CPU が手札強度からビッドを選ぶ。
func (g *Calabresella) cpuChooseBid(playerIdx int) CalabresellaBid {
	if g.config.CpuDifficulty == CalabresellaCpuDifficultyEasy {
		return CalabresellaBidNone
	}
	strength := g.handBidStrength(playerIdx)
	highest := g.highestBid()
	// 非常に強い手札は solo、まずまずなら chiamo。
	if strength >= 26 && CalabresellaBidSolo > highest {
		return CalabresellaBidSolo
	}
	if strength >= 18 && CalabresellaBidChiamo > highest {
		return CalabresellaBidChiamo
	}
	return CalabresellaBidNone
}

// handBidStrength 手札のビッド強度を見積もる (高ランク札とエースの多さ)。
func (g *Calabresella) handBidStrength(playerIdx int) int {
	p := g.players[playerIdx]
	total := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		total += calabresellaStrength(p.GetCard(i).GetValue())
	}
	return total
}

// --- Discard (monte exchange) ---

// startDiscard ソリストへ monte を渡し、捨て札フェーズへ移る。CPU ソリストは自動で捨てる。
func (g *Calabresella) startDiscard() {
	g.phase = CalabresellaPhaseDiscard
	g.monteTaken = true
	// monte の内容は取得時点で全員へ公開される情報なので、棋譜に残して
	// Web/CUI プレゼンターが伏せ札の中身を明示できるようにする。
	revealed := append([]*Card(nil), g.monte...)
	for _, c := range g.monte {
		g.players[g.soloistIdx].AddCard(c)
	}
	g.monte = nil
	calabresellaSortHand(g.players[g.soloistIdx])
	g.appendLog(g.soloistIdx, "monte_take",
		fmt.Sprintf("%s takes the monte", g.playerName(g.soloistIdx)), revealed)
	g.discardCount = 0
	g.currentPlayerIdx = g.soloistIdx

	if !g.players[g.soloistIdx].GetIsHuman() {
		g.cpuDiscardAll()
	}
}

// PlayerDiscard 人間ソリストが捨て札で 1 枚を捨てる。CalabresellaMonteSize 回呼ぶと交換完了。
func (g *Calabresella) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CalabresellaPhaseDiscard {
		return ErrWrongPhase
	}
	if !g.players[g.soloistIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	soloist := g.players[g.soloistIdx]
	if cardIndex < 0 || cardIndex >= soloist.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	g.discardOne(cardIndex)
	if g.discardCount >= CalabresellaMonteSize {
		g.startPlay()
	}
	return nil
}

// discardOne ソリストの cardIndex の札を捨てる (獲得トリックとして保持し得点計算に含める)。
func (g *Calabresella) discardOne(cardIndex int) {
	card := g.players[g.soloistIdx].RemoveCard(cardIndex)
	// 捨て札はソリストの獲得札として扱う (トレセッテ系の慣習: 交換で捨てた札の得点はソリストに帰属)。
	g.players[g.soloistIdx].AddTrick([]*Card{card})
	g.roundThirds[g.soloistIdx] += calabresellaThirds(card.GetValue())
	g.appendLog(g.soloistIdx, "discard",
		fmt.Sprintf("%s discards %s", g.playerName(g.soloistIdx), cardStr(card)), []*Card{card})
	g.discardCount++
}

// cpuDiscardAll CPU ソリストが最も弱い札から CalabresellaMonteSize 枚を捨てる。
func (g *Calabresella) cpuDiscardAll() {
	for g.discardCount < CalabresellaMonteSize {
		idx := g.cpuSelectDiscard()
		g.discardOne(idx)
	}
	g.startPlay()
}

// cpuSelectDiscard CPU ソリストが捨てる札 (最も弱く得点も低い札) のインデックスを選ぶ。
func (g *Calabresella) cpuSelectDiscard() int {
	p := g.players[g.soloistIdx]
	best := 0
	bestScore := 1 << 30
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		// 得点札は残したい: 得点を重み付けして保護する。
		score := calabresellaThirds(c.GetValue())*100 + calabresellaStrength(c.GetValue())
		if score < bestScore {
			bestScore = score
			best = i
		}
	}
	return best
}

// startPlay 捨て札完了後、プレイフェーズを開始する (ソリストが forehand としてリード)。
func (g *Calabresella) startPlay() {
	g.sortAllHands()
	g.leadPlayerIdx = g.soloistIdx
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = CalabresellaPhasePlay
}

// --- Play ---

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Calabresella) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != CalabresellaPhasePlay {
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
func (g *Calabresella) CpuPlay() {
	if g.gameEndFlag || g.phase != CalabresellaPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	cardIdx := g.cpuSelectPlayCard(idx)
	played := g.players[idx].RemoveCard(cardIdx)
	g.playCard(idx, played)
}

// playCard カードをプレイする共通処理。
func (g *Calabresella) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &CalabresellaTrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", g.playerName(playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == CalabresellaPlayerCnt {
		g.phase = CalabresellaPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % CalabresellaPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定し、得点を加算する。
func (g *Calabresella) ResolveTrick() {
	if g.phase != CalabresellaPhaseTrickEnd || len(g.currentTrick) != CalabresellaPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	thirds := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		thirds += calabresellaThirds(tc.Card.GetValue())
	}
	g.players[winnerIdx].AddTrick(trickCards)
	bonus := ""
	if g.trickNumber >= CalabresellaTrickCount {
		thirds += CalabresellaUltimaThirds
		bonus = " +ultima"
	}
	g.roundThirds[winnerIdx] += thirds
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (+%d/3%s)", g.playerName(winnerIdx), g.trickNumber, thirds, bonus), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= CalabresellaTrickCount {
		g.lastTrickWinner = winnerIdx
		g.phase = CalabresellaPhaseRoundEnd
	} else {
		g.phase = CalabresellaPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Calabresella) NextTrick() {
	if g.phase != CalabresellaPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = CalabresellaPhasePlay
}

// ScoreRound ラウンド結果を判定し、累積点へ加算してマッチ終了を判定する。
// ソリストが過半 (CalabresellaWinThirds 以上) を獲得すれば勝ち。ステーク分の点数が
// ソリストと連合の 2 人の間で移動する。
func (g *Calabresella) ScoreRound() {
	if g.phase != CalabresellaPhaseRoundEnd {
		return
	}
	stake := int(g.winningBid)
	if stake < 1 {
		stake = 1
	}
	soloistThirds := g.roundThirds[g.soloistIdx]
	soloistWon := soloistThirds >= CalabresellaWinThirds
	for i := 0; i < CalabresellaPlayerCnt; i++ {
		if i == g.soloistIdx {
			if soloistWon {
				g.playerScores[i] += stake * CalabresellaCoalitionSize
			} else {
				g.playerScores[i] -= stake * CalabresellaCoalitionSize
			}
		} else {
			if soloistWon {
				g.playerScores[i] -= stake
			} else {
				g.playerScores[i] += stake
			}
		}
	}
	result := "loses"
	if soloistWon {
		result = "wins"
	}
	g.appendLog(-1, "round_score",
		fmt.Sprintf("round %d: soloist(%s) %s (%d/3, stake=%d)",
			g.roundNumber, g.playerName(g.soloistIdx), result, soloistThirds, stake), nil)
	g.checkGameEnd()
}

// checkGameEnd 目標点到達でマッチ終了を判定する。
func (g *Calabresella) checkGameEnd() {
	leader, best := -1, -1<<30
	for i := 0; i < CalabresellaPlayerCnt; i++ {
		if g.playerScores[i] > best {
			best = g.playerScores[i]
			leader = i
		}
	}
	if best >= g.config.TargetPoints && leader >= 0 {
		g.gameEndFlag = true
		g.winnerPlayer = leader
		g.phase = CalabresellaPhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", g.playerName(leader)), nil)
	}
}

// --- Trick / play helpers ---

// validatePlay マストフォロー (切り札なし) を検証する。
func (g *Calabresella) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && g.playerHasSuit(playerIdx, leadSuit) {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	return nil
}

// playerHasSuit プレイヤーが指定スートのカードを持っているか。
func (g *Calabresella) playerHasSuit(playerIdx, design int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する。切り札がないため、リードスートの最強札が勝つ。
func (g *Calabresella) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	winnerStrength := calabresellaStrength(g.currentTrick[0].Card.GetValue())
	for _, tc := range g.currentTrick[1:] {
		if tc.Card.GetDesign() == leadSuit && calabresellaStrength(tc.Card.GetValue()) > winnerStrength {
			winnerIdx = tc.PlayerIdx
			winnerStrength = calabresellaStrength(tc.Card.GetValue())
		}
	}
	return winnerIdx
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Calabresella) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	var valid []int
	for i := 0; i < player.GetCardsSize(); i++ {
		if g.validatePlay(playerIdx, player.GetCard(i)) == nil {
			valid = append(valid, i)
		}
	}
	return valid
}

// isCoalition playerIdx が連合 (非ソリスト) 側か。
func (g *Calabresella) isCoalition(playerIdx int) bool {
	return playerIdx != g.soloistIdx
}

// sameSide a と b が同じ陣営 (両方ソリスト or 両方連合) か。
func (g *Calabresella) sameSide(a, b int) bool {
	return g.isCoalition(a) == g.isCoalition(b)
}

// --- Card helpers (Tressette rank/points, inline; casino-tagged Tressette は extra から到達不能) ---

// calabresellaStrength トリックの強さ。3 が最強 (9)、4 が最弱 (0)。
//
//	3 > 2 > A > K > Q > J > 7 > 6 > 5 > 4
func calabresellaStrength(value int) int {
	switch value {
	case 3:
		return 9
	case 2:
		return 8
	case 1: // Ace
		return 7
	case 13: // King
		return 6
	case 12: // Queen
		return 5
	case 11: // Jack
		return 4
	case 7:
		return 3
	case 6:
		return 2
	case 5:
		return 1
	default: // 4
		return 0
	}
}

// calabresellaThirds カードの得点を 1/3 点 単位で返す。A=3、2/3/J/Q/K=1、その他=0。
func calabresellaThirds(value int) int {
	switch value {
	case 1: // Ace = 1 point = 3 thirds
		return 3
	case 2, 3, 11, 12, 13: // 2,3,J,Q,K = 1/3 point each
		return 1
	default:
		return 0
	}
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Calabresella) sortAllHands() {
	for _, p := range g.players {
		calabresellaSortHand(p)
	}
}

// calabresellaSortHand 手札をスート→強さ順にソートする。
func calabresellaSortHand(p *CalabresellaPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return calabresellaStrength(cards[i].GetValue()) > calabresellaStrength(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// playerName プレイヤー名を返す。
func (g *Calabresella) playerName(idx int) string {
	if idx < 0 || idx >= len(g.players) {
		return fmt.Sprintf("Player %d", idx)
	}
	if g.players[idx].GetIsHuman() {
		return "You"
	}
	return fmt.Sprintf("CPU %d", idx)
}

// calabresellaBidName ビッドの表示名を返す。
func calabresellaBidName(bid CalabresellaBid) string {
	switch bid {
	case CalabresellaBidChiamo:
		return "chiamo"
	case CalabresellaBidSolo:
		return "solo"
	default:
		return "pass"
	}
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Calabresella) indexOfPlayerInTrick(playerIdx int) int {
	for i, tc := range g.currentTrick {
		if tc.PlayerIdx == playerIdx {
			return i
		}
	}
	return -1
}

// findHumanIdx 人間プレイヤーのインデックス (-1=なし)。
func (g *Calabresella) findHumanIdx() int {
	for i, p := range g.players {
		if p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// appendLog 棋譜にエントリを追加する。
func (g *Calabresella) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: len(g.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- CPU AI (play) ---

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Calabresella) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == CalabresellaCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 陣営 (ソリスト vs 連合) を意識した戦略プレイ。
func (g *Calabresella) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]

	// リード: 得点・強さの低い札を出して温存する。
	if len(g.currentTrick) == 0 {
		return g.minBy(player, valid, func(c *Card) int {
			return calabresellaThirds(c.GetValue())*100 + calabresellaStrength(c.GetValue())
		})
	}

	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.trickWinner()
	topStrength := calabresellaStrength(g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card.GetValue())
	partnerWinning := g.sameSide(winnerIdx, playerIdx)
	trickThirds := 0
	for _, tc := range g.currentTrick {
		trickThirds += calabresellaThirds(tc.Card.GetValue())
	}

	var follows []int
	for _, idx := range valid {
		if player.GetCard(idx).GetDesign() == leadSuit {
			follows = append(follows, idx)
		}
	}

	if len(follows) == 0 {
		// ボイド: 得点・強さの低い札を捨てて温存する。
		return g.minBy(player, valid, func(c *Card) int {
			return calabresellaThirds(c.GetValue())*100 + calabresellaStrength(c.GetValue())
		})
	}

	winners := calabresellaFilter(follows, func(idx int) bool {
		return calabresellaStrength(player.GetCard(idx).GetValue()) > topStrength
	})

	if partnerWinning {
		// 味方が勝っている: 上書きせず得点札を渡す。
		nonWinners := calabresellaFilter(follows, func(idx int) bool {
			return calabresellaStrength(player.GetCard(idx).GetValue()) < topStrength
		})
		if len(nonWinners) > 0 {
			return g.maxBy(player, nonWinners, func(c *Card) int {
				return calabresellaThirds(c.GetValue())*100 - calabresellaStrength(c.GetValue())
			})
		}
		return g.minBy(player, follows, func(c *Card) int { return calabresellaStrength(c.GetValue()) })
	}

	// 相手が勝っている: 得点があり勝てるなら最小限の札で取りに行く。
	if trickThirds > 0 && len(winners) > 0 {
		return g.minBy(player, winners, func(c *Card) int { return calabresellaStrength(c.GetValue()) })
	}
	// 取れない/取る価値がない: 得点・強さの低い札でダックする。
	return g.minBy(player, follows, func(c *Card) int {
		return calabresellaThirds(c.GetValue())*100 + calabresellaStrength(c.GetValue())
	})
}

// minBy score が最小となるインデックスを返す。
func (g *Calabresella) minBy(player *CalabresellaPlayer, indices []int, score func(*Card) int) int {
	best := indices[0]
	bestScore := score(player.GetCard(best))
	for _, idx := range indices[1:] {
		if s := score(player.GetCard(idx)); s < bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// maxBy score が最大となるインデックスを返す。
func (g *Calabresella) maxBy(player *CalabresellaPlayer, indices []int, score func(*Card) int) int {
	best := indices[0]
	bestScore := score(player.GetCard(best))
	for _, idx := range indices[1:] {
		if s := score(player.GetCard(idx)); s > bestScore {
			bestScore = s
			best = idx
		}
	}
	return best
}

// calabresellaFilter 述語を満たすインデックスを抽出する。
func calabresellaFilter(indices []int, pred func(int) bool) []int {
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
func (g *Calabresella) GetHint() *CalabresellaHint {
	human := g.findHumanIdx()
	if human < 0 {
		return nil
	}
	switch g.phase {
	case CalabresellaPhaseBid:
		if g.currentBidderIdx != human {
			return nil
		}
		bid := g.cpuChooseBidForHint(human)
		return &CalabresellaHint{Reason: calabresellaBidHintReason(bid)}
	case CalabresellaPhaseDiscard:
		if g.soloistIdx != human {
			return nil
		}
		idx := g.cpuSelectDiscard()
		return &CalabresellaHint{CardIndices: []int{idx}, Reason: "discard_low"}
	case CalabresellaPhasePlay:
		if g.currentPlayerIdx != human {
			return nil
		}
		valid := g.getValidPlayIndices(human)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuPlaySmart(human, valid)
		return &CalabresellaHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
	}
	return nil
}

// cpuChooseBidForHint ヒント用にビッド推奨を計算する (Easy 難易度でも強度から推奨する)。
func (g *Calabresella) cpuChooseBidForHint(playerIdx int) CalabresellaBid {
	strength := g.handBidStrength(playerIdx)
	highest := g.highestBid()
	if strength >= 26 && CalabresellaBidSolo > highest {
		return CalabresellaBidSolo
	}
	if strength >= 18 && CalabresellaBidChiamo > highest {
		return CalabresellaBidChiamo
	}
	return CalabresellaBidNone
}

// calabresellaBidHintReason ビッド推奨に対応するヒント理由キーを返す。
func calabresellaBidHintReason(bid CalabresellaBid) string {
	switch bid {
	case CalabresellaBidSolo:
		return "bid_solo"
	case CalabresellaBidChiamo:
		return "bid_chiamo"
	default:
		return "bid_pass"
	}
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Calabresella) playHintReason(playerIdx, chosenIdx int) string {
	if len(g.currentTrick) == 0 {
		return "lead_low"
	}
	card := g.players[playerIdx].GetCard(chosenIdx)
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	topStrength := calabresellaStrength(g.currentTrick[g.indexOfPlayerInTrick(winnerIdx)].Card.GetValue())
	if calabresellaStrength(card.GetValue()) > topStrength {
		return "follow_win"
	}
	if g.sameSide(winnerIdx, playerIdx) {
		return "give_partner"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Calabresella) GetPhase() CalabresellaPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Calabresella) SetPhase(phase CalabresellaPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Calabresella) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Calabresella) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Calabresella) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Calabresella) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Calabresella) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Calabresella) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Calabresella) GetCurrentTrick() []*CalabresellaTrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Calabresella) SetCurrentTrick(trick []*CalabresellaTrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Calabresella) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Calabresella) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Calabresella) GetDealerIdx() int { return g.dealerIdx }

// GetForehandIdx forehand インデックス取得
func (g *Calabresella) GetForehandIdx() int { return g.forehandIdx }

// GetSoloistIdx ソリストインデックス取得 (-1=未確定)
func (g *Calabresella) GetSoloistIdx() int { return g.soloistIdx }

// SetSoloistIdx ソリストインデックス設定 (テスト用)
func (g *Calabresella) SetSoloistIdx(idx int) { g.soloistIdx = idx }

// GetWinningBid 確定ビッド取得
func (g *Calabresella) GetWinningBid() CalabresellaBid { return g.winningBid }

// SetWinningBid 確定ビッド設定 (テスト用)
func (g *Calabresella) SetWinningBid(b CalabresellaBid) { g.winningBid = b }

// GetCurrentBidderIdx 現在のビッド手番インデックス取得
func (g *Calabresella) GetCurrentBidderIdx() int { return g.currentBidderIdx }

// GetPlayerScores プレイヤー別累積点取得
func (g *Calabresella) GetPlayerScores() [CalabresellaPlayerCnt]int { return g.playerScores }

// SetPlayerScores プレイヤー別累積点設定 (テスト用)
func (g *Calabresella) SetPlayerScores(s [CalabresellaPlayerCnt]int) { g.playerScores = s }

// GetRoundThirds 現ラウンドのプレイヤー別 1/3 点取得
func (g *Calabresella) GetRoundThirds() [CalabresellaPlayerCnt]int { return g.roundThirds }

// SetRoundThirds 現ラウンドのプレイヤー別 1/3 点設定 (テスト用)
func (g *Calabresella) SetRoundThirds(s [CalabresellaPlayerCnt]int) { g.roundThirds = s }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Calabresella) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *Calabresella) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayerCnt プレイヤー数取得
func (g *Calabresella) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Calabresella) GetPlayer(i int) *CalabresellaPlayer {
	if i < 0 || i >= len(g.players) {
		return nil
	}
	return g.players[i]
}

// IsHumanTurn 現在の手番 (プレイ) が人間か。
func (g *Calabresella) IsHumanTurn() bool {
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// IsHumanBidTurn 現在のビッド手番が人間か。
func (g *Calabresella) IsHumanBidTurn() bool {
	if g.phase != CalabresellaPhaseBid {
		return false
	}
	if g.currentBidderIdx < 0 || g.currentBidderIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentBidderIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Calabresella) GetConfig() CalabresellaConfig { return g.config }

// SetConfig 設定変更
func (g *Calabresella) SetConfig(cfg CalabresellaConfig) { g.config = cfg }

// GetActionLog 棋譜取得
func (g *Calabresella) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Calabresella) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != CalabresellaPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// calabresellaJSON is the JSON wire format for Calabresella.
type calabresellaJSON struct {
	TrumpCards       *TrumpCards                            `json:"tc"`
	Players          []*CalabresellaPlayer                  `json:"ps"`
	Config           CalabresellaConfig                     `json:"cf"`
	Phase            CalabresellaPhase                      `json:"ph"`
	RoundNumber      int                                    `json:"rn"`
	TrickNumber      int                                    `json:"tn"`
	CurrentPlayerIdx int                                    `json:"ci"`
	CurrentTrick     []*CalabresellaTrickCard               `json:"ct"`
	LeadPlayerIdx    int                                    `json:"li"`
	DealerIdx        int                                    `json:"di"`
	ForehandIdx      int                                    `json:"fh"`
	SoloistIdx       int                                    `json:"so"`
	WinningBid       CalabresellaBid                        `json:"wb"`
	CurrentBidderIdx int                                    `json:"cbi"`
	Bids             [CalabresellaPlayerCnt]CalabresellaBid `json:"bd"`
	BidActed         [CalabresellaPlayerCnt]bool            `json:"ba"`
	Monte            []*Card                                `json:"mo"`
	MonteTaken       bool                                   `json:"mt"`
	DiscardCount     int                                    `json:"dn"`
	PlayerScores     [CalabresellaPlayerCnt]int             `json:"sc"`
	RoundThirds      [CalabresellaPlayerCnt]int             `json:"rt"`
	LastTrickWinner  int                                    `json:"lt"`
	GameEndFlag      bool                                   `json:"ge"`
	WinnerPlayer     int                                    `json:"wp"`
	ActionLog        []*ActionLogEntry                      `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Calabresella) MarshalJSON() ([]byte, error) {
	return json.Marshal(calabresellaJSON{
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
		SoloistIdx:       g.soloistIdx,
		WinningBid:       g.winningBid,
		CurrentBidderIdx: g.currentBidderIdx,
		Bids:             g.bids,
		BidActed:         g.bidActed,
		Monte:            g.monte,
		MonteTaken:       g.monteTaken,
		DiscardCount:     g.discardCount,
		PlayerScores:     g.playerScores,
		RoundThirds:      g.roundThirds,
		LastTrickWinner:  g.lastTrickWinner,
		GameEndFlag:      g.gameEndFlag,
		WinnerPlayer:     g.winnerPlayer,
		ActionLog:        g.actionLog,
	})
}

// calabresellaMaxSliceLen caps slice sizes during deserialisation.
const calabresellaMaxSliceLen = 5000

// errCalabresellaOversized is the single sentinel error for oversized input arrays.
var errCalabresellaOversized = errors.New("calabresella: input array exceeds maximum allowed size")

// errCalabresellaInvalidPlayers is returned when restored state lacks exactly CalabresellaPlayerCnt players.
var errCalabresellaInvalidPlayers = errors.New("calabresella: invalid player count")

// errCalabresellaInvalidTrick is returned when a restored trick card or its card is nil / out of range.
var errCalabresellaInvalidTrick = errors.New("calabresella: invalid trick card")

// errCalabresellaInvalidMonte is returned when a restored monte card is nil.
var errCalabresellaInvalidMonte = errors.New("calabresella: invalid monte card")

// errCalabresellaInvalidIndex is returned when a restored index field is out of range.
var errCalabresellaInvalidIndex = errors.New("calabresella: index field out of range")

// errCalabresellaInvalidPhase is returned when a restored phase is out of range.
var errCalabresellaInvalidPhase = errors.New("calabresella: phase out of range")

// errCalabresellaInvalidBid is returned when a restored bid value is out of range.
var errCalabresellaInvalidBid = errors.New("calabresella: bid value out of range")

// calabresellaInRange reports whether v is in [0, CalabresellaPlayerCnt).
func calabresellaInRange(v int) bool { return v >= 0 && v < CalabresellaPlayerCnt }

// calabresellaInRangeOrUnset reports whether v is -1 (unset) or in [0, CalabresellaPlayerCnt).
func calabresellaInRangeOrUnset(v int) bool { return v == -1 || calabresellaInRange(v) }

// calabresellaValidBid reports whether b is a defined bid value.
func calabresellaValidBid(b CalabresellaBid) bool {
	return b >= CalabresellaBidNone && b <= CalabresellaBidSolo
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *Calabresella) UnmarshalJSON(data []byte) error {
	var j calabresellaJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > calabresellaMaxSliceLen || len(j.CurrentTrick) > calabresellaMaxSliceLen ||
		len(j.ActionLog) > calabresellaMaxSliceLen || len(j.Monte) > calabresellaMaxSliceLen {
		return errCalabresellaOversized
	}
	if len(j.Players) != CalabresellaPlayerCnt {
		return errCalabresellaInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errCalabresellaInvalidPlayers
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return errCalabresellaInvalidTrick
		}
		if !calabresellaInRange(tc.PlayerIdx) {
			return errCalabresellaInvalidTrick
		}
	}
	for _, c := range j.Monte {
		if c == nil {
			return errCalabresellaInvalidMonte
		}
	}
	// 範囲必須のインデックス [0, PlayerCnt)。
	if !calabresellaInRange(j.CurrentPlayerIdx) || !calabresellaInRange(j.DealerIdx) ||
		!calabresellaInRange(j.ForehandIdx) || !calabresellaInRange(j.CurrentBidderIdx) {
		return errCalabresellaInvalidIndex
	}
	// -1 (未設定) 許容のインデックス。
	if !calabresellaInRangeOrUnset(j.LeadPlayerIdx) || !calabresellaInRangeOrUnset(j.SoloistIdx) ||
		!calabresellaInRangeOrUnset(j.LastTrickWinner) || !calabresellaInRangeOrUnset(j.WinnerPlayer) {
		return errCalabresellaInvalidIndex
	}
	// フェーズ依存の厳格化: discard 以降は soloist、play 以降は lead が確定していな
	// ければ後続処理で g.players[-1] にアクセスして panic するため確定を要求する。
	if j.Phase >= CalabresellaPhaseDiscard && !calabresellaInRange(j.SoloistIdx) {
		return errCalabresellaInvalidIndex
	}
	if j.Phase >= CalabresellaPhasePlay && !calabresellaInRange(j.LeadPlayerIdx) {
		return errCalabresellaInvalidIndex
	}
	if j.DiscardCount < 0 || j.DiscardCount > CalabresellaMonteSize {
		return errCalabresellaInvalidIndex
	}
	if int(j.Phase) < CalabresellaPhaseMin || int(j.Phase) > CalabresellaPhaseMax {
		return errCalabresellaInvalidPhase
	}
	if !calabresellaValidBid(j.WinningBid) {
		return errCalabresellaInvalidBid
	}
	for _, b := range j.Bids {
		if !calabresellaValidBid(b) {
			return errCalabresellaInvalidBid
		}
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newCalabresellaDeck()
	}
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.roundNumber = j.RoundNumber
	g.trickNumber = j.TrickNumber
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.currentTrick = j.CurrentTrick
	if g.currentTrick == nil {
		g.currentTrick = make([]*CalabresellaTrickCard, 0)
	}
	g.leadPlayerIdx = j.LeadPlayerIdx
	g.dealerIdx = j.DealerIdx
	g.forehandIdx = j.ForehandIdx
	g.soloistIdx = j.SoloistIdx
	g.winningBid = j.WinningBid
	g.currentBidderIdx = j.CurrentBidderIdx
	g.bids = j.Bids
	g.bidActed = j.BidActed
	g.monte = j.Monte
	g.monteTaken = j.MonteTaken
	g.discardCount = j.DiscardCount
	g.playerScores = j.PlayerScores
	g.roundThirds = j.RoundThirds
	g.lastTrickWinner = j.LastTrickWinner
	g.gameEndFlag = j.GameEndFlag
	g.winnerPlayer = j.WinnerPlayer
	g.actionLog = j.ActionLog
	return nil
}
