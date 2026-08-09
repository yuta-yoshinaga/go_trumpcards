//go:build !js || !wasm || extra

// Package domain サウザンド / トゥシオンツ (Thousand / Tysiąc) のドメインモデル。
//
// Tysiąc はポーランド・東欧の 3 人用ビッド式トリックテイキングゲーム。24 枚デッキ
// (9, J, Q, K, 10, A の 4 スート) から各自 7 枚を配り、残り 3 枚は talon (widow) として
// 伏せる。プレイヤーは順に 100 点から +10 刻みでビッドし (パスで降りる)、最後まで残った
// 1 人が contract を持つ declarer となる。declarer は talon 3 枚を取り (10 枚)、各相手へ
// 1 枚ずつ渡す (8 枚 × 3)。リード時に同スートの K+Q を揃えていれば、その K か Q をリード
// することで結婚 (marriage) を宣言でき、そのスートが即座に切り札となり結婚点 (♠40/♣60/
// ♦80/♥100) を得る。ラウンド開始時は切り札なし。マストフォロー + ボイド時は切り札強制 +
// オーバートランプ義務。ラウンド終了時、declarer は (カード点 + 結婚点) が contract 以上で
// +contract、未満で −contract。非 declarer は (カード点 + 結婚点) を 10 単位に丸めて加算。
// 累積 1000 点で勝利。
//
// カードポイント: A=11, 10=10, K=4, Q=3, J=2, 9=0 (合計 120 点)。
// トリック強度 (切り札・非切り札共通): A > 10 > K > Q > J > 9。
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
)

// TysiacPlayerCnt プレイヤー数 (人間 1 + CPU 2)
const TysiacPlayerCnt = 3

// TysiacHandSize 各プレイヤーの配り札枚数 (talon 交換前)
const TysiacHandSize = 7

// TysiacTalonSize talon (widow) の枚数
const TysiacTalonSize = 3

// TysiacTrickCount 1 ラウンドのトリック数 (8 枚 × 3 人 = 24 枚 / 3 = 8)
const TysiacTrickCount = 8

// TysiacMinBid 最低ビッド (contract)
const TysiacMinBid = 100

// TysiacBidStep ビッドの刻み幅
const TysiacBidStep = 10

// TysiacWinTarget マッチ勝利に必要な累積点
const TysiacWinTarget = 1000

// TysiacPhase ゲームフェーズ
type TysiacPhase int

// Tysiąc のフェーズ定数
const (
	// TysiacPhaseBid ビッド (auction) フェーズ
	TysiacPhaseBid TysiacPhase = 0
	// TysiacPhaseTalon talon 交換フェーズ
	TysiacPhaseTalon TysiacPhase = 1
	// TysiacPhasePlay トリックプレイフェーズ
	TysiacPhasePlay TysiacPhase = 2
	// TysiacPhaseTrickEnd トリック終了フェーズ
	TysiacPhaseTrickEnd TysiacPhase = 3
	// TysiacPhaseRoundEnd ラウンド終了フェーズ
	TysiacPhaseRoundEnd TysiacPhase = 4
	// TysiacPhaseGameEnd ゲーム終了フェーズ
	TysiacPhaseGameEnd TysiacPhase = 5
)

// TysiacPhaseMin フェーズ下限 (検証用)
const TysiacPhaseMin = int(TysiacPhaseBid)

// TysiacPhaseMax フェーズ上限 (検証用)
const TysiacPhaseMax = int(TysiacPhaseGameEnd)

// TysiacHint ヒント情報
type TysiacHint struct {
	CardIndices []int  // 推奨カードインデックス (play フェーズ)
	Reason      string // ヒント理由キー
}

// Tysiac サウザンドのゲームクラス
type Tysiac struct {
	trumpCards       *TrumpCards
	players          []*TysiacPlayer
	config           TysiacConfig
	phase            TysiacPhase
	roundNumber      int
	trickNumber      int
	currentPlayerIdx int
	currentTrick     []*TrickCard
	leadPlayerIdx    int
	dealerIdx        int
	forehandIdx      int                   // ディーラーの左隣 (ビッド開始 & talon 交換後の最初のリード)
	declarerIdx      int                   // contract を持つ declarer (-1=未確定)
	contract         int                   // 確定した contract 値
	currentBid       int                   // 現在の最高ビッド額
	bidPassed        [TysiacPlayerCnt]bool // 既にパスしたか
	talon            []*Card               // talon (widow) 3 枚
	talonTaken       bool                  // declarer が talon を取得済みか
	discardCount     int                   // talon 交換で配り終えた枚数 (0..TysiacTalonSize)
	trumpSuit        int                   // 切り札スート (0=未設定)
	playerScores     [TysiacPlayerCnt]int  // 累積ゲーム点
	roundCardPts     [TysiacPlayerCnt]int  // 現ラウンドのプレイヤー別カード得点
	roundMarriage    [TysiacPlayerCnt]int  // 現ラウンドのプレイヤー別結婚点
	lastTrickWinner  int                   // 最終トリック勝者 (-1=未確定)
	gameEndFlag      bool
	winnerPlayer     int // -1=未確定
	actionLogBase
}

// NewTysiac コンストラクタ
func NewTysiac(trumpCards *TrumpCards, players []*TysiacPlayer, config TysiacConfig) *Tysiac {
	return &Tysiac{
		trumpCards:      trumpCards,
		players:         players,
		config:          config,
		winnerPlayer:    -1,
		lastTrickWinner: -1,
		declarerIdx:     -1,
	}
}

// NewDefaultTysiac 標準の 3 人構成 (人間 1, CPU 2) と既定設定で生成する。
func NewDefaultTysiac() *Tysiac {
	players := make([]*TysiacPlayer, TysiacPlayerCnt)
	players[0] = NewTysiacPlayer(true)
	for i := 1; i < TysiacPlayerCnt; i++ {
		players[i] = NewTysiacPlayer(false)
	}
	return NewTysiac(newTysiacDeck(), players, DefaultTysiacConfig())
}

// newTysiacDeck Tysiąc 用 24 枚デッキ (9, J, Q, K, 10, A = 値 9,11,12,13,10,1 × 4 スート) を生成する。
// classic タグの NewTrumpCardsBelote 等は extra ワーカーからは到達不能なため、ここでインライン構築する。
func newTysiacDeck() *TrumpCards {
	// NewTrumpCardsEuchre は build-tag 無しの TrumpCards.go にあり extra ワーカーからも
	// 到達可能で、Tysiąc と完全に同一の 24 枚構成 (A,9,10,J,Q,K × 4 スート) を返す。
	return NewTrumpCardsEuchre()
}

// Reset ゲーム初期化
func (g *Tysiac) Reset() {
	g.gameEndFlag = false
	g.winnerPlayer = -1
	g.roundNumber = 1
	g.dealerIdx = 0
	g.playerScores = [TysiacPlayerCnt]int{}
	g.actionLog = nil
	g.startRound()
}

// NextRound 次のラウンドを開始する
func (g *Tysiac) NextRound() {
	if g.phase != TysiacPhaseRoundEnd {
		return
	}
	g.roundNumber++
	g.dealerIdx = (g.dealerIdx + 1) % TysiacPlayerCnt
	g.startRound()
}

// startRound 手札を配り、talon を伏せ、ビッドフェーズを開始する。
func (g *Tysiac) startRound() {
	g.trickNumber = 1
	g.currentTrick = nil
	g.roundCardPts = [TysiacPlayerCnt]int{}
	g.roundMarriage = [TysiacPlayerCnt]int{}
	g.lastTrickWinner = -1
	g.trumpSuit = 0
	g.declarerIdx = -1
	g.contract = 0
	g.talonTaken = false
	g.discardCount = 0
	g.talon = nil
	g.bidPassed = [TysiacPlayerCnt]bool{}
	for _, p := range g.players {
		p.ResetRound()
	}
	g.trumpCards.Replenish()
	g.trumpCards.Shuffle()
	g.deal()

	g.forehandIdx = (g.dealerIdx + 1) % TysiacPlayerCnt
	g.sortAllHands()

	// ビッド開始: forehand から 100 点。
	g.currentBid = TysiacMinBid
	g.currentPlayerIdx = g.forehandIdx
	g.phase = TysiacPhaseBid
}

// deal 各プレイヤーへ 7 枚を配り、残り 3 枚を talon にする。
func (g *Tysiac) deal() {
	for i := 0; i < TysiacHandSize; i++ {
		for j := 0; j < TysiacPlayerCnt; j++ {
			idx := (g.dealerIdx + 1 + j) % TysiacPlayerCnt
			if c := g.trumpCards.DrawCard(); c != nil {
				g.players[idx].AddCard(c)
			}
		}
	}
	g.talon = make([]*Card, 0, TysiacTalonSize)
	for i := 0; i < TysiacTalonSize; i++ {
		if c := g.trumpCards.DrawCard(); c != nil {
			g.talon = append(g.talon, c)
		}
	}
}

// --- Bidding ---

// activeBidders まだパスしていないビッダー数を返す。
func (g *Tysiac) activeBidders() int {
	n := 0
	for _, passed := range g.bidPassed {
		if !passed {
			n++
		}
	}
	return n
}

// PlayerBid 人間がビッドする。raise=true で +10、false でパス。
func (g *Tysiac) PlayerBid(raise bool) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TysiacPhaseBid {
		return ErrWrongPhase
	}
	if !g.players[g.currentPlayerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	g.applyBid(g.currentPlayerIdx, raise)
	return nil
}

// CpuBid 現在の手番が CPU の場合に 1 回ビッドする。
func (g *Tysiac) CpuBid() {
	if g.gameEndFlag || g.phase != TysiacPhaseBid {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	raise := g.cpuWantsRaise(idx)
	g.applyBid(idx, raise)
}

// applyBid ビッド (raise/pass) を適用し、必要なら declarer を確定して talon フェーズへ進める。
func (g *Tysiac) applyBid(playerIdx int, raise bool) {
	if raise {
		g.currentBid += TysiacBidStep
		g.appendLog(playerIdx, "bid",
			fmt.Sprintf("%s bids %d", playerName(g.players, playerIdx), g.currentBid), nil)
	} else {
		g.bidPassed[playerIdx] = true
		g.appendLog(playerIdx, "bid_pass",
			fmt.Sprintf("%s passes", playerName(g.players, playerIdx)), nil)
	}

	if g.activeBidders() <= 1 {
		g.finalizeAuction()
		return
	}
	g.currentPlayerIdx = g.nextBidder(playerIdx)
}

// nextBidder playerIdx の次でまだパスしていないプレイヤーを返す。
func (g *Tysiac) nextBidder(playerIdx int) int {
	for i := 1; i <= TysiacPlayerCnt; i++ {
		cand := (playerIdx + i) % TysiacPlayerCnt
		if !g.bidPassed[cand] {
			return cand
		}
	}
	return playerIdx
}

// finalizeAuction ビッドを締め、declarer と contract を確定して talon フェーズへ移行する。
func (g *Tysiac) finalizeAuction() {
	declarer := -1
	for i := 0; i < TysiacPlayerCnt; i++ {
		if !g.bidPassed[i] {
			declarer = i
			break
		}
	}
	if declarer < 0 {
		// 全員パス: forehand が最低 contract を引き受ける。
		declarer = g.forehandIdx
		g.currentBid = TysiacMinBid
	}
	g.declarerIdx = declarer
	g.contract = g.currentBid
	g.appendLog(declarer, "declarer",
		fmt.Sprintf("%s is declarer with contract %d", playerName(g.players, declarer), g.contract), nil)

	g.startTalon()
}

// cpuWantsRaise CPU がさらにビッドを上げたいか判定する。
func (g *Tysiac) cpuWantsRaise(playerIdx int) bool {
	maxBid := g.cpuMaxBid(playerIdx)
	// 現在の最高ビッドが自分のものなら上げない (連続上げ防止)。
	return g.currentBid+TysiacBidStep <= maxBid
}

// cpuMaxBid CPU が手札強度から見積もる上限ビッド額。
func (g *Tysiac) cpuMaxBid(playerIdx int) int {
	p := g.players[playerIdx]
	pts := 0
	for i := 0; i < p.GetCardsSize(); i++ {
		pts += tysiacCardPoints(p.GetCard(i))
	}
	// 手札のカード点 (最大 ~70 程度) + 結婚ポテンシャル。
	marriage := g.handMarriagePotential(playerIdx)
	estimate := pts + marriage/2
	maxBid := TysiacMinBid + (estimate/10)*TysiacBidStep
	// 大きな結婚がなければ控えめに (≤120 目安)。
	cap := TysiacMinBid + 2*TysiacBidStep // 120
	if marriage >= tysiacMarriagePoints(CardDesignDiamond) {
		cap = TysiacMinBid + 6*TysiacBidStep // 160
	}
	if maxBid > cap {
		maxBid = cap
	}
	if g.config.CpuDifficulty == TysiacCpuDifficultyEasy {
		// Easy は強気にならず最低限。
		return TysiacMinBid
	}
	return maxBid
}

// handMarriagePotential 手札中の K+Q ペアによる結婚点合計を返す。
func (g *Tysiac) handMarriagePotential(playerIdx int) int {
	total := 0
	for _, suit := range tysiacSuits() {
		if g.playerHasCard(playerIdx, suit, 13) && g.playerHasCard(playerIdx, suit, 12) {
			total += tysiacMarriagePoints(suit)
		}
	}
	return total
}

// --- Talon exchange ---

// startTalon talon を declarer に渡し交換フェーズへ移る。CPU declarer は自動で交換する。
func (g *Tysiac) startTalon() {
	g.phase = TysiacPhaseTalon
	g.talonTaken = true
	for _, c := range g.talon {
		g.players[g.declarerIdx].AddCard(c)
	}
	g.talon = nil
	tysiacSortHand(g.players[g.declarerIdx])
	g.appendLog(g.declarerIdx, "talon_take",
		fmt.Sprintf("%s takes the talon", playerName(g.players, g.declarerIdx)), nil)
	g.discardCount = 0
	g.currentPlayerIdx = g.declarerIdx

	if !g.players[g.declarerIdx].GetIsHuman() {
		g.cpuDiscardAll()
	}
}

// PlayerDiscard 人間 declarer が talon 交換で 1 枚を次の相手に渡す。2 回呼ぶと交換完了。
func (g *Tysiac) PlayerDiscard(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TysiacPhaseTalon {
		return ErrWrongPhase
	}
	if !g.players[g.declarerIdx].GetIsHuman() {
		return ErrNotHumanTurn
	}
	declarer := g.players[g.declarerIdx]
	if cardIndex < 0 || cardIndex >= declarer.GetCardsSize() {
		return NewDomainError(ErrInvalidCard, "カードインデックスが範囲外です")
	}
	g.giveDiscard(cardIndex)
	if g.discardCount >= TysiacTalonSize-1 {
		g.startPlay()
	}
	return nil
}

// giveDiscard declarer の cardIndex の札を次の相手へ渡す。
func (g *Tysiac) giveDiscard(cardIndex int) {
	// discardCount 0 -> 最初の相手, 1 -> 2 人目の相手。
	recipients := g.opponentsOf(g.declarerIdx)
	recipient := recipients[g.discardCount]
	card := g.players[g.declarerIdx].RemoveCard(cardIndex)
	g.players[recipient].AddCard(card)
	tysiacSortHand(g.players[recipient])
	g.appendLog(g.declarerIdx, "discard",
		fmt.Sprintf("%s gives a card to %s", playerName(g.players, g.declarerIdx), playerName(g.players, recipient)), nil)
	g.discardCount++
}

// opponentsOf playerIdx 以外の 2 人を昇順で返す。
func (g *Tysiac) opponentsOf(playerIdx int) []int {
	var out []int
	for i := 0; i < TysiacPlayerCnt; i++ {
		if i != playerIdx {
			out = append(out, i)
		}
	}
	return out
}

// cpuDiscardAll CPU declarer が最も得点の低い 2 枚を相手へ渡す。
func (g *Tysiac) cpuDiscardAll() {
	for g.discardCount < TysiacTalonSize-1 {
		idx := g.cpuSelectDiscard()
		g.giveDiscard(idx)
	}
	g.startPlay()
}

// cpuSelectDiscard CPU declarer が渡す札 (最も弱い札) のインデックスを選ぶ。
func (g *Tysiac) cpuSelectDiscard() int {
	p := g.players[g.declarerIdx]
	best := 0
	bestScore := 1 << 30
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		// K/Q は結婚維持のため残したい: 高スコアで保護。
		score := tysiacCardPoints(c)*10 + tysiacStrength(c.GetValue())
		if c.GetValue() == 13 || c.GetValue() == 12 {
			score += 100
		}
		if score < bestScore {
			bestScore = score
			best = i
		}
	}
	return best
}

// startPlay talon 交換完了後、プレイフェーズを開始する (declarer が forehand としてリード)。
func (g *Tysiac) startPlay() {
	g.sortAllHands()
	g.leadPlayerIdx = g.declarerIdx
	g.currentPlayerIdx = g.leadPlayerIdx
	g.phase = TysiacPhasePlay
}

// --- Play ---

// PlayerPlay 人間プレイヤーがカードをプレイする。
func (g *Tysiac) PlayerPlay(cardIndex int) error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	if g.phase != TysiacPhasePlay {
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
	g.maybeDeclareMarriage(g.currentPlayerIdx, card)
	played := player.RemoveCard(cardIndex)
	g.playCard(g.currentPlayerIdx, played)
	return nil
}

// CpuPlay 現在の手番が CPU の場合に 1 ターン実行する。
func (g *Tysiac) CpuPlay() {
	if g.gameEndFlag || g.phase != TysiacPhasePlay {
		return
	}
	idx := g.currentPlayerIdx
	if g.players[idx].GetIsHuman() {
		return
	}
	cardIdx := g.cpuSelectPlayCard(idx)
	card := g.players[idx].GetCard(cardIdx)
	// **出せる札が無ければ何もしない。**セレクタは候補ゼロのとき 0 を返し、
	// 手札が空なら GetCard(0) も RemoveCard(0) も nil を返す。ここは RemoveCard
	// より前に card を読むので、ガードも前に置かないと maybeDeclareMarriage が
	// nil を触って HTTP ハンドラごと落ちる (#4606)。
	if card == nil {
		return
	}
	g.maybeDeclareMarriage(idx, card)
	played := g.players[idx].RemoveCard(cardIdx)
	if played == nil {
		return
	}
	g.playCard(idx, played)
}

// maybeDeclareMarriage リード時に K/Q を出し同スートの相棒を持っていれば結婚を宣言し
// そのスートを切り札に設定して結婚点を加算する。
func (g *Tysiac) maybeDeclareMarriage(playerIdx int, card *Card) {
	if len(g.currentTrick) != 0 {
		return // リード時のみ宣言可能
	}
	v := card.GetValue()
	if v != 13 && v != 12 {
		return
	}
	suit := card.GetDesign()
	partner := 12
	if v == 12 {
		partner = 13
	}
	if !g.playerHasCard(playerIdx, suit, partner) {
		return
	}
	g.trumpSuit = suit
	pts := tysiacMarriagePoints(suit)
	g.roundMarriage[playerIdx] += pts
	g.appendLog(playerIdx, "marriage",
		fmt.Sprintf("%s declares a %s marriage (+%d, trump=%s)",
			playerName(g.players, playerIdx), tysiacSuitName(suit), pts, tysiacSuitName(suit)), nil)
}

// playCard カードをプレイする共通処理。
func (g *Tysiac) playCard(playerIdx int, card *Card) {
	g.currentTrick = append(g.currentTrick, &TrickCard{PlayerIdx: playerIdx, Card: card})
	g.appendLog(playerIdx, "play", fmt.Sprintf("%s plays %s", playerName(g.players, playerIdx), cardStr(card)), []*Card{card})

	if len(g.currentTrick) == TysiacPlayerCnt {
		g.phase = TysiacPhaseTrickEnd
	} else {
		g.currentPlayerIdx = (g.currentPlayerIdx + 1) % TysiacPlayerCnt
	}
}

// ResolveTrick トリックを解決して勝者を決定し、得点を加算する。
func (g *Tysiac) ResolveTrick() {
	if g.phase != TysiacPhaseTrickEnd || len(g.currentTrick) != TysiacPlayerCnt {
		return
	}
	winnerIdx := g.trickWinner()
	trickCards := make([]*Card, len(g.currentTrick))
	pts := 0
	for i, tc := range g.currentTrick {
		trickCards[i] = tc.Card
		pts += tysiacCardPoints(tc.Card)
	}
	g.players[winnerIdx].AddTrick(trickCards)
	g.roundCardPts[winnerIdx] += pts
	g.appendLog(winnerIdx, "trick_win",
		fmt.Sprintf("%s wins trick %d (+%d)", playerName(g.players, winnerIdx), g.trickNumber, pts), trickCards)

	g.leadPlayerIdx = winnerIdx
	if g.trickNumber >= TysiacTrickCount {
		g.lastTrickWinner = winnerIdx
		g.phase = TysiacPhaseRoundEnd
	} else {
		g.phase = TysiacPhaseTrickEnd
	}
}

// NextTrick 次のトリックを開始する。
func (g *Tysiac) NextTrick() {
	if g.phase != TysiacPhaseTrickEnd {
		return
	}
	g.currentTrick = nil
	g.currentPlayerIdx = g.leadPlayerIdx
	g.trickNumber++
	g.phase = TysiacPhasePlay
}

// ScoreRound ラウンド結果を判定し、累積点へ加算してマッチ終了を判定する。
func (g *Tysiac) ScoreRound() {
	if g.phase != TysiacPhaseRoundEnd {
		return
	}
	for i := 0; i < TysiacPlayerCnt; i++ {
		total := g.roundCardPts[i] + g.roundMarriage[i]
		if i == g.declarerIdx {
			if total >= g.contract {
				g.playerScores[i] += g.contract
			} else {
				g.playerScores[i] -= g.contract
			}
		} else {
			g.playerScores[i] += tysiacRoundTo10(total)
		}
	}
	g.appendLog(-1, "round_score",
		fmt.Sprintf("round %d scored: declarer(%s) contract=%d",
			g.roundNumber, playerName(g.players, g.declarerIdx), g.contract), nil)
	g.checkGameEnd()
}

// tysiacRoundTo10 最も近い 10 単位に丸める。
func tysiacRoundTo10(n int) int {
	return ((n + 5) / 10) * 10
}

// checkGameEnd 目標点 (1000) 到達でマッチ終了を判定する。
func (g *Tysiac) checkGameEnd() {
	leader, best := -1, -1<<30
	for i := 0; i < TysiacPlayerCnt; i++ {
		if g.playerScores[i] > best {
			best = g.playerScores[i]
			leader = i
		}
	}
	if best >= g.config.TargetPoints && leader >= 0 {
		g.gameEndFlag = true
		g.winnerPlayer = leader
		g.phase = TysiacPhaseGameEnd
		g.appendLog(-1, "game_end", fmt.Sprintf("%s wins the match!", playerName(g.players, leader)), nil)
	}
}

// --- Trick / play helpers ---

// validatePlay マストフォロー + ボイド時の切り札強制 + オーバートランプ義務を検証する。
func (g *Tysiac) validatePlay(playerIdx int, card *Card) error {
	if len(g.currentTrick) == 0 {
		return nil
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	hasLeadSuit := g.playerHasSuit(playerIdx, leadSuit)
	if hasLeadSuit && card.GetDesign() != leadSuit {
		return NewDomainError(ErrInvalidPlay, "リードスートに従ってください")
	}
	if !hasLeadSuit {
		hasTrump := g.trumpSuit != 0 && g.playerHasSuit(playerIdx, g.trumpSuit)
		if hasTrump && card.GetDesign() != g.trumpSuit {
			return NewDomainError(ErrInvalidPlay, "切り札を出してください")
		}
	}
	// オーバートランプ義務: 切り札を出す場合 (ボイド時の強制切り札でも切り札リードへの
	// フォローでも)、場の最高切り札を上回れるなら勝てる札を出さねばならない。
	if g.trumpSuit != 0 && card.GetDesign() == g.trumpSuit {
		highestTrump := g.highestTrumpRankInTrick()
		if highestTrump > 0 && g.canOvertrump(playerIdx, highestTrump) &&
			g.tysiacRank(card) <= highestTrump {
			return NewDomainError(ErrInvalidPlay, "高い切り札で勝ってください")
		}
	}
	return nil
}

// highestTrumpRankInTrick 現在のトリック中の最高切り札ランクを返す (0=切り札なし)。
func (g *Tysiac) highestTrumpRankInTrick() int {
	best := 0
	for _, tc := range g.currentTrick {
		if g.trumpSuit != 0 && tc.Card.GetDesign() == g.trumpSuit {
			if r := g.tysiacRank(tc.Card); r > best {
				best = r
			}
		}
	}
	return best
}

// canOvertrump playerIdx が rank より高い切り札を持つか。
func (g *Tysiac) canOvertrump(playerIdx, rank int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == g.trumpSuit && g.tysiacRank(c) > rank {
			return true
		}
	}
	return false
}

// playerHasSuit プレイヤーが指定スートのカードを持っているか。
func (g *Tysiac) playerHasSuit(playerIdx, design int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		if p.GetCard(i).GetDesign() == design {
			return true
		}
	}
	return false
}

// playerHasCard プレイヤーが指定スート・ランクの札を持っているか。
func (g *Tysiac) playerHasCard(playerIdx, design, value int) bool {
	p := g.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c.GetDesign() == design && c.GetValue() == value {
			return true
		}
	}
	return false
}

// trickWinner トリックの勝者を決定する。切り札があれば最強切り札、なければ
// リードスートの最強札が勝つ。
func (g *Tysiac) trickWinner() int {
	if len(g.currentTrick) == 0 {
		return 0
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winnerIdx := g.currentTrick[0].PlayerIdx
	bestRank := g.tysiacRank(g.currentTrick[0].Card)
	for _, tc := range g.currentTrick[1:] {
		if g.trumpSuit != 0 && tc.Card.GetDesign() == g.trumpSuit {
			// 切り札は常に評価対象。
		} else if tc.Card.GetDesign() != leadSuit {
			continue
		}
		if r := g.tysiacRank(tc.Card); r > bestRank {
			bestRank = r
			winnerIdx = tc.PlayerIdx
		}
	}
	return winnerIdx
}

// tysiacRank トリック比較用ランク。切り札は非切り札より常に強い。
func (g *Tysiac) tysiacRank(card *Card) int {
	if g.trumpSuit != 0 && card.GetDesign() == g.trumpSuit {
		return 100 + tysiacStrength(card.GetValue())
	}
	return tysiacStrength(card.GetValue())
}

// tysiacStrength カード強度。A > 10 > K > Q > J > 9。
func tysiacStrength(value int) int {
	switch value {
	case 1: // Ace
		return 6
	case 10:
		return 5
	case 13: // King
		return 4
	case 12: // Queen
		return 3
	case 11: // Jack
		return 2
	default: // 9
		return 1
	}
}

// tysiacCardPoints カードポイント。A=11, 10=10, K=4, Q=3, J=2, 9=0。
func tysiacCardPoints(card *Card) int {
	switch card.GetValue() {
	case 1:
		return 11
	case 10:
		return 10
	case 13:
		return 4
	case 12:
		return 3
	case 11:
		return 2
	default:
		return 0
	}
}

// tysiacMarriagePoints スート別結婚点。♠=40, ♣=60, ♦=80, ♥=100。
func tysiacMarriagePoints(suit int) int {
	switch suit {
	case CardDesignSpade:
		return 40
	case CardDesignClover:
		return 60
	case CardDesignDiamond:
		return 80
	case CardDesignHeart:
		return 100
	default:
		return 0
	}
}

// tysiacSuits スート一覧を返す。
func tysiacSuits() []int {
	return []int{CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond}
}

// tysiacSuitName スート表示名 (英語) を返す。
func tysiacSuitName(suit int) string {
	switch suit {
	case CardDesignSpade:
		return "Spades"
	case CardDesignClover:
		return "Clubs"
	case CardDesignHeart:
		return "Hearts"
	case CardDesignDiamond:
		return "Diamonds"
	default:
		return "None"
	}
}

// getValidPlayIndices プレイ可能なカードのインデックスリストを返す。
func (g *Tysiac) getValidPlayIndices(playerIdx int) []int {
	player := g.players[playerIdx]
	return collectValidIndices(player.GetCardsSize(), func(i int) bool {
		return g.validatePlay(playerIdx, player.GetCard(i)) == nil
	})
}

// --- Misc helpers ---

// sortAllHands 全プレイヤーの手札をソートする。
func (g *Tysiac) sortAllHands() {
	for _, p := range g.players {
		tysiacSortHand(p)
	}
}

// tysiacSortHand 手札をスート→強さ順にソートする。
func tysiacSortHand(p *TysiacPlayer) {
	cards := make([]*Card, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		cards[i] = p.GetCard(i)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].GetDesign() != cards[j].GetDesign() {
			return cards[i].GetDesign() < cards[j].GetDesign()
		}
		return tysiacStrength(cards[i].GetValue()) > tysiacStrength(cards[j].GetValue())
	})
	p.Reset()
	for _, c := range cards {
		p.AddCard(c)
	}
}

// indexOfPlayerInTrick currentTrick 内で playerIdx の札の位置を返す (-1=なし)。
func (g *Tysiac) indexOfPlayerInTrick(playerIdx int) int {
	for i, tc := range g.currentTrick {
		if tc.PlayerIdx == playerIdx {
			return i
		}
	}
	return -1
}

// trickTopRank 現在のトリック勝者の札のランクを返す。見つからない場合は極小値。
func (g *Tysiac) trickTopRank(winnerIdx int) int {
	idx := g.indexOfPlayerInTrick(winnerIdx)
	if idx < 0 {
		return -1 << 30
	}
	return g.tysiacRank(g.currentTrick[idx].Card)
}

// --- CPU AI (play) ---

// cpuSelectPlayCard CPU がプレイするカードのインデックスを選ぶ。
func (g *Tysiac) cpuSelectPlayCard(playerIdx int) int {
	valid := g.getValidPlayIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == TysiacCpuDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	return g.cpuPlaySmart(playerIdx, valid)
}

// cpuPlaySmart 得点を意識した戦略プレイ。リード時に結婚可能なら K を優先する。
func (g *Tysiac) cpuPlaySmart(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		// 結婚可能なスートがあれば K をリードして宣言する。
		if idx := g.cpuMarriageLead(playerIdx, valid); idx >= 0 {
			return idx
		}
		return pickLowest(player, valid, func(c *Card) int {
			return tysiacCardPoints(c)*100 + g.tysiacRank(c)
		})
	}
	winnerIdx := g.trickWinner()
	topRank := g.trickTopRank(winnerIdx)
	trickPts := 0
	for _, tc := range g.currentTrick {
		trickPts += tysiacCardPoints(tc.Card)
	}
	winners := tysiacFilter(valid, func(idx int) bool { return g.tysiacRank(player.GetCard(idx)) > topRank })
	if trickPts > 0 && len(winners) > 0 {
		return pickLowest(player, winners, func(c *Card) int { return g.tysiacRank(c) })
	}
	return pickLowest(player, valid, func(c *Card) int {
		return tysiacCardPoints(c)*100 + g.tysiacRank(c)
	})
}

// cpuMarriageLead 結婚を宣言できる K のインデックスを返す (-1=なし)。最も高得点のスートを優先する。
func (g *Tysiac) cpuMarriageLead(playerIdx int, valid []int) int {
	player := g.players[playerIdx]
	best := -1
	bestPts := 0
	for _, idx := range valid {
		c := player.GetCard(idx)
		if c.GetValue() != 13 {
			continue
		}
		suit := c.GetDesign()
		if !g.playerHasCard(playerIdx, suit, 12) {
			continue
		}
		if pts := tysiacMarriagePoints(suit); pts > bestPts {
			bestPts = pts
			best = idx
		}
	}
	return best
}

// tysiacFilter 述語を満たすインデックスを抽出する。
func tysiacFilter(indices []int, pred func(int) bool) []int {
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
func (g *Tysiac) GetHint() *TysiacHint {
	human := findHumanIdx(g.players)
	if human < 0 {
		return nil
	}
	switch g.phase {
	case TysiacPhaseBid:
		if g.currentPlayerIdx != human {
			return nil
		}
		if g.cpuWantsRaise(human) {
			return &TysiacHint{Reason: "bid_raise"}
		}
		return &TysiacHint{Reason: "bid_pass"}
	case TysiacPhaseTalon:
		if g.declarerIdx != human {
			return nil
		}
		idx := g.cpuSelectDiscard()
		return &TysiacHint{CardIndices: []int{idx}, Reason: "talon_discard"}
	case TysiacPhasePlay:
		if g.currentPlayerIdx != human {
			return nil
		}
		valid := g.getValidPlayIndices(human)
		if len(valid) == 0 {
			return nil
		}
		idx := g.cpuPlaySmart(human, valid)
		return &TysiacHint{CardIndices: []int{idx}, Reason: g.playHintReason(human, idx)}
	}
	return nil
}

// playHintReason プレイヒントの理由キーを判定する。
func (g *Tysiac) playHintReason(playerIdx, chosenIdx int) string {
	card := g.players[playerIdx].GetCard(chosenIdx)
	if len(g.currentTrick) == 0 {
		if card.GetValue() == 13 && g.playerHasCard(playerIdx, card.GetDesign(), 12) {
			return "lead_marriage"
		}
		return "lead_low"
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	if card.GetDesign() != leadSuit && card.GetDesign() != g.trumpSuit {
		return "discard_low"
	}
	winnerIdx := g.trickWinner()
	if g.tysiacRank(card) > g.trickTopRank(winnerIdx) {
		return "follow_win"
	}
	return "follow_duck"
}

// --- State getters ---

// GetPhase 現在のフェーズ取得
func (g *Tysiac) GetPhase() TysiacPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Tysiac) SetPhase(phase TysiacPhase) { g.phase = phase }

// GetRoundNumber ラウンド番号取得
func (g *Tysiac) GetRoundNumber() int { return g.roundNumber }

// SetRoundNumber ラウンド番号設定 (テスト用)
func (g *Tysiac) SetRoundNumber(n int) { g.roundNumber = n }

// GetTrickNumber トリック番号取得
func (g *Tysiac) GetTrickNumber() int { return g.trickNumber }

// SetTrickNumber トリック番号設定 (テスト用)
func (g *Tysiac) SetTrickNumber(n int) { g.trickNumber = n }

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *Tysiac) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// SetCurrentPlayerIdx プレイヤーインデックス設定 (テスト用)
func (g *Tysiac) SetCurrentPlayerIdx(idx int) { g.currentPlayerIdx = idx }

// GetCurrentTrick 現在のトリック取得
func (g *Tysiac) GetCurrentTrick() []*TrickCard { return g.currentTrick }

// SetCurrentTrick トリック設定 (テスト用)
func (g *Tysiac) SetCurrentTrick(trick []*TrickCard) { g.currentTrick = trick }

// GetLeadPlayerIdx リードプレイヤーインデックス取得
func (g *Tysiac) GetLeadPlayerIdx() int { return g.leadPlayerIdx }

// SetLeadPlayerIdx リードプレイヤーインデックス設定 (テスト用)
func (g *Tysiac) SetLeadPlayerIdx(idx int) { g.leadPlayerIdx = idx }

// GetDealerIdx ディーラーインデックス取得
func (g *Tysiac) GetDealerIdx() int { return g.dealerIdx }

// GetForehandIdx forehand インデックス取得
func (g *Tysiac) GetForehandIdx() int { return g.forehandIdx }

// GetDeclarerIdx declarer インデックス取得 (-1=未確定)
func (g *Tysiac) GetDeclarerIdx() int { return g.declarerIdx }

// SetDeclarerIdx declarer インデックス設定 (テスト用)
func (g *Tysiac) SetDeclarerIdx(idx int) { g.declarerIdx = idx }

// GetContract 確定 contract 取得
func (g *Tysiac) GetContract() int { return g.contract }

// SetContract contract 設定 (テスト用)
func (g *Tysiac) SetContract(c int) { g.contract = c }

// GetCurrentBid 現在の最高ビッド取得
func (g *Tysiac) GetCurrentBid() int { return g.currentBid }

// SetCurrentBid 現在の最高ビッド設定 (テスト用)
func (g *Tysiac) SetCurrentBid(b int) { g.currentBid = b }

// GetTrumpSuit 切り札スート取得 (0=未設定)
func (g *Tysiac) GetTrumpSuit() int { return g.trumpSuit }

// SetTrumpSuit 切り札スート設定 (テスト用)
func (g *Tysiac) SetTrumpSuit(suit int) { g.trumpSuit = suit }

// GetPlayerScores プレイヤー別累積点取得
func (g *Tysiac) GetPlayerScores() [TysiacPlayerCnt]int { return g.playerScores }

// SetPlayerScores プレイヤー別累積点設定 (テスト用)
func (g *Tysiac) SetPlayerScores(s [TysiacPlayerCnt]int) { g.playerScores = s }

// GetRoundCardPoints 現ラウンドのカード得点取得
func (g *Tysiac) GetRoundCardPoints() [TysiacPlayerCnt]int { return g.roundCardPts }

// SetRoundCardPoints 現ラウンドのカード得点設定 (テスト用)
func (g *Tysiac) SetRoundCardPoints(s [TysiacPlayerCnt]int) { g.roundCardPts = s }

// GetRoundMarriage 現ラウンドの結婚点取得
func (g *Tysiac) GetRoundMarriage() [TysiacPlayerCnt]int { return g.roundMarriage }

// SetRoundMarriage 現ラウンドの結婚点設定 (テスト用)
func (g *Tysiac) SetRoundMarriage(s [TysiacPlayerCnt]int) { g.roundMarriage = s }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *Tysiac) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerPlayer 勝利プレイヤー取得 (-1=未確定)
func (g *Tysiac) GetWinnerPlayer() int { return g.winnerPlayer }

// GetPlayerCnt プレイヤー数取得
func (g *Tysiac) GetPlayerCnt() int { return len(g.players) }

// GetPlayer プレイヤー取得
func (g *Tysiac) GetPlayer(i int) *TysiacPlayer {
	return getPlayer(g.players, i)
}

// IsHumanTurn 現在の手番 (プレイ) が人間か。
func (g *Tysiac) IsHumanTurn() bool {
	return isHumanTurn(g.players, g.currentPlayerIdx)
}

// IsHumanBidTurn 現在のビッド手番が人間か。
func (g *Tysiac) IsHumanBidTurn() bool {
	if g.phase != TysiacPhaseBid {
		return false
	}
	if g.currentPlayerIdx < 0 || g.currentPlayerIdx >= len(g.players) {
		return false
	}
	return g.players[g.currentPlayerIdx].GetIsHuman()
}

// GetConfig 設定取得
func (g *Tysiac) GetConfig() TysiacConfig { return g.config }

// SetConfig 設定変更
func (g *Tysiac) SetConfig(cfg TysiacConfig) { g.config = cfg }

// GetPlayableIndices プレイ可能なカードのインデックス一覧を返す。
func (g *Tysiac) GetPlayableIndices(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(g.players) || g.phase != TysiacPhasePlay {
		return nil
	}
	return g.getValidPlayIndices(playerIdx)
}

// --- JSON ---

// tysiacJSON is the JSON wire format for Tysiac.
type tysiacJSON struct {
	TrumpCards       *TrumpCards           `json:"tc"`
	Players          []*TysiacPlayer       `json:"ps"`
	Config           TysiacConfig          `json:"cf"`
	Phase            TysiacPhase           `json:"ph"`
	RoundNumber      int                   `json:"rn"`
	TrickNumber      int                   `json:"tn"`
	CurrentPlayerIdx int                   `json:"ci"`
	CurrentTrick     []*TrickCard          `json:"ct"`
	LeadPlayerIdx    int                   `json:"li"`
	DealerIdx        int                   `json:"di"`
	ForehandIdx      int                   `json:"fh"`
	DeclarerIdx      int                   `json:"dc"`
	Contract         int                   `json:"co"`
	CurrentBid       int                   `json:"cb"`
	BidPassed        [TysiacPlayerCnt]bool `json:"bp"`
	Talon            []*Card               `json:"tl"`
	TalonTaken       bool                  `json:"tt"`
	DiscardCount     int                   `json:"dn"`
	TrumpSuit        int                   `json:"ts"`
	PlayerScores     [TysiacPlayerCnt]int  `json:"sc"`
	RoundCardPts     [TysiacPlayerCnt]int  `json:"rp"`
	RoundMarriage    [TysiacPlayerCnt]int  `json:"rm"`
	LastTrickWinner  int                   `json:"lt"`
	GameEndFlag      bool                  `json:"ge"`
	WinnerPlayer     int                   `json:"wp"`
	ActionLog        []*ActionLogEntry     `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Tysiac) MarshalJSON() ([]byte, error) {
	return json.Marshal(tysiacJSON{
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
		Contract:         g.contract,
		CurrentBid:       g.currentBid,
		BidPassed:        g.bidPassed,
		Talon:            g.talon,
		TalonTaken:       g.talonTaken,
		DiscardCount:     g.discardCount,
		TrumpSuit:        g.trumpSuit,
		PlayerScores:     g.playerScores,
		RoundCardPts:     g.roundCardPts,
		RoundMarriage:    g.roundMarriage,
		LastTrickWinner:  g.lastTrickWinner,
		GameEndFlag:      g.gameEndFlag,
		WinnerPlayer:     g.winnerPlayer,
		ActionLog:        g.actionLog,
	})
}

// tysiacMaxSliceLen caps slice sizes during deserialisation.
const tysiacMaxSliceLen = 5000

// errTysiacOversized is the single sentinel error for oversized input arrays.
var errTysiacOversized = errors.New("tysiac: input array exceeds maximum allowed size")

// errTysiacInvalidPlayers is returned when restored state lacks exactly TysiacPlayerCnt players.
var errTysiacInvalidPlayers = errors.New("tysiac: invalid player count")

// errTysiacInvalidTrick is returned when a restored trick card or its card is nil.
var errTysiacInvalidTrick = errors.New("tysiac: invalid trick card")

// errTysiacInvalidTalon is returned when a restored talon card is nil.
var errTysiacInvalidTalon = errors.New("tysiac: invalid talon card")

// errTysiacInvalidIndex is returned when a restored index field is out of range.
var errTysiacInvalidIndex = errors.New("tysiac: index field out of range")

// errTysiacInvalidPhase is returned when a restored phase is out of range.
var errTysiacInvalidPhase = errors.New("tysiac: phase out of range")

// errTysiacInvalidTrump is returned when a restored trump suit is out of range.
var errTysiacInvalidTrump = errors.New("tysiac: trump suit out of range")

// tysiacInRange reports whether v is in [0, TysiacPlayerCnt).
func tysiacInRange(v int) bool { return v >= 0 && v < TysiacPlayerCnt }

// tysiacInRangeOrUnset reports whether v is -1 (unset) or in [0, TysiacPlayerCnt).
func tysiacInRangeOrUnset(v int) bool { return v == -1 || tysiacInRange(v) }

// UnmarshalJSON implements json.Unmarshaler.
func (g *Tysiac) UnmarshalJSON(data []byte) error {
	var j tysiacJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) > tysiacMaxSliceLen || len(j.CurrentTrick) > tysiacMaxSliceLen ||
		len(j.ActionLog) > tysiacMaxSliceLen || len(j.Talon) > tysiacMaxSliceLen {
		return errTysiacOversized
	}
	if len(j.Players) != TysiacPlayerCnt {
		return errTysiacInvalidPlayers
	}
	for _, p := range j.Players {
		if p == nil {
			return errTysiacInvalidPlayers
		}
	}
	for _, tc := range j.CurrentTrick {
		if tc == nil || tc.Card == nil {
			return errTysiacInvalidTrick
		}
		if !tysiacInRange(tc.PlayerIdx) {
			return errTysiacInvalidTrick
		}
	}
	for _, c := range j.Talon {
		if c == nil {
			return errTysiacInvalidTalon
		}
	}
	// 範囲必須のインデックス (0..PlayerCnt)。
	if !tysiacInRange(j.CurrentPlayerIdx) || !tysiacInRange(j.DealerIdx) || !tysiacInRange(j.ForehandIdx) {
		return errTysiacInvalidIndex
	}
	// -1 (未設定) 許容のインデックス。
	if !tysiacInRangeOrUnset(j.LeadPlayerIdx) || !tysiacInRangeOrUnset(j.DeclarerIdx) ||
		!tysiacInRangeOrUnset(j.LastTrickWinner) || !tysiacInRangeOrUnset(j.WinnerPlayer) {
		return errTysiacInvalidIndex
	}
	if j.DiscardCount < 0 || j.DiscardCount > TysiacTalonSize {
		return errTysiacInvalidIndex
	}
	if int(j.Phase) < TysiacPhaseMin || int(j.Phase) > TysiacPhaseMax {
		return errTysiacInvalidPhase
	}
	// trumpSuit: 0=未設定 許容、それ以外は [Spade, Diamond]。
	if j.TrumpSuit != 0 && (j.TrumpSuit < CardDesignSpade || j.TrumpSuit > CardDesignDiamond) {
		return errTysiacInvalidTrump
	}
	if err := j.Config.Validate(); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = newTysiacDeck()
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
	g.contract = j.Contract
	g.currentBid = j.CurrentBid
	g.bidPassed = j.BidPassed
	g.talon = j.Talon
	g.talonTaken = j.TalonTaken
	g.discardCount = j.DiscardCount
	g.trumpSuit = j.TrumpSuit
	g.playerScores = j.PlayerScores
	g.roundCardPts = j.RoundCardPts
	g.roundMarriage = j.RoundMarriage
	g.lastTrickWinner = j.LastTrickWinner
	g.gameEndFlag = j.GameEndFlag
	g.winnerPlayer = j.WinnerPlayer
	g.actionLog = j.ActionLog
	return nil
}
